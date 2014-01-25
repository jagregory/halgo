package navigator

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

func New(uri string) Navigator {
	return Navigator{
		rootUri:         uri,
		linksToNavigate: []link{},
		HttpClient:      http.DefaultClient,
	}
}

type link struct {
	rel    string
	params Params
}

type Navigator struct {
	HttpClient      HttpClient
	linksToNavigate []link
	rootUri         string
}

func (n Navigator) Link(rel string) Navigator {
	return n.LinkExpand(rel, nil)
}

func (n Navigator) LinkExpand(rel string, params Params) Navigator {
	links := make([]link, 0, len(n.linksToNavigate)+1)
	copy(n.linksToNavigate, links)
	links = append(links, link{rel: rel})

	return Navigator{
		HttpClient:      n.HttpClient,
		linksToNavigate: links,
		rootUri:         n.rootUri,
	}
}

func (n *Navigator) url() (string, error) {
	url := n.rootUri

	for _, link := range n.linksToNavigate {
		links, err := n.getLinks(url)
		if err != nil {
			return "", err
		}

		if _, ok := links[link.rel]; !ok {
			return "", LinkNotFoundError{link.rel}
		}

		url, err = links.HrefParams(link.rel, link.params)
		if err != nil {
			return "", err
		}
	}

	return url, nil
}

func (n Navigator) Get() (*http.Response, error) {
	url, err := n.url()
	if err != nil {
		return nil, err
	}

	return n.HttpClient.Get(url)
}

func (n Navigator) PostForm(data url.Values) (*http.Response, error) {
	url, err := n.url()
	if err != nil {
		return nil, err
	}

	return n.HttpClient.PostForm(url, data)
}

func (n Navigator) Patch(bodyType string, body io.Reader) (*http.Response, error) {
	url, err := n.url()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", bodyType)

	return n.HttpClient.Do(req)
}

func (n Navigator) Post(bodyType string, body io.Reader) (*http.Response, error) {
	url, err := n.url()
	if err != nil {
		return nil, err
	}

	return n.HttpClient.Post(url, bodyType, body)
}

func (n Navigator) Unmarshal(v interface{}) error {
	res, err := n.Get()
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &v)
}

type halResponseBody struct {
	Links `json:"_links"`
}

func (n Navigator) getLinks(uri string) (Links, error) {
	res, err := n.HttpClient.Get(uri)
	if err != nil {
		return Links{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return Links{}, err
	}

	var m halResponseBody

	if err := json.Unmarshal(body, &m); err != nil {
		return Links{}, err
	}

	return m.Links, nil
}
