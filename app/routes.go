package main

import (
	"appengine"
	"appengine/datastore"
	"appengine/file"
	"appengine/urlfetch"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hoisie/mustache"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

func UserFetchHandler(w http.ResponseWriter, r *http.Request) {
	dir := path.Join(os.Getenv("PWD"), "templates")
	profile := path.Join(dir, "profile.html")
	layout := path.Join(dir, "profileLayout.html")
	vars := mux.Vars(r)
	c := appengine.NewContext(r)

	db := DB{c}
	match, _ := regexp.MatchString("^[0-9]+$", vars["user"])
	var err error
	var userMetaTemp interface{}
	var storedUserData interface{}
	var data string

	if match {
		userId, _ := strconv.ParseInt(vars["user"], 10, 64)
		userMetaTemp, err = db.GetUserMeta(userId)
		storedUserData, _ = db.GetUserData(userId)
	} else {
		temp := []StoredUserMeta{}
		q := datastore.NewQuery("UserMeta").Filter("VanityUrl =", strings.ToLower(vars["user"])).Limit(1)
		k, _ := q.GetAll(c, &temp)
		if len(temp) > 0 {
			userMetaTemp = temp[0]
			storedUserData, _ = db.GetUserData(k[0].IntID())
		} else {
			user404 := path.Join(dir, "user404.html")
			userData := map[string]string{"user": vars["user"]}
			data = mustache.RenderFileInLayout(user404, layout, userData)
			w.WriteHeader(http.StatusNotFound)
		}
	}

	if err == datastore.ErrNoSuchEntity {
		user404 := path.Join(dir, "user404.html")
		userData := map[string]string{"user": vars["user"]}
		data = mustache.RenderFileInLayout(user404, layout, userData)
		w.WriteHeader(http.StatusNotFound)
	} else if err != nil {
		fmt.Fprint(w, err.Error())
	} else if userMetaTemp != nil {

		userMeta := userMetaTemp.(StoredUserMeta)

		userData := map[string]interface{}{
			"username":    userMeta.Username,
			"userId":      userMeta.UserId,
			"description": userMeta.Description,
			"location":    userMeta.Location,
			"avatarUrl":   userMeta.AvatarUrl,
			"loops":       strconv.FormatInt(userMeta.Current.Loops, 10),
			"followers":   strconv.FormatInt(userMeta.Current.Followers, 10),
			"data":        storedUserData,
			"previous":    userMeta.Previous,
		}

		if userMeta.Background != "" {
			color := strings.SplitAfterN(userMeta.Background, "0x", 2)
			userData["profileBackground"] = color[1]
		} else {
			userData["profileBackground"] = "00BF8F"
		}

		data = mustache.RenderFileInLayout(profile, layout, userData)
	}

	fmt.Fprint(w, data)
}

func UserStoreHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	vineApi := VineRequest{c}
	db := DB{c}
	_, err := GetQueuedUser(r.FormValue("id"), c)
	data := make(map[string]bool)

	if err != datastore.ErrNoSuchEntity && err != nil {
		c.Errorf("got UserStore err: %v", err)
	}

	user, apiErr := vineApi.GetUser(r.FormValue("id"))

	if err == datastore.ErrNoSuchEntity {

		if apiErr != nil {
			data["exists"] = false
			data["queued"] = false
		} else {
			QueueUser(user.UserIdStr, c)
			data["exists"] = true
			data["queued"] = true
		}

		data["stored"] = false

	} else {
		_, err := db.GetUserMeta(user.UserId)
		if err == datastore.ErrNoSuchEntity {
			data["stored"] = false
		} else {
			data["stored"] = true
		}
		data["exists"] = true
		data["queued"] = false
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	dir := path.Join(os.Getenv("PWD"), "templates")
	aboutPage := path.Join(dir, "about.html")
	layout := path.Join(dir, "pageLayout.html")

	db := DB{appengine.NewContext(r)}
	totalUsers, _ := db.GetTotalUsers()
	stats := map[string]interface{}{"users": totalUsers}
	stats["lastUpdated"] = db.GetLastUpdated()
	data := mustache.RenderFileInLayout(aboutPage, layout, stats)

	fmt.Fprint(w, data)
}

func DiscoverHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	vineApi := VineRequest{c}
	db := DB{c}
	var recentUsers []*VineUser
	var recentVerified []StoredUserMeta

	recent := datastore.NewQuery("Queue").Order("-Discovered").Limit(5).KeysOnly()
	k, _ := recent.GetAll(c, nil)
	for i, _ := range k {
		user, err := vineApi.GetUser(strconv.FormatInt(k[i].IntID(), 10))
		if err == nil {
			recentUsers = append(recentUsers, user)
		}
	}
	verified := datastore.NewQuery("UserMeta").Filter("Verified =", true).Limit(5).KeysOnly()
	k, _ = verified.GetAll(c, nil)
	for i, _ := range k {
		user, err := db.GetUserMeta(k[i].IntID())
		if err == nil {
			recentVerified = append(recentVerified, user.(StoredUserMeta))
		}
	}
	data := map[string]interface{}{"recentUsers": recentUsers, "recentVerified": recentVerified}
	dir := path.Join(os.Getenv("PWD"), "templates")
	discover := path.Join(dir, "discover.html")
	layout := path.Join(dir, "pageLayout.html")
	page := mustache.RenderFileInLayout(discover, layout, data)
	fmt.Fprint(w, page)
}

func TopHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	db := DB{c}

	dir := path.Join(os.Getenv("PWD"), "templates")
	top := path.Join(dir, "top.html")
	layout := path.Join(dir, "pageLayout.html")
	data := db.GetTop()
	data["LastUpdated"] = db.GetLastUpdated()
	page := mustache.RenderFileInLayout(top, layout, data)
	fmt.Fprint(w, page)
}

func RandomHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("UserMeta").KeysOnly()
	keys, _ := q.GetAll(c, nil)

	keys = Shuffle(keys)
	var user QueuedUser
	key := datastore.NewKey(c, "Queue", "", keys[0].IntID(), nil)
	err := datastore.Get(c, key, &user)
	if err != nil {
		c.Errorf("got err %v", err)
	} else {
		c.Infof("got user %v", user)
	}
	http.Redirect(w, r, "/u/"+user.UserID, 301)
}

func DonateHandler(w http.ResponseWriter, r *http.Request) {
	dir := path.Join(os.Getenv("PWD"), "templates")
	donate := path.Join(dir, "donate.html")
	layout := path.Join(dir, "pageLayout.html")
	page := mustache.RenderFileInLayout(donate, layout, nil)
	fmt.Fprint(w, page)
}

func ExportHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	db := DB{c}
	vars := mux.Vars(r)

	if r.Method == "GET" {
		userId, err := strconv.ParseInt(vars["user"], 10, 64)
		if err != nil {
			c.Errorf("got err: %v", err)
			http.Redirect(w, r, "/404", 301)
			return
		}
		userMeta, err := db.GetUserMeta(userId)
		if err == datastore.ErrNoSuchEntity {
			http.Redirect(w, r, "/404", 301)
			return
		}
		user := userMeta.(StoredUserMeta)
		data := map[string]string{"username": user.Username, "userId": vars["user"], "captcha": Config["captchaPublic"]}
		dir := path.Join(os.Getenv("PWD"), "templates")
		export := path.Join(dir, "export.html")
		layout := path.Join(dir, "pageLayout.html")
		page := mustache.RenderFileInLayout(export, layout, data)
		fmt.Fprint(w, page)
	} else if r.Method == "POST" {
		client := urlfetch.Client(c)
		url := "https://www.google.com/recaptcha/api/siteverify?secret=" + Config["captchaPrivate"]
		url += "&response=" + r.FormValue("g-recaptcha-response") + "&remoteip=" + r.RemoteAddr
		req, _ := http.NewRequest("GET", url, nil)
		resp, err := client.Do(req)
		if err != nil {
			c.Errorf("got err: %v", err)
			return
		}
		body, _ := ioutil.ReadAll(resp.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["success"].(bool) {
			export := Export{c}
			export.User(vars["user"], w)
		} else {
			fmt.Fprint(w, "Seems like your CAPTCHA was wrong. Please press back and try again.")
		}
	}
}

func PopularFetchHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	vineApi := VineRequest{c}

	users, err := vineApi.GetPopularUsers()
	for _, v := range users {
		if _, err := GetQueuedUser(v, c); err == datastore.ErrNoSuchEntity {
			QueueUser(v, c)
		}
	}
	fmt.Fprint(w, "queuing popular users: %v w/ err %v", users, err)
	c.Infof("queueing popular users: %v w/ err %v", users, err)
}

func CronFetchHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	q := datastore.NewQuery("Queue").KeysOnly()
	keys, _ := q.GetAll(c, nil)
	db := DB{c}

	for _, v := range keys {
		db.FetchUser(strconv.FormatInt(v.IntID(), 10))
	}

	c.Infof("Finished cron fetch")

	fmt.Fprint(w, "fetching users")
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	dir := path.Join(os.Getenv("PWD"), "templates")
	notFound := path.Join(dir, "404.html")
	layout := path.Join(dir, "pageLayout.html")
	data := map[string]string{"url": r.RequestURI}
	page := mustache.RenderFileInLayout(notFound, layout, data)
	w.WriteHeader(404)
	fmt.Fprint(w, page)
}

func StartupHandler(w http.ResponseWriter, r *http.Request) {
	if len(Config) == 0 {
		c := appengine.NewContext(r)
		client := &http.Client{
			Transport: &oauth2.Transport{
				Source: google.AppEngineTokenSource(c, storage.ScopeReadOnly),
				Base: &urlfetch.Transport{
					Context: c,
				},
			},
		}
		bucket, _ := file.DefaultBucketName(c)
		ctx := cloud.NewContext("davine-web", client)
		rc, err := storage.NewReader(ctx, bucket, "config.yaml")
		if err != nil {
			c.Errorf("error reading config: %v", err.Error())
			return
		}
		configFile, err := ioutil.ReadAll(rc)
		rc.Close()
		if err != nil {
			c.Errorf("error reading config: %v", err.Error())
			return
		}

		c.Infof("loaded config file: %v", configFile)
		yaml.Unmarshal(configFile, &Config)
		c.Infof("loaded config struct: %v", Config)
	}
}
