package cewrap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
)

type TapHandler struct {
	http.ServeMux
	upstream *url.URL
	logger   *slog.Logger

	rootSet bool
}

type option func(*TapHandler)

// NewTapHandler creates a new tap handler with the given options.
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

// NewStraightTroughProxy creates a proxy that
// only calls the upstream service.
func NewStraightTroughProxy(u *url.URL) http.Handler {
	return httputil.NewSingleHostReverseProxy(u)
}

// Interface check.
var _ http.Handler = (*TapHandler)(nil)

func WithTap(paths []string, ntf NewTapFunc) option {
	return func(t *TapHandler) {
		// Modify the logger
		var l = t.logger
		for i, p := range paths {
			l = l.With(fmt.Sprintf("tap_%d", i), p)
		}
		h := t.newTappingHandler(ntf, l)
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

func (t *TapHandler) newTappingHandler(newTapFunc NewTapFunc, logger *slog.Logger) http.HandlerFunc {
	spy := &spy{
		t:      t,
		ntf:    newTapFunc,
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
	ntf    NewTapFunc
	logger *slog.Logger
}

// rewrite logs the call and starts the tap.
func (s *spy) rewrite(pr *httputil.ProxyRequest) {
	s.logger.Info("called")
	pr.SetURL(s.t.upstream)

	// Inject tap and start it.
	t := s.ntf()
	pr.Out = pr.Out.Clone(putTap(pr.Out.Context(), t))
	t.Start(pr)
}

func (s *spy) modifyResponse(r *http.Response) error {
	s.logger.Info("called")
	// Retrieve the tap and end it.
	t := getTap(r.Request.Context())
	if t == nil {
		return errors.New("no tap found in request")
	}
	t.End(r)
	return nil
}

// loggingTap logs its methods calls.
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

type contextKey int

var (
	tapKey    contextKey = 1
	loggerKey contextKey = 2
)

type NewTapFunc func() Tap

type Tap interface {
	Start(pr *httputil.ProxyRequest)
	End(r *http.Response) error
}

func putTap(ctx context.Context, tap Tap) context.Context {
	return context.WithValue(ctx, tapKey, tap)
}

func getTap(ctx context.Context) Tap {
	t, ok := ctx.Value(tapKey).(Tap)
	if !ok {
		panic(fmt.Sprintf("tapKey has incorrect type: %s", reflect.TypeOf(t).String()))
	}
	return t
}

type noopTap struct {
}

func (t *noopTap) Start(pr *httputil.ProxyRequest) {}
func (t *noopTap) End(r *http.Response) error      { return nil }

func newNoopTap() Tap {
	return &noopTap{}
}

type BodyRecordingTap struct {
	StatusCode int
	ReqMethod  string
	ReqHeader  http.Header
	RespHeader http.Header
	InBody     *bytes.Buffer
	OutBody    *bytes.Buffer
}

func (t *BodyRecordingTap) Start(pr *httputil.ProxyRequest) {
	// Save the request body.
	t.InBody = &bytes.Buffer{}
	b := &bytes.Buffer{}
	b.ReadFrom(io.TeeReader(pr.Out.Body, t.InBody))
	pr.Out.Body = io.NopCloser(b)
	
	// Save the headers.
	t.ReqHeader = pr.Out.Header.Clone()
	t.ReqMethod = pr.Out.Method
}

func (t *BodyRecordingTap) End(r *http.Response) error {
	// Save the response body.
	t.OutBody = &bytes.Buffer{}
	b := &bytes.Buffer{}
	b.ReadFrom(io.TeeReader(r.Body, t.OutBody))
	r.Body = io.NopCloser(b)

	// Here we can collect the data
	t.StatusCode = r.StatusCode
	t.RespHeader = r.Header.Clone()
	return nil
}
