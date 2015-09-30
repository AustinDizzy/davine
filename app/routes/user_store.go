package routes

import (
	"encoding/json"
	"net/http"

	"app/data"

	"github.com/austindizzy/vine-go"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

//UserStoreHandler is the http request handler for /user/store.
//It stores the specified Vine user if we have not discovered it
//already.
//The JSON response from this route is parsed by the client
//and the client is redirected to the user's profile if necessary.
func UserStoreHandler(w http.ResponseWriter, r *http.Request) {
	var (
		c       = appengine.NewContext(r)
		vineAPI = vine.NewRequest(urlfetch.Client(c))
		db      = data.NewRequest(c)
		data    = make(map[string]bool)
	)
	u, err := db.GetQueuedUser(r.FormValue("id"))

	if err != datastore.ErrNoSuchEntity && err != nil {
		log.Errorf(c, "got UserStore err: %v", err)
	}

	user, apiErr := vineAPI.GetUser(r.FormValue("id"))

	if err == datastore.ErrNoSuchEntity || u == nil {
		if apiErr != nil {
			log.Infof(c, "Got apiErr: %v", apiErr)
			data["exists"] = false
			data["queued"] = false
		} else {
			db.QueueUser(user.UserIdStr)
			data["exists"] = true
			data["queued"] = true
		}

		data["stored"] = false

	} else {
		_, err := db.GetUserRecord(user.UserId)
		if err == datastore.ErrNoSuchEntity {
			data["stored"] = false
		} else {
			data["stored"] = true
		}
		data["exists"] = true
		data["queued"] = false
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
