package main

import (
    "appengine"
    "appengine/urlfetch"
    "net/http"
    "io/ioutil"
    "errors"
    "encoding/json"
    "regexp"
)

type VineRequest struct {
	Context appengine.Context
}

const (
    VINE_API = "https://api.vineapp.com"
)

func (v *VineRequest) get(url string) (map[string]interface{}, error) {
	if v.Context == nil {
		return nil, errors.New("Google AppEngine Context Required")
	} else {
		c := v.Context
		client := urlfetch.Client(c)
		req, _ := http.NewRequest("GET", VINE_API + url, nil)
		req.Header.Set("x-vine-client", "vinewww/1.0")
		resp, err := client.Do(req)
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