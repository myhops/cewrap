package cewrap

import (
	"bytes"
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
		Downstream:    u,
		Client:        &http.Client{},
		Sink:          sink,
		ChangeMethods: DefaultChangeMethods,
		Logger:        slog.Default(),
		Source: "https://testservice.example.com/testapi",
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/testapi/path", bytes.NewBufferString("Hallo daar"))
	s.Handler()(rr, req)
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
