package main

import (
    "appengine"
    "appengine/datastore"
    "time"
)

func QueueUser(userId string, c appengine.Context) {
	key := datastore.NewKey(c, "Queue", userId, 0, nil)
	data := QueuedUser{userId, time.Now()}
	datastore.Put(c, key, &data)
}