package main

import (
	"appengine"
	"appengine/urlfetch"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
)

type VineRequest struct {
	Context appengine.Context
}

var (
	ErrUserDoesntExist = errors.New("That record doesn't exist.")
)

const (
	VINE_API = "https://api.vineapp.com"
)

func (v *VineRequest) get(url string) (*VineUser, error) {
	if v.Context == nil {
		return nil, errors.New("Google AppEngine Context Required")
	} else {
		c := v.Context
		client := urlfetch.Client(c)
		req, _ := http.NewRequest("GET", VINE_API+url, nil)
		req.Header.Set("x-vine-client", "vinewww/1.0")
		resp, err := client.Do(req)
		if err == nil {
			jsonData, _ := ioutil.ReadAll(resp.Body)
			data := new(VineUserWrapper)
			err = json.Unmarshal(jsonData, &data)
			if data.Success {
				return data.Data, nil
			} else {
				return nil, errors.New(data.Error)
			}
		} else {
			return nil, err
		}
	}
}

func (v *VineRequest) GetUser(userId string) (*VineUser, error) {
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
