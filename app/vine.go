package main

import (
	"appengine"
	"appengine/urlfetch"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
)

type VineRequest struct {
	Context appengine.Context
}

var (
	ErrUserDoesntExist = errors.New("That record does not exist.")
)

const (
	VINE_API = "https://api.vineapp.com"
)

func (v *VineRequest) get(url string) ([]byte, error) {
	if v.Context == nil {
		return nil, errors.New("Google AppEngine Context Required")
	} else {
		c := v.Context
		client := urlfetch.Client(c)
		req, _ := http.NewRequest("GET", VINE_API+url, nil)
		req.Header.Set("x-vine-client", "vinewww/1.0")
		resp, err := client.Do(req)
		if err == nil {
			return ioutil.ReadAll(resp.Body)
		} else {
			return nil, err
		}
	}
}

func (v *VineRequest) GetUser(userId string) (*VineUser, error) {
	url := "/users/profiles/"

	if v.IsVanity(userId) {
		url += userId
	} else {
		url += "vanity/" + userId
	}

	resp, err := v.get(url)
	if err != nil {
		return nil, err
	} else {
		data := new(VineUserWrapper)
		err = json.Unmarshal(resp, &data)
		if data.Success {
			return data.Data, nil
		} else {
			return nil, errors.New(data.Error)
		}
	}
}

func (v *VineRequest) GetPopularUsers(users chan string, length int) error {
	resp, err := v.get("/timelines/popular?size=" + strconv.Itoa(length))
	if err != nil {
		return err
	} else {
		data := new(VinePopularWrapper)
		err = json.Unmarshal(resp, &data)
		if data.Success {
			for _, v := range data.Data.Records {
				users <- v.UserIdStr
			}
			close(users)
			return nil
		} else {
			return errors.New(data.Error)
		}
	}
}

func (v *VineRequest) IsVanity(user string) bool {
    match, _ := regexp.MatchString("^[0-9]+$", user)
    return !match
}