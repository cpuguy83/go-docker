package container

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/cpuguy83/go-docker/container/containerapi"
	"github.com/cpuguy83/go-docker/errdefs"
	"github.com/cpuguy83/go-docker/httputil"

	"github.com/cpuguy83/go-docker/version"
)

// ListFilter represents filters to process on the container list.
type ListFilter struct {
	Ancestor  []string `json:"ancestor,omitempty"`
	Before    []string `json:"before,omitempty"`
	Expose    []string `json:"expose,omitempty"`
	Exited    []string `json:"exited,omitempty"`
	Health    []string `json:"health,omitempty"`
	ID        []string `json:"id,omitempty"`
	Isolation []string `json:"isolation,omitempty"`
	IsTask    []string `json:"is-task,omitempty"`
	Label     []string `json:"label,omitempty"`
	Name      []string `json:"name,omitempty"`
	Network   []string `json:"network,omitempty"`
	Publish   []string `json:"publish,omitempty"`
	Since     []string `json:"since,omitempty"`
	Status    []string `json:"status,omitempty"`
	Volume    []string `json:"volume,omitempty"`
}

// ListConfig holds the options for listing containers
type ListConfig struct {
	All    bool
	Limit  int
	Size   bool
	Filter ListFilter
}

// ListOption is used as functional arguments to list containers
// ListOption configure an InspectConfig.
type ListOption func(config *ListConfig)

// List fetches a list of containers.
func (s *Service) List(ctx context.Context, opts ...ListOption) ([]containerapi.Container, error) {
	cfg := ListConfig{
		Limit: -1,
	}
	for _, o := range opts {
		o(&cfg)
	}

	withListConfig := func(req *http.Request) error {
		q := req.URL.Query()
		q.Add("all", strconv.FormatBool(cfg.All))
		q.Add("limit", strconv.Itoa(cfg.Limit))
		q.Add("size", strconv.FormatBool(cfg.Size))
		filterJSON, err := json.Marshal(cfg.Filter)

		if err != nil {
			return err
		}
		q.Add("filters", string(filterJSON))

		req.URL.RawQuery = q.Encode()
		return nil
	}

	var containers []containerapi.Container

	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return s.tr.Do(ctx, http.MethodGet, version.Join(ctx, "/containers/json"), withListConfig)
	})
	if err != nil {
		return containers, err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return containers, nil
	}

	if err := json.Unmarshal(data, &containers); err != nil {
		return containers, errdefs.Wrap(err, "error unmarshalling container json")
	}

	return containers, nil
}
