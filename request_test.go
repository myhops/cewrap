package cewrap

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCloneRequest(t *testing.T) {
	const (
		srv1Body = "svr1 called"
		srv2Body = "svr2 called"
	)

	// Setup a downstream server.
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(srv2Body))
		t.Logf(srv2Body)
	}))
	defer srv2.Close()

	// Setup an upstream server
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NewSimpleProxy(srv2.URL, nil, nil).Do(w, r)
	}))
	defer srv1.Close()

	resp, err := http.Get(srv1.URL)
	if err != nil {
		t.Errorf("get error: %s", err.Error())
	}
	if resp.StatusCode < 200 && resp.StatusCode >= 300 {
		t.Errorf("status code: %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("read body failed: %s", err.Error())
	}
	if string(b) != srv2Body {
		t.Errorf("body mismatch, want %s, got %s", srv2Body, string(b))
	}
}
