package main

import (
	"log"
	"net/http"
	"time"

	"github.com/kianjones9/checkpoint.af/internal/api"
)

func logRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("%s %s", r.Method, r.URL.Path)
		next(w, r)
		log.Printf("%s %s completed in %v", r.Method, r.URL.Path, time.Since(start))
	}
}

func main() {

	client := &http.Client{}

	http.HandleFunc("/save", logRequest(api.Save(client)))
	http.HandleFunc("/rollback", logRequest(api.Rollback))
	http.HandleFunc("/migrate", logRequest(api.Migrate))

	log.Println("listening on 0.0.0.0:8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
