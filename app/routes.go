package main

import (
	"appengine"
	"appengine/datastore"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hoisie/mustache"
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

	_, err := GetQueuedUser(r.FormValue("id"), c)
	data := make(map[string]bool)

	if err == datastore.ErrNoSuchEntity {

		_, err := vineApi.GetUser(r.FormValue("id"))

		if err != nil {
			data["exists"] = false
			data["queued"] = false
		} else {
		    go QueueUser(r.FormValue("id"), c)
			data["exists"] = true
			data["queued"] = true
		}

		data["stored"] = false

	} else {
		data["exists"] = true
		data["stored"] = true
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
    dir := path.Join(os.Getenv("PWD"), "templates")
    discover := path.Join(dir, "discover.html")
    layout := path.Join(dir, "pageLayout.html")
    page := mustache.RenderFileInLayout(discover, layout)
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

func CronFetchHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	q := datastore.NewQuery("Queue").KeysOnly()
	keys, _ := q.GetAll(c, nil)
	db := DB{c}

	for _, v := range keys {
		go db.FetchUser(strconv.FormatInt(v.IntID(), 10))
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