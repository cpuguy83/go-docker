package image

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/version"
)

// ExportBundleConfig holds the options for exporting images.
type ExportBundleConfig struct {
	Names []string
}

// ExportBundleOption is used as functional arguments to export images.
// ExportBundleOption configure a ExportBundleConfig.
type ExportBundleOption func(config *ExportBundleConfig)

func (s *Service) ExportBundle(ctx context.Context, opts ...ExportBundleOption) ([]byte, error) {
	cfg := ExportBundleConfig{}
	for _, o := range opts {
		o(&cfg)
	}

	withExportBundleConfig := func(req *http.Request) error {
		q := req.URL.Query()
		for _, name := range cfg.Names {
			q.Add("names", name)
		}
		req.URL.RawQuery = q.Encode()
		return nil
	}

	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return s.tr.Do(ctx, http.MethodGet, version.Join(ctx, "/images/get"), withExportBundleConfig)
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
