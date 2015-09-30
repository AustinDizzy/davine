package main

import (
	"net/http"

	"app/admin"
	"app/routes"

	"github.com/gorilla/mux"

	_ "google.golang.org/appengine/remote_api"
)

func init() {
	router := mux.NewRouter()
	router.HandleFunc("/", routes.IndexHandler)
	router.HandleFunc("/user", routes.UserStoreHandler).Methods("POST").Queries("id", "")
	router.HandleFunc("/u/{user}", routes.UserFetchHandler).Methods("GET")
	router.HandleFunc("/about", routes.AboutHandler).Methods("GET")
	router.HandleFunc("/top", routes.TopHandler).Methods("GET")
	router.HandleFunc("/discover", routes.DiscoverHandler).Methods("GET")
	router.HandleFunc("/search", routes.SearchHandler).Methods("GET", "POST")
	router.HandleFunc("/random/user", routes.RandomHandler).Methods("GET")
	router.HandleFunc("/donate", routes.DonateHandler).Methods("GET")
	router.HandleFunc("/x/{user}", routes.UserExportHandler).Methods("GET", "POST")
	router.HandleFunc("/sign-up", routes.SignUpHandler).Methods("GET", "POST")
	router.PathPrefix("/api/").HandlerFunc(routes.APIRouter).Methods("GET", "POST")
	router.NotFoundHandler = http.HandlerFunc(routes.NotFoundHandler)

	router.HandleFunc("/admin/dashboard", admin.Handler).Methods("GET", "POST")

	router.HandleFunc("/cron/fetch", routes.CronFetchHandler).Methods("POST")
	router.HandleFunc("/cron/popular", routes.PopularFetchHandler).Methods("GET")
	router.HandleFunc("/cron/explore", routes.CronExploreHandler).Methods("GET", "POST")
	router.HandleFunc("/cron/report", routes.CronReportHandler).Methods("POST")
	router.HandleFunc("/cron/flush", routes.CronFlushHandler).Methods("GET")
	router.HandleFunc("/cron/import", routes.CronImportHandler).Methods("POST")
	router.HandleFunc("/cron/purge", routes.CronPurgeHandler).Methods("POST")

	router.HandleFunc("/_ah/start", routes.StartupHandler).Methods("GET")
	router.HandleFunc("/_ah/warmup", routes.StartupHandler).Methods("GET")
	router.HandleFunc("/_ah/mail/{email}", routes.EmailHandler).Methods("POST")

	http.Handle("/", router)
}

func main() {
	//just so we can compile
}
