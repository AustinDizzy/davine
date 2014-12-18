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
    key := datastore.NewKey(c, "Queue", r.FormValue("id"), 0, nil)
    data := make(map[string]bool)
    
    if err := datastore.Get(c, key, nil); err != nil && err == datastore.ErrNoSuchEntity {
        
        go QueueUser(r.FormValue("id"), c)
        
        data["exists"] = false
        data["queued"] = true
        
    } else {
        data["exists"] = true
        data["queued"] = false
    }
    
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