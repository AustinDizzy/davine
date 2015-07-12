package main

import (
	"appengine"
	"appengine/datastore"
	"appengine/search"
	"appengine/taskqueue"
	"archive/zip"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Export struct {
	Context appengine.Context
}

func (x *Export) User(userIdStr string, w http.ResponseWriter) {
	db := DB{x.Context}
	userId, _ := strconv.ParseInt(userIdStr, 10, 64)

	user, err := db.GetUser(userId)
	if err != nil {
		x.Context.Errorf("got err on export: %v", err)
		return
	}

	userMeta, _ := json.MarshalIndent(user.UserMeta, "", "  ")
	userData, _ := json.MarshalIndent(user.UserData, "", "  ")

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

		t := taskqueue.NewPOSTTask("/cron/fetch", map[string][]string{
			"id": {user.UserIdStr},
			"n":  {"0"},
		})
		t.Name = user.UserIdStr + "-0"

		if _, err = taskqueue.Add(c, t, ""); err != nil {
			c.Errorf("Error adding user %s to taskqueue: %v", user.UserIdStr, err)
		}

		if _, err = datastore.Put(c, key, &data); err != nil {
			c.Errorf("got datastore err on QueueUser: %v", err)
		}
	} else {
		c.Errorf("got QueueUser err: %v", err)
	}
}

func GetQueuedUser(userId string, c appengine.Context) (user *QueuedUser, err error) {
	vineApi := VineRequest{c}
	if vineApi.IsVanity(userId) {
		var temp []*QueuedUser
		q := datastore.NewQuery("Queue").Filter("UserID =", strings.ToLower(userId)).Limit(1)
		k, e := q.GetAll(c, &temp)
		if len(k) != 0 {
			user = temp[0]
		}
		err = e
	} else {
		intId, _ := strconv.ParseInt(userId, 10, 64)
		key := datastore.NewKey(c, "Queue", "", intId, nil)
		err = datastore.Get(c, key, &user)
	}
	return
}

func UserQueueExist(userId int64, c appengine.Context) bool {
	k := datastore.NewKey(c, "Queue", "", userId, nil)
	q := datastore.NewQuery("Queue").Filter("__key__ =", k).Limit(1).KeysOnly()
	keys, _ := q.GetAll(c, nil)
	return len(keys) > 0
}

func SearchUsers(c appengine.Context, query string) ([]UserRecord, error) {
	db := DB{c}
	index, err := search.Open("users")
	if err != nil {
		return nil, err
	}

	var users []UserRecord

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
		userRecord, err := db.GetUserRecord(id)
		if err != nil {
			return users, err
		}
		users = append(users, *userRecord)
	}

	return users, nil
}

func GetAppUsers(c appengine.Context) ([]*AppUser, []*AppUser, error) {
	var enterpriseUsers []*AppUser
	var emailReportUsers []*AppUser
	q := datastore.NewQuery("AppUser").KeysOnly()
	keys, _ := q.GetAll(c, nil)

	for _, v := range keys {
		u := new(AppUser)
		if err := datastore.Get(c, v, u); err == nil {
			if u.Type == "enterprise" {
				enterpriseUsers = append(enterpriseUsers, u)
			} else if u.Type == "email-report" {
				emailReportUsers = append(emailReportUsers, u)
			}
		} else {
			c.Errorf("got err: %v", err)
			return enterpriseUsers, emailReportUsers, err
		}
	}
	return enterpriseUsers, emailReportUsers, nil
}
