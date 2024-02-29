package cewrap

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

type proxyRequest struct {
	s      *Source
	logger *slog.Logger
}

func (p *proxyRequest) callDownstream(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	logger := p.logger.With(slog.String("receiver_method", "callDownstream"))

	// Build a client request from the server request.
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	cr, err := p.buildDownstreamRequest(ctx, r)
	if err != nil {
		return fmt.Errorf("error building downstream request: %w", err)
	}
	logger.Info("build the client request",
		slog.Group("client_request",
			slog.String("host", cr.Host),
			slog.String("path", cr.URL.Path),
		),
	)

	// Call the downstream service.
	resp, err := p.s.client.Do(cr)
	if err != nil {
		return fmt.Errorf("error calling downstream service: %w", err)
	}

	// Save the body.
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error reading downstream response body: %w", err)
	}

	logger.Info("called the downstream service")
	// Create the response and write it out to the responseWriter.
	err = p.writeResponse(w, resp, bytes.NewReader(body))
	if err != nil {
		logger.Error("error sending the response", slog.String("err", err.Error()))
		return fmt.Errorf("error sending the response: %w", err)
	}

	return nil
}

// writeResponse writes the info from resp to w.
func (p *proxyRequest) writeResponse(w http.ResponseWriter, resp *http.Response, body io.Reader) error {
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

func (p *proxyRequest) buildDownstreamRequest(ctx context.Context, r *http.Request) (*http.Request, error) {
	// Get the body.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body.Close()

	// Build the downstream path.
	du, err := url.JoinPath(p.s.downstream.String(), r.URL.Path)
	if err != nil {
		return nil, err
	}

	// Create the request.
	req, err := http.NewRequestWithContext(ctx, r.Method, du, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Copy the headers.
	for k, h := range r.Header {
		for _, hh := range h {
			req.Header.Add(k, hh)
		}
	}
	return req, nil
}


type serviceRequest struct {
	proxyRequest

	responseBody []byte
	method       string
	requestPath  string
	contentType  string
}

// callDownstream calls the downstream service and writes the response to the original caller.
//
// When err == nil, then w cannot be used anymore.
func (s *serviceRequest) callDownstream(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	err := s.proxyRequest.callDownstream(ctx, w, r)
	if err != nil {
		return err
	}

	if !s.s.isEmitEvent(r.Method) {
		return nil
	}

	// Save event data.
	// s.responseBody = body
	// s.contentType = resp.Header.Get("content-type")
	// s.saveRequestData(cr)
	return nil
}

func (s *serviceRequest) emitEvent(ctx context.Context) error {
	const typeSuffix = "_handled"

	evt := cloudevents.NewEvent()
	id := uuid.New()
	evt.SetID(id.String())
	evt.SetSource(s.s.source)
	evt.SetType(s.s.typePrefix + "." + strings.ToLower(s.method) + typeSuffix)
	evt.SetSubject(s.requestPath)
	if s.s.dataSchema != "" {
		evt.SetDataSchema(s.s.dataSchema)
	}
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

	s.logger.Info("about to send event")
	return s.sendEvent(ctx, evt)
}

func (s *serviceRequest) sendEvent(ctx context.Context, evt cloudevents.Event) error {
	evtCtx, evtCancel := context.WithTimeout(ctx, time.Second)
	defer evtCancel()
	result := s.s.sink.Send(evtCtx, evt)

	if !cloudevents.IsACK(result) {
		return result
	}
	return nil
}

func (s *serviceRequest) buildDownstreamRequest(ctx context.Context, r *http.Request) (*http.Request, error) {
	// Get the body.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body.Close()

	// Build the downstream path.
	du, err := url.JoinPath(s.s.downstream.String(), r.URL.Path)
	if err != nil {
		return nil, err
	}

	// Create the request.
	req, err := http.NewRequestWithContext(ctx, r.Method, du, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Copy the headers.
	for k, h := range r.Header {
		for _, hh := range h {
			req.Header.Add(k, hh)
		}
	}
	return req, nil
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

// saveRequestData saves data for the event.
//
// It uses the req to derive the subject and the type.
func (s *serviceRequest) saveRequestData(r *http.Request) {
	us := &url.URL{}
	us.Scheme = r.URL.Scheme
	us.Host = r.Host
	us.Path = r.URL.Path

	s.requestPath = us.Path
	if strings.Index(s.requestPath, s.s.pathPrefix) == 0 {
		s.requestPath = s.requestPath[len(s.s.pathPrefix):]
	}

	s.method = r.Method
}
