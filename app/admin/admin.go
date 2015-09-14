package admin

import (
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/cloud/storage"
	"ronoaldo.gopkg.net/aetools"

	"appengine"
	"appengine/file"
	"appengine/taskqueue"
)

type AdminTask struct {
	c   appengine.Context
	ctx context.Context
}

type appengineContext struct{}

func NewTask(c appengine.Context) *AdminTask {
	t := &AdminTask{c: c}
	t.ctx = getContext(t.ctx, c)
	return t
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
	bucket, _ := file.DefaultBucketName(a.c)
	rc, err := storage.NewReader(a.ctx, bucket, name)
	if err != nil {
		a.c.Errorf("Error reading %s in %s: %v", name, bucket, err)
		return err
	}
	defer rc.Close()
	return aetools.Load(a.c, rc, aetools.LoadSync)
}

func getContext(p context.Context, c appengine.Context) context.Context {
	return context.WithValue(p, appengineContext{}, c)
}
