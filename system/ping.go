package system

import (
	"context"
)

type Ping struct {
	APIVersion     string
	OSType         string
	Experimental   bool
	BuilderVersion string
}

func (s *Service) Ping(ctx context.Context) (Ping, error) {
	resp, err := s.tr.Do(ctx, "GET", "/_ping")
	if err != nil {
		return Ping{}, err
	}
	defer resp.Body.Close()

	var p Ping
	p.APIVersion = resp.Header.Get("API-Version")
	p.OSType = resp.Header.Get("OSType")
	p.Experimental = resp.Header.Get("Docker-Experimental") == "true"
	p.BuilderVersion = resp.Header.Get("Builder-Version")

	return p, nil
}
