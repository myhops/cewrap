package cewrap

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestStraightTrough(t *testing.T) {
	cases := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "root",
			method: http.MethodGet,
			path:   "/",
		},
		{
			name:   "some path",
			method: http.MethodGet,
			path:   "/some/path",
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
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
	newAllTap := func() Tap {
		return &loggingTap{l: logger.With("tap", "allTap")}
	}

	newGetTap := func() Tap {
		return &loggingTap{l: logger.With("tap", "getTap")}
	}

	newPutTap := func() Tap {
		return &BodyRecordingTap{}
	}

	cases := []struct {
		name    string
		method  string
		path    string
		options []option
		body    string
	}{
		{
			name:    "root",
			method:  http.MethodGet,
			path:    "/get",
			options: []option{WithLogger(logger), WithTap([]string{"/get"}, newAllTap)},
		},
		{
			name:    "some path tap",
			method:  http.MethodPut,
			path:    "/some/path",
			options: []option{WithLogger(logger), WithTap([]string{"/some/path"}, newGetTap)},
		},
		{
			name:    "some path getTap",
			method:  http.MethodGet,
			path:    "/some/path",
			options: []option{WithLogger(logger), WithTap([]string{"/some/path"}, newGetTap)},
		},
		{
			name:   "some path multiTap",
			method: http.MethodPut,
			path:   "/some/path",
			options: []option{
				WithLogger(logger),
				WithTap([]string{"PUT /some/path"}, newGetTap),
				WithTap([]string{"/some/path"}, newAllTap),
			},
		},
		{
			name:    "put path",
			method:  http.MethodGet,
			path:    "/put/path",
			options: []option{WithLogger(logger), WithTap([]string{"/put/path"}, newPutTap)},
			body: "Hallo Put",
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
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

			req, err := http.NewRequest(cc.method, reqURL, strings.NewReader(cc.body))
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
