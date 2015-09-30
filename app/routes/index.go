package routes

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"app/config"
	"app/counter"
	"app/data"
	"app/page"

	"github.com/qedus/nds"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

//IndexHandler is the http request handler for "/"
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	var (
		c        = appengine.NewContext(r)
		cnfg     = config.Load(c)
		p        = page.New("index.html")
		pageData = map[string]interface{}{
			"captcha": cnfg["captchaPublic"],
		}
		db           = data.NewRequest(c)
		err          error
		popFeed      popUsers
		featuredUser struct {
			UserID, PostID string
		}
	)

	for _, k := range []string{"TotalLoops", "TotalPosts", "TotalVerified", "24hLoops", "24hPosts", "24hUsers"} {
		if pageData[k], err = counter.Count(c, k); err != nil {
			log.Errorf(c, "Error getting %s: %v", k, err)
		}
	}

	popKey := datastore.NewKey(c, "_popusers_", "popusers", 0, nil)
	if err := nds.Get(c, popKey, &popFeed); err == nil {
		var users []*data.UserRecord
		for len(users) < 6 {
			rand.Seed(time.Now().Unix())
			r := rand.Intn(len(popFeed.Users))
			userRecord, err := db.GetUserRecord(popFeed.Users[r])
			if err != nil {
				log.Errorf(c, "error getting %s user record: %v", popFeed.Users[r], err)
			} else {
				users = append(users, userRecord)
				popFeed.Users = append(popFeed.Users[:r], popFeed.Users[r+1:]...)
			}
		}
		pageData["popusers"] = users
	} else {
		log.Errorf(c, "popusers memcache err: %v", err)
	}

	featuredKey := datastore.NewKey(c, "_featuredUser_", "featuredUser", 0, nil)
	if err := nds.Get(c, featuredKey, &featuredUser); err == nil {
		if s, err := strconv.ParseInt(featuredUser.UserID, 10, 64); err == nil {
			pageData["featuredUser"], err = db.GetUserRecord(s)
			if err != nil {
				log.Errorf(c, "error loading featured user %s: %v", featuredUser.UserID, err)
			}
		}
		pageData["featuredPost"] = featuredUser.PostID
	}

	p.LoadData(pageData)
	p.Write(w)
}
