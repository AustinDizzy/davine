package main

import (
	"appengine"
	"appengine/datastore"
	"appengine/search"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

type DB struct {
	Context appengine.Context
}

func (db *DB) FetchUser(userId string) {
    //Step 1. Get Vine user's data from the Vine API
	vineApi := VineRequest{db.Context}
	vineUser, err := vineApi.GetUser(userId)

	if err != nil {
		if err.Error() == ErrUserDoesntExist {
			db.UnqueueUser(userId)
		} else {
		    db.Context.Errorf("got error getting user %s from vine: %v", userId, err)
		}
		return
	} else if vineUser == nil {
		db.Context.Errorf("failed fetch on user %v. got err %v", userId, err)
		return
	} else if vineUser.Private == 1 {
	    db.Context.Infof("user %s is private", userId)
		return
	}

	recordKey := datastore.NewKey(db.Context, "UserRecord", "", vineUser.UserId, nil)

    //Step 2. Add user to user search index.
    userIndex := &UserIndex{
		Username:    vineUser.Username,
		Location:    vineUser.Location,
		Description: vineUser.Description,
	}

	if len(vineUser.VanityUrls) != 0 {
		userIndex.VanityUrl = strings.ToLower(vineUser.VanityUrls[0])
	}

	index, err := search.Open("users")
	if err != nil {
		db.Context.Errorf(err.Error())
	} else {
		index.Put(db.Context, vineUser.UserIdStr, userIndex)
	}

    //Step 3. Write records (user {meta, record, data}).
    if userRecord, err := db.GetUserRecord(vineUser.UserId); err == nil {
        var userMeta []*UserMeta

        if vineUser.Username != userRecord.Username {
            userMeta = append(userMeta, &UserMeta{vineUser.UserId, "username", userRecord.Username, time.Now()})
        }

        if vineUser.Location != userRecord.Location {
            userMeta = append(userMeta, &UserMeta{vineUser.UserId, "location", userRecord.Location, time.Now()})
        }

        if vineUser.Description != userRecord.Description {
            userMeta = append(userMeta, &UserMeta{vineUser.UserId, "description", userRecord.Description, time.Now()})
        }

        if vineUser.Location != userRecord.Location {
            userMeta = append(userMeta, &UserMeta{vineUser.UserId, "location", userRecord.Location, time.Now()})
        }

        if (vineUser.Verified != 0) != userRecord.Verified {
            userMeta = append(userMeta, &UserMeta{vineUser.UserId, "verified", strconv.FormatBool(userRecord.Verified), time.Now()})
        }

        var metaKey *datastore.Key
        for _, m := range userMeta {
            metaKey = datastore.NewIncompleteKey(db.Context, "UserMeta", recordKey)
            if key, err := datastore.Put(db.Context, metaKey, m); err != nil {
                db.Context.Errorf("got error storing user meta %s - %v: %v", userId, key, err)
            }
        }
    }

    userRecord := &UserRecord{
        UserId: vineUser.UserIdStr,
        Username: vineUser.Username,
        Description: vineUser.Description,
        Location: vineUser.Location,
        ProfileBackground: vineUser.ProfileBackground,
        AvatarUrl: vineUser.AvatarUrl,
        FollowerCount: vineUser.FollowerCount,
        FollowingCount: vineUser.FollowingCount,
        LoopCount: vineUser.LoopCount,
        PostCount: vineUser.AuthoredPostCount,
        RevineCount: (vineUser.PostCount - vineUser.AuthoredPostCount),
        LikeCount: vineUser.LikeCount,
        Private: (vineUser.Private != 0),
        Verified: (vineUser.Verified != 0),
        Explicit: (vineUser.ExplicitContent != 0),
    }
    if len(vineUser.VanityUrls) != 0 {
        userRecord.Vanity = strings.ToLower(vineUser.VanityUrls[0])
    }
    if _, err := datastore.Put(db.Context, recordKey, &userRecord); err != nil {
        db.Context.Errorf("got error storing user record %s: %v", userId, err)
    }

    dataKey := datastore.NewIncompleteKey(db.Context, "UserData", recordKey)
    userData := &UserData{
        UserId: vineUser.UserId,
        Recorded: time.Now(),
        Followers: vineUser.FollowerCount,
        Following: vineUser.FollowingCount,
        Loops: vineUser.LoopCount,
        Posts: vineUser.AuthoredPostCount,
        Revines: (vineUser.PostCount - vineUser.AuthoredPostCount),
        Likes: vineUser.LikeCount,
    }
    if key, err := datastore.Put(db.Context, dataKey, &userData); err != nil {
        db.Context.Errorf("got error storing user data %s - %v: %v", userId, key, err)
    }
}

func (db *DB) GetUser(userId int64) (user *UserRecord, err error) {
    user, err = db.GetUserRecord(userId)
    if err != nil {
        return
    }
    user.userData, err = db.GetUserData(userId)

    if err != nil {
        return
    }

    user.userMeta, err = db.GetUserMeta(userId)
    return
}

func (db *DB) GetUserRecord(userId int64) (*UserRecord, error) {

    user := new(UserRecord)

    recordKey := datastore.NewKey(db.Context, "UserRecord", "", userId, nil)
    err := datastore.Get(db.Context, recordKey, &user)

    if err != nil {
        return nil, err
    } else {
        return user, nil
    }
}

func (db *DB) GetUserData(userId int64) (userData []*UserData, err error) {

    record := datastore.NewKey(db.Context, "UserRecord", "", userId, nil)
	dataQuery := datastore.NewQuery("UserData").Ancestor(record).Order("Recorded")

	_, err = dataQuery.GetAll(db.Context, &userData)
	return
}

func (db *DB) GetUserMeta(userId int64) (userMeta []*UserMeta, err error) {

    record := datastore.NewKey(db.Context, "UserRecord", "", userId, nil)
	dataQuery := datastore.NewQuery("UserMeta").Ancestor(record).Order("Recorded")

	_, err = dataQuery.GetAll(db.Context, &userData)
	return
}

func (db *DB) GetTotalUsers() (int, error) {

	var metaStats MetaStats

	key := datastore.NewKey(db.Context, "__Stat_Kind_IsRootEntity__", "UserMeta", 0, nil)
	err := datastore.Get(db.Context, key, &metaStats)

	return metaStats.Count, err
}

func (db *DB) GetTop() (data map[string]interface{}) {

	var topOverall, topFollowed, topLooped, topPosts, topRevines []StoredUserMeta

	//top overall
	q := datastore.NewQuery("UserMeta").Order("-Current.Followers").Limit(10)
	q.GetAll(db.Context, &topOverall)

	sort.Sort(ByOverall(topOverall))

	//top followed
	q = datastore.NewQuery("UserMeta").Order("-Current.Followers").Limit(10)
	q.GetAll(db.Context, &topFollowed)

	//top looped
	q = datastore.NewQuery("UserMeta").Order("-Current.Loops").Limit(10)
	q.GetAll(db.Context, &topLooped)

	//top posts
	q = datastore.NewQuery("UserMeta").Order("-Current.AuthoredPosts").Limit(5)
	q.GetAll(db.Context, &topPosts)

	//top Revines
	q = datastore.NewQuery("UserMeta").Order("-Current.Revines").Limit(5)
	q.GetAll(db.Context, &topRevines)

	lastUpdated := db.GetLastUpdated()

	data = map[string]interface{}{
		"topOverall":  topOverall,
		"topFollowed": topFollowed,
		"topLooped":   topLooped,
		"topPosts":    topPosts,
		"topRevines":  topRevines,
		"lastUpdated": lastUpdated,
	}
	return
}

func (a ByOverall) Len() int      { return len(a) }
func (a ByOverall) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByOverall) Less(i, j int) bool {
	return a[i].Current.Followers > a[j].Current.Followers && a[i].Current.Loops > a[j].Current.Loops && a[i].Current.Following < a[j].Current.Following
}

