// Command servly boots the API service on :8080. This is the only place in
// the module that touches a real network listener or the real wall clock —
// everything under package servly is pure and deterministic, wired here
// with RealClock for production use.
package main

import (
	"log"
	"net/http"
	"time"

	servly "github.com/WorkSim/servly-template"
)

func main() {
	store := servly.NewStore([]servly.Record{
		{ID: 1, Name: "alpha"},
		{ID: 2, Name: "bravo"},
		{ID: 3, Name: "charlie"},
	})

	tokens := servly.NewTokenStore()
	tokens.Issue("demo-token", servly.Token{Key: "demo", ExpiresAt: time.Now().UTC().Add(24 * time.Hour)})

	handler := servly.NewServer(store, tokens, servly.RealClock{})

	log.Println("servly listening on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
