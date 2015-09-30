package data

import (
	"app/utils"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/austindizzy/vine-go"
	"github.com/hoisie/mustache"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/search"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
)

func (db *DB) GetQueuedUser(userId string) (user *QueuedUser, err error) {
	if vine.IsVanity(userId) {
		var temp []*QueuedUser
		q := datastore.NewQuery("Queue").Filter("UserID =", strings.ToLower(userId)).Limit(1)
		k, e := q.GetAll(db.Context, &temp)
		if len(k) != 0 {
			user = temp[0]
		}
		err = e
	} else {
		intId, _ := strconv.ParseInt(userId, 10, 64)
		key := datastore.NewKey(db.Context, "Queue", "", intId, nil)
		err = datastore.Get(db.Context, key, &user)
	}
	return
}

func (db *DB) QueueUser(userId string) {
	vineApi := vine.NewRequest(urlfetch.Client(db.Context))
	user, err := vineApi.GetUser(userId)
	if err == nil {

		key := datastore.NewKey(db.Context, "Queue", "", user.UserId, nil)

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
		t.Name = user.UserIdStr + "-0-" + utils.GenSlug()

		if _, err = taskqueue.Add(db.Context, t, ""); err != nil {
			log.Errorf(db.Context, "Error adding user %s to taskqueue: %v", user.UserIdStr, err)
		}

		if _, err = datastore.Put(db.Context, key, &data); err != nil {
			log.Errorf(db.Context, "got datastore err on QueueUser: %v", err)
		}
	} else {
		log.Errorf(db.Context, "got QueueUser err: %v", err)
	}
}

func (db *DB) UserQueueExist(userId int64) bool {
	k := datastore.NewKey(db.Context, "Queue", "", userId, nil)
	q := datastore.NewQuery("Queue").Filter("__key__ =", k).Limit(1).KeysOnly()
	keys, _ := q.GetAll(db.Context, nil)
	return len(keys) > 0
}

func (db *DB) GenSummaryChart(user *UserRecord) (string, error) {
	dir := path.Join(os.Getenv("PWD"), "templates")
	template := path.Join(dir, "weeklyreport.chart")
	var loops, followers, dates string

	for i := 1; (len(user.UserData)-i-1 > -1) && (i <= 7); i++ {
		if i > 1 {
			loops += ","
			followers += ","
			dates += ","
		}
		u := user.UserData[len(user.UserData)-i]
		v := user.UserData[len(user.UserData)-i-1]
		loops += fmt.Sprintf("%d", u.Loops-v.Loops)
		followers += fmt.Sprintf("%d", u.Followers-v.Followers)
		dates += fmt.Sprintf("\"%d/%d\"", u.Recorded.Month(), u.Recorded.Day())
	}

	data := map[string]string{
		"loops":     loops,
		"followers": followers,
		"dates":     dates,
	}

	log.Infof(db.Context, "opts: %#v", data)

	opts := &url.Values{}
	opts.Add("options", mustache.RenderFile(template, data))
	opts.Add("width", "500")
	opts.Add("scale", "0.5")

	client := urlfetch.Client(db.Context)
	resp, err := client.Get(fmt.Sprintf("http://export.highcharts.com/?%s", opts.Encode()))
	b, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Infof(db.Context, "got highcharts error: %v", string(b[:]))
	}

	return base64.StdEncoding.EncodeToString(b), err
}

func (db *DB) SearchUsers(query string) ([]UserRecord, error) {
	index, err := search.Open("users")
	if err != nil {
		return nil, err
	}

	var users []UserRecord

	opts := &search.SearchOptions{
		Limit:   100,
		IDsOnly: true,
	}

	for t := index.Search(db.Context, query, opts); ; {
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
