package main

import (
	"app/admin"
	"app/config"
	"app/counter"
	"app/email"
	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"appengine/taskqueue"
	"appengine/urlfetch"
	"appengine/user"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	"github.com/hoisie/mustache"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	var (
		dir    = path.Join(os.Getenv("PWD"), "templates")
		index  = path.Join(dir, "index.html")
		layout = path.Join(dir, "layout.html")
		c      = appengine.NewContext(r)
		cnfg   = config.Load(c)
		data   = map[string]interface{}{
			"title":   PageTitle,
			"captcha": cnfg["captchaPublic"],
		}
		err error
	)

	for _, k := range []string{"TotalLoops", "TotalPosts", "TotalVerified", "24hLoops", "24hPosts", "24hUsers"} {
		if data[k], err = counter.Count(c, k); err != nil {
			c.Errorf("Error getting %s: %v", k, err)
		}
	}

	if popusers, err := memcache.Get(c, "popusers"); err == nil {
		var users []*VineUser
		var dec *gob.Decoder
		userIDs := strings.Split(string(popusers.Value[:]), ",")
		for len(users) < 6 {
			rand.Seed(time.Now().Unix())
			r := rand.Intn(len(userIDs))

			key, err := memcache.Get(c, userIDs[r])
			dec = gob.NewDecoder(bytes.NewReader(key.Value))
			var u *VineUser
			err = dec.Decode(&u)
			if err == nil {
				users = append(users, u)
				userIDs = append(userIDs[:r], userIDs[r+1:]...)
				//above removes already chosen user from userID array
			}
		}
		data["popusers"] = users
	} else {
		c.Errorf("popusers memcache err: %v", err)
	}

	if featuredUser, err := memcache.Get(c, "featuredUser"); err == nil {
		var user *VineUser
		r := strings.Split(string(featuredUser.Value[:]), ";")
		u, _ := memcache.Get(c, r[0])
		dec := gob.NewDecoder(bytes.NewReader(u.Value))
		dec.Decode(&user)

		data["featuredUser"] = user
		data["featuredPost"] = r[1]
	}

	fmt.Fprint(w, mustache.RenderFileInLayout(index, layout, data))
}

func UserFetchHandler(w http.ResponseWriter, r *http.Request) {
	dir := path.Join(os.Getenv("PWD"), "templates")
	profile := path.Join(dir, "profile.html")
	layout := path.Join(dir, "layout.html")
	vars := mux.Vars(r)
	c := appengine.NewContext(r)

	db := DB{c}
	match, _ := regexp.MatchString("^[0-9]+$", vars["user"])
	var userRecord *UserRecord
	var err error
	var data string

	if match {
		userId, _ := strconv.ParseInt(vars["user"], 10, 64)
		userRecord, err = db.GetUser(userId)
	} else {
		q := datastore.NewQuery("UserRecord").Filter("Vanity =", strings.ToLower(vars["user"])).KeysOnly().Limit(1)
		k, _ := q.GetAll(c, nil)

		if len(k) > 0 {
			userRecord, _ = db.GetUser(k[0].IntID())
		} else {
			user404 := path.Join(dir, "user404.html")
			userData := map[string]string{"user": vars["user"]}
			data = mustache.RenderFileInLayout(user404, layout, userData)
			w.WriteHeader(http.StatusNotFound)
		}
	}

	if err == datastore.ErrNoSuchEntity {
		user404 := path.Join(dir, "user404.html")
		userData := map[string]string{"user": vars["user"]}
		data = mustache.RenderFileInLayout(user404, layout, userData)
		w.WriteHeader(http.StatusNotFound)
	} else if err != nil {
		c.Errorf("got error on fetching user %s: %v", vars["user"], err)
		fmt.Fprint(w, err.Error())
	} else if userRecord != nil {

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

		data = mustache.RenderFileInLayout(profile, layout, userData)
	}

	fmt.Fprint(w, data)
}

func UserStoreHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	vineApi := VineRequest{c}
	db := DB{c}
	u, err := GetQueuedUser(r.FormValue("id"), c)
	data := make(map[string]bool)

	if err != datastore.ErrNoSuchEntity && err != nil {
		c.Errorf("got UserStore err: %v", err)
	}

	user, apiErr := vineApi.GetUser(r.FormValue("id"))

	if err == datastore.ErrNoSuchEntity || u == nil {
		if apiErr != nil {
			c.Infof("Got apiErr: %v", apiErr)
			data["exists"] = false
			data["queued"] = false
		} else {
			QueueUser(user.UserIdStr, c)
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

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	dir := path.Join(os.Getenv("PWD"), "templates")
	aboutPage := path.Join(dir, "about.html")
	layout := path.Join(dir, "layout.html")

	db := DB{appengine.NewContext(r)}
	totalUsers, _ := db.GetTotalUsers()
	stats := map[string]interface{}{"users": totalUsers}
	data := mustache.RenderFileInLayout(aboutPage, layout, stats)

	fmt.Fprint(w, data)
}

func DiscoverHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	data := map[string]interface{}{
		"title": "Discover - " + PageTitle,
	}
	data["totalUsers"], _ = counter.Count(c, "TotalUsers")
	data["24hUsers"], _ = counter.Count(c, "24hUsers")
	data["totalVerified"], _ = counter.Count(c, "TotalVerified")
	data["totalExplicit"], _ = counter.Count(c, "TotalExplicit")

	dir := path.Join(os.Getenv("PWD"), "templates")
	discover := path.Join(dir, "discover.html")
	layout := path.Join(dir, "layout.html")
	page := mustache.RenderFileInLayout(discover, layout, data)
	fmt.Fprint(w, page)
}

func TopHandler(w http.ResponseWriter, r *http.Request) {
	var (
		c      = appengine.NewContext(r)
		db     = DB{c}
		cnfg   = config.Load(c)
		dir    = path.Join(os.Getenv("PWD"), "templates")
		top    = path.Join(dir, "top.html")
		layout = path.Join(dir, "layout.html")
		data   = db.GetTop()
	)

	data["title"] = "Top - " + PageTitle
	data["captcha"] = cnfg["captchaPublic"]
	page := mustache.RenderFileInLayout(top, layout, data)

	fmt.Fprint(w, page)
}

func RandomHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("UserMeta").KeysOnly()
	keys, err := q.GetAll(c, nil)
	if err != nil {
		c.Errorf("got err %v", err)
		return
	}
	randomKey := RandomKey(keys)
	var user QueuedUser
	key := datastore.NewKey(c, "Queue", "", randomKey.IntID(), nil)
	err = datastore.Get(c, key, &user)
	if err != nil {
		c.Errorf("got err %v", err)
	} else {
		c.Infof("got user %v", user)
	}
	http.Redirect(w, r, "/u/"+user.UserID, 301)
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	dir := path.Join(os.Getenv("PWD"), "templates")
	search := path.Join(dir, "search.html")
	layout := path.Join(dir, "layout.html")
	data := map[string]interface{}{
		"query": r.FormValue("q"),
		"count": 0,
		"title": "Search for \"" + r.FormValue("q") + "\" - " + PageTitle,
	}
	if len(r.FormValue("q")) > 0 {
		results, err := SearchUsers(c, r.FormValue("q"))
		if err != nil {
			c.Errorf("got err on search: %v", err)
		}

		switch r.FormValue("s") {
		case "overall":
			sort.Sort(ByOverall(results))
			break
		case "followers":
			sort.Sort(ByFollowers(results))
			break
		case "loops":
			sort.Sort(ByLoops(results))
			break
		case "posts":
			sort.Sort(ByPosts(results))
			break
		case "revines":
			sort.Sort(ByRevines(results))
			break
		}

		if r.Method == "GET" {
			data["count"] = len(results)
			data["results"] = results
		} else if r.Method == "POST" {
			jsonData, _ := json.Marshal(results)
			fmt.Fprint(w, string(jsonData))
			return
		}
	}

	page := mustache.RenderFileInLayout(search, layout, data)
	fmt.Fprint(w, page)
}

func DonateHandler(w http.ResponseWriter, r *http.Request) {
	dir := path.Join(os.Getenv("PWD"), "templates")
	donate := path.Join(dir, "donate.html")
	layout := path.Join(dir, "layout.html")
	page := mustache.RenderFileInLayout(donate, layout, nil)
	fmt.Fprint(w, page)
}

func ExportHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	db := DB{c}
	vars := mux.Vars(r)
	cnfg := config.Load(c)

	if r.Method == "GET" {
		StartupHandler(w, r)
		userId, err := strconv.ParseInt(vars["user"], 10, 64)
		if err != nil {
			c.Errorf("got err: %v", err)
			http.Redirect(w, r, "/404", 301)
			return
		}
		userRecord, err := db.GetUserRecord(userId)
		if err == datastore.ErrNoSuchEntity {
			http.Redirect(w, r, "/404", 301)
			return
		}
		data := map[string]string{"username": userRecord.Username, "userId": userRecord.UserId, "captcha": cnfg["captchaPublic"]}
		dir := path.Join(os.Getenv("PWD"), "templates")
		export := path.Join(dir, "export.html")
		layout := path.Join(dir, "layout.html")
		page := mustache.RenderFileInLayout(export, layout, data)
		fmt.Fprint(w, page)
	} else if r.Method == "POST" {
		captcha := verifyCaptcha(c, map[string]string{
			"response": r.FormValue("g-recaptcha-response"),
			"remoteip": r.RemoteAddr,
		})
		if captcha {
			export := Export{c}
			export.User(vars["user"], w)
		} else {
			fmt.Fprint(w, "Seems like your CAPTCHA was wrong. Please press back and try again.")
		}
	}
}

func PopularFetchHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	vineApi := VineRequest{c}
	start := time.Now()
	var popfeedUsers []*memcache.Item
	var userFeed []string

	users, err := vineApi.GetPopularUsers(60)
	for _, u := range users {
		if _, err := GetQueuedUser(u.UserIdStr, c); err == datastore.ErrNoSuchEntity {
			QueueUser(u.UserIdStr, c)
		}
		var d bytes.Buffer
		if user, err := vineApi.GetUser(u.UserIdStr); err == nil {
			enc := gob.NewEncoder(&d)
			user.ProfileBackground = strings.TrimPrefix(user.ProfileBackground, "0x")
			enc.Encode(user)

			userFeed = append(userFeed, u.UserIdStr)
			popfeedUsers = append(popfeedUsers, &memcache.Item{
				Key:   u.UserIdStr,
				Value: d.Bytes(),
			})
		}
	}

	popfeedUsers = append(popfeedUsers, &memcache.Item{
		Key:   "popusers",
		Value: []byte(strings.Join(userFeed, ",")),
	})

	memcache.AddMulti(c, popfeedUsers)
	finish := time.Since(start)
	fmt.Fprint(w, "queuing popular users: %v w/ err %v", users, err)
	c.Infof("queueing popular users: %v w/ err %v. Took %s", users, err, finish)
}

func CronExploreHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if r.Method == "GET" {
		feeds := []*taskqueue.Task{
			taskqueue.NewPOSTTask("/cron/explore", map[string][]string{
				"feed": {"/timelines/popular"},
			}),
			taskqueue.NewPOSTTask("/cron/explore", map[string][]string{
				"feed": {"/timelines/promoted"},
			}),
		}
		for i := 1; !appengine.IsDevAppServer() && i <= 17; i++ {
			feeds = append(feeds, []*taskqueue.Task{
				taskqueue.NewPOSTTask("/cron/explore", map[string][]string{
					"feed": {fmt.Sprintf("/timelines/channels/%d/popular", i)},
				}),
				taskqueue.NewPOSTTask("/cron/explore", map[string][]string{
					"feed": {fmt.Sprintf("/timelines/channels/%d/recent", i)},
				}),
			}...)
		}
		if _, err := taskqueue.AddMulti(c, feeds, "explore"); err != nil {
			c.Errorf("error tasking explore: %v", err)
		}
	} else if r.Method == "POST" {
		vineApi := VineRequest{c}
		userIDs, err := vineApi.ScrapeUserIDs(r.FormValue("feed"))
		if err != nil {
			c.Errorf("Error scraping %s: %v", r.FormValue("feed"), err)
		} else {
			c.Infof("%d users", len(userIDs))
			for _, u := range userIDs {
				if _, err := GetQueuedUser(u, c); err == datastore.ErrNoSuchEntity {
					QueueUser(u, c)
				}
			}
		}
	}
}

func CronFetchHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	db := DB{c}
	start := time.Now()
	n, _ := strconv.Atoi(r.FormValue("n"))

	t := taskqueue.NewPOSTTask("/cron/fetch", map[string][]string{
		"id": {r.FormValue("id")},
		"n":  {strconv.Itoa(n + 1)},
	})
	t.Name = r.FormValue("id") + "-" + strconv.Itoa(n+1)

	db.FetchUser(r.FormValue("id"))

	finish := time.Since(start)
	PostValue(c, "cron fetch", finish.Seconds()*1000.0)

	if appengine.IsDevAppServer() {
		t.Delay = (10 * time.Minute) - finish
	} else {
		t.Delay = (24 * time.Hour) - finish
	}

	if _, err := taskqueue.Add(c, t, ""); err != nil {
		c.Errorf("Error adding user %s to taskqueue: %v", r.FormValue("id"), err)
	}

	w.WriteHeader(200)
}

