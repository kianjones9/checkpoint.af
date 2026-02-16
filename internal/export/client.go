package export

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"gocloud.dev/blob"

	_ "gocloud.dev/blob/azureblob" // Azure
	_ "gocloud.dev/blob/fileblob"  // local filesystem
	_ "gocloud.dev/blob/gcsblob"   // GCS
	_ "gocloud.dev/blob/memblob"   // In-Memory
	_ "gocloud.dev/blob/s3blob"    // S3
)

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
	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

func ProcessResponse(resp *http.Response, w io.Writer) error {
	// process results
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("Response status code not 2XX")
	}
	defer resp.Body.Close()

	_, err := io.Copy(w, resp.Body)
	if err != nil {
		return err
	}

	fmt.Println("Succeeded, response status code: ", resp.StatusCode)
	return nil
}

func BuildWriter(ctx context.Context, dest string) (*blob.Bucket, io.WriteCloser, error) {

	uri, err := url.Parse(dest)
	if err != nil {
		return nil, nil, err
	}

	var w io.WriteCloser = nil
	var bucket *blob.Bucket

	// fileblob needs no_tmp_dir to avoid cross-device rename errors in containers
	if uri.Scheme == "file" {
		q := uri.Query()
		q.Set("no_tmp_dir", "true")
		uri.RawQuery = q.Encode()
	}

	switch uri.Scheme {

	// currently supports file, s3, gcs, etc.: https://gocloud.dev/howto/blob/#services
	default:
		bucket, err = blob.OpenBucket(ctx, uri.String())
		if err != nil {
			return nil, nil, err
		}

		w, err = bucket.NewWriter(ctx, time.Now().UTC().Format("20060102-150405")+".af", nil)
		if err != nil {
			return nil, nil, err
		}
	}
	return bucket, w, nil
}
