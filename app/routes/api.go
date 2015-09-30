package routes

import (
	"encoding/json"
	"net/http"
	"strings"

	"app/counter"

	"google.golang.org/appengine"
)

//APIRouter is the http request router for all "/api" requests
func APIRouter(w http.ResponseWriter, r *http.Request) {
	var (
		c    = appengine.NewContext(r)
		path = strings.Split(strings.TrimPrefix(r.URL.Path, "/api/"), "/")
		data = make(map[string]interface{})
		err  error
	)

	r.ParseForm()
	switch path[0] {
	case "statistics":
		data["totalLoops"], err = counter.Count(c, "TotalLoops")
		data["totalPosts"], err = counter.Count(c, "TotalPosts")
		data["totalUsers"], err = counter.Count(c, "TotalUsers")
		data["totalVerified"], err = counter.Count(c, "TotalVerified")
		data["totalExpliicit"], err = counter.Count(c, "TotalExplicit")
		data["24hLoops"], err = counter.Count(c, "24hLoops")
		data["24hPosts"], err = counter.Count(c, "24hPosts")
		data["24hUsers"], err = counter.Count(c, "24hUsers")
		if err != nil {
			data["error"] = err.Error()
		} else {
			data["success"] = true
		}
	}
	json.NewEncoder(w).Encode(data)
}
