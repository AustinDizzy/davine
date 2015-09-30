package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"app/data"
	"app/utils"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
)

//CronFetchHandler is the http request handler for /cron/fetch.
//It is the endpoint which initiates a data fetch for a specific user,
//then retasks the user for another data fetch in 24hrs.
func CronFetchHandler(w http.ResponseWriter, r *http.Request) {
	var (
		c     = appengine.NewContext(r)
		db    = data.NewRequest(c)
		start = time.Now()
		n, _  = strconv.Atoi(r.FormValue("n"))
		t     = taskqueue.NewPOSTTask("/cron/fetch", map[string][]string{
			"id": {r.FormValue("id")},
			"n":  {strconv.Itoa(n + 1)},
		})
	)

	t.Name = fmt.Sprintf("%s-%d-%s", r.FormValue("id"), n+1, utils.GenSlug())

	err := db.FetchUser(r.FormValue("id"))

	finish := time.Since(start)
	utils.PostValue(c, "cron fetch", finish.Seconds()*1000.0)

	if appengine.IsDevAppServer() {
		t.Delay = (10 * time.Minute) - finish
	} else {
		t.Delay = (24 * time.Hour) - finish
	}

	if _, taskErr := taskqueue.Add(c, t, ""); taskErr != nil {
		log.Errorf(c, "Error adding user %s to taskqueue: %v", r.FormValue("id"), taskErr)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
