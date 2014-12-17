package main

import (
    "time"
)

type QueuedUser struct {
	UserID     string
	Discovered time.Time
}

type StoredUserMeta struct {
    UserID  int64
    Username    string
    Location string
    Description string
    Backgroudn string
    Previous struct{
        Usernames []struct {
           Username string
           Discovered   time.Time
        }
        Location []struct {
            Location string
            Discovered  time.Time
        }
        Description []struct {
            Description string
            Discovered  time.Time
        }
        Background  []struct {
            Background string
            Discovered time.Time
        }
    }
    VanityUrl   string
    Verified    bool
    VerifiedDate    time.Time
    AvatarUrl   string
}

type StoredUserData struct {
    UserID int64
    LastUpdated time.Time
    Followers []int64
    Following []int64
    Loops []int64
    AuthoredPostCount []int64
    PostCount []int64
    Likes []int64
    Updated []time.Time
}