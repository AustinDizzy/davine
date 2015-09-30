package admin

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type FeaturedUser struct {
	UserID, PostID string
}

type AppUser struct {
	Email      string
	Type       string
	Active     bool
	UserIdStr  string
	AuthKey    string
	Discovered time.Time
}

func getAppUsers(c context.Context) ([]*AppUser, []*AppUser, error) {
	var enterpriseUsers []*AppUser
	var emailReportUsers []*AppUser
	q := datastore.NewQuery("AppUser").KeysOnly()
	keys, _ := q.GetAll(c, nil)

	for _, v := range keys {
		u := new(AppUser)
		if err := datastore.Get(c, v, u); err == nil {
			if u.Type == "enterprise" {
				enterpriseUsers = append(enterpriseUsers, u)
			} else if u.Type == "email-report" {
				emailReportUsers = append(emailReportUsers, u)
			}
		} else {
			log.Errorf(c, "got err: %v", err)
			return enterpriseUsers, emailReportUsers, err
		}
	}
	return enterpriseUsers, emailReportUsers, nil
}
