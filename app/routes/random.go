package routes

import (
	"net/http"

	"app/data"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

//RandomHandler is the http request handler for /random/user.
//It picks a random user record and redirects the http request
//to the randomly selected user's profile page.
func RandomHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: fix randomHandler in production.
	//Can't iterate over all 100k+ keys in less than 10 seconds, so request
	//is cut off by appengine's 10s limit and crashes
	var (
		c    = appengine.NewContext(r)
		q    = datastore.NewQuery("UserRecord").KeysOnly()
		user data.UserRecord
		err  error
	)
	keys, err := q.GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "got err %v", err)
		return
	}
	randomKey := data.RandomKey(keys)
	err = datastore.Get(c, randomKey, &user)
	if err != nil {
		log.Errorf(c, "got err %v", err)
	} else {
		log.Infof(c, "got user %v", user)
	}
	http.Redirect(w, r, "/u/"+user.UserId, 301)
}
