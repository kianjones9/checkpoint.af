package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/kianjones9/checkpoint.af/internal/export"
)

// save
func Save(client *http.Client) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		// expect args as json fields in request body
		payload, err := io.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		config := export.Config{}

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
		resp, err := client.Do(req)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadGateway)
			return
		}

		bucket, w, err := export.BuildWriter(req.Context(), config.Destination, config.AgentId, config.Overwrite)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		err = export.ProcessResponse(resp, w)
		if err != nil {
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
		log.Print("save complete")
	}
}

// rollback (unimplemented)
func Rollback(http.ResponseWriter, *http.Request) {}

// migrate (unimplemented)
func Migrate(http.ResponseWriter, *http.Request) {}
