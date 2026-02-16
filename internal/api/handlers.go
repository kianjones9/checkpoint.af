package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/kianjones9/checkpoint.af/internal/export"
)

type Config struct {
	BearerToken string `json:"Authorization"`
	AgentId     string `json:"agent_id"`
	Host        string `json:"host"`
	Destination string `json:"destination"`
}

// save
func Save(writer http.ResponseWriter, req *http.Request) {

	// expect args as json fields in request body
	payload, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	config := Config{}

	err = json.Unmarshal(payload, &config)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	req, err = export.BuildRequest(config.Host, config.AgentId, config.BearerToken)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	// perform request
	// TODO: optimize client here
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("BuildWriter dest=%q", config.Destination)
	bucket, w, err := export.BuildWriter(req.Context(), config.Destination)
	if err != nil {
		log.Printf("BuildWriter error: %v", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("BuildWriter ok, bucket=%v writer=%v", bucket, w)

	err = export.ProcessResponse(resp, w)
	if err != nil {
		log.Printf("ProcessResponse error: %v", err)
		bucket.Close()
		w.Close()
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if err := w.Close(); err != nil {
		log.Printf("writer close error: %v", err)
	}
	if err := bucket.Close(); err != nil {
		log.Printf("bucket close error: %v", err)
	}
	log.Println("save complete")
}

// rollback (unimplemented)
func Rollback(http.ResponseWriter, *http.Request) {}

// migrate (unimplemented)
func Migrate(http.ResponseWriter, *http.Request) {}
