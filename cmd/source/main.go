package main

import (
	"log"
	"net/http"

	"github.com/myhops/cewrap"
)

func main() {
	opts, err := getOptions()
	if err != nil {
		log.Fatalf("options failed: %v", err)
	}
	s := cewrap.NewSource(opts.downstream, opts.sink, nil, nil)

	http.ListenAndServe(":"+opts.port, http.HandlerFunc(s.Handle))
}
