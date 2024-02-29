package cewrap

import (
	"net/http"
	"time"
)

const (
	DefaultReadHeaderTimeout time.Duration = 3 * time.Second
)

// Returns a server with sensible defaults.
func NewServer() *http.Server {
	srv := &http.Server{
		ReadHeaderTimeout: DefaultReadHeaderTimeout,
	}

	return srv
}
