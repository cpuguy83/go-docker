package version

import (
	"context"
	"os"
	"path"
)

type apiVersion struct{}

// WithAPIVersion stores the API version to make a request with in the provided context
func WithAPIVersion(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, apiVersion{}, version)
}

// APIVersion gets the API version from the passed in context
// If no version is set then an empty string will be returned.
func APIVersion(ctx context.Context) string {
	v := ctx.Value(apiVersion{})
	if v == nil {
		return ""
	}
	return v.(string)
}

// Join adds the API version stored in the context to the provided uri
func Join(ctx context.Context, uri string) string {
	v := APIVersion(ctx)
	if len(v) > 0 && v[0] != 'v' {
		v = "v" + v
	}
	return path.Join("/", v, uri)
}

// FromEnv sets the API version to use from the DOCKER_API_VERSION environment variable
// This is like how the Docker CLI sets a specific API version.
func FromEnv(ctx context.Context) context.Context {
	return WithAPIVersion(ctx, os.Getenv("DOCKER_API_VERSION"))
}

// Negotiate looks at the version currently in ctx and the verion of the server.
// It returns a context with the best api version to use.
func Negotiate(ctx context.Context, srv string) context.Context {
	if srv == "" {
		srv = "v1.12"
	}

	v := APIVersion(ctx)
	if v == "" {
		v = srv
	}

	if LessThan(srv, v) {
		return WithAPIVersion(ctx, srv)
	}
	return WithAPIVersion(ctx, v)
}
