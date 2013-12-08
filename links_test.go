package halgo

import (
	"encoding/json"
	"testing"
)

var hrefTests = []struct {
	name     string
	expected string
	url      string
}{
	{"normal", "/example", "/example"},
	{"parameterised", "/example", "/example{?q}"},
}

func TestHref(t *testing.T) {
	for _, test := range hrefTests {
		links := NewLinks(Link(test.name, test.url))
		href, err := links.Href(test.name)
		if err != nil {
			t.Error(err)
		}
		if href != test.expected {
			t.Errorf("%s: Expected href to be '%s', got '%s'", test.name, test.expected, href)
		}
	}
}

var hrefParamsTests = []struct {
	name     string
	expected string
	url      string
	params   Params
}{
	{"nil parameters", "/example", "/example{?q}", nil},
	{"empty parameters", "/example", "/example{?q}", Params{}},
	{"mismatched parameters", "/example", "/example{?q}", Params{"c": "test"}},
	{"single parameter", "/example?q=test", "/example{?q}", Params{"q": "test"}},
	{"multiple parameters", "/example?q=test&page=1", "/example{?q,page}", Params{"q": "test", "page": 1}},
}

func TestHrefParams(t *testing.T) {
	for _, test := range hrefParamsTests {
		links := NewLinks(Link(test.name, test.url))
		href, err := links.HrefParams(test.name, test.params)
		if err != nil {
			t.Error(err)
		}
		if href != test.expected {
			t.Errorf("%s: Expected href to be '%s', got '%s'", test.name, test.expected, href)
		}
	}
}

type MyResource struct {
	Links
	Name string
}

func TestMarshalLinksToJSON(t *testing.T) {
	res := MyResource{
		Name:  "James",
		Links: NewLinks(Link("self", "abc")),
	}

	b, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != `{"_links":{"self":{"href":"abc"}},"Name":"James"}` {
		t.Errorf("Unexpected JSON %s", b)
	}
}

func TestEmptyMarshalLinksToJSON(t *testing.T) {
	res := MyResource{
		Name:  "James",
		Links: Links{},
	}

	b, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != `{"Name":"James"}` {
		t.Errorf("Unexpected JSON %s", b)
	}
}

func TestUnmarshalLinksToJSON(t *testing.T) {
	res := MyResource{}
	err := json.Unmarshal([]byte(`{"_links":{"self":{"href":"abc"}},"Name":"James"}`), &res)
	if err != nil {
		t.Fatal(err)
	}

	if res.Name != "James" {
		t.Error("Expected name to be unmarshaled")
	}

	href, err := res.Href("self")
	if err != nil {
		t.Fatal(err)
	}
	if href != "abc" {
		t.Errorf("Expected self to be abc, got %s", href)
	}
}
