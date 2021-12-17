package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/context/ctxhttp"
)

// Client is a Fronius HTTP client.
type Client struct {
	host   string
	client *http.Client
}

// NewClient creates a Fronius HTTP client.
func NewClient(host string) Client {
	return Client{
		host:   host,
		client: &http.Client{},
	}
}

func (c Client) readArchive(ctx context.Context, q url.Values) (result archiveResponse, err error) {
	u := url.URL{Scheme: "http", Host: c.host, Path: "/solar_api/v1/GetArchiveData.cgi", RawQuery: q.Encode()}

	res, err := ctxhttp.Get(ctx, c.client, u.String())
	if err != nil {
		return result, err
	}

	defer func() {
		if cErr := res.Body.Close(); cErr != nil {
			err = cErr
		}
	}()

	if res.StatusCode != http.StatusOK {
		return result, fmt.Errorf("%w: %d", ErrStatusNotOk, res.StatusCode)
	}

	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return result, err
	}

	return result, err
}
