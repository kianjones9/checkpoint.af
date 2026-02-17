package export

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gocloud.dev/blob"

	_ "gocloud.dev/blob/azureblob" // Azure
	_ "gocloud.dev/blob/fileblob"  // local filesystem
	_ "gocloud.dev/blob/gcsblob"   // GCS
	_ "gocloud.dev/blob/memblob"   // In-Memory
	_ "gocloud.dev/blob/s3blob"    // S3
)

type Config struct {
	BearerToken string `json:"Authorization"`
	AgentId     string `json:"agent_id"`
	Host        string `json:"host"`
	Destination string `json:"destination"`
	Overwrite   bool   `json:"overwrite"`
}

func BuildRequest(host string, agentId string, bearerToken string) (*http.Request, error) {

	// format inputs for use
	urlPrefix := "v1/agents"
	urlSuffix := "export"

	endpoint := url.URL{
		Scheme: "https",
		Host:   host,
		Path:   urlPrefix + "/" + agentId + "/" + urlSuffix,
	}

	authHeader := "Bearer " + bearerToken

	// construct request
	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", authHeader)

	return req, nil
}

func ProcessResponse(resp *http.Response, w io.Writer) error {
	// process results
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("upstream returned %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	_, err := io.Copy(w, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func BuildWriter(ctx context.Context, dest string, agentId string, overwrite bool) (*blob.Bucket, io.WriteCloser, error) {

	uri, err := url.Parse(dest)
	if err != nil {
		return nil, nil, err
	}

	var w io.WriteCloser
	var bucket *blob.Bucket

	prefix := ""

	// fileblob needs no_tmp_dir to avoid cross-device rename errors in containers
	if uri.Scheme == "file" {
		q := uri.Query()
		q.Set("no_tmp_dir", "true")
		uri.RawQuery = q.Encode()
	} else {
		prefix = strings.TrimPrefix(uri.Path, "/")
		if prefix != "/" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
	}

	switch uri.Scheme {

	// currently supports file, s3, gcs, etc.: https://gocloud.dev/howto/blob/#services
	default:
		bucket, err = blob.OpenBucket(ctx, uri.String())
		if err != nil {
			return nil, nil, err
		}
		bucket = blob.PrefixedBucket(bucket, prefix)

		filename := agentId
		if !overwrite {
			filename += "_" + time.Now().UTC().Format("20060102-150405")
		}
		filename += ".af"

		w, err = bucket.NewWriter(ctx, filename, nil)
		if err != nil {
			return nil, nil, err
		}
	}
	return bucket, w, nil
}
