# An experimental Docker client

This is an experimental client for Docker. Most of the low-level details are finalized and just requires implementing
the API calls on top of that.

The goals of this project are:

1. Abstract away connection semantics - the official client is kind of broken in this regard, where abstractions for
connection and transport protocol are co-mingled. This makes it difficult to, for instance, take advantage of HTTP/2
in the official client, or implement things like SSH forwarding (this exists in the official client, but not with great
effort).
2. Semantic versioning - The official client is built such that it is very difficult to not make breaking changes when
adding new things, therefore it is difficult to version semantically (without being at some insanely high and meaningless
version).
3. Easy to consume in go.mod - Traditionally the official client is... painful... to import, particularly with go modules.
This is caused by a number of project level issues, such as (lack of) version tagging. We are trying to fix this upstream
but we have to make certain not to break too much or introduce things that are not maintainable.
4. Do not import from the upstream types -- the upstream repo has a lot of history, things move around, have lots of 
transitive dependencies, and in general is just slow to import due to that history. Instead we define the required types
directly in this repo, even if it's only as a copy of the existing ones.
5. Easy to reason about errors - You should be able to know exactly what kind of error was returned without having to
muck around with the http response
6. Integrate with your choice tracing/metrics frameworks

## Usage

All client operations are dependent on a transport. Transports are defined in the transport package. You can implement
your own, here is the interface:

```go
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
```

`Do` is the main function to implement, it takes an HTTP method, a uri (e.g. `/containers/json`), and a lits of options
for configuring an `*http.Request` (e.g. to add request headers, query params, etc.)

`DoRaw` is used only for endpoints that need to "hijack" the http connection (ie. drop all HTTP semantics and drop to a
raw, bi-directional stream). This is used for container attach.

The package contains a default transport that you can use directly, or wrap, as well as helpers for creating it from
`DOCKER_HOST` style connection strings.

Once you have a transport you can create a client:

```go
// create a transport that connects over /var/run/docker.sock
tr := transport.DefaultUnixTransport()
client := NewClient(WithTransport(tr))
```

Or if you don't provide a transport, the default for the platform will be used.

Perform actions on a container:

```go
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
```
