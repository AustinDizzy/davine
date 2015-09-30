package routes

import (
	"net/http"

	"app/counter"
	"app/data"
	"app/page"

	"google.golang.org/appengine"
)

//DiscoverHandler is the http request handler for /discover.
func DiscoverHandler(w http.ResponseWriter, r *http.Request) {
	var (
		c    = appengine.NewContext(r)
		db   = data.NewRequest(c)
		data = map[string]interface{}{
			"title": "Discover",
		}
		p = page.New("discover.html")
	)
	data["totalUsers"], _ = counter.Count(c, "TotalUsers")
	data["24hUsers"], _ = counter.Count(c, "24hUsers")
	data["totalVerified"], _ = counter.Count(c, "TotalVerified")
	data["totalExplicit"], _ = counter.Count(c, "TotalExplicit")
	data["recentUsers"], _ = db.GetRecentUsers(12)
	data["recentExplicit"], _ = db.GetRecentUsers(4, "Explicit =", true)
	data["recentVerified"], _ = db.GetRecentUsers(4, "Verified =", true)

	p.LoadData(data)
	p.Write(w)
}
