package cewrap

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"log/slog"

	cloudevents "github.com/cloudevents/sdk-go/v2"

)

type EventData struct {
	Method       string `json:"method"`
	Resource     string `json:"resource"`
	ResourceData any    `json:"resource_data"`
}

type Source struct {
	// The downstream service.
	Downstream *url.URL
	// sink is the url that sinks the events
	Sink cloudevents.Client
	// HTTP client for sending the downstream requests.
	Client *http.Client
	// Methods that indicate a change and will generate an event.
	ChangeMethods []string

	// Event source.
	Source string
	// Type prefix for the event type field.
	TypePrefix string
	// Dataschema for the event.
	Dataschema string

	Logger slog.Logger
}

var DefaultChangeMethods = []string{
	http.MethodPost,
	http.MethodDelete,
	http.MethodPatch,
	http.MethodPut,
}

func (s *Source) isChange(method string) bool {
	for _, m := range s.ChangeMethods {
		if method == m {
			return true
		}
	}
	return false
}

func NewSource(downstream, sink string, client *http.Client, changeMethods []string) *Source {
	s := &Source{}
	if client == nil {
		s.Client = http.DefaultClient
	}
	var err error
	s.Downstream, err = url.Parse(downstream)
	if err != nil {
		return nil
	}
	s.Sink, err = cloudevents.NewClientHTTP(cloudevents.WithTarget(sink))
	if err != nil {
		return nil
	}
	s.ChangeMethods = changeMethods
	if len(s.ChangeMethods) == 0 {
		s.ChangeMethods = DefaultChangeMethods
	}
	return s
}

func (s *Source) buildDownstreamRequest(ctx context.Context, r *http.Request) (*http.Request, error) {
	// Get the body.
	body := bytes.Buffer{}
	if _, err := io.Copy(&body, r.Body); err != nil {
		return nil, err
	}

	// Create the request.
	req, err := http.NewRequestWithContext(ctx, r.Method, s.Downstream.String(), &body)
	if err != nil {
		return nil, err
	}

	// header filter
	hf := func(h string) bool {
		switch strings.ToLower(h) {
		case "content-length":
			return false
		default:
			return true
		}
	}

	// Copy the headers.
	for k, h := range r.Header {
		if hf(k) {
			for _, hh := range h {
				req.Header.Add(k, hh)
			}
		}
	}

	return req, nil
}

// Handle handles the requests.
//
// It passes the request to the downstream service and generates a cloud event
// and sends it to the sink.
func (s *Source) Handle(w http.ResponseWriter, r *http.Request) {
	// Build a client request from the server request.
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	cr, err := s.buildDownstreamRequest(ctx, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Call the downstream service.
	resp, err := s.Client.Do(cr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Create the response and write it out to the responseWriter.
	err = s.writeResponse(w, resp)
	if err != nil {
		// Write an internal error.
	}

	if resp.StatusCode < 200 && resp.StatusCode >= 300 {
		// We are done.
		return
	}
	if !s.isChange(r.Method) {
		return
	}
	// Only run when we have an events sink.
	if s.Sink == nil {
		return
	}
	evt := cloudevents.NewEvent()
	evt.SetSource("bla.source")

	us := &url.URL{}
	us.Scheme = r.URL.Scheme
	us.Host = r.Host
	us.Path = r.URL.Path
	us.Scheme = "http"

	sub := us.String()
	evt.SetSubject(sub)
	evt.SetType("my.type")
	evt.SetData("application/json",
		EventData{
			Method:   r.Method,
			Resource: sub,
			ResourceData: map[string]any{
				"name": "Peter",
			},
		},
	)
	s.sendEvent(r.Context(), evt)
}

func (s *Source) sendEvent(ctx context.Context, evt cloudevents.Event) error {
	// Build the cloud event and emit it.
	if s.Sink == nil {
		return nil
	}

	evtCtx, evtCancel := context.WithTimeout(ctx, time.Second)
	defer evtCancel()
	result := s.Sink.Send(evtCtx, evt)

	if !cloudevents.IsACK(result) {
		return result
	}
	return nil
}

func (s *Source) writeResponse(w http.ResponseWriter, resp *http.Response) error {
	// Copy the headers.
	// header filter
	hf := func(h string) bool {
		switch strings.ToLower(h) {
		case "content-length":
			return false
		default:
			return true
		}
	}

	// Copy the headers.
	for k, h := range resp.Header {
		if hf(k) {
			for _, hh := range h {
				w.Header().Add(k, hh)
			}
		}
	}

	// Write the content.
	body := bytes.Buffer{}
	n, err := io.Copy(&body, resp.Body)
	if err != nil {
		return err
	}
	// Set content length
	w.Header().Set("Content-Length", strconv.FormatInt(n, 10))

	// Write the headers with the status code.
	w.WriteHeader(resp.StatusCode)

	// Write the body.
	_, err = w.Write(body.Bytes())
	if err != nil {
		return err
	}
	return nil
}
