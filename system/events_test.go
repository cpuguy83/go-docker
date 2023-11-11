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

	imgRef := "hello-world:latest"

	f, err = svc.Events(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: It would be nice to not have to trigger a pull to get an event.
	remote, err := image.ParseRef(imgRef)
	if err != nil {
		t.Fatal(err)
	}
	if err := image.NewService(tr).Pull(ctx, remote); err != nil {
		t.Fatal(err)
	}

	ev, err := f()
	if err != nil {
		t.Fatal(err)
	}

	if ev.Type != "image" {
		t.Fatalf("expected type to be image, got %s", ev.Type)
	}
	if ev.Action != "pull" {
		t.Fatalf("expected action to be image pull, got %s", ev.Action)
	}
	if ev.Actor.ID != imgRef {
		t.Fatalf("expected actor to be %s, got %s", imgRef, ev.Actor.ID)
	}

	cancel()
	if _, err := f(); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
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
