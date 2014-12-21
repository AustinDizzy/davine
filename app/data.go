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

    if data.Private == 1 {
        return
    }

    var userMeta StoredUserMeta
    var userData StoredUserData

    userId := data.UserId

    userMetaTemp, err := db.GetUserMeta(userId)

    if err == datastore.ErrNoSuchEntity {
        userMeta = StoredUserMeta{
            Username: data.Username,
            Location: data.Location,
            Description: data.Description,
            Verified: data.Verified == 1,
            AvatarUrl: data.AvatarUrl,
            Background: data.ProfileBackground,
        }
        if data.VanityUrls != nil {
            userMeta.VanityUrl = data.VanityUrls[0]
        }

        if userMeta.Verified {
            userMeta.VerifiedDate = time.Now()
        }

        userData = StoredUserData{
            LastUpdated: time.Now(),
            Followers: []int64{data.FollowerCount,},
            Following: []int64{data.FollowingCount,},
            Loops: []int64{data.LoopCount,},
            AuthoredPosts: []int64{data.AuthoredPostCount,},
            Reposts: []int64{data.PostCount - data.AuthoredPostCount,},
            Likes: []int64{data.LikeCount,},
            Updated: []time.Time{time.Now(),},
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
            userData.Reposts = append(userData.Reposts, data.PostCount - data.AuthoredPostCount)
            userData.Likes = append(userData.Followers, data.LikeCount)
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