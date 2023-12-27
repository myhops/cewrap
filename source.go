package cewrap

import (
	"log/slog"
	"net/http"
	"net/url"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
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
	// Dataschema for the event.
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

		// Create and init a serviceRequest.
		svcReq := &serviceRequest{}
		svcReq.logger = logger.With(slog.String("request", r.URL.Path))
		svcReq.s = s

		ctx := r.Context()
		err := svcReq.callDownstream(ctx, w, r)
		if err != nil {
			// write error
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("error calling downstream", slog.String("err", err.Error()))
			return
		}
		logger.Info("successfully proxied request")

		emit := func() {
			// Emit the event.
			logger.Info("emitting event")
			err = svcReq.emitEvent(ctx)
			if err != nil {
				logger.Error("emitEvent failed", slog.String("err", err.Error()))
			}
			logger.Info("emitted event")
		}
		// Check if an event needs to be emitted.
		if s.isEmitEvent(r.Method) {
			if s.asyncEmit {
				go emit()
				return
			}
			emit()
			return
		}
		logger.Info("skip emitting event")
	}
}
