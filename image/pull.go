package image

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/version"
)

// PullConfig is the configuration for pulling an image.
type PullConfig struct {
	// When supplied, will be used to retrieve credentials for the given domain.
	// If you want this to work with docker CLI auth you can use https://github.com/cpuguy83/dockercfg/blob/3ee4ae1349920b54391faf7d2cff00385a3c6a39/auth.go#L23-L32
	CredsFunction func(string) (string, string, error)
	// Platform is the platform spec to use when pulling the image.
	// Example: "linux/amd64", "linux/arm64", "linux/arm/v7", etc.
	// This is used to filter a manifest list to find the right image.
	//
	// If not set, the default platform for the daemon will be used.
	Platform string
	// NewProgressDecoder is used to decode the pull progress.
	// The created decoder will be called until it returns io.EOF or some other error.
	// Any error other than io.EOF will cause the Pull function to return with that error.
	// If nil, the progress will be discarded.
	NewProgressDecoder func(rdr io.Reader) Decoder
}

// Decoder is used to decode the pull a stream of messages.
type Decoder interface {
	Decode(context.Context) error
}

// PullProgressDecoderV1 is a decoder for the v1 progress message.
type PullProgressDecoderV1 struct {
	dec    *json.Decoder
	msg    *PullProgressMessageV1
	notify func(context.Context, *PullProgressMessageV1) error
	err    error
}

// PullProgressMessageV1 represents a message received from the Docker daemon during a pull operation.
type PullProgressMessageV1 struct {
	Status   string `json:"status"`
	Detail   string `json:"progressDetail"`
	ID       string `json:"id"`
	Progress struct {
		Current int64 `json:"current"`
		Total   int64 `json:"total"`
	} `json:"progress"`
}

// Decode decodes the message and calls the notify function with the message.
func (d *PullProgressDecoderV1) Decode(ctx context.Context) error {
	if d.err != nil {
		return d.err
	}
	if d.msg == nil {
		d.msg = &PullProgressMessageV1{}
	} else {
		*d.msg = PullProgressMessageV1{}
	}

	if err := d.dec.Decode(d.msg); err != nil {
		d.err = err
		return err
	}

	d.notify(ctx, d.msg)
	return nil
}

// WithPullProgressV1Decoder returns a PullOption that sets a decoder to decode using the PullProgressV1Decoder.
// notifier must be set otherwise a panic will occur during pull.
func WithPullProgressV1Decoder(notifier func(context.Context, *PullProgressMessageV1) error) PullOption {
	return func(cfg *PullConfig) error {
		cfg.NewProgressDecoder = func(rdr io.Reader) Decoder {
			return &PullProgressDecoderV1{
				notify: notifier,
				dec:    json.NewDecoder(rdr),
			}
		}
		return nil
	}
}

// PullOption is a function that can be used to modify the pull config.
// It is used during `Pull` as functional arguments.
type PullOption func(config *PullConfig) error

// Pull pulls an image from a remote registry.
func (s *Service) Pull(ctx context.Context, remote Remote, opts ...PullOption) error {
	cfg := PullConfig{}

	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return err
		}
	}

	withConfig := func(req *http.Request) error {
		q := req.URL.Query()
		if cfg.Platform != "" {
			q.Set("platform", cfg.Platform)
		}
		q.Set("tag", remote.Tag)
		from := remote.Locator
		if remote.Host != "" && remote.Host != dockerDomain {
			from = remote.Host + "/" + remote.Locator
		}
		q.Set("fromImage", from)
		req.URL.RawQuery = q.Encode()

		if cfg.CredsFunction != nil {
			username, password, err := cfg.CredsFunction(resolveRegistryHost(remote.Host))
			if err != nil {
				return err
			}
			var ac authConfig
			if username == "" || username == "<token>" {
				ac.IdentityToken = password
			} else {
				ac.Username = username
				ac.Password = password
			}

			auth, err := json.Marshal(&ac)
			if err != nil {
				return err
			}
			req.Header = map[string][]string{}
			req.Header.Set("X-Registry-Auth", base64.URLEncoding.EncodeToString(auth))
		}

		return nil
	}

	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return s.tr.Do(ctx, http.MethodPost, version.Join(ctx, "/images/create"), withConfig)
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if cfg.NewProgressDecoder == nil {
		io.Copy(ioutil.Discard, resp.Body)
		return nil
	}

	dec := cfg.NewProgressDecoder(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := dec.Decode(ctx); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

type authConfig struct {
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	IdentityToken string `json:"identitytoken,omitempty"`
}
