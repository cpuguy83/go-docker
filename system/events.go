package system

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cpuguy83/go-docker/version"
)

type eventAPI struct {
	Type     string     `json:"type"`
	Action   string     `json:"action"`
	Scope    string     `json:"scope,omitempty"`
	TimeNano int64      `json:"timeNano"`
	Actor    EventActor `json:"actor"`
}

// Event represents a docker event.
type Event struct {
	Type   string
	Action string
	Scope  string
	Time   time.Time
	Actor  EventActor
}

// EventActor represents the actor of a docker event with attributes specific to that actor.
type EventActor struct {
	ID         string            `json:"id,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// EventConfig is used to configure the event stream.
type EventConfig struct {
	Since        *time.Time
	Until        *time.Time
	FieldFilters FieldFilter
}

type FieldFilter struct {
	fields map[string]map[string]bool
}

// WithEventsBetween is an EventOption that sets the time range for the event stream.
func WithEventsBetween(since, until time.Time) EventOption {
	return func(cfg *EventConfig) {
		cfg.Since = &since
		cfg.Until = &until
	}
}

func (f *FieldFilter) Add(key, value string) {
	if f.fields == nil {
		f.fields = make(map[string]map[string]bool)
	}
	if _, ok := f.fields[key]; !ok {
		f.fields[key] = make(map[string]bool)
	}
	f.fields[key][value] = true
}

// EventOption is a function that can be passed to Events to configure the event stream.
type EventOption func(*EventConfig)

// WithFilters is an EventOption that adds filters to the event stream.
// These filters are passed to the docker daemon and are used to filter the events returned.
func WithEventFilters(f FieldFilter) EventOption {
	return func(cfg *EventConfig) {
		cfg.FieldFilters = f
	}
}

// WithAddEventFilter is an EventOption that adds a filter to the event stream.
// If the key already exists in the filters list, the new value is appended to the list of values for that key.
func WithAddEventFilter(key, value string) EventOption {
	return func(cfg *EventConfig) {
		cfg.FieldFilters.Add(key, value)
	}
}

// Events returns a function that can be called to get the next event.
// The function will block until an event is available.
// The returned event is only valid until the next call to the function.
//
// Canceling the context will stop the event stream.
// Once cancelled the next call to the returned function will return the context error.
func (s *Service) Events(ctx context.Context, opts ...EventOption) (func() (*Event, error), error) {
	var cfg EventConfig
	for _, o := range opts {
		o(&cfg)
	}

	resp, err := s.tr.Do(ctx, http.MethodGet, version.Join(ctx, "/events"), func(req *http.Request) error {
		q := url.Values{}
		if len(cfg.FieldFilters.fields) > 0 {
			dt, err := json.Marshal(cfg.FieldFilters.fields)
			if err != nil {
				return err
			}
			q.Set("filters", string(dt))
		}
		if cfg.Since != nil {
			q.Set("since", strconv.FormatInt(cfg.Since.Unix(), 10))
		}

		if cfg.Until != nil {
			q.Set("until", strconv.FormatInt(cfg.Until.Unix(), 10))
		}
		if len(q) > 0 {
			req.URL.RawQuery = q.Encode()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(resp.Body)
	evAPI := &eventAPI{}
	ev := &Event{}

	return func() (*Event, error) {
		*evAPI = eventAPI{}
		if err := dec.Decode(evAPI); err != nil {
			return nil, err
		}

		ev.Type = evAPI.Type
		ev.Action = evAPI.Action
		ev.Scope = evAPI.Scope
		ev.Time = time.Unix(0, evAPI.TimeNano)
		ev.Actor = evAPI.Actor
		return ev, nil
	}, nil
}
