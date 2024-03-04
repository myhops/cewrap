package cewrap

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

// type Proxy 

type ProxiedRequest struct {
	request  *SavedRequest
	response *SavedResponse
}

// Event creates an event from the proxied call.
func (p *ProxiedRequest) Event(
	source string,
	prefix,
	suffix,
	dataSchema string) *cloudevents.Event {
	evt := cloudevents.NewEvent()
	evt.SetID(uuid.NewString())
	evt.SetSource(source)
	evt.SetType(fmt.Sprintf("%s.%s%s", prefix, p.request.method, suffix))
	evt.SetSubject(p.request.path)

	if dataSchema != "" {
		evt.SetDataSchema(dataSchema)
	}

	const jsonType = "application/json"

	// Set the data
	contentType := p.response.header.Get("Content-Type")
	if strings.Index(contentType, jsonType) == 0 {
		// Copied from Event.SetData for data is not a byte array.
		evt.SetDataContentType(jsonType)
		evt.DataEncoded = p.response.body
		evt.DataBase64 = false
	} else {
		evt.SetData(contentType, p.response.body)
	}

	return &evt
}

type SavedRequest struct {
	body   []byte
	header http.Header
	method string
	path   string
}

func SaveRequest(r *http.Request) (*SavedRequest, error) {
	res := &SavedRequest{}

	// Get the method.
	res.method = r.Method

	// Get the body.
	if r.Body != nil {
		defer r.Body.Close()
		var err error
		res.body, err = io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
	}

	// Save the path
	res.path = r.URL.Path
	// Save the header
	res.header = r.Header.Clone()
	return res, nil
}

func (s *SavedRequest) Request(ctx context.Context, downstream string) (*http.Request, error) {
	// Get the full downstream url.
	du, err := url.JoinPath(downstream, s.path)
	if err != nil {
		return nil, err
	}
	// Create the request.
	req, err := http.NewRequestWithContext(ctx, s.method, du, bytes.NewReader(s.body))
	if err != nil {
		return nil, err
	}
	// Copy all headers except for content-length.
	cl := http.CanonicalHeaderKey("Content-Length")
	for k, vv := range s.header {
		if k == cl {
			continue
		}
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	return req, nil
}

type SavedResponse struct {
	body       []byte
	header     http.Header
	statusCode int
}

func SaveResponse(resp *http.Response) (*SavedResponse, error) {
	res := &SavedResponse{}

	res.statusCode = resp.StatusCode
	res.header = resp.Header.Clone()

	// Get the body.
	if resp.Body != nil {
		var err error
		res.body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (s *SavedResponse) Write(w http.ResponseWriter) error {
	// Copy all headers except for content-length.
	cl := http.CanonicalHeaderKey("Content-Length")
	for k, vv := range s.header {
		if k == cl {
			continue
		}
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	// Set the status code.
	w.WriteHeader(s.statusCode)
	// Write the body
	_, err := w.Write(s.body)
	return err
}
