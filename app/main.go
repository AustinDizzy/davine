package main

import (
	"appengine"
	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

var Config map[string]string

func init() {
	router := mux.NewRouter()
	router.HandleFunc("/user", UserStoreHandler).Methods("POST").Queries("id", "")
	router.HandleFunc("/u/{user}", UserFetchHandler).Methods("GET")
	router.HandleFunc("/about", AboutHandler).Methods("GET")
	router.HandleFunc("/top", TopHandler).Methods("GET")
	router.HandleFunc("/discover", DiscoverHandler).Methods("GET")
	router.HandleFunc("/search", SearchHandler).Methods("GET", "POST")
	router.HandleFunc("/random/user", RandomHandler).Methods("GET")
	router.HandleFunc("/donate", DonateHandler).Methods("GET")
	router.HandleFunc("/x/{user}", ExportHandler).Methods("GET", "POST")
	router.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	router.HandleFunc("/admin/dashboard", AdminHandler).Methods("GET", "POST")

	router.HandleFunc("/cron/fetch", CronFetchHandler).Methods("POST")
	router.HandleFunc("/cron/popular", PopularFetchHandler).Methods("GET")

	if appengine.IsDevAppServer() {
		configFile, _ := ioutil.ReadFile(path.Join(os.Getenv("PWD"), "config.yaml"))
		yaml.Unmarshal(configFile, &Config)
	}

	router.HandleFunc("/_ah/start", StartupHandler).Methods("GET")
	router.HandleFunc("/_ah/warmup", StartupHandler).Methods("GET")

	http.Handle("/", router)
}

func main() {
	//just so we can compile
}
