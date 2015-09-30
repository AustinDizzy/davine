package routes

import (
	"net/http"

	"app/config"
	"app/counter"
	"app/data"
	"app/page"
	"app/utils"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

//AboutHandler is the http request handler for /about.
func AboutHandler(w http.ResponseWriter, r *http.Request) {
	var (
		p        = page.New("about.html")
		pageData = map[string]interface{}{
			"title": "About",
		}
		c   = appengine.NewContext(r)
		err error
	)
	for _, k := range []string{"TotalUsers", "TotalLoops", "TotalPosts"} {
		pageData[k], err = counter.Count(c, k)
		if err != nil {
			log.Errorf(c, "got counter err: %v", err)
		}
	}
	p.LoadData(pageData)
	p.Write(w)
}

//TopHandler is the http request handler for /top.
func TopHandler(w http.ResponseWriter, r *http.Request) {
	var (
		c        = appengine.NewContext(r)
		db       = data.NewRequest(c)
		cnfg     = config.Load(c)
		p        = page.New("top.html")
		pageData = db.GetTop()
	)

	pageData["title"] = "Top"
	pageData["captcha"] = cnfg["captchaPublic"]
	p.LoadData(pageData)
	p.Write(w)
}

//DonateHandler is the http request handler for /donate.
func DonateHandler(w http.ResponseWriter, r *http.Request) {
	page.New("donate.html").Write(w)
}

//NotFoundHandler is the http request handler for /404.
//It is the default "404" page and is rendered when a request route is
//not specified.
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	var (
		p        = page.New("404.html")
		pageData = map[string]string{"url": r.RequestURI}
	)
	w.WriteHeader(404)
	p.LoadData(pageData)
	p.Write(w)
}

//StartupHandler is the http request handler for /_ah/warmup.
//It is automatically called by AppEngine when a new instance start up.
func StartupHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	config.Load(c)
	utils.PostCount(c, "new instance", 1)
}
