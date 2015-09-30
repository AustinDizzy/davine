package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"app/admin"
	"app/data"
	"app/email"

	"github.com/dustin/go-humanize"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/user"
)

//CronReportHandler is the http request handler for /cron/report.
//When called from a task, it sends a specific user's email report to
//the specified email address, then retasks the report for 7 days from now.
func CronReportHandler(w http.ResponseWriter, r *http.Request) {
	var (
		id, _   = strconv.ParseInt(r.FormValue("id"), 10, 64)
		c       = appengine.NewContext(r)
		msg     = email.New()
		db      = data.NewRequest(c)
		u, err  = db.GetUser(id)
		appUser = admin.AppUser{}
		key     = datastore.NewKey(c, "AppUser", r.FormValue("email"), 0, nil)
	)

	log.Infof(c, "generating email for %s at %s", r.FormValue("id"), r.FormValue("email"))

	if err != nil {
		log.Errorf(c, "error sending user report for %s: %v", r.FormValue("id"), err)
		return
	}

	err = datastore.Get(c, key, &appUser)
	if err != nil && !(err == datastore.ErrNoSuchEntity && user.IsAdmin(c)) {
		log.Errorf(c, "error reading appUser for %s: %v", r.FormValue("email"), err)
		return
	}

	if len(u.UserData) < 7 {
		log.Infof(c, "user %s has only %d data entries", u.Username, len(u.UserData))
		t := taskqueue.NewPOSTTask("/cron/report", map[string][]string{
			"id":    []string{r.FormValue("id")},
			"email": []string{r.FormValue("email")},
		})

		t.Delay = 7 * 24 * time.Hour

		tk, err := taskqueue.Add(c, t, "reports")
		if err != nil {
			log.Infof(c, "error queuing email report for %s: %v", r.FormValue("email"), err)
		} else if user.IsAdmin(c) {
			log.Infof(c, "task created: %s", tk.Name)
		}
		return
	}

	mon := []string{"Jan", "Feb", "Mar", "Apr", "May", "June", "Jul", "Aug", "Sept", "Oct", "Nov", "Dec"}
	d := u.UserData[len(u.UserData)-7:]

	chart, err := db.GenSummaryChart(u)
	reportData := map[string]interface{}{
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
		reportData["key"] = strings.Split(appUser.AuthKey, ";")[1]
	}

	for _, v := range []string{"newPosts", "newLoops", "newFollowers"} {
		if !strings.Contains(reportData[v].(string), "-") {
			reportData[v] = "+" + reportData[v].(string)
		}
	}

	msg.LoadTemplate(1, reportData)
	msg.To = []string{r.FormValue("email")}
	msg.Send(c)

	t := taskqueue.NewPOSTTask("/cron/report", map[string][]string{
		"id":    []string{r.FormValue("id")},
		"email": []string{r.FormValue("email")},
	})

	t.Delay = 7 * 24 * time.Hour

	if _, err := taskqueue.Add(c, t, "reports"); err != nil {
		log.Infof(c, "error queuing email report for %s: %v", r.FormValue("email"), err)
	}
}
