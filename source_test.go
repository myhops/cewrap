package cewrap

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/client/test"
)

func TestHandle(t *testing.T) {
	// Create a dummy server.
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("custom-header", "value")
		w.Write([]byte("Hi there"))
	}))
	defer svr.Close()

	// Create a dummy event sink.
	sink, echan := test.NewMockSenderClient(t, 1, client.WithUUIDs(), client.WithTimeNow())

	u, _ := url.Parse(svr.URL)
	s := &Source{
		downstream:    u,
		client:        &http.Client{},
		sink:          sink,
		changeMethods: DefaultChangeMethods,
		logger:        slog.Default(),
		source:        "https://testservice.example.com/testapi",
		// PathPrefix: "/testapi",
	}
	_ = s

	rr := httptest.NewRecorder()
	// req := httptest.NewRequest(http.MethodPost, "/testapi/path", bytes.NewBufferString("Hallo daar"))
	// s.Handler()(rr, req)
	buf := make([]byte, rr.Body.Len())
	rr.Body.Read(buf)
	// bs := string(buf)

	time.Sleep(time.Second)

	// Read the events from the channel
	var evt cloudevents.Event
	var text string
	select {
	case evt = <-echan:
		t.Logf("event received: %v", evt)
		text = string(evt.Data())

	default:
		t.Errorf("no event received")
	}
	t.Logf("done, %s", text)
}

func TestHandleWithWith(t *testing.T) {
	// Create a dummy server.
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("custom-header", "value")
		w.Write([]byte("Hi there"))
	}))
	defer svr.Close()

	// Create a dummy event sink.
	sink, echan := test.NewMockSenderClient(t, 1, client.WithUUIDs(), client.WithTimeNow())

	// u, _ := url.Parse(svr.URL)
	s := NewSource(
		WithDownstream(svr.URL),
		WithHTTPClient(&http.Client{}),
		WithSink(sink),
		WithChangeMethods(DefaultChangeMethods),
		// WithLogger(slog.Default()),
		WithSource("https://testservice.example.com/testapi"),
	)
	_ = s

	// s := &Source{
	// 	downstream:    u,
	// 	client:        &http.Client{},
	// 	sink:          sink,
	// 	changeMethods: DefaultChangeMethods,
	// 	logger:        slog.Default(),
	// 	source: "https://testservice.example.com/testapi",
	// 	// PathPrefix: "/testapi",
	// }

	rr := httptest.NewRecorder()
	// req := httptest.NewRequest(http.MethodPost, "/testapi/path", bytes.NewBufferString("Hallo daar"))
	// s.Handler()(rr, req)
	buf := make([]byte, rr.Body.Len())
	rr.Body.Read(buf)
	// bs := string(buf)

	time.Sleep(time.Second)

	// Read the events from the channel
	var evt cloudevents.Event
	var text string
	select {
	case evt = <-echan:
		t.Logf("event received: %v", evt)
		text = string(evt.Data())

	default:
		t.Errorf("no event received")
	}
	t.Logf("done, %s", text)
}
