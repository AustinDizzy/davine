package main

import (
    "net/http"
    "github.com/gorilla/mux"
)

func init() {
    router := mux.NewRouter()
    router.HandleFunc("/user", UserStoreHandler).Methods("POST").Queries("id", "")
    router.HandleFunc("/u/{user}", UserFetchHandler).Methods("GET")
    
    router.HandleFunc("/cron/fetch", CronFetchHandler).Methods("GET")

    http.Handle("/", router)
}