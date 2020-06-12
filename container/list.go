package container

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"

	"github.com/cpuguy83/go-docker/common/filters"
	"github.com/cpuguy83/go-docker/container/containerapi"
	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/version"
)

type ListConfig struct {
	Size    bool
	All     bool
	Since   string
	Before  string
	Limit   int
	Filters filters.Args
}

type ListOption func(config *ListConfig)

// List returns the list of containers in the docker host.
func (s *Service) List(ctx context.Context, opts ...ListOption) ([]containerapi.ContainerSummary, error) {
	c := ListConfig{
		Limit: -1,
	}
	for _, o := range opts {
		o(&c)
	}

	var (
		containers []containerapi.ContainerSummary
		query      = url.Values{}
	)

	if c.All {
		query.Set("all", "1")
	}
	if c.Limit != -1 {
		query.Set("limit", strconv.Itoa(c.Limit))
	}
	if len(c.Since) > 0 {
		query.Set("since", c.Since)
	}
	if len(c.Before) > 0 {
		query.Set("before", c.Before)
	}
	if c.Size {
		query.Set("size", "1")
	}
	if c.Filters.Len() > 0 {
		// TODO: if version is less than 1.22 then the encoded format is different
		filterJSON, err := filters.ToJSON(c.Filters)
		if err != nil {
			return nil, err
		}
		query.Set("filters", filterJSON)
	}

	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return s.tr.Do(ctx, http.MethodGet, version.Join(ctx, "/containers/json"), func(req *http.Request) error {
			req.URL.RawQuery = query.Encode()
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&containers)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling containers json")
	}

	return containers, nil
}
