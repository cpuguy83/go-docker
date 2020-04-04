package system

import (
	"context"
	"net/http"

	"github.com/cpuguy83/go-docker/httputil"
)

type Ping struct {
	APIVersion     string
	OSType         string
	Experimental   bool
	BuilderVersion string
}

func (s *Service) Ping(ctx context.Context) (Ping, error) {
	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return s.tr.Do(ctx, "GET", "/_ping")
	})
	var p Ping
	if resp != nil {
		defer resp.Body.Close()

		p.APIVersion = resp.Header.Get("API-Version")
		p.OSType = resp.Header.Get("OSType")
		p.Experimental = resp.Header.Get("Docker-Experimental") == "true"
		p.BuilderVersion = resp.Header.Get("Builder-Version")
	}

	// We are intentionally returning a populated ping response even if there is an error
	//  since this data may have been returned by the API.
	return p, err
}
