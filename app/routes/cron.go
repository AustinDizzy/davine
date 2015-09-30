package routes

import (
	"net/http"

	"app/admin"
	"app/counter"
	"app/utils"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

//CronFlushHandler is the http request handler for /cron/flush.
//It flushes all "24hr" stats at the end of each day, or as
//configured in cron.yaml and logs the previous values for analysis
//and metrics purposes.
func CronFlushHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	for _, k := range []string{"24hLoops", "24hPosts", "24hUsers"} {
		if n, err := counter.Count(c, k); err != nil {
			log.Errorf(c, "got err sending stat %s: %v", k, n)
		} else {
			utils.PostCount(c, k, int(n))
		}
		if err := counter.Delete(c, k); err != nil {
			log.Errorf(c, "got err flushing %s: %v", k, err)
		}
	}
}

//CronImportHandler is the http request handler for /cron/import.
//Utilizing tasks, it imports raw datastore records and is only
//for admin testing purposes.
func CronImportHandler(w http.ResponseWriter, r *http.Request) {
	var (
		c = appengine.NewContext(r)
		t = admin.NewTask(c)
	)
	t.LoadCtx(r)

	log.Infof(c, "starting import of %s", r.FormValue("file"))
	err := t.LoadGSData(r.FormValue("file"))
	if err != nil {
		log.Errorf(c, "error importing %s: %v", r.FormValue("file"), err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

//CronPurgeHandler is the http request handler for /cron/purge.
//Using tasks, it purges all datastore records of a specified kind.
//It is only used for admin testing purposes.
func CronPurgeHandler(w http.ResponseWriter, r *http.Request) {
	var (
		c = appengine.NewContext(r)
		q = datastore.NewQuery(r.FormValue("v")).KeysOnly()
	)
	log.Infof(c, "purging %s", r.FormValue("v"))
	t := q.Run(c)
	n, g := 0, false
	var a []*datastore.Key
	for !g {
		k, err := t.Next(nil)
		if len(a) < 500 && k != nil {
			a = append(a, k)
		}

		if len(a) == 500 || err == datastore.Done {
			if err == datastore.Done {
				g = true
			}
			err := datastore.DeleteMulti(c, a)
			if err != nil {
				log.Errorf(c, "Error deleting %d keys from %s: %v", len(a), r.FormValue("v"), err)
			} else {
				log.Infof(c, "%d deleted", len(a))
				n += len(a)
				a = nil
			}
		}
	}
	log.Infof(c, "%d %s entities deleted successfully.", n, r.FormValue("v"))
}
