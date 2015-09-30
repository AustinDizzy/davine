package data

import (
	"archive/zip"
	"encoding/json"
	"net/http"
	"strconv"

	"google.golang.org/appengine/log"
)

func (db *DB) ExportUser(userIdStr string, w http.ResponseWriter) {
	userId, _ := strconv.ParseInt(userIdStr, 10, 64)

	user, err := db.GetUser(userId)
	if err != nil {
		log.Errorf(db.Context, "got err on export: %v", err)
		return
	}

	userMeta, _ := json.MarshalIndent(user.UserMeta, "", "  ")
	userData, _ := json.MarshalIndent(user.UserData, "", "  ")
	user.UserMeta, user.UserData = nil, nil
	userJson, _ := json.MarshalIndent(user, "", "  ")

	w.Header().Add("Content-Type", "application/zip")
	zipWriter := zip.NewWriter(w)

	var files = []struct {
		Name, Data string
	}{
		{"UserRecord.json", string(userJson)},
		{"UserData.json", string(userData)},
		{"UserMeta.json", string(userMeta)},
	}
	for _, file := range files {
		f, err := zipWriter.Create(file.Name)
		if err != nil {
			log.Errorf(db.Context, err.Error())
		}
		_, err = f.Write([]byte(file.Data))
		if err != nil {
			log.Errorf(db.Context, err.Error())
		}
	}

	err = zipWriter.Close()
	if err != nil {
		log.Errorf(db.Context, err.Error())
	}
}
