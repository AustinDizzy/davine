package admin

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"app/config"
	"app/data"
	"app/page"
	"app/utils"

	"github.com/austindizzy/vine-go"
	"github.com/qedus/nds"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
	"google.golang.org/appengine/user"
)

//Handler is the request handler for /admin/dashboard.
func Handler(w http.ResponseWriter, r *http.Request) {
	var (
		c       = appengine.NewContext(r)
		db      = data.NewRequest(c)
		vineAPI = vine.NewRequest(urlfetch.Client(c))
	)
	if user.Current(c) == nil {
		url, _ := user.LoginURL(c, "/admin/dashboard")
		http.Redirect(w, r, url, 301)
		return
	} else if !user.IsAdmin(c) {
		w.WriteHeader(401)
		return
	}

	if r.Method == "GET" {
		enterprise, reports, _ := getAppUsers(c)
		var (
			p    = page.New("admin.html")
			data = map[string]interface{}{
				"title":           "Admin Dashboard",
				"config":          config.Load(c),
				"enterpriseUsers": enterprise,
				"reportUsers":     reports,
			}
		)
		p.LoadData(data)
		p.Write(w)
	} else if r.Method == "POST" {
		switch r.FormValue("op") {
		case "TaskUsers":
			t := NewTask(c)
			t.LoadCtx(r)
			t.BatchTaskUsers(strings.Split(r.FormValue("v"), "\n")...)
		case "TaskImport":
			log.Infof(c, "Tasking %s for import", r.FormValue("v"))
			var tasks []*taskqueue.Task
			for k, v := range strings.Split(r.FormValue("v"), "\n") {
				t := taskqueue.NewPOSTTask("/cron/import", map[string][]string{
					"file": {strings.TrimSpace(v)},
				})
				t.Name = fmt.Sprintf("%s-%s", strings.Split(v, ".")[0], utils.GenSlug())
				t.Delay = time.Minute * time.Duration(k) * 4
				t.RetryOptions = &taskqueue.RetryOptions{RetryLimit: 0, AgeLimit: t.Delay + (2 * time.Second)}
				tasks = append(tasks, t)
			}
			if _, err := taskqueue.AddMulti(c, tasks, ""); err != nil {
				log.Errorf(c, "error adding import tasks: %v", err)
			}
		case "UnqueueUser":
			log.Infof(c, "unqueuing %v", r.FormValue("v"))
			db.UnqueueUser(r.FormValue("v"))
		case "BatchUsers":
			users := strings.Split(r.FormValue("v"), ",")
			log.Infof(c, "queueing users: %v", users)
			for _, v := range users {
				db.QueueUser(strings.TrimSpace(v))
			}
		case "FeaturedUser":
			if user, err := vineAPI.GetUser(r.FormValue("user")); err == nil {
				key := datastore.NewKey(c, "_featuredUser_", "featuredUser", 0, nil)
				featuredUser := &featuredUser{UserID: user.UserIdStr, PostID: r.FormValue("vine")}
				if _, err := nds.Put(c, key, featuredUser); err != nil {
					log.Errorf(c, "error setting featured user: %v", err)
					http.Error(w, err.Error(), 500)
					return
				}
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "DumpKind":
			t := NewTask(c)
			t.LoadCtx(r)
			t.DumpData(r.FormValue("v"), w)
			return
		case "PurgeData":
			t := taskqueue.NewPOSTTask("/cron/purge", map[string][]string{
				"v": {strings.TrimSpace(r.FormValue("v"))},
			})
			t.Delay = 45 * time.Second
			t.Name = "purge-" + r.FormValue("v") + "-" + utils.GenSlug()
			if _, err := taskqueue.Add(c, t, ""); err != nil {
				log.Errorf(c, "error adding purge task: %v", err)
			}
		case "LoadData":
			file, _, err := r.FormFile("file")
			if err != nil {
				log.Errorf(c, "error loading file: %v", err)
				return
			}
			t := NewTask(c)
			t.LoadCtx(r)
			if err := t.LoadData(r.FormValue("v"), file); err != nil {
				log.Errorf(c, "Error loading data: %v", err)
			}
		}
		fmt.Fprintf(w, "{\"op\":\"%v\",\"success\":true}", r.FormValue("op"))
	}
}
