package image

import (
	"net/url"
	"path"
	"strings"

	"github.com/cpuguy83/go-docker/errdefs"
)

const (
	dockerDomain = "docker.io"
	legacyDomain = "index.docker.io"
)

// Remote represents are remote repository reference.
// Tag may include a tag or digest.
type Remote struct {
	// Host is the registry hostname.
	// Host may be empty, which will cause the daemon to use the default registry (e.g. docker.io).
	Host string
	// Locator is the repository name, without the host.
	Locator string
	// Tag is the tag or digest of the image.
	Tag string
}

func (r Remote) String() string {
	s := path.Join(r.Host, r.Locator)
	if r.Tag != "" {
		if strings.Contains(r.Tag, ":") {
			// Should be a digest
			s = s + "@" + r.Tag
		} else if r.Tag != "" {
			s += ":" + r.Tag
		}
	}

	return s
}

// ParseRef takes an image ref and parses it into a `Remote` struct.
// This can handle the non-canonicalized docker reference format.
func ParseRef(ref string) (Remote, error) {
	var r Remote
	if ref == "" {
		return r, errdefs.Invalid("invalid reference: " + ref)
	}
	u, err := url.Parse("dummy://" + ref)
	if err != nil {
		tagIdx := strings.LastIndex(ref, ":")
		if tagIdx > strings.LastIndex(ref, "/") {
			r.Tag = ref[tagIdx+1:]
			ref = ref[:tagIdx]
		}
		var err2 error
		u, err2 = url.Parse("dummy://" + ref)
		if err2 != nil {
			return Remote{}, errdefs.AsInvalid(err)
		}
	}

	if u.Scheme != "dummy" {
		// Something is very wrong if this happened
		return r, errdefs.Invalid("invalid reference: " + ref)
	}

	switch {
	case u.Path == "":
		r.Host, r.Locator = splitDockerDomain(u.Host)
	case u.Host == "":
		r.Host, r.Locator = splitDockerDomain(u.Path)
	default:
		r.Host, r.Locator = splitDockerDomain(ref)
	}

	l, t, ok := strings.Cut(r.Locator, ":")
	if ok {
		r.Locator = l
		r.Tag = t
	}
	if r.Tag == "" {
		r.Tag = "latest"
	}
	return r, nil
}

// splitDockerDomain splits a repository name to domain and remotename string.
// If no valid domain is found, the default domain is used.
func splitDockerDomain(name string) (host, locator string) {
	i := strings.IndexRune(name, '/')
	var domain string
	if i == -1 || (!strings.ContainsAny(name[:i], ".:") && name[:i] != "localhost") {
		return dockerDomain, name
	} else {
		domain = name[:i]
		locator = name[i+1:]
	}
	if domain == legacyDomain {
		domain = dockerDomain
	}
	return domain, locator
}

func resolveRegistryHost(host string) string {
	switch host {
	case "index.docker.io", "docker.io", "https://index.docker.io/v1/", "registry-1.docker.io":
		return "https://index.docker.io/v1/"
	}
	return host
}
