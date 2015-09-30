package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"app/data"
	"app/page"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

//SearchHandler is the http request handler for /search.
//On POST, it searches across the user database for the specified query.
//On GET, it renders the search page to initiate a query.
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	var (
		c        = appengine.NewContext(r)
		p        = page.New("search.html")
		db       = data.NewRequest(c)
		pageData = map[string]interface{}{
			"query": r.FormValue("q"),
			"count": 0,
			"title": "Search for \"" + r.FormValue("q") + "\"",
		}
	)
	if len(r.FormValue("q")) > 0 {
		results, err := db.SearchUsers(r.FormValue("q"))
		if err != nil {
			log.Errorf(c, "got err on search: %v", err)
		}

		switch r.FormValue("s") {
		case "overall":
			sort.Sort(data.ByOverall(results))
			break
		case "followers":
			sort.Sort(data.ByFollowers(results))
			break
		case "loops":
			sort.Sort(data.ByLoops(results))
			break
		case "posts":
			sort.Sort(data.ByPosts(results))
			break
		case "revines":
			sort.Sort(data.ByRevines(results))
			break
		}

		if r.Method == "GET" {
			pageData["count"] = len(results)
			pageData["results"] = results
		} else if r.Method == "POST" {
			jsonData, _ := json.Marshal(results)
			fmt.Fprint(w, string(jsonData))
			return
		}
	}

	p.LoadData(pageData)
	p.Write(w)
}
