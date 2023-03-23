package buildkitopt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"testing"

	"github.com/cpuguy83/go-docker/transport"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
)

func TestDial(t *testing.T) {
	t.Parallel()
	tr, err := transport.DefaultTransport()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	c, err := client.New(ctx, "", FromDocker(tr)...)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	def, err := llb.Image("alpine:latest").Run(llb.Args([]string{"/bin/sh", "-c", "cat /etc/os-release"})).Marshal(ctx)
	if err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan *client.SolveStatus, 1)
	go func() {
		for status := range ch {
			for _, st := range status.Statuses {
				t.Log(st.Name, st.ID, st.Total, st.Completed)
			}
			for _, v := range status.Logs {
				t.Log(v.Timestamp, v.Vertex, string(v.Data))
			}
		}
	}()
	_, err = c.Solve(ctx, def, client.SolveOpt{SharedKey: hex.EncodeToString(key[:n])}, ch)
	if err != nil {
		t.Fatal(err)
	}

	<-ch
}
