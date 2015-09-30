package routes

import (
	"fmt"
	"net/http"
	"regexp"

	"app/data"

	"github.com/austindizzy/vine-go"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
)

//CronExploreHandler is the http request handler for /cron/explore.
//When called from a cron job, it tasks all available Vine feeds to be
//scraped for user IDs. When called from a task, it scrapes the feed
//utilizing regular expressions for user IDs to store.
func CronExploreHandler(w http.ResponseWriter, r *http.Request) {
	var (
		c  = appengine.NewContext(r)
		db = data.NewRequest(c)
	)

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
			log.Errorf(c, "error tasking explore: %v", err)
		}
	} else if r.Method == "POST" {
		userIDs, err := scrapeUserIDs(c, r.FormValue("feed"))
		if err != nil {
			log.Errorf(c, "Error scraping %s: %v", r.FormValue("feed"), err)
		} else {
			log.Infof(c, "%d users", len(userIDs))
			for _, u := range userIDs {
				if _, err := db.GetQueuedUser(u); err == datastore.ErrNoSuchEntity {
					db.QueueUser(u)
				}
			}
		}
	}
}

func scrapeUserIDs(c context.Context, feed string) ([]string, error) {
	v := vine.NewRequest(urlfetch.Client(c))
	resp, err := v.Get(feed)

	if err != nil {
		return nil, err
	}

	users := []string{}
	regex := regexp.MustCompile(`(?:\"userId\"\: )([0-9]*)(?:,)`)
	for _, u := range regex.FindAllStringSubmatch(string(resp), -1) {
		users = append(users, u[1])
	}
	return removeDuplicates(users), nil
}

func removeDuplicates(a []string) []string {
	found := make(map[interface{}]bool)
	j := 0
	for i, x := range a {
		if !found[x] {
			found[x] = true
			a[j] = a[i]
			j++
		}
	}
	return a[:j]
}
