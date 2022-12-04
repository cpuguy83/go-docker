package image

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/cpuguy83/go-docker/errdefs"
	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/version"
)

// ExportConfig is the configuration for exporting an image.
type ExportConfig struct {
	Refs []string
}

// ExportOption is a functional option for configuring an image export.
type ExportOption func(*ExportConfig) error

// WithExportRefs adds the given image refs to the list of refs to export.
func WithExportRefs(refs ...string) ExportOption {
	return func(cfg *ExportConfig) error {
		cfg.Refs = append(cfg.Refs, refs...)
		return nil
	}
}

// Export exports an image(s) from the daemon.
// The returned reader is a tar archive of the exported image(s).
//
// Note: The way the docker API works, this will always return a reader.
// If there is an error it will be in that reader.
// TODO: Figure out how to deal with this case. For sure upstream moby should be fixed to return an error status code if there is an error.
// TODO: Right now the moby daemon writes the response header immediately before even validating any of the image refs.
func (s *Service) Export(ctx context.Context, opts ...ExportOption) (io.ReadCloser, error) {
	var cfg ExportConfig
	for _, o := range opts {
		if err := o(&cfg); err != nil {
			return nil, err
		}
	}

	if len(cfg.Refs) == 0 {
		return nil, errdefs.Invalid("no refs provided")
	}

	withNames := func(req *http.Request) error {
		q := req.URL.Query()
		for _, ref := range cfg.Refs {
			q.Add("names", ref)
		}
		req.URL.RawQuery = q.Encode()
		return nil
	}

	ctx = httputil.WithResponseLimitIfEmpty(ctx, httputil.UnlimitedResponseLimit)
	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return s.tr.Do(ctx, http.MethodGet, version.Join(ctx, "/images/get"), withNames)
	})
	if err != nil {
		return nil, err
	}

	if resp.Header.Get("Content-Type") != "application/x-tar" {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected content type: %s", resp.Header.Get("Content-Type"))
	}
	return resp.Body, nil
}
