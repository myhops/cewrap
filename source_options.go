package cewrap

import (
	"log/slog"
	"net/http"
	"net/url"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type SourceOption interface {
	apply(s *Source)
}

type downStream string

func (ds downStream) apply(s *Source) {
	if uu, err := url.Parse(string(ds)); err == nil {
		s.downstream = uu
	}
}

// WithDownstream sets the url for the downstream service.
func WithDownstream(u string) SourceOption {
	return downStream(u)
}

type sink struct{ c cloudevents.Client }

func (si sink) apply(s *Source) { s.sink = si.c }

// WithSink sets the sink for the Cloud events to send to.
func WithSink(s cloudevents.Client) SourceOption { return &sink{c: s} }

type httpClient struct {
	c *http.Client
}

func (c *httpClient) apply(s *Source) { s.client = c.c }

// WithHTTPClient sets the http client to use for calling the downstream service.
func WithHTTPClient(c *http.Client) SourceOption {
	return &httpClient{c: c}
}

type changeMethods []string

func (c changeMethods) apply(s *Source) { s.changeMethods = c }

// WithChangeMethods sets the methods that should emit an event.
func WithChangeMethods(m []string) SourceOption {
	return changeMethods(m)
}

type source string

func (ss source) apply(s *Source) { s.source = string(ss) }

// WithSource sets the source field for the clound event.
func WithSource(v string) SourceOption {
	return source(v)
}

type prefix string

func (p prefix) apply(s *Source) { s.typePrefix = string(p) }

// WithTypePrefix sets the prefix of the type field of the cloud event.
//
// A prefix of thisPrefix will result in cloud events with the 
// typefield set to thisPrefix.post_handled when the method was a post.
func WithTypePrefix(v string) SourceOption {
	return prefix(v)
}

type pathPrefix string

func (p pathPrefix) apply(s *Source) { s.pathPrefix = string(p) }

// WithPathPrefix sets the prefix that will be removed from the request path.
//
// The resulting path will be the subject of the cloud event.
func WithPathPrefix(v string) SourceOption {
	return pathPrefix(v)
}

type dataSchema string

func (d dataSchema) apply(s *Source) { s.dataSchema = string(d) }

// WithDataschema sets the dataschema of the cloud event.
func WithDataschema(v string) SourceOption {
	return dataSchema(v)
}

type loggerOption struct{ l *slog.Logger }

func (l loggerOption) apply(s *Source) { s.logger = l.l }

// // WithLogger sets the logger.
// func WithLogger(l *slog.Logger) SourceOption {
// 	return loggerOption{l: l}
// }
