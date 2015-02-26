package halgo

import (
	"net/http"
)

type extract struct {
	rel    string
	header http.Header
}

func (link *extract) AddHeader(header string, value string) {
	link.header.Add(header, value)
}

func (link *extract) SetHeader(header string, value string) {
	link.header.Set(header, value)
}

func (link extract) Fetch(n navigator, url string) (string, error) {
	return n.getEmbedded(url, link.rel, link.header)
}

var _ Operation = (*extract)(nil) // Static check on interface
