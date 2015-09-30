package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"app/data"
	"app/page"

	"github.com/austindizzy/vine-go"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

//UserFetchHandler is the http request handler for /u/(userID|vanity).
//This is the "profile" route that renders all the user's acrued data.
func UserFetchHandler(w http.ResponseWriter, r *http.Request) {
	var (
		p          = page.New("profile.html")
		vars       = mux.Vars(r)
		c          = appengine.NewContext(r)
		db         = data.NewRequest(c)
		err        error
		userRecord *data.UserRecord
	)

	if !vine.IsVanity(vars["user"]) {
		userID, _ := strconv.ParseInt(vars["user"], 10, 64)
		userRecord, err = db.GetUser(userID)
	} else {
		q := datastore.NewQuery("UserRecord").Filter("Vanity =", strings.ToLower(vars["user"])).KeysOnly().Limit(1)
		k, _ := q.GetAll(c, nil)

		if len(k) > 0 {
			userRecord, _ = db.GetUser(k[0].IntID())
		} else {
			err = datastore.ErrNoSuchEntity
		}
	}

	if err == datastore.ErrNoSuchEntity {
		p = page.New("user404.html")
		p.LoadData(map[string]string{"user": vars["user"]})
		w.WriteHeader(http.StatusNotFound)
		p.Write(w)
		return
	} else if err != nil {
		log.Errorf(c, "got error on fetching user %s: %v", vars["user"], err)
		w.WriteHeader(http.StatusInternalServerError)
		return
		//TODO: Write a 500 error page.
	}

	if userRecord != nil {

		userData := userRecord

		if userData.ProfileBackground != "" {
			color := strings.SplitAfterN(userData.ProfileBackground, "0x", 2)
			userData.ProfileBackground = color[1]
		} else {
			userData.ProfileBackground = "00BF8F"
		}

		jsonStr, err := json.Marshal(userData.UserData)
		if err == nil {
			userData.UserDataJsonStr = string(jsonStr)
		}
		jsonStr, err = json.Marshal(userData.UserMeta)
		if err == nil {
			userData.UserMetaJsonStr = string(jsonStr)
		}

		p.LoadData(userData)
		p.Write(w)
	}
}
