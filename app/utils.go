package main

import (
	"app/config"
	"appengine"
	"appengine/urlfetch"
	"crypto/rand"
	"github.com/stathat/go"
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
