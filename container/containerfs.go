package container

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cpuguy83/go-docker/errdefs"
	"github.com/cpuguy83/go-docker/httputil"
	"github.com/cpuguy83/go-docker/version"
)

type Stat struct {
	Name       string      `json:"name"`
	Size       int64       `json:"size"`
	Mode       os.FileMode `json:"mode"`
	Mtime      time.Time   `json:"mtime"`
	LinkTarget string      `json:"linkTarget"`
}

func (s *Stat) Reset() {
	*s = Stat{}
}

func (c *Container) archivePath() string {
	return "/containers/" + c.ID() + "/archive"
}

// Stat retrieves the stat information for a file or directory in the
// container's filesystem.
func (c *Container) Stat(ctx context.Context, p string, stat *Stat) error {
	pathOpt := func(r *http.Request) error {
		q := r.URL.Query()
		q.Set("path", filepath.ToSlash(p))
		r.URL.RawQuery = q.Encode()
		return nil
	}

	resp, err := httputil.DoRequest(ctx, func(ctx context.Context) (*http.Response, error) {
		return c.tr.Do(ctx, http.MethodHead, version.Join(ctx, c.archivePath()), pathOpt)
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return decodeStatHeader(resp.Header, stat)
}

type DownloadConfig struct {
	// If provided Stat will be populated with the file's metadata.
	// This can be useful for avoiding an extra Stat call tog et information
	// about the content being downloaded since this is returned in the
	// response headers when downloading.
	Stat *Stat
}

type DownloadOption func(*DownloadConfig)

func decodeStatHeader(hdr http.Header, stat *Stat) error {
	const hdrKey = "X-Docker-Container-Path-Stat"
	v := hdr.Get(hdrKey)

	if v == "" {
		return nil // no stat header, nothing to decode
	}

	dt, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return errdefs.Wrap(err, "error decoding container stat header from base64")
	}

	if err := json.Unmarshal(dt, &stat); err != nil {
		return errdefs.Wrap(err, "error unmarshaling container stat")
	}
	return nil
}

// Download retrieves a file or directory from the container's filesystem.
// The returned reader is a tar archive.
func (c *Container) Download(ctx context.Context, p string, opts ...DownloadOption) (io.ReadCloser, error) {
	var cfg DownloadConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	pathOpt := func(r *http.Request) error {
		q := r.URL.Query()
		q.Set("path", filepath.ToSlash(p))
		r.URL.RawQuery = q.Encode()
		return nil
	}

	resp, err := c.tr.Do(ctx, http.MethodGet, version.Join(ctx, c.archivePath()), pathOpt)
	if err != nil {
		return nil, err
	}

	if err := httputil.CheckResponseError(resp); err != nil {
		resp.Body.Close()
		return nil, err
	}

	if cfg.Stat != nil {
		if err := decodeStatHeader(resp.Header, cfg.Stat); err != nil {
			resp.Body.Close()
			return nil, err
		}
	}

	return resp.Body, nil
}

type UploadConfig struct {
	// Overwrite forces the upload to overwrite existing files, particularly when
	// the file type is different (e.g. write a file over where a directory exists).
	Overwrite bool

	// CopyUIDGID indicates whether to copy the UID and GID from the source file
	CopyUIDGID bool
}

type UploadOption func(*UploadConfig)

// Upload uploads a file or directory to the container's filesystem.
// The reader passed in must be a tar archive containing the files to upload.
func (c *Container) Upload(ctx context.Context, p string, r io.Reader, opts ...UploadOption) error {
	var cfg UploadConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	withOpts := func(r *http.Request) error {
		q := r.URL.Query()
		if !cfg.Overwrite {
			q.Set("noOverwriteDirNonDir", "false")
		}

		if cfg.CopyUIDGID {
			q.Set("copyUIDGID", "true")
		}

		return nil
	}

	pathOpt := func(r *http.Request) error {
		q := r.URL.Query()
		q.Set("path", filepath.ToSlash(p))
		r.URL.RawQuery = q.Encode()
		return nil
	}

	withReader := func(req *http.Request) error {
		req.Body = io.NopCloser(r)
		return nil
	}

	resp, err := c.tr.Do(ctx, http.MethodPut, version.Join(ctx, c.archivePath()), pathOpt, withReader, withOpts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := httputil.CheckResponseError(resp); err != nil {
		return err
	}
	return nil
}
