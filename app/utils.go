package main

import (
	"crypto/rand"
	mr "math/rand"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func GenKey() string {
	k := make([]byte, 32)
	rand.Read(k)
	return string(k[:])
}

func GenSlug() string {
	b := make([]rune, 5)
	mr.Seed(time.Now().UTC().UnixNano())
	for i := range b {
		b[i] = letters[mr.Intn(len(letters))]
	}
	return string(b[:])
}
