package main

import (
    "appengine"
    "appengine/urlfetch"
    "io/ioutil"
    "errors"
    "encoding/json"
)

type VineRequest struct {
	AESession appengine.Context
}

const (
    VINE_API = "https://api.vineapp.com"
)

func (v *VineRequest) get(url string) (map[string]interface{}, error) {
	if v.AESession == nil {
		return nil, errors.New("Google AppEngine Context Required")
	} else {
		c := v.AESession
		client := urlfetch.Client(c)
		resp, err := client.Get(VINE_API + url)
		if err == nil {
			jsonData, _ := ioutil.ReadAll(resp.Body)
			var data interface{}
			err = json.Unmarshal(jsonData, &data)
			d := data.(map[string]interface{})
			return d["data"].(map[string]interface{}), nil
		} else {
			return nil, err
		}
	}
}