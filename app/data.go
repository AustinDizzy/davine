package main

import (
	"appengine"
	"appengine/datastore"
	"strings"
	"time"
	"sort"
	"math/rand"
)

type DB struct {
	Context appengine.Context
}

func (db *DB) FetchUser(user string) {
	vineApi := VineRequest{db.Context}
	data, err := vineApi.GetUser(user)

	if data == nil {
	    db.Context.Errorf("failed fetch on user %v. got err %v", user, err)
	    return
	} else if data.Private == 1 {
	    return
	}

	var userMeta StoredUserMeta
	var userData StoredUserData

	userId := data.UserId

	userMetaTemp, err := db.GetUserMeta(userId)

	if err == datastore.ErrNoSuchEntity {
		userMeta = StoredUserMeta{
			Username:    data.Username,
			Location:    data.Location,
			Description: data.Description,
			Verified:    data.Verified == 1,
			AvatarUrl:   data.AvatarUrl,
			Background:  data.ProfileBackground,
		}
		if len(data.VanityUrls) != 0 {
			userMeta.VanityUrl = strings.ToLower(data.VanityUrls[0])
		}

		if userMeta.Verified {
			userMeta.VerifiedDate = time.Now()
		}

		userMeta.Current = StoredUserMetaCurrent{
			Followers:     data.FollowerCount,
			Following:     data.FollowingCount,
			Loops:         data.LoopCount,
			AuthoredPosts: data.AuthoredPostCount,
			Revines:       data.PostCount - data.AuthoredPostCount,
			Likes:         data.LikeCount,
		}

		userData = StoredUserData{
			LastUpdated:   time.Now(),
			Followers:     []int64{data.FollowerCount},
			Following:     []int64{data.FollowingCount},
			Loops:         []int64{data.LoopCount},
			AuthoredPosts: []int64{data.AuthoredPostCount},
			Revines:       []int64{data.PostCount - data.AuthoredPostCount},
			Likes:         []int64{data.LikeCount},
			Updated:       []time.Time{time.Now()},
		}

	} else {

		userMeta = userMetaTemp.(StoredUserMeta)

		if userMeta.Location != data.Location {
			userMeta.Previous.Location = append(userMeta.Previous.Location, PreviousLocation{userMeta.Location, time.Now()})
			userMeta.Location = data.Location
		}

		if userMeta.Username != data.Username {
			userMeta.Previous.Username = append(userMeta.Previous.Username, PreviousUsername{userMeta.Username, time.Now()})
			userMeta.Username = data.Username
		}

		if userMeta.Description != data.Description {
			userMeta.Previous.Description = append(userMeta.Previous.Description, PreviousDescription{userMeta.Description, time.Now()})
			userMeta.Description = data.Description
		}

		if userMeta.Background != data.ProfileBackground {
			userMeta.Previous.Background = append(userMeta.Previous.Background, PreviousBackground{userMeta.Background, time.Now()})
			userMeta.Background = data.ProfileBackground
		}

		userDataTemp, err := db.GetUserData(userId)
		userData = userDataTemp.(StoredUserData)

		if err != datastore.ErrNoSuchEntity {
			userData.LastUpdated = time.Now()
			userData.Followers = append(userData.Followers, data.FollowerCount)
			userData.Following = append(userData.Following, data.FollowingCount)
			userData.Loops = append(userData.Loops, data.LoopCount)
			userData.AuthoredPosts = append(userData.AuthoredPosts, data.AuthoredPostCount)
			userData.Revines = append(userData.Revines, data.PostCount-data.AuthoredPostCount)
			userData.Likes = append(userData.Likes, data.LikeCount)
			userData.Updated = append(userData.Updated, time.Now())
		}
	}

	dataKey := datastore.NewKey(db.Context, "UserData", "", userId, nil)
	metaKey := datastore.NewKey(db.Context, "UserMeta", "", userId, nil)

	datastore.Put(db.Context, dataKey, &userData)
	datastore.Put(db.Context, metaKey, &userMeta)
}

func (db *DB) GetUserData(user int64) (interface{}, error) {

	data := StoredUserData{}

	key := datastore.NewKey(db.Context, "UserData", "", user, nil)
	err := datastore.Get(db.Context, key, &data)

	if err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

func (db *DB) GetUserMeta(user int64) (interface{}, error) {

	meta := StoredUserMeta{}

	key := datastore.NewKey(db.Context, "UserMeta", "", user, nil)
	err := datastore.Get(db.Context, key, &meta)

	if err != nil {
		return nil, err
	} else {
		return meta, nil
	}
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
	    "topOverall": topOverall,
	    "topFollowed": topFollowed,
	    "topLooped": topLooped,
	    "topPosts": topPosts,
	    "topRevines": topRevines,
	    "lastUpdated": lastUpdated,
	}
	return
}

func (a ByOverall) Len() int           { return len(a) }
func (a ByOverall) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByOverall) Less(i, j int) bool {
    return a[i].Current.Followers > a[j].Current.Followers && a[i].Current.Loops > a[j].Current.Loops && a[i].Current.Following < a[j].Current.Following
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

func Shuffle(a []*datastore.Key) []*datastore.Key {
    b := a
    for i := range a {
        j := rand.Intn(i + 1)
        b[i], b[j] = b[j], b[i]
    }
    return b
}