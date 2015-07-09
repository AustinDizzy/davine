package main

import (
	"time"
)

const PageTitle = "Davine - Open Data Analytics for Vine"

type QueuedUser struct {
	UserID     string
	Discovered time.Time
}

type VinePopularWrapper struct {
	Data    PopularPage `json:"data"`
	Success bool        `json:"success"`
	Error   string      `json:"error"`
}

type VineUserWrapper struct {
	Data    *VineUser `json:"data"`
	Success bool      `json:"success"`
	Error   string    `json:"error"`
}

type VineUser struct {
	Username          string   `json:"username"`
	FollowerCount     int64    `json:"followerCount"`
	Verified          int      `json:"verified"`
	VanityUrls        []string `json:"vanityUrls"`
	LoopCount         int64    `json:"loopCount"`
	AvatarUrl         string   `json:"avatarUrl"`
	AuthoredPostCount int64    `json:"authoredPostCount"`
	UserId            int64    `json:"userId"`
	UserIdStr         string   `json:"userIdStr"`
	PostCount         int64    `json:"postCount"`
	ProfileBackground string   `json:"profileBackground"`
	LikeCount         int64    `json:"likeCount"`
	Private           int      `json:"private"`
	Location          string   `json:"location"`
	FollowingCount    int64    `json:"followingCount"`
	ExplicitContent   int      `json:"explicitContent"`
	Description       string   `json:"description"`
}

type UserRecord struct {
    UserId            string
    Discovered        time.Time `datastore:",noindex"`
	Username          string `datastore:",noindex"`
	Vanity            string
	Description       string `datastore:",noindex"`
	Location          string `datastore:",noindex"`
	ProfileBackground string `datastore:",noindex"`
	AvatarUrl         string `datastore:",noindex"`
	FollowerCount     int64 `datastore:",noindex"`
	FollowingCount    int64 `datastore:",noindex"`
	LoopCount         int64 `datastore:",noindex"`
	PostCount         int64 `datastore:",noindex"`
	RevineCount       int64 `datastore:",noindex"`
	LikeCount         int64 `datastore:",noindex"`
	Private           bool `datastore:",noindex"`
	Verified          bool
	Explicit          bool
	UserData          []*UserData `datastore:"-"`
	UserMeta          []*UserMeta `datastore:"-"`
	UserDataJsonStr   string `datastore:"-"`
	UserMetaJsonStr   string `datastore:"-"`
}

type UserData struct {
    UserId        int64
    Recorded      time.Time
    Followers     int64
	Following     int64 `datastore:",noindex"`
	Loops         int64
	Posts         int64
	Revines       int64
	Likes         int64
}

type UserMeta struct {
    UserId  int64
    Record  string
    Value   string    `datastore:",noindex"`
    Updated time.Time
}

type StoredUserMeta struct {
	Username, UserId string
	Location         string `datastore:",noindex"`
	Description      string `datastore:",noindex"`
	Background       string `datastore:",noindex"`
	Current          StoredUserMetaCurrent
	Previous         struct {
		Username    []PreviousUsername    `datastore:",noindex"`
		Location    []PreviousLocation    `datastore:",noindex"`
		Description []PreviousDescription `datastore:",noindex"`
		Background  []PreviousBackground  `datastore:",noindex"`
	} `datastore:",noindex"`
	VanityUrl    string
	Verified     bool
	VerifiedDate time.Time `datastore:",noindex"`
	AvatarUrl    string    `datastore:",noindex"`
}

type StoredUserMetaCurrent struct {
	Followers     int64
	Following     int64
	Loops         int64
	AuthoredPosts int64
	Revines       int64
	Likes         int64
}

type StoredUserData struct {
	LastUpdated   time.Time
	Followers     []int64     `datastore:",noindex"`
	Following     []int64     `datastore:",noindex"`
	Loops         []int64     `datastore:",noindex"`
	AuthoredPosts []int64     `datastore:",noindex"`
	Revines       []int64     `datastore:",noindex"`
	Likes         []int64     `datastore:",noindex"`
	Updated       []time.Time `datastore:",noindex"`
}

type PopularPage struct {
	AnchorStr string           `json:"anchorStr"`
	Records   []*PopularRecord `json:"records"`
	NextPage  int              `json:"nextPage"`
	Size      int              `json:"size"`
}

type PopularRecord struct {
	UserId    int64  `json:"userId"`
	UserIdStr string `json:"userIdStr"`
}


type PreviousUsername struct {
	Username string    `datastore:",noindex"`
	Modified time.Time `datastore:",noindex"`
}

type PreviousLocation struct {
	Location string    `datastore:",noindex"`
	Modified time.Time `datastore:",noindex"`
}

type PreviousDescription struct {
	Description string    `datastore:",noindex"`
	Modified    time.Time `datastore:",noindex"`
}

type PreviousBackground struct {
	Background string    `datastore:",noindex"`
	Modified   time.Time `datastore:",noindex"`
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

type FeaturedUser struct {
	UserIDStr, VineID string
}

type AppUser struct {
	Email      string
	Type       string
	Active     bool
	UserIdStr  string
	Key        string
	Discovered time.Time
}
