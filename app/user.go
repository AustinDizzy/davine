package main

import (
	"appengine"
	"appengine/datastore"
	"strings"
	"time"
	"regexp"
	"strconv"
)

func QueueUser(userId string, c appengine.Context) {
    vineApi := VineRequest{c}
    user, err := vineApi.GetUser(userId)
    if err == nil {

        key := datastore.NewKey(c, "Queue", "", user.UserId, nil)

        var data QueuedUser
        if len(user.VanityUrls) > 0 {
            data = QueuedUser{strings.ToLower(user.VanityUrls[0]), time.Now()}
        } else {
            data = QueuedUser{user.UserIdStr, time.Now()}
        }

        _, err := datastore.Put(c, key, &data)
        if err != nil {
            c.Errorf("got datastore err on QueueUser: %v", err)
        }
    } else {
        c.Errorf("got QueueUser err: %v", err)
    }
}

func GetQueuedUser(userId string, c appengine.Context) (user *QueuedUser, err error) {
    match, _ := regexp.MatchString("^[0-9]+$", userId)
    if match {
        intId, _ := strconv.ParseInt(userId, 10, 64)
        key := datastore.NewKey(c, "Queue", "", intId, nil)
        err = datastore.Get(c, key, &user)
    } else {
        temp := []*QueuedUser{}
        q := datastore.NewQuery("Queue").Filter("UserID =", strings.ToLower(userId)).Limit(1)
        k, _ := q.GetAll(c, &temp)
        if len(k) != 0 {
            user = temp[0]
        } else {
            err = datastore.ErrNoSuchEntity
        }
    }

    return
}
