package docker

import "context"

type apiVersionKey struct{}

func WithAPIVersion(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, apiVersionKey{}, v)
}

func getAPIVersion(ctx context.Context, defaultVersion string) string {
	v := ctx.Value(apiVersionKey{})
	if v == nil {
		return defaultVersion
	}
	return v.(string)
}
