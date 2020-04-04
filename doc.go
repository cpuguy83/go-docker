/*
package docker provides a client for accessing the Docker API.

*Usage*

All client operations are dependent on a transport. Transports are defined in the transport package. You can implement your own, here is the interface:

	// RequestOpt is as functional arguments to configure an HTTP request for a Doer.
	type RequestOpt func(*http.Request) error

	// Doer performs an http request for Client
	// It is the Doer's responsibility to deal with setting the host details on
	// the request
	// It is expected that one Doer connects to one Docker instance.
	type Doer interface {
		// Do typically performs a normal http request/response
		Do(ctx context.Context, method string, uri string, opts ...RequestOpt) (*http.Response, error)
		// DoRaw performs the request but passes along the response as a bi-directional stream
		DoRaw(ctx context.Context, method string, uri string, opts ...RequestOpt) (io.ReadWriteCloser, error)
	}

Do is the main function to implement, it takes an HTTP method, a uri (e.g. /containers/json), and a lits of options for configuring an *http.Request (e.g. to add request headers, query params, etc.)

DoRaw is used only for endpoints that need to “hijack” the http connection (ie. drop all HTTP semantics and drop to a raw, bi-directional stream). This is used for container attach.

The package contains a default transport that you can use directly, or wrap, as well as helpers for creating it from DOCKER_HOST style connection strings.

Once you have a transport you can create a client:

// create a transport that connects over /var/run/docker.sock
tr := transport.DefaultUnixTransport()
client := NewClient(WithTransport(tr))

Or if you don’t provide a transport, the default for the platform will be used.

Perform actions on a container:

	s := client.ContainerService()
	c, err := s.Create(ctx, container.WithCreateImage("busybox:latest"), container.WithCreateCmd("/bin/echo", "hello"))
	if err != nil {
			// handle error
	}

	cStdout, err := c.StdoutPipe(ctx)
	if err != nil {
			// handle error
	}
	defer cStdout.Close()

	if err := c.Start(ctx); err != nil {
			// handle error
	}

	io.Copy(os.Stdout, cStdout)

	if err := s.Remove(ctx, c.ID(), container.WithRemoveForce); err != nil {
			// handle error
	}

*/
package docker
