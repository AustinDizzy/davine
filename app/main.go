package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func init() {
	router := mux.NewRouter()
	router.HandleFunc("/user", UserStoreHandler).Methods("POST").Queries("id", "")
	router.HandleFunc("/u/{user}", UserFetchHandler).Methods("GET")
	router.HandleFunc("/about", AboutHandler).Methods("GET")
	router.HandleFunc("/top", TopHandler).Methods("GET")
	router.HandleFunc("/discover", DiscoverHandler).Methods("GET")
	router.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	router.HandleFunc("/cron/fetch", CronFetchHandler).Methods("GET")

	http.Handle("/", router)
}

func main() {
    //just so we can compile
}