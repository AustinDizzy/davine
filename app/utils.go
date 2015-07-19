package main

import (
	"app/config"
	"appengine"
	"appengine/urlfetch"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/hoisie/mustache"
	"github.com/stathat/go"
	"io/ioutil"
	"net/url"
	"os"
	"path"
)

func genRand(dict string, n int) string {
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dict[v%byte(len(dict))]
	}

	return string(bytes)
}

func GenKey() string {
	dict := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	dict += "abcdefghijklmnopqrstuvwxyz"
	dict += "1234567890=+~-"
	return genRand(dict, 64)
}

func GenSlug() string {
	dict := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	dict += "1234567890"
	dict += "abcdefghijklmnopqrstuvwxyz"
	return genRand(dict, 6)
}

func PostValue(c appengine.Context, key string, value float64) {
	rt := urlfetch.Client(c).Transport
	cnfg := config.Load(c)
	stathat.DefaultReporter = stathat.NewReporter(100000, 10, rt)

	if appengine.IsDevAppServer() {
		key = key + " [DEV]"
	}
	if err := stathat.PostEZValue(key, cnfg["stathatKey"], value); err != nil {
		c.Errorf("Error posting %v value %v to stathat: %v", key, value, err)
	}
}

func PostCount(c appengine.Context, key string, count int) {
	rt := urlfetch.Client(c).Transport
	cnfg := config.Load(c)
	stathat.DefaultReporter = stathat.NewReporter(100000, 10, rt)

	if appengine.IsDevAppServer() {
		key = key + " [DEV]"
	}
	if err := stathat.PostEZCount(key, cnfg["stathatKey"], count); err != nil {
		c.Errorf("Error posting %v value %v to stathat: %v", key, count, err)
	}
}

func GenSummaryChart(c appengine.Context, user *UserRecord) (string, error) {
	dir := path.Join(os.Getenv("PWD"), "templates")
	template := path.Join(dir, "weeklyreport.chart")
	var loops, followers, dates string

	for i := 1; (len(user.UserData)-i-1 > -1) && (i <= 7); i++ {
		if i > 1 {
			loops += ","
			followers += ","
			dates += ","
		}
		u := user.UserData[len(user.UserData)-i]
		v := user.UserData[len(user.UserData)-i-1]
		loops += fmt.Sprintf("%d", u.Loops-v.Loops)
		followers += fmt.Sprintf("%d", u.Followers-v.Followers)
		dates += fmt.Sprintf("\"%d/%d\"", u.Recorded.Month(), u.Recorded.Day())
	}

	data := map[string]string{
		"loops":     loops,
		"followers": followers,
		"dates":     dates,
	}

	c.Infof("opts: %#v", data)

	opts := &url.Values{}
	opts.Add("options", mustache.RenderFile(template, data))
	opts.Add("width", "500")
	opts.Add("scale", "0.5")

	client := urlfetch.Client(c)
	resp, err := client.Get(fmt.Sprintf("http://export.highcharts.com/?%s", opts.Encode()))
	b, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		c.Infof("got highcharts error: %v", string(b[:]))
	}

	return base64.StdEncoding.EncodeToString(b), err
}
