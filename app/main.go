package main

import (
    "fmt"
    "net/http"
    "github.com/gorilla/mux"
)

func init() {
    router := mux.NewRouter()
    router.HandleFunc("/user", handler).Methods("GET")
    
    http.Handle("/", router)
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Hello, world!")
}