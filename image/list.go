package image

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/image/imageapi"
	"github.com/cpuguy83/go-docker/version"
)

// ListFilter represents filters to process on the image list. See the official
// docker docs for the meaning of each field
// https://docs.docker.com/engine/api/v1.41/#operation/ImageList
type ListFilter struct {
	Before    []string `json:"before,omitempty"`
	Dangling  []string `json:"dangling,omitempty"`
	Label     []string `json:"label,omitempty"`
	Reference []string `json:"reference,omitempty"`
	Since     []string `json:"since,omitempty"`
}

// ListConfig holds the options for listing images.
type ListConfig struct {
	All     bool
	Digests bool
	Filter  ListFilter
}

// ListOption is used as functional arguments to list images.
// ListOption configure a ListConfig.
type ListOption func(config *ListConfig)

// List lists images.
func (s *Service) List(ctx context.Context, opts ...ListOption) ([]imageapi.Image, error) {
	cfg := ListConfig{}
	for _, o := range opts {
		o(&cfg)
	}

	withListConfig := func(req *http.Request) error {
		q := req.URL.Query()
		q.Add("all", strconv.FormatBool(cfg.All))
		q.Add("digests", strconv.FormatBool(cfg.Digests))
		filterJSON, err := json.Marshal(cfg.Filter)

		if err != nil {
			return err
		}
		q.Add("filters", string(filterJSON))

		req.URL.RawQuery = q.Encode()
		return nil
	}

	var images []imageapi.Image

	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return s.tr.Do(ctx, http.MethodGet, version.Join(ctx, "/images/json"), withListConfig)
	})
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &images); err != nil {
		return nil, fmt.Errorf("unmarshaling container json: %s", err)
	}

	return images, nil
}
