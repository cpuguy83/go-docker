package image

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

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
	// ConsumeProgress is called after a pull response is received to consume the progress messages from the response body.
	// ConSumeProgress should not return until EOF is reached on the passed in stream or it may cause the pull to be cancelled.
	// If this is not set, progress messages are discarded.
	ConsumeProgress StreamConsumer
}

func WithPullPlatform(platform string) PullOption {
	return func(cfg *PullConfig) error {
		cfg.Platform = platform
		return nil
	}
}

// StreamConsumer is a function that consumes a stream of data, typically a stream of messages.
type StreamConsumer func(context.Context, io.Reader) error

// PullProgressDecoderV1 is a decoder for the v1 progress message.
// PullProgressMessageV1 represents a message received from the Docker daemon during a pull operation.
type PullProgressMessage struct {
	Status   string `json:"status,omitempty"`
	Progress string `json:"progress,omitempty"`
	ID       string `json:"id"`
	Detail   struct {
		Current int64 `json:"current"`
		Total   int64 `json:"total"`
	} `json:"progressDetail,omitempty"`
}

// PullProgressMessageHandler is used with `WithPullProgressMessage` to handle progress messages.
type PullProgressMessageHandler func(context.Context, PullProgressMessage) error

// PullProgressDigest wraps the past in handler with a progress callback suitable for `WithPullProgressMessage`
// The past in callback is called when the digest of a pulled image is received.
func PullProgressDigest(h func(ctx context.Context, digest string) error) PullProgressMessageHandler {
	return func(ctx context.Context, msg PullProgressMessage) error {
		_, right, ok := strings.Cut(msg.Status, "Digest:")
		if !ok {
			return nil
		}
		return h(ctx, strings.TrimSpace(right))
	}
}

// PullProgressHandlers makes a PullProgressMessageHandler from a list of PullProgressMessageHandlers.
// Handlers are executed in the order they are passed in.
// An error in a handler will stop execution of the remaining handlers.
func PullProgressHandlers(handlers ...PullProgressMessageHandler) PullProgressMessageHandler {
	return func(ctx context.Context, msg PullProgressMessage) error {
		for _, h := range handlers {
			if err := h(ctx, msg); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithPullProgressMessage returns a PullOption that sets a pull progress consumer.
// The passed in callback will be called for each progress message.
func WithPullProgressMessage(h PullProgressMessageHandler) PullOption {
	return func(cfg *PullConfig) error {
		cfg.ConsumeProgress = func(ctx context.Context, r io.Reader) error {
			dec := json.NewDecoder(r)
			msg := &PullProgressMessage{}
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				if err := dec.Decode(msg); err != nil {
					if err == io.EOF {
						return nil
					}
					return err
				}

				if err := h(ctx, *msg); err != nil {
					return err
				}
				*msg = PullProgressMessage{}
			}
		}
		return nil
	}
}

// PullOption is a function that can be used to modify the pull config.
// It is used during `Pull` as functional arguments.
type PullOption func(config *PullConfig) error

// Pull pulls an image from a remote registry.
// It is up to the caller to set a response size limit as the normal default limit is not used in this case.
func (s *Service) Pull(ctx context.Context, remote Remote, opts ...PullOption) error {
	cfg := PullConfig{}

	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return err
		}
	}

	if cfg.ConsumeProgress == nil {
		cfg.ConsumeProgress = func(ctx context.Context, r io.Reader) error {
			_, err := io.Copy(ioutil.Discard, r)
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

	// Set unlimited response size since this is going to be consumed by a progress reader.
	// It's also pretty important to read the full body.
	ctx = httputil.WithResponseLimitIfEmpty(ctx, httputil.UnlimitedResponseLimit)
	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return s.tr.Do(ctx, http.MethodPost, version.Join(ctx, "/images/create"), withConfig)
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return cfg.ConsumeProgress(ctx, resp.Body)
}

type authConfig struct {
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	IdentityToken string `json:"identitytoken,omitempty"`
}
