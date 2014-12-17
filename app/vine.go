package main

import (
    "appengine"
    "appengine/urlfetch"
    "io"
    "errors"
)

type VineRequest struct {
	AESession appengine.Context
}

const (
    VINE_API = "https://api.vineapp.com"
)

func (v *VineRequest) get(url string) (io.ReadCloser, error) {
	if v.AESession == nil {
		return nil, errors.New("Google AppEngine Context Required")
	} else {
		c := v.AESession
		client := urlfetch.Client(c)
		resp, err := client.Get(VINE_API + url)
		if err == nil {
			return resp.Body, nil
		} else {
			return nil, err
		}
	}
}