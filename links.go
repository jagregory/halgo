package halgo

import (
	"encoding/json"
	"fmt"
	"github.com/jtacoma/uritemplates"
)

type Links struct {
	Items HyperlinkSet `json:"_links,omitempty"`
	// Curies CurieSet
}

// Create a set of hyperlinks
func NewLinks(links ...Hyperlink) Links {
	return Links{Items: links}
}

// A hyperlink with a href/URL and a relationship
type Hyperlink struct {
	Rel  string
	Href string
}

// Create a hyperlink to a URL
func Link(rel string, href string) Hyperlink {
	return Hyperlink{rel, href}
}

// Create a hyperlink to a URL with a format string
func Linkf(rel string, format string, args ...interface{}) Hyperlink {
	return Link(rel, fmt.Sprintf(format, args...))
}

// Create a rel:self hyperlink to a url
func Self(href string) Hyperlink {
	return Link("self", href)
}

// Create a rel:self hyperlink with a format string
func Selff(format string, args ...interface{}) Hyperlink {
	return Self(fmt.Sprintf(format, args...))
}

// Set of hyperlinks
type HyperlinkSet []Hyperlink

func (l HyperlinkSet) MarshalJSON() ([]byte, error) {
	out := make(map[string]map[string]string)

	for _, link := range l {
		out[link.Rel] = map[string]string{"href": link.Href}
	}

	return json.Marshal(out)
}

func (l *HyperlinkSet) UnmarshalJSON(d []byte) error {
	out := make(map[string]map[string]string)

	if err := json.Unmarshal(d, &out); err != nil {
		return err
	}

	for rel, link := range out {
		*l = append(*l, Hyperlink{rel, link["href"]})
	}

	return nil
}

type Params map[string]interface{}

// Find the href of a link by its relationship. Returns
// "" if a link doesn't exist.
func (l Links) Href(rel string) (string, error) {
	return l.HrefParams(rel, nil)
}

// Find the href of a link by its relationship, expanding any URI Template
// parameters with params. Returns "" if a link doesn't exist.
func (l Links) HrefParams(rel string, params map[string]interface{}) (string, error) {
	for _, link := range l.Items {
		if link.Rel == rel {
			println(link.Href)
			template, err := uritemplates.Parse(link.Href)
			if err != nil {
				return "", err
			}

			return template.Expand(params)
		}
	}

	return "", nil
}