func CronReportHandler(w http.ResponseWriter, r *http.Request) {
	var (
		id, _   = strconv.ParseInt(r.FormValue("id"), 10, 64)
		c       = appengine.NewContext(r)
		msg     = email.New()
		db      = DB{c}
		u, err  = db.GetUser(id)
		appUser = new(AppUser)
		key     = datastore.NewKey(c, "AppUser", r.FormValue("email"), 0, nil)
	)

	c.Infof("generating email for %s at %s", r.FormValue("id"), r.FormValue("email"))

	if err != nil {
		c.Errorf("error sending user report for %s: %v", r.FormValue("id"), err)
		return
	}

	if err = datastore.Get(c, key, &appUser); err != nil && !(err == datastore.ErrNoSuchEntity && user.IsAdmin(c)) {
		c.Errorf("error reading appUser for %s: %v", r.FormValue("email"), err)
		return
	}

	mon := []string{"Jan", "Feb", "Mar", "Apr", "May", "June", "Jul", "Aug", "Sept", "Oct", "Nov", "Dec"}
	d := u.UserData[len(u.UserData)-7:]

	chart, err := GenSummaryChart(c, u)
	data := map[string]interface{}{
		"dateStart":    fmt.Sprintf("%s. %d", mon[d[0].Recorded.Month()], d[0].Recorded.Day()),
		"dateEnd":      fmt.Sprintf("%s. %d", mon[d[len(d)-1].Recorded.Month()], d[len(d)-1].Recorded.Day()),
		"newPosts":     humanize.Comma(d[len(d)-1].Posts - d[0].Posts),
		"newLoops":     humanize.Comma(d[len(d)-1].Loops - d[0].Loops),
		"newFollowers": humanize.Comma(d[len(d)-1].Followers - d[0].Followers),
		"loops":        humanize.Comma(u.LoopCount),
		"followers":    humanize.Comma(u.FollowerCount),
		"posts":        humanize.Comma(u.PostCount),
		"following":    humanize.Comma(u.FollowingCount),
		"revines":      humanize.Comma(u.RevineCount),
		"chartData":    chart,
		"user":         u,
	}

	if len(appUser.AuthKey) > 0 {
		data["key"] = strings.Split(appUser.AuthKey, ";")[1]
	}

	for _, v := range []string{"newPosts", "newLoops", "newFollowers"} {
		if !strings.Contains(data[v].(string), "-") {
			data[v] = "+" + data[v].(string)
		}
	}

	msg.LoadTemplate(1, data)
	msg.To = []string{r.FormValue("email")}
	msg.Send(c)

	t := taskqueue.NewPOSTTask("/cron/report", map[string][]string{
		"id":    []string{r.FormValue("id")},
		"email": []string{r.FormValue("email")},
	})

	t.Delay = 7 * 24 * time.Hour

	if _, err := taskqueue.Add(c, t, "reports"); err != nil {
		c.Infof("error queuing email report for %s: %v", r.FormValue("email"), err)
	}
}

func CronFlushHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	for _, k := range []string{"24hLoops", "24hPosts", "24hUsers"} {
		if n, err := counter.Count(c, k); err != nil {
			c.Errorf("got err sending stat %s: %v", k, n)
		} else {
			PostCount(c, k, int(n))
		}
		if err := counter.Delete(c, k); err != nil {
			c.Errorf("got err flushing %s: %v", k, err)
		}
	}
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	dir := path.Join(os.Getenv("PWD"), "templates")
	notFound := path.Join(dir, "404.html")
	layout := path.Join(dir, "layout.html")
	data := map[string]string{"url": r.RequestURI}
	page := mustache.RenderFileInLayout(notFound, layout, data)
	w.WriteHeader(404)
	fmt.Fprint(w, page)
}

func StartupHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	PostCount(c, "new instance", 1)
}

func EmailHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if msg, err := email.Read(r.Body); err != nil {
		c.Errorf("err reading email: %v", err)
	} else {
		switch strings.Split(msg.Header.Get("To"), "@")[0] {
		case "share":
			regex := regexp.MustCompile(`(?:vine.co/u/)([0-9]+)`)
			matches := regex.FindAllStringSubmatch(msg.Body.Text, 1)
			if len(matches) > 0 {
				client := urlfetch.Client(c)
				url := fmt.Sprintf("http://%s/user?id=%s", appengine.DefaultVersionHostname(c), matches[0][1])
				req, _ := http.NewRequest("POST", url, nil)
				resp, err := client.Do(req)
				if err != nil {
					c.Errorf("got err: %v", err)
				} else {
					body, _ := ioutil.ReadAll(resp.Body)
					var data map[string]bool
					json.Unmarshal(body, &data)
					if data["exists"] {
						msg := email.New()
						if data["stored"] {
							msg.LoadTemplate(2, map[string]interface{}{
								"stored": strconv.FormatBool(data["stored"]),
								"id":     matches[0][1],
							})
						} else {
							msg.LoadTemplate(2, map[string]interface{}{
								"id": matches[0][1],
							})
						}
						msg.Send(c)
						PostCount(c, "shared via email", 1)
					}
				}
			}
		}
	}
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	vineApi := VineRequest{c}
	data := map[string]interface{}{}
	appUser := new(AppUser)
	if r.Method == "GET" {
		if r.FormValue("type") == "activate" && len(r.FormValue("key")) > 0 && len(r.FormValue("email")) > 0 {
			key := datastore.NewKey(c, "AppUser", r.FormValue("email"), 0, nil)
			if err := datastore.Get(c, key, appUser); err != nil {
				c.Infof("error activating user %s: %v\ndata: %v", r.FormValue("email"), err, r.Form)
			} else {
				if strings.Split(appUser.AuthKey, ";")[1] == r.FormValue("key") {
					appUser.Active = true
					if _, err := datastore.Put(c, key, appUser); err != nil {
						c.Errorf("error saving activated user %s: %v", r.FormValue("email"), err)
						return
					} else {
						t := taskqueue.NewPOSTTask("/cron/report", map[string][]string{
							"id":    []string{appUser.UserIdStr},
							"email": []string{r.FormValue("email")},
						})

						if _, err := taskqueue.Add(c, t, "reports"); err != nil {
							c.Errorf("error tasking user %s report: %v", appUser.UserIdStr, err)
							fmt.Fprintf(w, "There was a problem confirming your subscription. Please try again and contact us if this problem persists.")
						} else {
							fmt.Fprintf(w, "Your email subscription is now activated. You may close this page.")
						}
					}
				} else {
					c.Infof("authKey: %s\nsuppliedKey: %s", strings.Split(appUser.AuthKey, ";")[1], r.FormValue("key"))
					http.Error(w, "The supplied key did not match with our records.", http.StatusBadRequest)
				}
			}
		} else {
			dir := path.Join(os.Getenv("PWD"), "templates")
			admin := path.Join(dir, "signup.html")
			layout := path.Join(dir, "layout.html")
			cnfg := config.Load(c)
			data := map[string]interface{}{
				"captcha": cnfg["captchaPublic"],
			}
			page := mustache.RenderFileInLayout(admin, layout, data)
			fmt.Fprint(w, page)
		}
	} else if r.Method == "POST" {
		captcha := verifyCaptcha(c, map[string]string{
			"response": r.FormValue("g-recaptcha-response"),
			"remoteip": r.RemoteAddr,
		})
		key := datastore.NewKey(c, "AppUser", r.FormValue("email"), 0, nil)
		if r.FormValue("type") == "enterprise" && len(r.FormValue("email")) > 0 {
			appUser = &AppUser{
				Email:      r.FormValue("email"),
				Type:       "enterprise",
				Active:     true,
				Discovered: time.Now(),
			}
			if captcha {
				if _, err := datastore.Put(c, key, appUser); err != nil {
					data["success"] = false
					data["error"] = err.Error()
				} else {
					data["success"] = true
				}
			} else {
				data["success"] = false
				data["error"] = "Captcha challenge failed."
			}
		} else if r.FormValue("type") == "email-report" && len(r.FormValue("email")) > 0 {
			vineUser, err := vineApi.GetUser(r.FormValue("user"))
			if !captcha {
				data["success"] = false
				data["error"] = "Captcha challenge failed."
			} else if err != nil {
				data["success"] = false
				data["error"] = err.Error()
			} else if !UserQueueExist(vineUser.UserId, c) {
				data["success"] = false
				data["error"] = "That user doesn't appear to exist in Davine's database yet. Please submit it to us first."
			} else {
				slug := GenSlug()
				appUser = &AppUser{
					Email:      r.FormValue("email"),
					Type:       "email-report",
					Active:     false,
					UserIdStr:  r.FormValue("user"),
					AuthKey:    slug + ";" + GenKey(),
					Discovered: time.Now(),
				}
				if _, err := datastore.Put(c, key, appUser); err == nil {
					data["success"] = true
					data["code"] = slug
				} else {
					c.Errorf("got appUser store err: %v", err)
				}
			}
		} else if r.FormValue("type") == "email-ping" {
			err := datastore.Get(c, key, appUser)
			if err == nil {
				if u, err := vineApi.GetUser(appUser.UserIdStr); err != nil {
					data["success"] = false
					data["error"] = err.Error()
				} else {
					authKey := strings.Split(appUser.AuthKey, ";")
					if strings.Contains(u.Description, authKey[0]) {
						data["success"] = true
						emailData := map[string]interface{}{
							"username": u.Username,
							"id":       u.UserIdStr,
							"link":     fmt.Sprintf("https://%s/sign-up?type=activate&key=%s&email=%s", appengine.DefaultVersionHostname(c), authKey[1], r.FormValue("email")),
						}
						msg := email.New()
						msg.LoadTemplate(0, emailData)
						msg.To = []string{r.FormValue("email")}
						if err := msg.Send(c); err != nil {
							c.Errorf("error sending %s email to %s: %v", u.UserIdStr, msg.To[0], err)
							data["success"] = false
							data["error"] = err.Error()
						}
					} else {
						data["success"] = false
					}
				}
			} else {
				data["success"] = false
				data["error"] = err.Error()
			}
		}
		json.NewEncoder(w).Encode(data)
	}
}

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	db := DB{c}
	vineApi := VineRequest{c}
	adminUser := user.Current(c)
	if adminUser == nil {
		url, _ := user.LoginURL(c, "/admin/dashboard")
		http.Redirect(w, r, url, 301)
		return
	} else if !adminUser.Admin {
		w.WriteHeader(401)
		return
	}

	if r.Method == "GET" {
		dir := path.Join(os.Getenv("PWD"), "templates")
		admin := path.Join(dir, "admin.html")
		layout := path.Join(dir, "layout.html")
		enterprise, reports, _ := GetAppUsers(c)
		data := map[string]interface{}{
			"config":          config.Load(c),
			"enterpriseUsers": enterprise,
			"reportUsers":     reports,
		}
		page := mustache.RenderFileInLayout(admin, layout, data)
		fmt.Fprint(w, page)
	} else if r.Method == "POST" {
		switch r.FormValue("op") {
		case "TaskUsers":
			admin.NewTask(c).BatchTaskUsers(strings.Split(r.FormValue("v"), "\n")...)
		case "UnqueueUser":
			c.Infof("unqueuing %v", r.FormValue("v"))
			db.UnqueueUser(r.FormValue("v"))
		case "BatchUsers":
			users := strings.Split(r.FormValue("v"), ",")
			c.Infof("queueing users: %v", users)
			for _, v := range users {
				QueueUser(strings.TrimSpace(v), c)
			}
		case "FeaturedUser":
			if user, err := vineApi.GetUser(r.FormValue("user")); err == nil {
				key := datastore.NewKey(c, "Misc", "featuredUser", 0, nil)
				featuredUser := &FeaturedUser{user.UserIdStr, r.FormValue("vine")}
				if _, err := datastore.Put(c, key, featuredUser); err != nil {
					http.Error(w, err.Error(), 500)
					return
				} else {
					items := []*memcache.Item{&memcache.Item{
						Key:   "featuredUser",
						Value: []byte(user.UserIdStr + ";" + r.FormValue("vine")),
					}}
					var d bytes.Buffer
					enc := gob.NewEncoder(&d)
					user.ProfileBackground = strings.TrimPrefix(user.ProfileBackground, "0x")
					enc.Encode(user)
					items = append(items, &memcache.Item{
						Key:   user.UserIdStr,
						Value: d.Bytes(),
					})
					memcache.AddMulti(c, items)
				}
			} else {
				http.Error(w, err.Error(), 500)
				return
			}
		case "DumpKind":
			admin.NewTask(c).DumpData(r.FormValue("v"), w)
			return
		case "LoadData":
			file, _, err := r.FormFile("file")
			if err != nil {
				c.Errorf("error loading file: %v", err)
				return
			}
			if err := admin.NewTask(c).LoadData(r.FormValue("v"), file); err != nil {
				c.Errorf("Error loading data: %v", err)
			}
		}
		fmt.Fprintf(w, "{\"op\":\"%v\",\"success\":true}", r.FormValue("op"))
	}
}
