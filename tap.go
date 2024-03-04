package cewrap

import "net/http"

type Tap interface {
	Start(serverRequest *http.Request) (clientRequest *http.Request, err error)
	Finish(responseWriter http.ResponseWriter, response *http.Response) error
}

type DummyTap struct {}

func (d *DummyTap) Start(serverRequest *http.Request) (clientRequest *http.Request, err error) {
	// Create a new client request.
	
	return serverRequest, nil
}

func (d *DummyTap) Finish(responseWriter http.ResponseWriter, response *http.Response) error {
	return nil
}