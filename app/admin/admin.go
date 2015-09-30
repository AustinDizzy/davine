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
	"google.golang.org/appengine/log"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
	"ronoaldo.gopkg.net/aetools"

	"google.golang.org/appengine/taskqueue"

	"appengine"
	"appengine/urlfetch"
)

type task struct {
	c   context.Context
	ctx appengine.Context
}

//NewTask initiats a new admin-spepcific task.
func NewTask(c context.Context) *task {
	return &task{c: c}
}

//LoadCtx loads a classic appengine.Context into the admin
//task given a http.Request.
func (a *task) LoadCtx(r *http.Request) {
	a.ctx = appengine.NewContext(r)
}

//BatchTaskUsers adds new user fetch tasks given an array of
//comma delimited rows, where the row is in the following format:
//	userID,n,delay (Ex: 9012415478910456,1,15m)
//		userID - the user's user ID.
//		n - the nth fetch task for this user
//		delay - the delay in which to execute the task, in time.Duration format
func (a *task) BatchTaskUsers(usersRow ...string) {
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
			log.Errorf(a.c, "Error parsing task delay %v: %v", u, err)
			continue
		}

		if _, err = taskqueue.Add(a.c, t, ""); err != nil {
			log.Errorf(a.c, "Error adding user %s to taskqueue: %v", u[0], err)
		}
	}
}

//DumpData dumps an entire datastore kind to JSON format
//and writes the file to the given http.ResponseWriter.
func (a *task) DumpData(kind string, w http.ResponseWriter) error {
	opts := &aetools.Options{Kind: kind, PrettyPrint: true}
	w.Header().Set("Content-Disposition", "attachment; filename=\""+kind+".json\"")
	err := aetools.Dump(a.ctx, w, opts)
	return err
}

//LoadData loads a specific datastore kind, in raw datastore
//record format, into the datastore.
func (a *task) LoadData(kind string, file io.Reader) error {
	opts := &aetools.Options{GetAfterPut: true, Kind: kind}
	return aetools.Load(a.ctx, file, opts)
}

func (a *task) LoadGSData(name string) error {
	if a.ctx == nil {
		return errors.New("context is nil")
	}
	var client *http.Client
	tokenSource, err := google.DefaultTokenSource(a.c, storage.ScopeReadOnly)
	if appengine.IsDevAppServer() {
		authFile, err := ioutil.ReadFile(path.Join(os.Getenv("PWD"), "client_secret.json"))
		if err != nil {
			return err
		}
		cfg, err := google.ConfigFromJSON(authFile, storage.ScopeReadOnly)
		if err != nil {
			return err
		}
		t, _ := tokenSource.Token()
		client = cfg.Client(a.c, t)
	} else {
		client = &http.Client{
			Transport: &oauth2.Transport{
				Source: tokenSource,
				Base: &urlfetch.Transport{
					Context: a.ctx,
				},
			},
		}
	}
	ctx := cloud.NewContext("davine-web", client)
	rc, err := storage.NewReader(ctx, "davine-web.appspot.com", name)
	if err != nil {
		log.Errorf(a.c, "Error reading %s in %s: %v", name, "davine-web.appspot.com", err)
		return err
	}
	err = aetools.Load(a.ctx, rc, aetools.LoadSync)
	rc.Close()
	return err
}
