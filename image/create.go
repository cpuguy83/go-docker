package image

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/version"
)

// CreateConfig holds the options for creating images.
type CreateConfig struct {
	FromImage string
	FromSrc   string
	Repo      string
	Tag       string
	Platform  string
}

// CreateOption is used as functional arguments to create images.
// CreateOption configure a CreateConfig.
type CreateOption func(config *CreateConfig)

func (s *Service) Create(ctx context.Context, opts ...CreateOption) error {
	cfg := CreateConfig{}
	for _, o := range opts {
		o(&cfg)
	}

	withCreateConfig := func(req *http.Request) error {
		q := req.URL.Query()
		q.Add("fromImage", cfg.FromImage)
		q.Add("fromSrc", cfg.FromSrc)
		q.Add("repo", cfg.Repo)
		q.Add("tag", cfg.Tag)
		q.Add("platform", cfg.Platform)
		req.URL.RawQuery = q.Encode()
		return nil
	}

	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return s.tr.Do(ctx, http.MethodPost, version.Join(ctx, "/images/create"), withCreateConfig)
	})
	if err != nil {
		return err
	}
	// Reading all is required otherwise the docker daemon considers the context
	// as cancelled and the pulling is aborted.
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
