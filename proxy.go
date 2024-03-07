package cewrap

// import (
// 	"context"
// 	"fmt"
// 	"log/slog"
// 	"net/http"
// 	"net/http/httputil"
// 	"net/url"
// 	"reflect"
// 	"slices"
// )

// type ProxyConfig struct {
// 	// Methods that should trigger an event.
// 	//
// 	// An empty or nil array triggers on all methods.
// 	Methods []string
// 	// Paths that should trigger an event.
// 	//
// 	// An empty or nil array triggers on all paths.
// 	Paths []string
// 	// Logger for standard logging.
// 	Logger *slog.Logger
// 	// Emitters emit events.
// 	Emitters []Emitter
// 	// Upstream service.
// 	Upstream string
// }

// type Proxy struct {
// 	rp *httputil.ReverseProxy

// 	selector http.ServeMux
// 	methods  []string
// 	upstream *url.URL
// }

// func noopHandlerFunc(_ http.ResponseWriter, _ *http.Request) {}

// func NewProxy(cfg *ProxyConfig) (*Proxy, error) {
// 	// Create the proxy
// 	p := &Proxy{}

// 	// Prepare the reverse proxy.
// 	rp := &httputil.ReverseProxy{
// 		Rewrite:        p.startTap(),
// 		ModifyResponse: p.endTap(),
// 	}

// 	s := http.NewServeMux()
// 	for _, p := range cfg.Paths {
// 		s.HandleFunc(p, noopHandlerFunc)
// 	}
// 	p.methods = append(p.methods, cfg.Methods...)

// 	//
// 	u, err := url.Parse(cfg.Upstream)
// 	if err != nil {
// 		return nil, err
// 	}
// 	p.upstream = u

// 	p.rp = rp
// 	return p, nil
// }

// func (p *Proxy) Handle() http.HandlerFunc {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		p.rp.ServeHTTP(w, r)
// 	})
// }

// func (p *Proxy) getTap(pr *httputil.ProxyRequest) Tap {
// 	var handleThis bool
// 	// Check if we need to handle this request.
// 	_, path := p.selector.Handler(pr.In)
// 	handleThis = len(path) > 0
// 	// Check if we need to handle this based on the method.
// 	handleThis = handleThis || slices.Contains(p.methods, pr.In.Method)

// 	t := newNoopTap()
// 	return t
// }

// func (p *Proxy) startTap() func(pr *httputil.ProxyRequest) {
// 	return func(pr *httputil.ProxyRequest) {
// 		// Set the upstream service.
// 		pr.SetURL(p.upstream)

// 		// Get the tap for this request.
// 		t := p.getTap(pr)
// 		// insert the tap into the context and start it.
// 		pr.Out = pr.Out.Clone(putTap(pr.Out.Context(), t))
// 		t.Start(pr)
// 	}
// }

// func (p *Proxy) endTap() func(r *http.Response) error {
// 	return func(r *http.Response) error {
// 		// Get the tap from the request context.
// 		t := getTap(r.Request.Context())
// 		if t == nil {
// 			return nil
// 		}
// 		return t.End(r)
// 	}
// }
