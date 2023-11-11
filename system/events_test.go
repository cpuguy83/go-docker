package system

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/cpuguy83/go-docker/image"
	"github.com/cpuguy83/go-docker/testutils"
)

func TestEvents(t *testing.T) {
	t.Parallel()

	tr, _ := testutils.NewDefaultTestTransport(t, true)
	svc := NewService(tr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	f, err := svc.Events(ctx)
	if err != nil {
		t.Fatal(err)
	}

	cancel()
	if _, err := f(); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	const imgRef = "hello-world:latest"

	// TODO: It would be nice to not have to trigger a pull to get an event.
	remote, err := image.ParseRef(imgRef)
	if err != nil {
		t.Fatal(err)
	}

	f, err = svc.Events(ctx)
	if err != nil {
		t.Fatal(err)
	}

	beforePull := time.Now()

	if err := image.NewService(tr).Pull(ctx, remote); err != nil {
		cancel()
		t.Fatal(err)
	}
	afterPull := time.Now()

	// Even if the test was not run in parallel, we can't guarantee that the event will be the first one
	// returned, so we loop until we get the event we want.
	for {
		ev, err := f()
		if err != nil {
			t.Error(err)
			break
		}

		t.Log(ev)

		if ev.Type != "image" {
			continue
		}
		if ev.Action != "pull" {
			continue
		}

		if ev.Actor.ID != imgRef {
			continue
		}

		if !ev.Time.After(beforePull) {
			continue
		}

		if !ev.Time.Before(afterPull) {
			continue
		}
		break
	}

	cancel()

	ctxT, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	for {
		if ctxT.Err() != nil {
			t.Fatal("timeout waiting for event")
		}
		if _, err := f(); !errors.Is(err, context.Canceled) {
			continue
		}
		break
	}

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	f, err = svc.Events(ctx, WithEventsBetween(time.Now(), time.Now()))
	if err != nil {
		t.Fatal(err)
	}

	_, err = f()
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF, got: %v", err)
	}
}
