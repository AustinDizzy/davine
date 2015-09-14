package admin

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	newurlfetch "google.golang.org/appengine/urlfetch"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
	"ronoaldo.gopkg.net/aetools"

	"appengine"
	"appengine/taskqueue"
)

type AdminTask struct {
	c   appengine.Context
	ctx context.Context
}

func NewTask(c appengine.Context) *AdminTask {
	return &AdminTask{c: c}
}

func (a *AdminTask) LoadCtx(c context.Context) {
	a.ctx = c
}

func (a *AdminTask) BatchTaskUsers(usersRow ...string) {
	var err error
	for _, user := range usersRow {
		u := strings.Split(user, ",")
		t := taskqueue.NewPOSTTask("/cron/fetch", map[string][]string{
			"id": {strings.TrimSpace(u[0])},
			"n":  {strings.TrimSpace(u[1])},
		})
		t.Name = u[0] + "-0"
		t.Delay, err = time.ParseDuration(strings.TrimSpace(u[2]))

		if err != nil {
			a.c.Errorf("Error parsing task delay %v: %v", u, err)
			continue
		}

		if _, err = taskqueue.Add(a.c, t, ""); err != nil {
			a.c.Errorf("Error adding user %s to taskqueue: %v", u[0], err)
		}
	}
}

func (a *AdminTask) DumpData(kind string, w http.ResponseWriter) error {
	opts := &aetools.Options{Kind: kind, PrettyPrint: true}
	w.Header().Set("Content-Disposition", "attachment; filename=\""+kind+".json\"")
	err := aetools.Dump(a.c, w, opts)
	return err
}

func (a *AdminTask) LoadData(kind string, file io.Reader) error {
	opts := &aetools.Options{GetAfterPut: true, Kind: kind}
	return aetools.Load(a.c, file, opts)
}

func (a *AdminTask) LoadGSData(name string) error {
	if a.ctx == nil {
		return errors.New("context is nil")
	}
	var client *http.Client
	if appengine.IsDevAppServer() {
		authFile, err := ioutil.ReadFile(path.Join(os.Getenv("PWD"), "client_secret.json"))
		if err != nil {
			return err
		}
		tokenSource, err := google.DefaultTokenSource(a.ctx, storage.ScopeReadOnly)
		cfg, err := google.ConfigFromJSON(authFile, storage.ScopeReadOnly)
		if err != nil {
			return err
		}
		t, _ := tokenSource.Token()
		client = cfg.Client(a.ctx, t)
	} else {
		client = &http.Client{
			Transport: &oauth2.Transport{
				Source: google.AppEngineTokenSource(a.ctx, storage.ScopeReadOnly),
				Base: &newurlfetch.Transport{
					Context: a.ctx,
				},
			},
		}
	}
	ctx := cloud.NewContext("davine-web", client)
	rc, err := storage.NewReader(ctx, "davine-web.appspot.com", name)
	if err != nil {
		a.c.Errorf("Error reading %s in %s: %v", name, "davine-web.appspot.com", err)
		return err
	}
	defer rc.Close()
	return aetools.Load(a.c, rc, aetools.LoadSync)
}
