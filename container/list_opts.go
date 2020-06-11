package container

import (
	"github.com/cpuguy83/go-docker/common/filters"
)

func WithListAll(cfg *ListConfig) {
	cfg.All = true
}

func WithListSize(cfg *ListConfig) {
	cfg.Size = true
}

func WithListSince(since string) ListOption {
	return func(config *ListConfig) {
		config.Since = since
	}
}

func WithListBefore(before string) ListOption {
	return func(config *ListConfig) {
		config.Before = before
	}
}

func WithListLimit(limit int) ListOption {
	return func(config *ListConfig) {
		config.Limit = limit
	}
}

func WithListFilters(filters filters.Args) ListOption {
	return func(config *ListConfig) {
		config.Filters = filters
	}
}
