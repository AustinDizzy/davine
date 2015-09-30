package routes

import (
	"net/http"
	"time"

	"app/data"

	"github.com/austindizzy/vine-go"
	"github.com/qedus/nds"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type popUsers struct {
	Users []int64
}

//PopularFetchHandler is the http request handler for /cron/popular.
//It explores the Vine popular feed(s), queues any user that we are not
//already tracking, and then adds the user to the front page's "popular users"
//feed utilizing github.com/qedus/nds for memcache persistence.
func PopularFetchHandler(w http.ResponseWriter, r *http.Request) {
	var (
		c       = appengine.NewContext(r)
		vineAPI = vine.NewRequest(urlfetch.Client(c))
		db      = data.NewRequest(c)
		start   = time.Now()
		popFeed = &popUsers{}
		err     error
	)

	users, err := vineAPI.GetPopularUsers(60)

	for _, u := range users {
		if !db.UserQueueExist(u.UserId) {
			db.QueueUser(u.UserIdStr)
		}
		popFeed.Users = append(popFeed.Users, u.UserId)
	}

	key := datastore.NewKey(c, "_popusers_", "popusers", 0, nil)
	_, err = nds.Put(c, key, popFeed)
	if err != nil {
		log.Errorf(c, "error storing popular users: %v", err)
	}

	finish := time.Since(start)
	if err == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	log.Infof(c, "queueing %d popular users: %v w/ err %v. Took %s", len(popFeed.Users), popFeed.Users, err, finish)
}
