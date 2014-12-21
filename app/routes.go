package main

import (
    "appengine"
    "appengine/datastore"
    "fmt"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/hoisie/mustache"
    "path"
    "os"
    "encoding/json"
)

func UserFetchHandler(w http.ResponseWriter, r *http.Request) {
    template := path.Join(path.Join(os.Getenv("PWD"), "templates"), "message.html.mustache")
    vars := mux.Vars(r)
    data := mustache.RenderFile(template, map[string]string{"message": "Hello, " + vars["user"]})
    fmt.Fprint(w, data)
}

func UserStoreHandler(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    vineApi := VineRequest{c}

    key := datastore.NewKey(c, "Queue", r.FormValue("id"), 0, nil)
    data := make(map[string]bool)

    if err := datastore.Get(c, key, nil); err != nil && err == datastore.ErrNoSuchEntity {

        go QueueUser(r.FormValue("id"), c)

        _, err := vineApi.GetUser(r.FormValue("id"))

        if err == ErrUserDoesntExist {
            data["exists"] = false
        } else {
            data["exists"] = true
        }

        data["stored"] = false
        data["queued"] = true

    } else {
        data["exists"] = true
        data["stored"] = true
        data["queued"] = false
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}

func CronFetchHandler(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

	q := datastore.NewQuery("Queue").KeysOnly()
	keys, _ := q.GetAll(c, nil)
	db := DB{c}

	for _, v := range keys {
	    go db.FetchUser(v.StringID())
	}

	fmt.Fprint(w, "fetching users")
}