package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/version"
)

// LoginConfig is the configuration for logging into a registry.
type LoginConfig struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Auth     string `json:"auth,omitempty"`

	ServerAddress string `json:"serveraddress,omitempty"`

	// IdentityToken is used to authenticate the user and get
	// an access token for the registry.
	IdentityToken string `json:"identitytoken,omitempty"`

	// RegistryToken is a bearer token to be sent to a registry
	RegistryToken string `json:"registrytoken,omitempty"`
}

type loginResponse struct {
	// An opaque token used to authenticate a user after a successful login
	// Required: true
	IdentityToken string `json:"IdentityToken"`
}

// LoginOption is a function that can be used to modify the login config.
type LoginOption func(config *LoginConfig) error

// Login logs in to a registry with the given credentials.
// Login may return an access token for the registry.
func (s *Service) Login(ctx context.Context, opts ...LoginOption) (string, error) {
	cfg := LoginConfig{}
	for _, o := range opts {
		if err := o(&cfg); err != nil {
			return "", err
		}
	}

	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return s.tr.Do(ctx, http.MethodPost, version.Join(ctx, "/auth"), httputil.WithJSONBody(&cfg))
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	var lr loginResponse
	if err := json.Unmarshal(data, &lr); err != nil {
		return "", fmt.Errorf("error unmarshaling response: %w", err)
	}

	return lr.IdentityToken, nil
}
