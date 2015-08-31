package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func init() {
	router := mux.NewRouter()
	router.HandleFunc("/", IndexHandler)
	router.HandleFunc("/user", UserStoreHandler).Methods("POST").Queries("id", "")
	router.HandleFunc("/u/{user}", UserFetchHandler).Methods("GET")
	router.HandleFunc("/about", AboutHandler).Methods("GET")
	router.HandleFunc("/top", TopHandler).Methods("GET")
	router.HandleFunc("/discover", DiscoverHandler).Methods("GET")
	router.HandleFunc("/search", SearchHandler).Methods("GET", "POST")
	router.HandleFunc("/random/user", RandomHandler).Methods("GET")
	router.HandleFunc("/donate", DonateHandler).Methods("GET")
	router.HandleFunc("/x/{user}", ExportHandler).Methods("GET", "POST")
	router.HandleFunc("/sign-up", SignUpHandler).Methods("GET", "POST")
	router.PathPrefix("/api/").HandlerFunc(ApiRouter).Methods("GET", "POST")
	router.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	router.HandleFunc("/admin/dashboard", AdminHandler).Methods("GET", "POST")

	router.HandleFunc("/cron/fetch", CronFetchHandler).Methods("POST")
	router.HandleFunc("/cron/popular", PopularFetchHandler).Methods("GET")
	router.HandleFunc("/cron/explore", CronExploreHandler).Methods("GET", "POST")
	router.HandleFunc("/cron/report", CronReportHandler).Methods("POST")
	router.HandleFunc("/cron/flush", CronFlushHandler).Methods("GET")

	router.HandleFunc("/_ah/start", StartupHandler).Methods("GET")
	router.HandleFunc("/_ah/warmup", StartupHandler).Methods("GET")
	router.HandleFunc("/_ah/mail/{email}", EmailHandler).Methods("POST")

	http.Handle("/", router)
}

func main() {
	//just so we can compile
}
