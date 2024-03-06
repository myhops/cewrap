package cewrap

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func TestStraightTrough(t *testing.T) {
	cases := []struct {
		name string
		method string
		path string
	}{
		{
			name: "root",
			method: http.MethodGet,
			path: "/",
		},
		{
			name: "some path",
			method: http.MethodGet,
			path: "/some/path",
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T)  {
			var responseBody = []byte("dasfadfaefacvv vasdfad asdfasd")
			// upstream server
			us := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write(responseBody)
			}))

			defer us.Close()

			upstream, _ := url.Parse(us.URL)
			th := NewTapHandler(upstream)

			ps := httptest.NewServer(th)
			defer ps.Close()

			reqURL, err := url.JoinPath(ps.URL, cc.path)
			if err != nil {
				t.Fatalf("error joining paths: %s", err.Error())
			}

			req, err := http.NewRequest(cc.method, reqURL, nil)
			if err != nil {
				t.Fatalf("error creating request: %s", err.Error())
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("error getting root: %s", err.Error())
			}

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				t.Fatalf("received bad status code: %s", resp.Status)
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("error reading body")
			}
			if !bytes.Equal(responseBody, body) {
				t.Fatalf("received incorrect body")
			}

		})
	}
}

func TestTapper(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tap := &loggingTap{l: logger.With("tap", "allTap")}

	getTap := &loggingTap{l: logger.With("tap", "getTap")}

	cases := []struct {
		name string
		method string
		path string
		options []option
	}{
		{
			name: "root",
			method: http.MethodGet,
			path: "/get",
			options: []option{WithLogger(logger), WithTap([]string{"/get"},tap)},
		},
		{
			name: "some path tap",
			method: http.MethodPut,
			path: "/some/path",
			options: []option{WithLogger(logger), WithTap([]string{"/some/path"}, tap)},
		},
		{
			name: "some path getTap",
			method: http.MethodGet,
			path: "/some/path",
			options: []option{WithLogger(logger), WithTap([]string{"/some/path"}, getTap)},
		},
		{
			name: "some path multiTap",
			method: http.MethodPut,
			path: "/some/path",
			options: []option{
					WithLogger(logger), 
					WithTap([]string{"PUT /some/path"}, getTap),
					WithTap([]string{"/some/path"}, tap),
				},
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T)  {
			var responseBody = []byte("dasfadfaefacvv vasdfad asdfasd")
			// upstream server
			us := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write(responseBody)
			}))

			defer us.Close()

			upstream, _ := url.Parse(us.URL)
			th := NewTapHandler(upstream, cc.options...)

			ps := httptest.NewServer(th)
			defer ps.Close()

			reqURL, err := url.JoinPath(ps.URL, cc.path)
			if err != nil {
				t.Fatalf("error joining paths: %s", err.Error())
			}

			req, err := http.NewRequest(cc.method, reqURL, nil)
			if err != nil {
				t.Fatalf("error creating request: %s", err.Error())
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("error getting %s: %s", cc.path, err.Error())
			}

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				t.Fatalf("received bad status code: %s", resp.Status)
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("error reading body")
			}
			if !bytes.Equal(responseBody, body) {
				t.Fatalf("received incorrect body")
			}

		})
	}
	t.Error()
}