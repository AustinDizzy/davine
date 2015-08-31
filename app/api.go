package main

import (
    "app/config"
    "app/counter"
    "appengine"
    "appengine/urlfetch"
    "io/ioutil"
    "net/http"
    "net/url"
    "encoding/json"
    "strings"
)

func ApiRouter(w http.ResponseWriter, r *http.Request) {
    var (
        c = appengine.NewContext(r)
        path = strings.Split(strings.TrimPrefix(r.URL.Path, "/api/"), "/")
        data = make(map[string]interface{})
        err error
    )

    r.ParseForm()
    switch(path[0]) {
        case "statistics":
            data["loops"], err = counter.Count(c, "TotalLoops")
            data["posts"], err = counter.Count(c, "TotalPosts")
            if err != nil {
                data["error"] = err.Error()
            } else {
                data["success"] = true
            }
    }
    json.NewEncoder(w).Encode(data)
}

func verifyCaptcha(c appengine.Context, vals map[string]string) bool {
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
		c.Errorf("got err: %v", err)
		return false
	}
	body, _ := ioutil.ReadAll(resp.Body)
	var data map[string]interface{}
	json.Unmarshal(body, &data)
	return data["success"].(bool)
}