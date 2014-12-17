package main

import (
    "fmt"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/hoisie/mustache"
    "path"
    "os"
)

func init() {
    router := mux.NewRouter()
    router.HandleFunc("/user", UserStoreHandler).Methods("GET")
    router.HandleFunc("/u/{user}", UserFetchHandler).Methods("GET")
    
    http.Handle("/", router)
}

func UserFetchHandler(w http.ResponseWriter, r *http.Request) {
    template := path.Join(path.Join(os.Getenv("PWD"), "templates"), "message.html.mustache")
    vars := mux.Vars(r)
    data := mustache.RenderFile(template, map[string]string{"message": "Hello, " + vars["user"]})
    fmt.Fprint(w, data)
}

func UserStoreHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Hello, world!")
}