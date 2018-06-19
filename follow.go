package halgo

import (
	"fmt"
	"net/http"
)

type follow struct {
	rel    string
	params P
	header http.Header
}

func (link *follow) AddHeader(header string, value string) {
	link.header.Add(header, value)
}

func (link *follow) SetHeader(header string, value string) {
	link.header.Set(header, value)
}

func (link follow) Fetch(n Nav, url string) (string, error) {
	links, err := n.getLinks(url, link.header)
	if err != nil {
		return "", fmt.Errorf("Error getting links (%s, %v): %v", url, links, err)
	}

	if _, ok := links.Items[link.rel]; !ok {
		return "", LinkNotFoundError{link.rel, links.Items}
	}

	url, err = links.HrefParams(link.rel, link.params)
	if err != nil {
		return "", fmt.Errorf("Error getting url (%v, %v): %v", link.rel, link.params, err)
	}

	if url == "" {
		return "", InvalidUrlError{url}
	}

	if err != nil {
		return "", fmt.Errorf("Error making url absolute: %v", err)
	}

	return url, nil
}

var _ Operation = (*follow)(nil) // Static check on interface
