package main

import (
    "time"
)

type QueuedUser struct {
	UserID     string
	Discovered time.Time
}

type VineUserWrapper struct {
    Data *VineUser
    Success bool `json:"success"`
    Error string `json:"error"`
}

type VineUser struct {
    Username string `json:"username"`
    FollowerCount int64 `json:"followerCount"`
    Verified int `json:"verified"`
    VanityUrls []string `json:"vanityUrls"`
    LoopCount int64 `json:"loopCount"`
    AvatarUrl string `json:"avatarUrl"`
    AuthoredPostCount int64 `json:"authoredPostCount"`
    UserId int64 `json:"userId"`
    UserIdStr string `json:"userIdStr"`
    PostCount int64 `json:"postCount"`
    ProfileBackground string `json:"profileBackground"`
    LikeCount int64 `json:"likeCount"`
    Private int `json:"private"`
    Location string `json:"location"`
    FollowingCount int64 `json:"followingCount"`
    ExplicitContent int `json:"explicitContent"`
    Description string `json:"description"`
}

type StoredUserMeta struct {
    Username    string
    Location string
    Description string
    Background string
    Previous struct{
        Username []PreviousUsername
        Location []PreviousLocation
        Description []PreviousDescription
        Background []PreviousBackground
    }
    VanityUrl   string
    Verified    bool
    VerifiedDate    time.Time
    AvatarUrl   string
}

type StoredUserData struct {
    LastUpdated time.Time
    Followers []int64
    Following []int64
    Loops []int64
    AuthoredPosts []int64
    Reposts []int64
    Likes []int64
    Updated []time.Time
}

type StoredUserDelta struct {
    Followers []int64
    Following []int64
    Loops []int64
    AuthoredPostCount []int64
    PostCount []int64
    Likes []int64
}

type PreviousUsername struct {
    Username string
    Modified   time.Time
}

type PreviousLocation struct {
    Location string
    Modified  time.Time
}

type PreviousDescription struct {
    Description string
    Modified  time.Time
}

type PreviousBackground  struct {
    Background string
    Modified time.Time
}