package main

import (
    "time"
)

type QueuedUser struct {
	UserID     string
	Discovered time.Time
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
    Followers []float64
    Following []float64
    Loops []float64
    AuthoredPosts []float64
    Reposts []float64
    Likes []float64
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