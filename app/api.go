package main

import (
	"app/config"
	"app/counter"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"google.golang.org/appengine"
)

func ApiRouter(w http.ResponseWriter, r *http.Request) {
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

func verifyCaptcha(c context.Context, vals map[string]string) bool {
	cnfg := config.Load(c)
	client := urlfetch.Client(c)
	uri, _ := url.Parse("https://www.google.com/recaptcha/api/siteverify")
	q := url.Values{}
	q.Add("secret", cnfg["captchaPrivate"])
	for k := range vals {
		q.Add(k, vals[k])
	}
	uri.RawQuery = q.Encode()
	req, _ := http.NewRequest("GET", uri.String(), nil)
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf(c, "got err: %v", err)
		return false
	}
	body, _ := ioutil.ReadAll(resp.Body)
	var data map[string]interface{}
	json.Unmarshal(body, &data)
	return data["success"].(bool)
}
