package halgo

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Navigator creates a new API navigator for the given URI. By default
// it will use http.DefaultClient as its mechanism for navigating
// relations.
//
//     nav := Navigator("http://example.com")
//
// If you want to supply your own navigator, just assign HttpClient after
// creation.
//
//     nav := Navigator("http://example.com")
//     nav.HttpClient = MyHttpClient{}
//
// Any Client you supply must implement halgo.HttpClient, which
// http.Client does implicitly.
//
// By creating decorators for the HttpClient, logging and caching clients
// are trivial. See LoggingHttpClient for an example.
func Navigator(uri string) navigator {
	return navigator{
		rootUri:    uri,
		path:       []relation{},
		HttpClient: http.DefaultClient,
	}
}

// relation is an instruction of a relation to follow and any params to
// expand with when executed.
type relation struct {
	rel    string
	params P
}

// navigator is the API navigator
type navigator struct {
	// HttpClient is used to execute requests. By default it's
	// http.DefaultClient. By decorating a HttpClient instance you can
	// easily write loggers or caching mechanisms.
	HttpClient HttpClient

	// path is the follow queue.
	path []relation

	// rootUri is where the navigation will begin from.
	rootUri string
}

// Follow adds a relation to the follow queue of the navigator.
func (n navigator) Follow(rel string) navigator {
	return n.Followf(rel, nil)
}

// Followf adds a relation to the follow queue of the navigator, with a
// set of parameters to expand on execution.
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

// url returns the URL of the tip of the follow queue. Will follow the
// usual pattern of requests.
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

// Get performs a GET request on the tip of the follow queue.
//
// When a navigator is evaluated it will first request the root, then
// request each relation on the queue until it reaches the tip. Once the
// tip is reached it will defer to the calling method. In the case of GET
// the last request will just be returned. For Post it will issue a post
// to the URL of the last relation. Any error along the way will terminate
// the walk and return immediately.
func (n navigator) Get() (*http.Response, error) {
	url, err := n.url()
	if err != nil {
		return nil, err
	}

	return n.HttpClient.Get(url)
}

// PostForm performs a POST request on the tip of the follow queue with
// the given form data.
//
// See GET for a note on how the navigator executes requests.
func (n navigator) PostForm(data url.Values) (*http.Response, error) {
	url, err := n.url()
	if err != nil {
		return nil, err
	}

	return n.HttpClient.PostForm(url, data)
}

// Patch parforms a PATCH request on the tip of the follow queue with the
// given bodyType and body content.
//
// See GET for a note on how the navigator executes requests.
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

// Post performs a POST request on the tip of the follow queue with the
// given bodyType and body content.
//
// See GET for a note on how the navigator executes requests.
func (n navigator) Post(bodyType string, body io.Reader) (*http.Response, error) {
	url, err := n.url()
	if err != nil {
		return nil, err
	}

	return n.HttpClient.Post(url, bodyType, body)
}

// Unmarshal is a shorthand for Get followed by json.Unmarshal. Handles
// closing the response body and unmarshalling the body.
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

// getLinks does a GET on a particular URL and try to deserialise it into
// a HAL links collection.
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
