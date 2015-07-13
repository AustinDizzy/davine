package admin

import (
	"appengine"
	"appengine/taskqueue"
	"github.com/ronoaldo/aetools"
	"io"
	"net/http"
	"strings"
	"time"
)

type AdminTask struct {
	c appengine.Context
}

func NewTask(c appengine.Context) *AdminTask {
	return &AdminTask{c}
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
