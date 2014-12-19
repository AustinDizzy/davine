package main

import (
    "appengine"
    "appengine/datastore"
    "time"
)

type DB struct {
    Context appengine.Context
}

func (db *DB) FetchUser(user string) {
    vineApi := VineRequest{db.Context}
    data, err := vineApi.GetUser(user)

    if data["private"].(float64) == 1.0 {
        return
    }

    var userMeta StoredUserMeta
    var userData StoredUserData

    userId := data["userIdStr"].(string)

    userMetaTemp, err := db.GetUserMeta(userId)

    if err == datastore.ErrNoSuchEntity {
        userMeta = StoredUserMeta{
            Username: data["username"].(string),
            Location: data["location"].(string),
            Description: data["description"].(string),
            Verified: data["verified"].(float64) == 1.0,
            AvatarUrl: data["avatarUrl"].(string),
        }
        if data["vanityUrls"] != nil {
            userMeta.VanityUrl = data["vanityUrls"].([]interface{})[0].(string)
        }

        if data["profileBackground"] != nil {
            userMeta.Background = data["profileBackground"].(string)
        }

        if userMeta.Verified {
            userMeta.VerifiedDate = time.Now()
        }

        userData = StoredUserData{
            LastUpdated: time.Now(),
            Followers: []float64{data["followerCount"].(float64),},
            Following: []float64{data["followingCount"].(float64),},
            Loops: []float64{data["loopCount"].(float64),},
            AuthoredPosts: []float64{data["authoredPostCount"].(float64),},
            Reposts: []float64{data["postCount"].(float64) - data["authoredPostCount"].(float64),},
            Likes: []float64{data["likeCount"].(float64),},
            Updated: []time.Time{time.Now(),},
        }

    } else {

        userMeta = userMetaTemp.(StoredUserMeta)

        if userMeta.Location != data["location"].(string) {
            userMeta.Previous.Location = append(userMeta.Previous.Location, PreviousLocation{userMeta.Location, time.Now()})
            userMeta.Location = data["location"].(string)
        }

        if userMeta.Username != data["username"].(string) {
            userMeta.Previous.Username = append(userMeta.Previous.Username, PreviousUsername{userMeta.Username, time.Now()})
            userMeta.Username = data["username"].(string)
        }

        if userMeta.Description != data["description"].(string) {
            userMeta.Previous.Description = append(userMeta.Previous.Description, PreviousDescription{userMeta.Description, time.Now()})
            userMeta.Description = data["description"].(string)
        }

        if data["profileBackground"] != nil {
            if userMeta.Background != data["profileBackground"].(string) {
                userMeta.Previous.Background = append(userMeta.Previous.Background, PreviousBackground{userMeta.Background, time.Now()})
                userMeta.Background = data["profileBackground"].(string)
            }
        }

        userDataTemp, err := db.GetUserData(userId)
        userData = userDataTemp.(StoredUserData)

        if err != datastore.ErrNoSuchEntity {
            userData.LastUpdated = time.Now()
            userData.Followers = append(userData.Followers, data["followerCount"].(float64))
            userData.Following = append(userData.Following, data["followingCount"].(float64))
            userData.Loops = append(userData.Loops, data["loopCount"].(float64))
            userData.AuthoredPosts = append(userData.AuthoredPosts, data["authoredPostCount"].(float64))
            userData.Reposts = append(userData.Reposts, data["postCount"].(float64) - data["authoredPostCount"].(float64))
            userData.Likes = append(userData.Followers, data["likeCount"].(float64))
            userData.Updated = append(userData.Updated, time.Now())
        }
    }

    dataKey := datastore.NewKey(db.Context, "UserData", userId, 0, nil)
    metaKey := datastore.NewKey(db.Context, "UserMeta", userId, 0, nil)

    datastore.Put(db.Context, dataKey, &userData)
    datastore.Put(db.Context, metaKey, &userMeta)
}

func (db *DB) GetUserData(user string) (interface{}, error) {

    data := StoredUserData{}

    key := datastore.NewKey(db.Context, "UserData", user, 0, nil)
    err := datastore.Get(db.Context, key, &data)

    if err != nil {
        return nil, err
    } else {
        return data, nil
    }
}

func (db *DB) GetUserMeta(user string) (interface{}, error) {

    meta := StoredUserMeta{}

	key := datastore.NewKey(db.Context, "UserMeta", user, 0, nil)
    err := datastore.Get(db.Context, key, &meta)

    if err != nil {
        return nil, err
    } else {
        return meta, nil
    }
}