package image

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/version"
)

// ImageRemoveConfig is the configuration for removing an image.
// Use ImageRemoveOption to configure this.
type ImageRemoveConfig struct {
	Force bool
}

func WithRemoveForce(cfg *ImageRemoveConfig) error {
	cfg.Force = true
	return nil
}

// ImageRemoveOption is a functional option for configuring an image remove.
type ImageRemoveOption func(config *ImageRemoveConfig) error

// ImageRemoved represents the response from removing an image.
type ImageRemoved struct {
	Deleted  []string
	Untagged []string
}

type removeStreamResponse struct {
	Deleted  string `json:"Deleted"`
	Untagged string `json:"Untagged"`
}

// Remove removes an image.
func (s *Service) Remove(ctx context.Context, ref string, opts ...ImageRemoveOption) (ImageRemoved, error) {
	var cfg ImageRemoveConfig
	for _, o := range opts {
		if err := o(&cfg); err != nil {
			return ImageRemoved{}, err
		}
	}

	withRemoveConfig := func(req *http.Request) error {
		q := req.URL.Query()
		q.Add("force", strconv.FormatBool(cfg.Force))
		req.URL.RawQuery = q.Encode()
		return nil
	}

	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return s.tr.Do(ctx, http.MethodDelete, version.Join(ctx, "/images/"+ref), withRemoveConfig)
	})
	if err != nil {
		return ImageRemoved{}, err
	}
	defer resp.Body.Close()

	var rmS []removeStreamResponse
	if err := json.NewDecoder(resp.Body).Decode(&rmS); err != nil {
		return ImageRemoved{}, fmt.Errorf("decoding response: %w", err)
	}

	var rm ImageRemoved
	for _, r := range rmS {
		rm.Deleted = append(rm.Deleted, r.Deleted)
		rm.Untagged = append(rm.Untagged, r.Untagged)
	}

	return rm, nil
}
