package cewrap

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type serviceRequest struct {
	responseBody []byte
	method       string
	requestPath  string
	contentType  string
}

func (s *serviceRequest) emitEvent(ctx context.Context) error {
	const typeSuffix = "_handled"

	evt := cloudevents.NewEvent()
	// id := uuid.New()
	// evt.SetID(id.String())
	// evt.SetSource(s.s.source)
	// evt.SetType(s.s.typePrefix + "." + strings.ToLower(s.method) + typeSuffix)
	// evt.SetSubject(s.requestPath)
	// if s.s.dataSchema != "" {
	// 	evt.SetDataSchema(s.s.dataSchema)
	// }
	evt.SetTime(time.Now())

	const jsonType = "application/json"

	// Set the data
	if strings.Index(s.contentType, jsonType) == 0 {
		// Copied from Event.SetData for data is not a byte array.
		evt.SetDataContentType(jsonType)
		evt.DataEncoded = s.responseBody
		evt.DataBase64 = false
	} else {
		evt.SetData(s.contentType, s.responseBody)
	}

	// return s.sendEvent(ctx, evt)
	return nil
}

// func (s *serviceRequest) sendEvent(ctx context.Context, evt cloudevents.Event) error {
// 	evtCtx, evtCancel := context.WithTimeout(ctx, time.Second)
// 	defer evtCancel()
// 	result := s.s.sink.Send(evtCtx, evt)

// 	if !cloudevents.IsACK(result) {
// 		return result
// 	}
// 	return nil
// }

func (s *serviceRequest) buildDownstreamRequest(ctx context.Context, r *http.Request) (*http.Request, error) {
	// Get the body.
	// body, err := io.ReadAll(r.Body)
	// if err != nil {
	// 	return nil, err
	// }
	// r.Body.Close()

	// Build the downstream path.
	// du, err := url.JoinPath(s.s.downstream.String(), r.URL.Path)
	// if err != nil {
	// 	return nil, err
	// }

	// Create the request.
	// req, err := http.NewRequestWithContext(ctx, r.Method, du, bytes.NewReader(body))
	// if err != nil {
	// 	return nil, err
	// }

	// Copy the headers.
	// for k, h := range r.Header {
	// 	for _, hh := range h {
	// 		req.Header.Add(k, hh)
	// 	}
	// }
	// return req, nil
	return nil, nil
}

// writeResponse writes the info from resp to w.
func (s *serviceRequest) writeResponse(w http.ResponseWriter, resp *http.Response, body io.Reader) error {
	// Copy the headers.
	for k, h := range resp.Header {
		for _, hh := range h {
			w.Header().Add(k, hh)
		}
	}

	// Write the headers with the status code.
	w.WriteHeader(resp.StatusCode)

	// Write the body.
	if _, err := io.Copy(w, body); err != nil {
		return fmt.Errorf("error parsing content-length: %w", err)
	}
	return nil
}
