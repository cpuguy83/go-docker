package transport

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
)

// DockerCLIConnectionConfig is the configuration for the DockerCLIConnectionOption
type DockerCLIConnectionConfig struct {
	// Env is the environment to use for the docker CLI
	// Useful for setting DOCKER_HOST or DOCKER_CONTEXT to specify the docker daemon that the docker cli should connect to
	// Defaults to os.Environ()
	Env []string
	// StderrPipe is the writer to use for the stderr of the docker CLI
	// This is the only means of getting error messages from the CLI.
	//
	// Deprecated: This is only here until validation is done on the cli connection.
	// Right now this is the only way to get failure details from the CLI.
	StderrPipe io.Writer
}

// DockerCLIConnectionOption is an option for the FromDockerCLI function
type DockerCLIConnectionOption func(*DockerCLIConnectionConfig) error

// FromDockerCLI creates a Transport from the docker CLI
// In this case, the docker CLI acts as a proxy to the docker daemon.
// Any protocol your CLI supports, this transport will support.
func FromDockerCLI(opts ...DockerCLIConnectionOption) *Transport {
	dial := func(ctx context.Context, _, _ string) (net.Conn, error) {
		cfg := DockerCLIConnectionConfig{
			Env: os.Environ(),
		}
		for _, o := range opts {
			if err := o(&cfg); err != nil {
				return nil, err
			}
		}

		cmd := exec.CommandContext(ctx, "docker", "system", "dial-stdio")
		cmd.Env = cfg.Env

		c1, c2 := net.Pipe()
		cmd.Stdin = c1
		cmd.Stdout = c1
		cmd.Stderr = cfg.StderrPipe

		if err := cmd.Start(); err != nil {
			c1.Close()
			c2.Close()
			return nil, fmt.Errorf("failed to start docker dial-stdio: %w", err)
		}
		go func() {
			cmd.Wait()
			c1.Close()
			c2.Close()
		}()

		// TODO: Validate that we can actually handshake with the server

		return c2, nil
	}

	tr := &http.Transport{
		DisableCompression: true,
		DialContext:        dial,
	}

	return &Transport{
		scheme: "http",
		c: &http.Client{
			Transport: tr,
		},
		host: ".",
		dial: func(ctx context.Context) (net.Conn, error) {
			return tr.DialContext(ctx, "", "")
		},
	}
}
