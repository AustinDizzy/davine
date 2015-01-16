package main

import (
	"appengine"
	"appengine/datastore"
	"appengine/search"
	"archive/zip"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Export struct {
	Context appengine.Context
}

func (x *Export) User(user string, w http.ResponseWriter) {
	db := DB{x.Context}
	userId, _ := strconv.ParseInt(user, 10, 64)

	userMetaTemp, err := db.GetUserMeta(userId)
	userMeta, _ := json.MarshalIndent(userMetaTemp.(StoredUserMeta), "", "  ")
	if err != nil {
		x.Context.Errorf("got err on export: %v", err)
		return
	}

	userDataTemp, err := db.GetUserData(userId)
	userData, _ := json.MarshalIndent(userDataTemp.(StoredUserData), "", "  ")
	if err != nil {
		x.Context.Errorf("got err on export: %v", err)
		return
	}

	w.Header().Add("Content-Type", "application/zip")
	zipWriter := zip.NewWriter(w)

	var files = []struct {
		Name, Data string
	}{
		{"UserMeta.json", string(userMeta)},
		{"UserData.json", string(userData)},
	}
	for _, file := range files {
		f, err := zipWriter.Create(file.Name)
		if err != nil {
			x.Context.Errorf(err.Error())
		}
		_, err = f.Write([]byte(file.Data))
		if err != nil {
			x.Context.Errorf(err.Error())
		}
	}

	err = zipWriter.Close()
	if err != nil {
		x.Context.Errorf(err.Error())
	}
}

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

func SearchUsers(c appengine.Context, query string) ([]StoredUserMeta, error) {
	db := DB{c}
	index, err := search.Open("users")
	if err != nil {
		return nil, err
	}

	var users []StoredUserMeta

	opts := &search.SearchOptions{
		Limit:   100,
		IDsOnly: true,
	}

	for t := index.Search(c, query, opts); ; {
		key, err := t.Next(nil)
		if err == search.Done {
			break
		} else if err != nil {
			return nil, err
		}
		id, _ := strconv.ParseInt(key, 10, 64)
		userMeta, err := db.GetUserMeta(id)

		users = append(users, userMeta.(StoredUserMeta))
	}

	return users, nil
}
