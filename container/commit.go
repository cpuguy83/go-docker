package container

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/cpuguy83/go-docker/container/containerapi"
	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/version"
	"github.com/pkg/errors"
)

// CommitOption is used as a funtional option when commiting a container to an
// an image.
type CommitOption func(*CommitConfig)

// CommitConfig is used by CommitOption to set options used for committing a
// container to an image.
type CommitConfig struct {
	Comment   string
	Author    string
	Changes   []string
	Message   string
	Pause     *bool
	Config    *containerapi.Config
	Reference *CommitImageReference
}

// CommitImageReference sets the image reference to use when committing an image.
type CommitImageReference struct {
	Repo string
	Tag  string
}

type containerCommitResponse struct {
	ID string `json:"Id"`
}

// Commit takes a snapshot of the container's filessystem and creates an image
// from it.
// TODO: Return an image type that can be inspected, etc.
func (c *Container) Commit(ctx context.Context, opts ...CommitOption) (string, error) {
	var cfg CommitConfig
	for _, o := range opts {
		o(&cfg)
	}

	var repo, tag string

	if cfg.Reference != nil {
		repo = cfg.Reference.Repo
		tag = cfg.Reference.Tag
	}

	withOptions := func(req *http.Request) error {
		q := req.URL.Query()
		q.Set("container", c.id)
		q.Set("repo", repo)
		q.Set("tag", tag)
		q.Set("author", cfg.Author)
		q.Set("comment", cfg.Comment)
		if cfg.Pause != nil {
			q.Set("pause", strconv.FormatBool(*cfg.Pause))
		}
		for _, c := range cfg.Changes {
			q.Add("changes", c)
		}

		req.URL.RawQuery = q.Encode()
		return nil
	}

	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return c.tr.Do(ctx, http.MethodPost, version.Join(ctx, "/commit"), withOptions)
	})
	if err != nil {
		return "", errors.Wrap(err, "error commiting container")
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "error reading response body")
	}

	var r containerCommitResponse
	if err := json.Unmarshal(b, &r); err != nil {
		return "", errors.Wrap(err, "error unmarshalling response")
	}
	return r.ID, nil
}
