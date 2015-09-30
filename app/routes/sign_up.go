package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"app/admin"
	"app/config"
	"app/data"
	"app/email"
	"app/page"
	"app/utils"

	"github.com/austindizzy/vine-go"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
)

//SignUpHandler is the http request handler for /sign-up.
//On GET, it renders the sign-up page for users to sign up
//to the weekly email service.
//On GET where type=activate, it confirms and activates a user's
//email address for the email service. Without activating, the
//user will not receive any email for the report they signed up for.
//On POST, it stores and validates the user's sign up submission.
//This also handles enterprise users who are interested in raw access
//to the Vine dataset.
func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var (
		c        = appengine.NewContext(r)
		vineApi  = vine.NewRequest(urlfetch.Client(c))
		appUser  = new(admin.AppUser)
		pageData = map[string]interface{}{}
	)
	if r.Method == "GET" {
		if r.FormValue("type") == "activate" && len(r.FormValue("key")) > 0 && len(r.FormValue("email")) > 0 {
			key := datastore.NewKey(c, "AppUser", r.FormValue("email"), 0, nil)
			if err := datastore.Get(c, key, appUser); err != nil {
				log.Infof(c, "error activating user %s: %v\ndata: %v", r.FormValue("email"), err, r.Form)
			} else {
				if strings.Split(appUser.AuthKey, ";")[1] == r.FormValue("key") {
					appUser.Active = true
					if _, err := datastore.Put(c, key, appUser); err != nil {
						log.Errorf(c, "error saving activated user %s: %v", r.FormValue("email"), err)
						return
					} else {
						t := taskqueue.NewPOSTTask("/cron/report", map[string][]string{
							"id":    []string{appUser.UserIdStr},
							"email": []string{r.FormValue("email")},
						})

						if _, err := taskqueue.Add(c, t, "reports"); err != nil {
							log.Errorf(c, "error tasking user %s report: %v", appUser.UserIdStr, err)
							fmt.Fprintf(w, "There was a problem confirming your subscription. Please try again and contact us if this problem persists.")
						} else {
							fmt.Fprintf(w, "Your email subscription is now activated. You may close this page.")
						}
					}
				} else {
					log.Infof(c, "authKey: %s\nsuppliedKey: %s", strings.Split(appUser.AuthKey, ";")[1], r.FormValue("key"))
					http.Error(w, "The supplied key did not match with our records.", http.StatusBadRequest)
				}
			}
		} else {
			p := page.New("signup.html")
			cnfg := config.Load(c)
			pageData = map[string]interface{}{
				"title":   "Sign Up",
				"captcha": cnfg["captchaPublic"],
			}
			p.LoadData(pageData)
			p.Write(w)
		}
	} else if r.Method == "POST" {
		captcha := utils.VerifyCaptcha(c, map[string]string{
			"response": r.FormValue("g-recaptcha-response"),
			"remoteip": r.RemoteAddr,
		})
		key := datastore.NewKey(c, "AppUser", r.FormValue("email"), 0, nil)
		if r.FormValue("type") == "enterprise" && len(r.FormValue("email")) > 0 {
			appUser = &admin.AppUser{
				Email:      r.FormValue("email"),
				Type:       "enterprise",
				Active:     true,
				Discovered: time.Now(),
			}
			if captcha {
				if _, err := datastore.Put(c, key, appUser); err != nil {
					pageData["success"] = false
					pageData["error"] = err.Error()
				} else {
					pageData["success"] = true
				}
			} else {
				pageData["success"] = false
				pageData["error"] = "Captcha challenge failed."
			}
		} else if r.FormValue("type") == "email-report" && len(r.FormValue("email")) > 0 {
			vineUser, err := vineApi.GetUser(r.FormValue("user"))
			if !captcha {
				pageData["success"] = false
				pageData["error"] = "Captcha challenge failed."
			} else if err != nil {
				pageData["success"] = false
				pageData["error"] = err.Error()
			} else if db := data.NewRequest(c); !db.UserQueueExist(vineUser.UserId) {
				pageData["success"] = false
				pageData["error"] = "That user doesn't appear to exist in Davine's database yet. Please submit it to us first."
			} else {
				slug := utils.GenSlug()
				appUser = &admin.AppUser{
					Email:      r.FormValue("email"),
					Type:       "email-report",
					Active:     false,
					UserIdStr:  r.FormValue("user"),
					AuthKey:    slug + ";" + utils.GenKey(),
					Discovered: time.Now(),
				}
				if _, err := datastore.Put(c, key, appUser); err == nil {
					pageData["success"] = true
					pageData["code"] = slug
				} else {
					log.Errorf(c, "got appUser store err: %v", err)
				}
			}
		} else if r.FormValue("type") == "email-ping" {
			err := datastore.Get(c, key, appUser)
			if err == nil {
				if u, err := vineApi.GetUser(appUser.UserIdStr); err != nil {
					pageData["success"] = false
					pageData["error"] = err.Error()
				} else {
					authKey := strings.Split(appUser.AuthKey, ";")
					if strings.Contains(u.Description, authKey[0]) {
						pageData["success"] = true
						emailData := map[string]interface{}{
							"username": u.Username,
							"id":       u.UserIdStr,
							"link":     fmt.Sprintf("https://%s/sign-up?type=activate&key=%s&email=%s", appengine.DefaultVersionHostname(c), authKey[1], r.FormValue("email")),
						}
						msg := email.New()
						msg.LoadTemplate(0, emailData)
						msg.To = []string{r.FormValue("email")}
						if err := msg.Send(c); err != nil {
							log.Errorf(c, "error sending %s email to %s: %v", u.UserIdStr, msg.To[0], err)
							pageData["success"] = false
							pageData["error"] = err.Error()
						}
					} else {
						pageData["success"] = false
					}
				}
			} else {
				pageData["success"] = false
				pageData["error"] = err.Error()
			}
		}
		json.NewEncoder(w).Encode(pageData)
	}
}
