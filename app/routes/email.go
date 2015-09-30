package routes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"app/email"
	"app/utils"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

//EmailHandler is the http request handler for /_ah/mail/*.
//It parses incoming emails sent to preconfigured mail addresses
//and handles them accordingly.
//For instance, sending Vine's "Invite by Email" to share@<domain>
//will queue the user.
//More email routes may come in the future.
func EmailHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if msg, err := email.Read(r.Body); err != nil {
		log.Errorf(c, "err reading email: %v", err)
	} else {
		switch strings.Split(msg.Header.Get("To"), "@")[0] {
		case "share":
			regex := regexp.MustCompile(`(?:vine.co/u/)([0-9]+)`)
			matches := regex.FindAllStringSubmatch(msg.Body.Text, 1)
			if len(matches) > 0 {
				client := urlfetch.Client(c)
				url := fmt.Sprintf("http://%s/user?id=%s", appengine.DefaultVersionHostname(c), matches[0][1])
				req, _ := http.NewRequest("POST", url, nil)
				resp, err := client.Do(req)
				if err != nil {
					log.Errorf(c, "got err: %v", err)
				} else {
					body, _ := ioutil.ReadAll(resp.Body)
					var data map[string]bool
					json.Unmarshal(body, &data)
					if data["exists"] {
						msg := email.New()
						if data["stored"] {
							msg.LoadTemplate(2, map[string]interface{}{
								"stored": strconv.FormatBool(data["stored"]),
								"id":     matches[0][1],
							})
						} else {
							msg.LoadTemplate(2, map[string]interface{}{
								"id": matches[0][1],
							})
						}
						msg.Send(c)
						utils.PostCount(c, "shared via email", 1)
					}
				}
			}
		}
	}
}
