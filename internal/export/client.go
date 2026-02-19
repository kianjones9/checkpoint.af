package export

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gocloud.dev/blob"
	"gocloud.dev/gcerrors"

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
	OnlyOnDiff  bool   `json:"only_on_diff"`
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

func ProcessResponseWithDeduplication(ctx context.Context, resp *http.Response, bucket *blob.Bucket, filename string) error {
	// process results
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("upstream returned %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	hasher := sha256.New()
	tee := io.TeeReader(resp.Body, hasher)

	_, err := io.Copy(&buf, tee)
	if err != nil {
		return err
	}

	newHash := hex.EncodeToString(hasher.Sum(nil))

	attrs, err := bucket.Attributes(ctx, filename)
	if err != nil && gcerrors.Code(err) != gcerrors.NotFound {
		return err
	}
	if err == nil && attrs.Metadata["sha256"] == newHash {
		return nil
	}

	w, err := bucket.NewWriter(ctx, filename, &blob.WriterOptions{
		Metadata: map[string]string{"sha256": newHash},
	})
	if err != nil {
		return err
	}

	_, err = io.Copy(w, &buf)
	if err != nil {
		w.Close()
		return err
	}

	return w.Close()
}

func BuildWriter(ctx context.Context, dest string, agentId string, overwrite bool) (*blob.Bucket, string, error) {

	uri, err := url.Parse(dest)
	if err != nil {
		return nil, "", err
	}

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

	filename := ""

	switch uri.Scheme {

	// currently supports file, s3, gcs, etc.: https://gocloud.dev/howto/blob/#services
	default:
		bucket, err = blob.OpenBucket(ctx, uri.String())
		if err != nil {
			return nil, "", err
		}
		bucket = blob.PrefixedBucket(bucket, prefix)

		filename = agentId
		if !overwrite {
			filename += "_" + time.Now().UTC().Format("20060102-150405")
		}
		filename += ".af"
	}
	return bucket, filename, nil
}
