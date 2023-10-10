package cewrap

import (
	"log/slog"
	"net/http"
	"net/url"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type sourceOptions struct {
	downstream    string
	sink          string
	client        *http.Client
	changeMethods []string
	source        string
	typePrefix    string
	pathPrefix    string
	dataSchema    string
	logger        *slog.Logger
}

type sourceOption func(s *Source)
type SourceOptions []sourceOption

func WithDownstream(u string) sourceOption {
	return func(s *Source) {
		if uu, err := url.Parse(u); err == nil {
			s.downstream = uu
		}
	}
}

func WithSink(sink cloudevents.Client) sourceOption {
	return func(s *Source) {
		s.sink = sink
	}
}

func WithHTTPClient(c *http.Client) sourceOption {
	return func(s *Source) {
		s.client = c
	}
}

func WithChangeMethods(m []string) sourceOption {
	return func(s *Source) {
		s.changeMethods = m
	}
}

func WithSource(v string) sourceOption {
	return func(s *Source) {
		s.source = v
	}
}

func WithTypePrefix(v string) sourceOption {
	return func(s *Source) {
		s.typePrefix = v
	}
}

func WithPathPrefix(v string) sourceOption {
	return func(s *Source) {
		s.pathPrefix = v
	}
}

func WithDataschema(v string) sourceOption {
	return func(s *Source) {
		s.dataschema = v
	}
}

func WithLogger(l *slog.Logger) sourceOption {
	return func(s *Source) {
		s.logger = l
	}
}
