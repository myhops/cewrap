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
func WithDownstream(u string) SourceOption {
	return downStream(u)
}

type sink struct{ c cloudevents.Client }

func (si sink) apply(s *Source) { s.sink = si.c }

func WithSink(s cloudevents.Client) SourceOption { return &sink{c: s} }

type httpClient struct {
	c *http.Client
}

func (c *httpClient) apply(s *Source) {
	s.client = c.c
}
func WithHTTPClient(c *http.Client) SourceOption {
	return &httpClient{c: c}
}

type changeMethods []string

func (c changeMethods) apply(s *Source) {
	s.changeMethods = c
}
func WithChangeMethods(m []string) SourceOption {
	return changeMethods(m)
}

type source string

func (ss source) apply(s *Source) { s.source = string(ss) }
func WithSource(v string) SourceOption {
	return source(v)
}

type prefix string

func (p prefix) apply(s *Source) { s.typePrefix = string(p) }
func WithTypePrefix(v string) SourceOption {
	return prefix(v)
}

type pathPrefix string

func (p pathPrefix) apply(s *Source) { s.pathPrefix = string(p) }
func WithPathPrefix(v string) SourceOption {
	return pathPrefix(v)
}

type dataSchema string

func (d dataSchema) apply(s *Source) { s.dataschema = string(d) }
func WithDataschema(v string) SourceOption {
	return dataSchema(v)
}

type loggerOption struct{ l *slog.Logger }

func (l loggerOption) apply(s *Source) { s.logger = l.l }
func WithLogger(l *slog.Logger) SourceOption {
	return loggerOption{l: l}
}