func (a ByFollowers) Len() int      { return len(a) }
func (a ByFollowers) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByFollowers) Less(i, j int) bool {
	return a[i].Current.Followers > a[j].Current.Followers
}

func (a ByLoops) Len() int      { return len(a) }
func (a ByLoops) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByLoops) Less(i, j int) bool {
	return a[i].Current.Loops > a[j].Current.Loops
}

func (a ByPosts) Len() int      { return len(a) }
func (a ByPosts) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByPosts) Less(i, j int) bool {
	return a[i].Current.AuthoredPosts > a[j].Current.AuthoredPosts
}

func (a ByRevines) Len() int      { return len(a) }
func (a ByRevines) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByRevines) Less(i, j int) bool {
	return a[i].Current.Revines > a[j].Current.Revines
}

func (db *DB) GetLastUpdatedUser() *StoredUserData {
	var lastUpdatedUser []*StoredUserData
	q := datastore.NewQuery("UserData").Order("-LastUpdated").Limit(1)
	q.GetAll(db.Context, &lastUpdatedUser)
	if len(lastUpdatedUser) == 0 {
		return nil
	} else {
		return lastUpdatedUser[0]
	}
}

func (db *DB) GetLastUpdated() time.Time {
	lastUpdatedUser := db.GetLastUpdatedUser()
	if lastUpdatedUser == nil {
		return time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		return lastUpdatedUser.LastUpdated
	}
}

func (db *DB) UnqueueUser(user string) {
	var key *datastore.Key
	vineApi := VineRequest{db.Context}
	if vineApi.IsVanity(user) {
		q := datastore.NewQuery("Queue").Filter("UserID =", user).KeysOnly()
		keys, err := q.GetAll(db.Context, nil)
		if err == nil {
			key = keys[0]
		} else {
			db.Context.Errorf("error removing %v from queue: %v", user, err)
		}
	} else {
		userId, _ := strconv.ParseInt(user, 10, 64)
		key = datastore.NewKey(db.Context, "Queue", "", userId, nil)
	}

	datastore.Delete(db.Context, key)
	db.Context.Infof("%v removed from queue.", user)
}

func RandomKey(a []*datastore.Key) *datastore.Key {
	rand.Seed(time.Now().UTC().UnixNano())
	return a[rand.Intn(len(a))]
}

func RemoveDuplicates(a []string) []string {
	found := make(map[string]bool)
	j := 0
	for i, x := range a {
		if !found[x] {
			found[x] = true
			a[j] = a[i]
			j++
		}
	}
	return a[:j]
}
