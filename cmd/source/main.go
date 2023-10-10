package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/myhops/cewrap"
)

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, nil)).With(
		slog.String("application", "cewrap/source"),
	)
}

func main() {
	logger := newLogger()
	opts, err := getOptions()
	if err != nil {
		logger.Error("failed to get options", slog.String("err", err.Error()))
		return
	}

	so, err := opts.getSourceOptions()
	if err != nil {
		logger.Error("error creating sink", slog.String("err", err.Error()))
		return
	}

	// Add the logger.
	so = append(so, cewrap.WithLogger(logger))
	// Create the source.
	s := cewrap.NewSource(so...)

	// Start server with the source.
	if err := http.ListenAndServe(":"+opts.port, s.Handler()); err != nil {
		logger.Error("server stopped", slog.String("err", err.Error()))
	}
}
