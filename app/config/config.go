package config

import (
	"appengine"
	"appengine/file"
	"appengine/urlfetch"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

var Config ConfigData

type ConfigData map[string]string

func Load(c ...appengine.Context) ConfigData {
	if Config != nil {
		return Config
	}
	var configFile []byte
	if appengine.IsDevAppServer() {
		configFile, _ = ioutil.ReadFile(path.Join(os.Getenv("PWD"), "config.yaml"))
	} else {
		client := &http.Client{
			Transport: &oauth2.Transport{
				Source: google.AppEngineTokenSource(c[0], storage.ScopeReadOnly),
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
