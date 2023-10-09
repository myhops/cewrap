package cewrap

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type EventData struct {
	ResourceData any `json:"resource_data"`
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

	Logger *slog.Logger
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

func NewSource(downstream, sink string, client *http.Client, changeMethods []string, logger *slog.Logger) *Source {
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
	if logger == nil {
		logger = slog.Default()
	}
	s.Logger = logger.With(
		slog.String("service", "Source"),
	)

	return s
}

func (s *Source) buildDownstreamRequest(ctx context.Context, r *http.Request) (*http.Request, error) {
	// Get the body.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// Build the downstream path.
	du, err := url.JoinPath(s.Downstream.String(), r.URL.Path)
	if err != nil {
		return nil, err
	}

	// Create the request.
	req, err := http.NewRequestWithContext(ctx, r.Method, du, bytes.NewReader(body))
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

// Handler returns a HandlerFunc that handles the requests.
//
// It passes the request to the downstream service and generates a cloud event
// and sends it to the sink.
//
// TODO: Save the response body as bytes because we need it for the event.
func (s *Source) Handler() http.HandlerFunc {
	// Initialize the variables common to all requests.
	logger := s.Logger.With(slog.String("operation", "Handle"))

	return func(w http.ResponseWriter, r *http.Request) {
		defer func(start time.Time) {
			logger.Info("Handle served", slog.Duration("duration", time.Since(start)))
		}(time.Now())

		// Build a client request from the server request.
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		cr, err := s.buildDownstreamRequest(ctx, r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("error building downstream request", slog.String("err", err.Error()))
			return
		}
		logger.Info("build the client request",
			slog.Group("client_request",
				slog.String("host", cr.Host),
				slog.String("path", cr.URL.Path),
			),
		)

		// Call the downstream service.
		resp, err := s.Client.Do(cr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("error calling downstream service", slog.String("err", err.Error()))
			return
		}
		logger.Info("called the downstream service")

		// Create the response and write it out to the responseWriter.
		err = s.writeResponse(w, resp)
		if err != nil {
			logger.Error("error sending the response", slog.String("err", err.Error()))
			return
		}

		if resp.StatusCode < 200 && resp.StatusCode >= 300 {
			// We are done.
			return // Write an internal error.
		}

		if !s.isChange(r.Method) {
			return
		}
		// Only run when we have an events sink.
		if s.Sink == nil {
			logger.Info("sink is nil")
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
			EventData{},
		)
		err = s.sendEvent(r.Context(), evt)
		if err != nil {
			logger.Error("sending event error", slog.String("err", err.Error()))
		}
		logger.Info("event sent", slog.String("event", evt.String()))
	}
}

// buildEvent builds a cloud event.
//
// It uses the req to derive the subject and the type.
// It uses the
func (s *Source) buildEvent(req *http.Request, resp *http.Response) (*cloudevents.Event, error) {
	evt := cloudevents.NewEvent()
	evt.SetSource("bla.source")

	us := &url.URL{}
	us.Scheme = req.URL.Scheme
	us.Host = req.Host
	us.Path = req.URL.Path
	us.Scheme = "http"

	sub := us.String()
	evt.SetSubject(sub)
	evt.SetType("my.type")
	evt.SetData("application/json",
		EventData{
		},
	)
	return &evt, nil
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
