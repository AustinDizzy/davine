package data

import "time"

type QueuedUser struct {
	UserID     string
	Discovered time.Time
}

type UserRecord struct {
	UserId            string
	Discovered        time.Time
	Username          string `datastore:",noindex"`
	Vanity            string
	Description       string `datastore:",noindex"`
	Location          string `datastore:",noindex"`
	ProfileBackground string `datastore:",noindex"`
	AvatarUrl         string `datastore:",noindex"`
	FollowerCount     int64
	FollowingCount    int64
	LoopCount         int64
	LoopVelocity      float64
	PostCount         int64
	RevineCount       int64
	LikeCount         int64
	Private           bool
	Verified          bool
	Explicit          bool
	UserData          []*UserData `datastore:"-"`
	UserMeta          []*UserMeta `datastore:"-"`
	UserDataJsonStr   string      `datastore:"-"`
	UserMetaJsonStr   string      `datastore:"-"`
}

type UserData struct {
	UserId    int64
	Recorded  time.Time
	Followers int64
	Following int64
	Loops     int64
	Velocity  float64
	Posts     int64
	Revines   int64
	Likes     int64
}

type UserMeta struct {
	UserId  int64
	Record  string
	Value   string `datastore:",noindex"`
	Updated time.Time
}

type MetaStats struct {
	Count     int       `datastore:"count"`
	Timestamp time.Time `datastore:"timestamp"`
}

type ByOverall []UserRecord
type ByFollowers []UserRecord
type ByLoops []UserRecord
type ByPosts []UserRecord
type ByRevines []UserRecord

type UserIndex struct {
	Username, Location, Description, VanityUrl string
}
