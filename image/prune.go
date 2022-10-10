package image

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/image/imageapi"
	"github.com/cpuguy83/go-docker/version"
)

// PruneFilter represents filters to process on the prune list. See the official
// docker docs for the meaning of each field
// https://docs.docker.com/engine/api/v1.41/#operation/ImagePrune
type PruneFilter struct {
	Dangling []string `json:"dangling,omitempty"`
	Label    []string `json:"label,omitempty"`
	NotLabel []string `json:"label!,omitempty"`
	Until    []string `json:"until,omitempty"`
}

// PruneConfig holds the options for pruning images.
type PruneConfig struct {
	Filters PruneFilter
}

// PruneOption is used as functional arguments to prune images. PruneOption
// configure a PruneConfig.
type PruneOption func(config *PruneConfig)

// prune prunes container images.
func (s *Service) Prune(ctx context.Context, opts ...PruneOption) (imageapi.Prune, error) {
	cfg := PruneConfig{}
	for _, o := range opts {
		o(&cfg)
	}

	withPruneConfig := func(req *http.Request) error {
		q := req.URL.Query()
		filterJSON, err := json.Marshal(cfg.Filters)
		if err != nil {
			return err
		}
		_ = filterJSON
		q.Add("filters", string(filterJSON))
		req.URL.RawQuery = q.Encode()
		return nil
	}

	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return s.tr.Do(ctx, http.MethodPost, version.Join(ctx, "/images/prune"), withPruneConfig)
	})
	if err != nil {
		return imageapi.Prune{}, fmt.Errorf("pruning images: %w", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return imageapi.Prune{}, err
	}
	var prune imageapi.Prune
	if err := json.Unmarshal(data, &prune); err != nil {
		return imageapi.Prune{}, fmt.Errorf("reading response body: %w", err)
	}
	return prune, nil
}
