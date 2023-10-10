package cewrap

import (
	"log/slog"
	"net/http"
	"net/url"
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

func (s *Source) isEmitEvent(method string) bool {
	return s.Sink != nil && s.isChange(method)
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
		
		if !s.isEmitEvent(r.Method) {
			return 
		}
		svcReq.emitEvent(ctx)
	}
}
