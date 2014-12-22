package main

import (
    "appengine"
    "appengine/datastore"
    "strings"
    "time"
)

func QueueUser(userId string, c appengine.Context) {
	key := datastore.NewKey(c, "Queue", strings.ToLower(userId), 0, nil)
	data := QueuedUser{strings.ToLower(userId), time.Now()}
	datastore.Put(c, key, &data)
}

func GetQueuedUser(userId string, c appengine.Context) (usesr QueuedUser, err error) {
    key := datastore.NewKey(c, "Queue", strings.ToLower(userId), 0, nil)
    var user *QueuedUser
    err = datastore.Get(c, key, &user)
    return
}