package config

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
	"gopkg.in/yaml.v2"

	"appengine"
	"appengine/file"
	"appengine/urlfetch"
)

var Config ConfigData

type (
	ConfigData       map[string]string
	appengineContext struct{}
)

func Load(c ...appengine.Context) ConfigData {
	if Config != nil {
		return Config
	}
	var configFile []byte
	if appengine.IsDevAppServer() {
		configFile, _ = ioutil.ReadFile(path.Join(os.Getenv("PWD"), "config.yaml"))
	} else {
		var context context.Context
		context = getContext(context, c[0])
		client := &http.Client{
			Transport: &oauth2.Transport{
				Source: google.AppEngineTokenSource(context, storage.ScopeReadOnly),
				Base: &urlfetch.Transport{
					Context: c[0],
				},
			},
		}
		bucket, _ := file.DefaultBucketName(c[0])
		ctx := cloud.NewContext("davine-web", client)
		rc, err := storage.NewReader(ctx, bucket, "config.yaml")
		if err != nil {
			c[0].Errorf("error reading config: %v", err.Error())
		}
		configFile, err = ioutil.ReadAll(rc)
		rc.Close()
		if err != nil {
			c[0].Errorf("error reading config: %v", err.Error())
		}
	}
	yaml.Unmarshal(configFile, &Config)
	return Config
}

func getContext(p context.Context, c appengine.Context) context.Context {
	return context.WithValue(p, appengineContext{}, c)
}
