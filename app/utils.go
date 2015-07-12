package main

import (
	"crypto/rand"
)

func genRand(dict string, n int) string {
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dict[v%byte(len(dict))]
	}

	return string(bytes)
}

func GenKey() string {
	dict := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	dict += "abcdefghijklmnopqrstuvwxyz"
	dict += "1234567890=+~-"
	return genRand(dict, 64)
}

func GenSlug() string {
	dict := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	dict += "1234567890"
	dict += "abcdefghijklmnopqrstuvwxyz"
	return genRand(dict, 6)
}
