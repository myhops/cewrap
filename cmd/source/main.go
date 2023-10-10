package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/myhops/cewrap"
)

func newLogger(o *options) *slog.Logger {
	var h slog.Handler
	if o.logFormat == "text" {
		h = slog.NewTextHandler(os.Stderr, nil)
	} else {
		h = slog.NewJSONHandler(os.Stderr, nil)
	}

	return slog.New(h).With(
		slog.String("application", "cewrap/source"),
	)
}

func logOptions(o *options, l *slog.Logger) {
	l.WithGroup("options").Info("used options",
		slog.String("dataschema", o.dataschema),
		slog.String("downstream", o.downstream),
		slog.String("pathPrefix", o.pathPrefix),
		slog.String("port", o.port),
		slog.String("sink", o.sink),
		slog.String("source", o.source),
		slog.String("typePrefix", o.typePrefix),
		slog.String("logFormat", o.logFormat),
		slog.String("logLevel", o.logLevel),
	)
}

func main() {
	opts, err := getOptions()
	if err != nil {
		slog.Default().Error("failed to get options", slog.String("err", err.Error()))
		return
	}
	logger := newLogger(opts)

	so, err := opts.getSourceOptions()
	if err != nil {
		logger.Error("error creating sink", slog.String("err", err.Error()))
		os.Exit(1)
	}

	// Add the logger.
	so = append(so, cewrap.WithLogger(logger))
	// Create the source.
	s := cewrap.NewSource(so...)

	// Log the current options.
	logOptions(opts, logger)

	// Start server with the source.
	la := ":" + opts.port
	logger.Info("starting server", slog.String("listen_address", la))
	if err := http.ListenAndServe(la, s.Handler()); err != nil {
		logger.Error("server stopped", slog.String("err", err.Error()))
	}
}
