package halgo

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
)

func createTestHttpServer() (*httptest.Server, map[string]int) {
	r := mux.NewRouter()
	hits := make(map[string]int)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hits["/"] += 1
		fmt.Fprintf(w, `{
      "_links": {
        "next": { "href": "http://%s/2nd" },
        "one": { "href": "http://%s/a/{id}", "templated": true }
      }
    }`, r.Host, r.Host)
	})

	r.HandleFunc("/2nd", func(w http.ResponseWriter, r *http.Request) {
		hits["/2nd"] += 1
		fmt.Sprintln(w, "OK")
		w.WriteHeader(200)
	})

	r.HandleFunc("/a/{id}", func(w http.ResponseWriter, r *http.Request) {
		hits["/a/"+mux.Vars(r)["id"]] += 1
		fmt.Sprintln(w, "OK")
		w.WriteHeader(200)
	})

	return httptest.NewServer(r), hits
}

func TestNavigatingToUnknownLink(t *testing.T) {
	ts, _ := createTestHttpServer()
	defer ts.Close()

	_, err := Navigator(ts.URL).Follow("missing").Get()
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

func TestGettingTheRoot(t *testing.T) {
	ts, hits := createTestHttpServer()
	defer ts.Close()

	nav := Navigator(ts.URL)
	res, err := nav.Get()
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected OK, got %d", res.StatusCode)
	}

	if res.Request.URL.String() != ts.URL {
		t.Errorf("Expected url to be %s, got %s", ts.URL, res.Request.URL)
	}

	if hits["/"] != 1 {
		t.Errorf("Expected 1 request to /, got %d", hits["/"])
	}
}

func TestFollowingALink(t *testing.T) {
	ts, hits := createTestHttpServer()
	defer ts.Close()

	nav := Navigator(ts.URL).Followf("one", P{"id": 1})
	res, err := nav.Get()
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected OK, got %d", res.StatusCode)
	}

	if res.Request.URL.String() != ts.URL+"/a/1" {
		t.Errorf("Expected url to be %s, got %s", ts.URL+"/a/1", res.Request.URL)
	}

	if hits["/"] != 1 {
		t.Errorf("Expected 1 request to /, got %d", hits["/"])
	}

	if hits["/a/1"] != 1 {
		t.Errorf("Expected 1 request to /a/1, got %d", hits["/a/1"])
	}
}

func TestFollowingATemplatedLink(t *testing.T) {
	ts, hits := createTestHttpServer()
	defer ts.Close()

	nav := Navigator(ts.URL).Follow("next")
	res, err := nav.Get()
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected OK, got %d", res.StatusCode)
	}

	if res.Request.URL.String() != ts.URL+"/2nd" {
		t.Errorf("Expected url to be %s, got %s", ts.URL+"/2nd", res.Request.URL)
	}

	if hits["/"] != 1 {
		t.Errorf("Expected 1 request to /, got %d", hits["/"])
	}

	if hits["/2nd"] != 1 {
		t.Errorf("Expected 1 request to /2nd, got %d", hits["/2nd"])
	}
}
