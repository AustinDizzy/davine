package main

import (
	"appengine"
	"appengine/datastore"
	"appengine/file"
	"appengine/memcache"
	"appengine/taskqueue"
	"appengine/urlfetch"
	"appengine/user"
	"bytes"
	"encoding/json"
	"encoding/gob"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hoisie/mustache"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
    dir := path.Join(os.Getenv("PWD"), "templates")
    index := path.Join(dir, "index.html")
    layout := path.Join(dir, "layout.html")
    c := appengine.NewContext(r)
    data := map[string]interface{}{
        "title": PageTitle,
    }

    if popusers, err := memcache.Get(c, "popusers"); err == nil {
        var users []*VineUser
        var dec *gob.Decoder
        userIDs := strings.Split(string(popusers.Value[:]), ",")
        for len(users) < 6 {
            rand.Seed(time.Now().Unix())
            r := rand.Intn(len(userIDs))

            key, err := memcache.Get(c, userIDs[r])
            dec = gob.NewDecoder(bytes.NewReader(key.Value))
            var u *VineUser
            err = dec.Decode(&u)
            if err == nil {
                users = append(users, u)
                userIDs = append(userIDs[:r], userIDs[r+1:]...)
                //above removes already chosen user from userID array
            }
        }
        data["popusers"] = users
    } else {
        c.Errorf("popusers memcache err: %v", err)
    }

    if featuredUser, err := memcache.Get(c, "featuredUser"); err == nil {
        var user *VineUser
        r := strings.Split(string(featuredUser.Value[:]), ";")
        u, _ := memcache.Get(c, r[0])
        dec := gob.NewDecoder(bytes.NewReader(u.Value))
        dec.Decode(&user)

        data["featuredUser"] = user
        data["featuredPost"] = r[1]
    }

    fmt.Fprint(w, mustache.RenderFileInLayout(index, layout, data))
}

func UserFetchHandler(w http.ResponseWriter, r *http.Request) {
	dir := path.Join(os.Getenv("PWD"), "templates")
	profile := path.Join(dir, "profile.html")
	layout := path.Join(dir, "layout.html")
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
	u, err := GetQueuedUser(r.FormValue("id"), c)
	data := make(map[string]bool)

	if err != datastore.ErrNoSuchEntity && err != nil {
		c.Errorf("got UserStore err: %v", err)
	}

	user, apiErr := vineApi.GetUser(r.FormValue("id"))

	if err == datastore.ErrNoSuchEntity || u == nil {
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
	layout := path.Join(dir, "layout.html")

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
	data := map[string]interface{}{
	    "recentUsers": recentUsers,
	    "recentVerified": recentVerified,
	    "title": "Discover - " + PageTitle,
	}
	dir := path.Join(os.Getenv("PWD"), "templates")
	discover := path.Join(dir, "discover.html")
	layout := path.Join(dir, "layout.html")
	page := mustache.RenderFileInLayout(discover, layout, data)
	fmt.Fprint(w, page)
}

func TopHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	db := DB{c}

	dir := path.Join(os.Getenv("PWD"), "templates")
	top := path.Join(dir, "top.html")
	layout := path.Join(dir, "layout.html")
	data := db.GetTop()
	data["title"] = "Top - " + PageTitle
	data["LastUpdated"] = db.GetLastUpdated()
	page := mustache.RenderFileInLayout(top, layout, data)
	fmt.Fprint(w, page)
}

func RandomHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("UserMeta").KeysOnly()
	keys, err := q.GetAll(c, nil)
    if err != nil {
        c.Errorf("got err %v", err)
        return
    }
	randomKey := RandomKey(keys)
	var user QueuedUser
	key := datastore.NewKey(c, "Queue", "", randomKey.IntID(), nil)
	err = datastore.Get(c, key, &user)
	if err != nil {
		c.Errorf("got err %v", err)
	} else {
		c.Infof("got user %v", user)
	}
	http.Redirect(w, r, "/u/"+user.UserID, 301)
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	dir := path.Join(os.Getenv("PWD"), "templates")
	search := path.Join(dir, "search.html")
	layout := path.Join(dir, "layout.html")
	data := map[string]interface{}{
		"query": r.FormValue("q"),
		"count": 0,
		"title": "Search for \"" + r.FormValue("q") + "\" - " + PageTitle,
	}
	if len(r.FormValue("q")) > 0 {
		results, err := SearchUsers(c, r.FormValue("q"))
		if err != nil {
			c.Errorf("got err on search: %v", err)
		}

		switch r.FormValue("s") {
		case "overall":
			sort.Sort(ByOverall(results))
			break
		case "followers":
			sort.Sort(ByFollowers(results))
			break
		case "loops":
			sort.Sort(ByLoops(results))
			break
		case "posts":
			sort.Sort(ByPosts(results))
			break
		case "revines":
			sort.Sort(ByRevines(results))
			break
		}

		if r.Method == "GET" {
			data["count"] = len(results)
			data["results"] = results
		} else if r.Method == "POST" {
			jsonData, _ := json.Marshal(results)
			fmt.Fprint(w, string(jsonData))
			return
		}
	}

	page := mustache.RenderFileInLayout(search, layout, data)
	fmt.Fprint(w, page)
}

func DonateHandler(w http.ResponseWriter, r *http.Request) {
	dir := path.Join(os.Getenv("PWD"), "templates")
	donate := path.Join(dir, "donate.html")
	layout := path.Join(dir, "layout.html")
	page := mustache.RenderFileInLayout(donate, layout, nil)
	fmt.Fprint(w, page)
}

func ExportHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	db := DB{c}
	vars := mux.Vars(r)

	if r.Method == "GET" {
	    StartupHandler(w, r)
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
    start := time.Now()
    var popfeedUsers []*memcache.Item
    var userFeed []string

	users, err := vineApi.GetPopularUsers(60)
	for _, u := range users {
		if _, err := GetQueuedUser(u.UserIdStr, c); err == datastore.ErrNoSuchEntity {
			QueueUser(u.UserIdStr, c)
		}
		var d bytes.Buffer
		if user, err := vineApi.GetUser(u.UserIdStr); err == nil {
		    enc := gob.NewEncoder(&d)
		    user.ProfileBackground = strings.TrimPrefix(user.ProfileBackground, "0x")
		    enc.Encode(user)
		    
		    userFeed = append(userFeed, u.UserIdStr)
		    popfeedUsers = append(popfeedUsers, &memcache.Item{
		        Key: u.UserIdStr,
		        Value: d.Bytes(),
            })
		}
	}

    popfeedUsers = append(popfeedUsers, &memcache.Item{
        Key: "popusers",
        Value: []byte(strings.Join(userFeed, ",")),
    })

    memcache.AddMulti(c, popfeedUsers)
    finish := time.Since(start)
	fmt.Fprint(w, "queuing popular users: %v w/ err %v", users, err)
	c.Infof("queueing popular users: %v w/ err %v. Took %s", users, err, finish)
}

func CronFetchHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	db := DB{c}
	start := time.Now()
	n, _ := strconv.Atoi(r.FormValue("n"))

	t := taskqueue.NewPOSTTask("/cron/fetch", map[string][]string{
	    "id":{r.FormValue("id")},
	    "n": {strconv.Itoa(n+1)},
	})
	t.Name = r.FormValue("id") + "-" + strconv.Itoa(n+1)

	db.FetchUser(r.FormValue("id"))

	finish := time.Since(start)

	if appengine.IsDevAppServer() {
	    t.Delay = (10 * time.Minute) - finish
	} else {
	    t.Delay = (24 * time.Hour) - finish
	}

	if _, err := taskqueue.Add(c, t, ""); err != nil {
	    c.Errorf("Error adding user %s to taskqueue: %v", r.FormValue("id"), err)
    }

	c.Infof("Finished cron fetch, took %s", finish)
	w.WriteHeader(200)
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	dir := path.Join(os.Getenv("PWD"), "templates")
	notFound := path.Join(dir, "404.html")
	layout := path.Join(dir, "layout.html")
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

func AdminHandler(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    db := DB{c}
    vineApi := VineRequest{c}
    adminUser := user.Current(c)
    var err error
    if adminUser == nil {
        url, _ := user.LoginURL(c, "/admin/dashboard")
        http.Redirect(w, r, url, 301)
        return
    } else if !adminUser.Admin {
        w.WriteHeader(401)
        return
    }

    if r.Method == "GET" {
        dir := path.Join(os.Getenv("PWD"), "templates")
        admin := path.Join(dir, "admin.html")
        layout := path.Join(dir, "layout.html")
        data := map[string]interface{}{"user": adminUser.String(), "config": Config,}
        page := mustache.RenderFileInLayout(admin, layout, data)
        fmt.Fprint(w, page)
    } else if r.Method == "POST" {
        switch r.FormValue("op") {
            case "TaskUsers":
                for _, user := range strings.Split(r.FormValue("v"), "\n") {
                    u := strings.Split(user, ",")
                    t := taskqueue.NewPOSTTask("/cron/fetch", map[string][]string{
                        "id": {strings.TrimSpace(u[0])},
                        "n": {strings.TrimSpace(u[1])},
                    })
                    t.Name = u[0] + "-0"
                    t.Delay, err = time.ParseDuration(strings.TrimSpace(u[2]))

                    if err != nil {
                        c.Errorf("Error parsing task delay %v: %v", u, err)
                        continue;
                    }

                    if _, err = taskqueue.Add(c, t, ""); err != nil {
                        c.Errorf("Error adding user %s to taskqueue: %v", u[0], err)
                    }
                }
            case "UnqueueUser":
                c.Infof("unqueuing %v", r.FormValue("v"))
                db.UnqueueUser(r.FormValue("v"))
            case "BatchUsers":
                users := strings.Split(r.FormValue("v"), ",")
                c.Infof("queueing users: %v", users)
                for _, v := range users {
                    QueueUser(strings.TrimSpace(v), c)
                }
            case "FeaturedUser":
                if user, err := vineApi.GetUser(r.FormValue("user")); err == nil {
                    key := datastore.NewKey(c, "Misc", "featuredUser", 0, nil)
                    featuredUser := &FeaturedUser{user.UserIdStr, r.FormValue("vine")}
                    if _, err := datastore.Put(c, key, featuredUser); err != nil {
                        http.Error(w, err.Error(), 500)
                        return
                    } else {
                        items := []*memcache.Item{&memcache.Item{
                            Key: "featuredUser",
                            Value: []byte(user.UserIdStr + ";" + r.FormValue("vine")),
                        }}
                        var d bytes.Buffer
                        enc := gob.NewEncoder(&d)
                        user.ProfileBackground = strings.TrimPrefix(user.ProfileBackground, "0x")
                        enc.Encode(user)
                        items = append(items, &memcache.Item{
                            Key: user.UserIdStr,
                            Value: d.Bytes(),
                        })
                        memcache.AddMulti(c, items)
                    }
                } else {
                    http.Error(w, err.Error(), 500)
                    return
                }
        }
        fmt.Fprintf(w, "{\"op\":\"%v\",\"success\":true}", r.FormValue("op"))
    }
}