package halgo

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

func Navigator(uri string) navigator {
	return navigator{
		rootUri:    uri,
		path:       []relation{},
		HttpClient: http.DefaultClient,
	}
}

type relation struct {
	rel    string
	params P
}

type navigator struct {
	HttpClient HttpClient
	path       []relation
	rootUri    string
}

func (n navigator) Follow(rel string) navigator {
	return n.Followf(rel, nil)
}

func (n navigator) Followf(rel string, params P) navigator {
	relations := make([]relation, 0, len(n.path)+1)
	copy(n.path, relations)
	relations = append(relations, relation{rel: rel, params: params})

	return navigator{
		HttpClient: n.HttpClient,
		path:       relations,
		rootUri:    n.rootUri,
	}
}

func (n navigator) url() (string, error) {
	url := n.rootUri

	for _, link := range n.path {
		links, err := n.getLinks(url)
		if err != nil {
			return "", err
		}

		if _, ok := links.Items[link.rel]; !ok {
			return "", LinkNotFoundError{link.rel}
		}

		url, err = links.HrefParams(link.rel, link.params)
		if err != nil {
			return "", err
		}
	}

	return url, nil
}

func (n navigator) Get() (*http.Response, error) {
	url, err := n.url()
	if err != nil {
		return nil, err
	}

	return n.HttpClient.Get(url)
}

func (n navigator) PostForm(data url.Values) (*http.Response, error) {
	url, err := n.url()
	if err != nil {
		return nil, err
	}

	return n.HttpClient.PostForm(url, data)
}

func (n navigator) Patch(bodyType string, body io.Reader) (*http.Response, error) {
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

func (n navigator) Post(bodyType string, body io.Reader) (*http.Response, error) {
	url, err := n.url()
	if err != nil {
		return nil, err
	}

	return n.HttpClient.Post(url, bodyType, body)
}

func (n navigator) Unmarshal(v interface{}) error {
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

func (n navigator) getLinks(uri string) (Links, error) {
	res, err := n.HttpClient.Get(uri)
	if err != nil {
		return Links{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return Links{}, err
	}

	var m Links

	if err := json.Unmarshal(body, &m); err != nil {
		return Links{}, err
	}

	return m, nil
}
