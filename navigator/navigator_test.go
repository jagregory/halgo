package navigator

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
)

func createTestHttpServer() *httptest.Server {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"_links":{}}`)
	})

	return httptest.NewServer(r)
}

func TestNavigatingToUnknownLink(t *testing.T) {
	ts := createTestHttpServer()
	defer ts.Close()

	nav := New(ts.URL)
	nav.HttpClient = LoggingHttpClient{nav.HttpClient}
	_, err := nav.Link("missing").Get()
	if err == nil {
		t.Fatal("Expected error to be raised for missing link")
	}

	if err.Error() != "Response didn't contain link with relation: missing" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}

	if _, ok := err.(LinkNotFoundError); !ok {
		t.Error("Expected error to be LinkNotFoundError")
	}
}
