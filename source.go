package cewrap

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

type Source struct {
	// The downstream service.
	downstream *url.URL
	// sink is the url that sinks the events
	sink cloudevents.Client
	// HTTP client for sending the downstream requests.
	client *http.Client
	// Methods that indicate a change and will generate an event.
	changeMethods []string

	// Event source.
	source string
	// Type prefix for the event type field.
	typePrefix string
	// Path prefix, when set, removes the prefix from the path that is set in the event source.
	pathPrefix string
	// DataSchema for the event.
	dataSchema string

	// When true, emit in a go routine.
	asyncEmit bool

	// logger
	logger *slog.Logger
}

var DefaultChangeMethods = []string{
	http.MethodPost,
	http.MethodDelete,
	http.MethodPatch,
	http.MethodPut,
}

func (s *Source) isChange(method string) bool {
	for _, m := range s.changeMethods {
		if method == m {
			return true
		}
	}
	return false
}

func NewSource(options ...SourceOption) *Source {
	s := &Source{}
	for _, opt := range options {
		opt.apply(s)
	}

	if s.client == nil {
		s.client = http.DefaultClient
	}
	if len(s.changeMethods) == 0 {
		s.changeMethods = DefaultChangeMethods
	}
	if s.logger == nil {
		s.logger = slog.Default().With(
			slog.String("service", "Source"),
		)
	}
	return s
}

func (s *Source) isEmitEvent(method string) bool {
	return s.sink != nil && s.isChange(method)
}

func (s *Source) proxyRequest(logger *slog.Logger, w http.ResponseWriter, r *http.Request) (*ProxiedRequest, error) {
	// Save the request.
	sreq, err := SaveRequest(r)
	if err != nil {
		// write error
		return nil, fmt.Errorf("error saving request: %w", err)
	}

	req, err := sreq.Request(r.Context(), s.downstream.String())
	if err != nil {
		// write error
		logger.Error("creating downstream request", slog.String("err", err.Error()))
		return nil, fmt.Errorf("error creating downstream request: %w", err)
	}

	// Call the downstream service.
	resp, err := s.client.Do(req)
	if err != nil {
		// write error
		return nil, fmt.Errorf("error calling downstream service: %w", err)
	}
	// Save the response.
	sresp, err := SaveResponse(resp)
	if err != nil {
		// write error
		return nil, fmt.Errorf("error saving response: %w", err)
	}

	// Write the response.
	if err := sresp.Write(w); err != nil {
		// write error

		logger.Error("saving response", slog.String("err", err.Error()))
		return nil, fmt.Errorf("error writing the response response: %w", err)
	}
	return &ProxiedRequest{
		request:  sreq,
		response: sresp,
	}, nil
}

func (s *Source) mayEmit(req *http.Request) bool {
	return false
}

func (s *Source) TappingHandler() http.HandlerFunc {
	tap := &DummyTap{}
	return func(w http.ResponseWriter, r *http.Request) {
		req, _ := tap.Start(r)

		// Call the downstream service.
		resp, err := s.client.Do(req)
		if err != nil {
			// write error
			return
		}
		tap.Finish(w, resp)
	}
}

// Handler returns a HandlerFunc that handles the requests.
//
// It passes the request to the downstream service and generates a cloud event
// and sends it to the sink.
func (s *Source) Handler() http.HandlerFunc {
	// Initialize the log variables common to all requests.
	logger := s.logger.With(slog.String("operation", "Handle"))

	return func(w http.ResponseWriter, r *http.Request) {
		defer func(start time.Time) {
			logger.Info("Handle served", slog.Duration("duration", time.Since(start)))
		}(time.Now())

		// Test if the request may emit an event.
		// if ! s.mayEmit(req) {

		// }

		pr, err := s.proxyRequest(s.logger, w, r)
		if err != nil {
			http.Error(w, fmt.Sprintf("error proxying the request: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		logger.Info("successfully proxied request")

		// Determine if an event needs to be emitted.
		if !s.isEmitEvent(pr.request.method) {
			return
		}

		// Create and emit the event.
		s.emitEvent(logger, pr)
	}
}

func (s *Source) emitEvent(logger *slog.Logger, pr *ProxiedRequest) {
	// Construct the event.
	evt := cloudevents.NewEvent()
	evt.SetID(uuid.NewString())
	evt.SetSource(s.source)
	evt.SetType(fmt.Sprintf("%s.%s%s", s.typePrefix, pr.request.method, "_handled"))
	evt.SetSubject(pr.request.path)

	if s.dataSchema != "" {
		evt.SetDataSchema(s.dataSchema)
	}

	const jsonType = "application/json"

	// Set the data
	contentType := pr.response.header.Get("Content-Type")
	if strings.Index(contentType, jsonType) == 0 {
		// Copied from Event.SetData for data is not a byte array.
		evt.SetDataContentType(jsonType)
		// evt.DataEncoded = p.response.body
		evt.DataBase64 = false
	} else {
		// evt.SetData(contentType, p.response.body)
	}

	emit := func() {
		// Emit the event.
		logger.Info("emitting event")
		// err = svcReq.emitEvent(ctx)
		// if err != nil {
		// 	logger.Error("emitEvent failed", slog.String("err", err.Error()))
		// }
		logger.Info("emitted event")
	}
	// Check if an event needs to be emitted.
	if s.asyncEmit {
		go emit()
		return
	}
	emit()
	logger.Info("skip emitting event")
}
