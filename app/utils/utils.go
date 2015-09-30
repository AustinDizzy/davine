package utils

import (
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"app/config"

	"github.com/stathat/go"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

func genRand(dict string, n int) string {
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dict[v%byte(len(dict))]
	}

	return string(bytes)
}

//GenKey returns a randomly generated 64 character string, using only
//alphanumeric characters and a small selection of special characters.
func GenKey() string {
	dict := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	dict += "abcdefghijklmnopqrstuvwxyz"
	dict += "1234567890=+~-"
	return genRand(dict, 64)
}

//GenSlug returns a randomly generated 6 character alphanumeric string.
func GenSlug() string {
	dict := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	dict += "1234567890"
	dict += "abcdefghijklmnopqrstuvwxyz"
	return genRand(dict, 6)
}

//PostValue sends a value stat to the configured StatHat account.
func PostValue(c context.Context, key string, value float64) {
	rt := urlfetch.Client(c).Transport
	cnfg := config.Load(c)
	stathat.DefaultReporter = stathat.NewReporter(100000, 10, rt)

	if appengine.IsDevAppServer() {
		key = key + " [DEV]"
	}
	if err := stathat.PostEZValue(key, cnfg["stathatKey"], value); err != nil {
		log.Errorf(c, "Error posting %v value %v to stathat: %v", key, value, err)
	}
}

//PostCount sends a count stat to the configured StatHat account.
func PostCount(c context.Context, key string, count int) {
	rt := urlfetch.Client(c).Transport
	cnfg := config.Load(c)
	stathat.DefaultReporter = stathat.NewReporter(100000, 10, rt)

	if appengine.IsDevAppServer() {
		key = key + " [DEV]"
	}
	if err := stathat.PostEZCount(key, cnfg["stathatKey"], count); err != nil {
		log.Errorf(c, "Error posting %v value %v to stathat: %v", key, count, err)
	}
}

//VerifyCaptcha verifies the supplied values are a valid solved captcha.
//Read the required parameters in the reCAPTCHA documentation, found
//here: https://developers.google.com/recaptcha/docs/verify
func VerifyCaptcha(c context.Context, vals map[string]string) bool {
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
