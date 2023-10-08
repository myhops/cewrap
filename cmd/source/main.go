package main

import (
	"log"
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
	opts, err := getOptions()
	if err != nil {
		log.Fatalf("options failed: %v", err)
	}
	s := cewrap.NewSource(opts.downstream, opts.sink, nil, opts.changeMethods, newLogger())

	http.ListenAndServe(":"+opts.port, http.HandlerFunc(s.Handle))
}
