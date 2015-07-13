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
	ErrUserDoesntExist = "That record does not exist."
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
		url += "vanity/" + userId
	} else {
		url += userId
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

func (v *VineRequest) GetPopularUsers(num int) ([]*PopularRecord, error) {
	resp, err := v.get("/timelines/popular?size=" + strconv.Itoa(num))
	if err != nil {
		return nil, err
	} else {
		data := new(VinePopularWrapper)
		err = json.Unmarshal(resp, &data)
		if data.Success {
			return data.Data.Records, err
		} else {
			return data.Data.Records, errors.New(data.Error)
		}
	}
}

func (v *VineRequest) IsVanity(user string) bool {
	match, _ := regexp.MatchString("^[0-9]+$", user)
	return !match
}

func (v *VineRequest) ScrapeUserIDs(feed string) ([]string, error) {
	v.Context.Infof("scraping %s", feed)
	resp, err := v.get(feed)
	if err != nil {
		return nil, err
	} else {
		users := []string{}
		regex := regexp.MustCompile(`(?:\"userId\"\: )([0-9]*)(?:,)`)
		for _, u := range regex.FindAllStringSubmatch(string(resp), -1) {
			users = append(users, u[1])
		}
		return RemoveDuplicates(users), nil
	}
}
