package config

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"google.golang.org/appengine/log"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
	"gopkg.in/yaml.v2"

	"google.golang.org/appengine"
	"google.golang.org/appengine/file"
	"google.golang.org/appengine/urlfetch"
)

var Config ConfigData

type (
	ConfigData map[string]string
)

func Load(c context.Context) ConfigData {
	if len(Config) != 0 {
		return Config
	}
	var configFile []byte
	if appengine.IsDevAppServer() {
		configFile, _ = ioutil.ReadFile(path.Join(os.Getenv("PWD"), "config.yaml"))
	} else {
		tokenSource, err := google.DefaultTokenSource(c, storage.ScopeReadOnly)
		client := &http.Client{
			Transport: &oauth2.Transport{
				Source: tokenSource,
				Base: &urlfetch.Transport{
					Context: c,
				},
			},
		}
		ctx := cloud.NewContext("davine-web", client)
		bucket, _ := file.DefaultBucketName(c)
		rc, err := storage.NewReader(ctx, bucket, "config.yaml")
		if err != nil {
			log.Errorf(c, "error reading config: %v", err.Error())
		}
		configFile, err = ioutil.ReadAll(rc)
		rc.Close()
		if err != nil {
			log.Errorf(c, "error reading config: %v", err.Error())
		}
	}
	yaml.Unmarshal(configFile, &Config)
	return Config
}
