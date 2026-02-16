package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/kianjones9/checkpoint.af/internal/api"
	"github.com/kianjones9/checkpoint.af/internal/export"
)

func main() {

	client := http.Client{}

	// user inputs
	host := flag.String("base_url", "api.letta.com", "hostname of Letta server (defaults to api.letta.com)")
	lettaApiKey := flag.String("api_key", os.Getenv("LETTA_API_KEY"), "api key for authenticating to Letta")
	agentId := flag.String("agent_id", "agent-<uuid4>", "ID of agent to snapshot")
	// TODO: support exporting to af.directory
	dest := flag.String("dest", "", "URI of desired destination (supports file, s3, gcs, azure, and other compatible backends: https://gocloud.dev/howto/blob/#services)")
	// only_on_diff
	checkpointServer := flag.String("checkpoint_server", "http://127.0.0.1:8080", "URI of checkpoint server if running (executes locally if not running)")

	flag.Parse()

	payload := api.Config{
		Host:        *host,
		Destination: *dest,
		BearerToken: *lettaApiKey,
		AgentId:     *agentId,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
		return
	}

	resp, err := http.Post(*checkpointServer+"/save", "application/json", bytes.NewBuffer(jsonPayload))
	if err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			fmt.Println("Success:", string(body))
		} else {
			fmt.Println("Failed:", resp.StatusCode, string(body))
		}
		return
	}

	req, err := export.BuildRequest(*host, *agentId, *lettaApiKey)
	if err != nil {
		log.Fatal(err)
	}

	// perform request
	resp, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	bucket, w, err := export.BuildWriter(req.Context(), *dest)
	if err != nil {
		log.Fatal(err)
	}
	defer bucket.Close()
	defer w.Close()

	export.ProcessResponse(resp, w)
}
