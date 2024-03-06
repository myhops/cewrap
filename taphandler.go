package cewrap

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type TapHandler struct {
	http.ServeMux
	upstream *url.URL
	logger   *slog.Logger

	rootSet bool
}

type option func(*TapHandler)

func NewTapHandler(upstream *url.URL, options ...option) *TapHandler {
	t := &TapHandler{
		upstream: upstream,
	}

	for _, opt := range options {
		opt(t)
	}

	if !t.rootSet {
		t.ServeMux.Handle("/", NewStraightTroughProxy(t.upstream))
	}
	return t
}

func NewStraightTroughProxy(u *url.URL) http.Handler {
	return httputil.NewSingleHostReverseProxy(u)
}

// Interface check.
var _ http.Handler = (*TapHandler)(nil)

func WithTap(paths []string, tap Tap) option {
	return func(t *TapHandler) {
		// Modify the logger
		var l = t.logger
		for i, p := range paths {
			l = l.With(fmt.Sprintf("tap_%d", i), p)
		}
		h := t.newTappingHandler(tap, l)
		for _, p := range paths {
			t.rootSet = t.rootSet || p == "/"
			t.ServeMux.Handle(p, h)
		}
	}
}

func WithLogger(l *slog.Logger) option {
	return func(t *TapHandler) {
		t.logger = l
	}
}

func (t *TapHandler) newTappingHandler(tap Tap, logger *slog.Logger) http.HandlerFunc {
	spy := &spy{
		t:      t,
		tap:    tap,
		logger: logger,
	}
	h := &httputil.ReverseProxy{
		Rewrite:        spy.rewrite,
		ModifyResponse: spy.modifyResponse,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a proxy with a spy.
		h.ServeHTTP(w, r)
	})
}

type spy struct {
	t      *TapHandler
	tap    Tap
	logger *slog.Logger
}

// rewrite logs the call and starts the tap.
func (s *spy) rewrite(pr *httputil.ProxyRequest) {
	s.logger.Info("called")
	pr.SetURL(s.t.upstream)
	s.tap.Start(pr)
}

func (s *spy) modifyResponse(r *http.Response) error {
	s.logger.Info("called")
	s.tap.End(r)
	return nil
}

type loggingTap struct {
	l *slog.Logger
}

func (t *loggingTap) Start(pr *httputil.ProxyRequest) {
	t.l.Info("loggingTap", "method", "Start")
}

func (t *loggingTap) End(r *http.Response) error {
	t.l.Info("loggingTap", "method", "Start")
	return nil
}
