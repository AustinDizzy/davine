package main

import (
    "appengine"
    "appengine/urlfetch"
    "io/ioutil"
    "errors"
    "encoding/json"
    "regexp"
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

func (v *VineRequest) GetUser(userId string) (map[string]interface{}, error) {
    url := "/users/profiles/"
    match, _ := regexp.MatchString("[0-9]+", userId)
    
    if match {
        url += userId   
    } else {
        url += "vanity/" + userId   
    }
    
    data, err := v.get(url)
    if err != nil {
        return nil, err   
    } else {
        return data, nil  
    }
}