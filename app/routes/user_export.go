package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"app/config"
	"app/data"
	"app/page"
	"app/utils"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

//UserExportHandler is the http request handler for /x/userID
//where userID is a valid user ID that Davine has previously discovered.
//It exports all the user's data using data.ExportUser.
func UserExportHandler(w http.ResponseWriter, r *http.Request) {
	var (
		c    = appengine.NewContext(r)
		db   = data.NewRequest(c)
		vars = mux.Vars(r)
		cnfg = config.Load(c)
	)

	if r.Method == "GET" {
		userID, err := strconv.ParseInt(vars["user"], 10, 64)
		if err != nil {
			log.Errorf(c, "got err: %v", err)
			http.Redirect(w, r, "/404", 301)
			return
		}
		userRecord, err := db.GetUserRecord(userID)
		if err == datastore.ErrNoSuchEntity {
			http.Redirect(w, r, "/404", 301)
			return
		}
		data := map[string]string{"username": userRecord.Username, "userId": userRecord.UserId, "captcha": cnfg["captchaPublic"]}
		p := page.New("export.html")
		p.LoadData(data)
		p.Write(w)
	} else if r.Method == "POST" {
		captcha := utils.VerifyCaptcha(c, map[string]string{
			"response": r.FormValue("g-recaptcha-response"),
			"remoteip": r.RemoteAddr,
		})
		if captcha {
			db.ExportUser(vars["user"], w)
		} else {
			fmt.Fprint(w, "Seems like your CAPTCHA was wrong. Please press back and try again.")
		}
	}
}
