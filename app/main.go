package main

import (
    "fmt"
    "net/http"
)

func init() {
    http.HandleFunc("/user", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Hello, world!")
}