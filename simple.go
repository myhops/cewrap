package cewrap

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

type SimpleProxy struct {
	client     *http.Client
	logger     *slog.Logger
	downstream string
}

func NewSimpleProxy(downstream string, client *http.Client, logger *slog.Logger) *SimpleProxy {
	res := &SimpleProxy{
		downstream: downstream,
	}
	if client == nil {
		res.client = http.DefaultClient
	}
	return res
}

// Do forwards the request r to the downstream url
// and writes the response verbatim to w.
func (p *SimpleProxy) Do(w http.ResponseWriter, r *http.Request) error {
	req, err := newClientRequest(r.Context(), p.downstream, r)
	if r.Body != nil {
		defer r.Body.Close()
	}
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	return passResponse(w, resp)
}

// newClientRequest creates a new client request from the server request.
//
// The url of the request consist of the downstream url appended with the path of
// the request.
func newClientRequest(ctx context.Context, downStream string, r *http.Request) (*http.Request, error) {
	// Build the url
	nus, err := url.JoinPath(downStream, r.URL.RawPath)
	if err != nil {
		return nil, err
	}

	// Clone the request.
	rr, err := http.NewRequestWithContext(ctx, r.Method, nus, r.Body)
	if err != nil {
		return nil, err
	}
	// Clone the headers.
	for k, v := range r.Header {
		rr.Header[k] = v
	}

	return rr, nil
}

// passResponse passes the response to w.
func passResponse(w http.ResponseWriter, r *http.Response) error {
	for k, v := range r.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(r.StatusCode)
	if _, err := io.Copy(w, r.Body); err != nil {
		return err
	}
	return nil
}
