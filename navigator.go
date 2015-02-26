package halgo

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Navigator is a mechanism for navigating HAL-compliant REST APIs. You
// start by creating a Navigator with a base URI, then Follow the links
// exposed by the API until you reach the place where you want to perform
// an action.
//
// For example, to request an API exposed at api.example.com and follow a
// link named products and GET the resulting page you'd do this:
//
//     res, err := Navigator("http://api.example.com").
//       Follow("products").
//       Get()
//
// To do the same thing but POST to the products page, you'd do this:
//
//     res, err := Navigator("http://api.example.com").
//       Follow("products").
//       Post("application/json", someContent)
//
// Multiple links followed in sequence.
//
//     res, err := Navigator("http://api.example.com").
//       Follow("products").
//       Follow("next")
//       Get()
//
// Links can also be expanded with Followf if they are URI templates.
//
//     res, err := Navigator("http://api.example.com").
//       Follow("products").
//       Followf("page", halgo.P{"number": 10})
//       Get()
//
// Navigation of relations is lazy. Requests will only be triggered when
// you execute a method which returns a result. For example, this doesn't
// perform any HTTP requests.
//
//     Navigator("http://api.example.com").
//       Follow("products")
//
// It's only when you add a call to Get, Post, PostForm, Patch, or
// Unmarshal to the end will any requests be triggered.
//
// By default a Navigator will use http.DefaultClient as its mechanism for
// making HTTP requests. If you want to supply your own HttpClient, you
// can assign to nav.HttpClient after creation.
//
//     nav := Navigator("http://api.example.com")
//     nav.HttpClient = MyHttpClient{}
//
// Any Client you supply must implement halgo.HttpClient, which
// http.Client does implicitly. By creating decorators for the HttpClient,
// logging and caching clients are trivial. See LoggingHttpClient for an
// example.
func Navigator(uri string) navigator {
	return navigator{
		rootUri:       uri,
		path:          []Operation{},
		sessionHeader: http.Header{},
		HttpClient:    http.DefaultClient,
	}
}

type Operation interface {
	Fetch(n navigator, url string) (string, error)
	SetHeader(header string, value string)
	AddHeader(header string, value string)
	// Not sure yet
}

// navigator is the API navigator
type navigator struct {
	// HttpClient is used to execute requests. By default it's
	// http.DefaultClient. By decorating a HttpClient instance you can
	// easily write loggers or caching mechanisms.
	HttpClient HttpClient

	// sessionHeader is a structure used to store the headers for
	// all requests made by the navigator (as opposed to those
	// that might be included on a per request basis)
	sessionHeader http.Header

	// path is the follow queue.
	path []Operation

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
	relations := append([]Operation{}, n.path...)
	relations = append(relations, &follow{
		rel:    rel,
		params: params,
		header: http.Header{},
	})

	return navigator{
		HttpClient:    n.HttpClient,
		sessionHeader: n.cloneHeader(),
		path:          relations,
		rootUri:       n.rootUri,
	}
}

// Follow adds a relation to the follow queue of the navigator.
func (n navigator) Extract(rel string) navigator {
	relations := append([]Operation{}, n.path...)
	relations = append(relations, &extract{
		rel:    rel,
		header: http.Header{},
	})

	return navigator{
		HttpClient:    n.HttpClient,
		sessionHeader: n.cloneHeader(),
		path:          relations,
		rootUri:       n.rootUri,
	}
}

// cloneHeader makes a new copy of the sessionHeaders.  This allows each
// navigator generated as a chain of operations to be truly
// independent of each other (and potentially diverge)
func (n navigator) cloneHeader() http.Header {
	h2 := make(http.Header, len(n.sessionHeader))
	for k, vv := range n.sessionHeader {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}

// SetSessionHeader sets a header to all requests in the chain
func (n navigator) SetSessionHeader(header string, value string) navigator {
	h := n.cloneHeader()
	h.Set(header, value)
	return navigator{
		HttpClient:    n.HttpClient,
		sessionHeader: h,
		path:          n.path,
		rootUri:       n.rootUri,
	}
}

// AddSessionHeader adds a header to all requests in the chain
func (n navigator) AddSessionHeader(header string, value string) navigator {
	h := n.cloneHeader()
	h.Add(header, value)
	return navigator{
		HttpClient:    n.HttpClient,
		sessionHeader: h,
		path:          n.path,
		rootUri:       n.rootUri,
	}
}

func (n navigator) SetRequestHeader(header string, value string) navigator {
	if len(n.path) == 0 {
		// No way to return an error, but this is almost certainly an error
		// because this does nothing (since there is no relation to add
		// this header to.
		return n
	}
	n.path[len(n.path)-1].SetHeader(header, value)
	return n
}

// AddFollowHeader adds a header to for a specific follow command
func (n navigator) AddRequestHeader(header string, value string) navigator {
	if len(n.path) == 0 {
		// No way to return an error, but this is almost certainly an error
		// because this does nothing (since there is no relation to add
		// this header to.
		return n
	}
	n.path[len(n.path)-1].AddHeader(header, value)
	return n
}

// Location follows the Location header from a response.  It makes the URI
// absolute, if necessary.
func (n navigator) Location(resp *http.Response) (navigator, error) {
	_, exists := resp.Header["Location"]
	if !exists {
		return n, fmt.Errorf("Response didn't contain a Location header")
	}
	loc := resp.Header.Get("Location")
	lurl, err := makeAbsoluteIfNecessary(loc, n.rootUri)
	if err != nil {
		return n, err
	}
	return navigator{
		HttpClient:    n.HttpClient,
		sessionHeader: n.cloneHeader(),
		path:          []Operation{},
		rootUri:       lurl,
	}, nil
}

func mergeHeaders(req *http.Request, headers ...http.Header) {
	for _, header := range headers {
		for k, vs := range header {
			for _, v := range vs {
				req.Header.Add(k, v)
			}
		}
	}
}

// url returns the URL of the tip of the follow queue. Will follow the
// usual pattern of requests.
func (n navigator) url() (string, error) {
	var err error
	url := n.rootUri

	for _, link := range n.path {
		url, err = link.Fetch(n, url)
		if err != nil {
			return "", err
		}
		url, err = makeAbsoluteIfNecessary(url, n.rootUri)
		if err != nil {
			return "", fmt.Errorf("Error making url %s absolute: %v", url, err)
		}
	}

	return url, nil
}

// makeAbsoluteIfNecessary takes the current url and the root url, and
// will make the current URL absolute by using the root's Host, Scheme,
// and credentials if current isn't already absolute.
func makeAbsoluteIfNecessary(current, root string) (string, error) {
	currentUri, err := url.Parse(current)
	if err != nil {
		return "", err
	}

	if currentUri.IsAbs() {
		return current, nil
	}

	rootUri, err := url.Parse(root)
	if err != nil {
		return "", err
	}

	currentUri.Scheme = rootUri.Scheme
	currentUri.Host = rootUri.Host
	currentUri.User = rootUri.User

	return currentUri.String(), nil
}

// Get performs a GET request on the tip of the follow queue.
//
// When a navigator is evaluated it will first request the root, then
// request each relation on the queue until it reaches the tip. Once the
// tip is reached it will defer to the calling method. In the case of GET
// the last request will just be returned. For Post it will issue a post
// to the URL of the last relation. Any error along the way will terminate
// the walk and return immediately.
func (n navigator) Get(headers ...http.Header) (*http.Response, error) {
	url, err := n.url()
	if err != nil {
		return nil, err
	}

	req, err := newHalRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	headers = append([]http.Header{n.sessionHeader}, headers...)
	mergeHeaders(req, headers...)

	return n.HttpClient.Do(req)
}

// Options performs an OPTIONS request on the tip of the follow queue.
func (n navigator) Options(headers ...http.Header) (*http.Response, error) {
	url, err := n.url()
	if err != nil {
		return nil, err
	}

	req, err := newHalRequest("OPTIONS", url, nil)
	if err != nil {
		return nil, err
	}

	headers = append([]http.Header{n.sessionHeader}, headers...)
	mergeHeaders(req, headers...)

	return n.HttpClient.Do(req)
}

// PostForm performs a POST request on the tip of the follow queue with
// the given form data.
//
// See GET for a note on how the navigator executes requests.
func (n navigator) PostForm(data url.Values, headers ...http.Header) (*http.Response, error) {
	url, err := n.url()
	if err != nil {
		return nil, err
	}

	req, err := newHalRequest("PATCH", url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	headers = append([]http.Header{n.sessionHeader}, headers...)
	mergeHeaders(req, headers...)

	return n.HttpClient.Do(req)
}

// Patch parforms a PATCH request on the tip of the follow queue with the
// given bodyType and body content.
//
// See GET for a note on how the navigator executes requests.
func (n navigator) Patch(bodyType string, body io.Reader, headers ...http.Header) (*http.Response, error) {
	url, err := n.url()
	if err != nil {
		return nil, err
	}

	req, err := newHalRequest("PATCH", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", bodyType)

	headers = append([]http.Header{n.sessionHeader}, headers...)
	mergeHeaders(req, headers...)

	return n.HttpClient.Do(req)
}

// Put parforms a PUT request on the tip of the follow queue with the
// given bodyType and body content.
//
// See GET for a note on how the navigator executes requests.
func (n navigator) Put(bodyType string, body io.Reader, headers ...http.Header) (*http.Response, error) {
	url, err := n.url()
	if err != nil {
		return nil, err
	}

	req, err := newHalRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", bodyType)

	headers = append([]http.Header{n.sessionHeader}, headers...)
	mergeHeaders(req, headers...)

	return n.HttpClient.Do(req)
}

// Post performs a POST request on the tip of the follow queue with the
// given bodyType and body content.
//
// See GET for a note on how the navigator executes requests.
func (n navigator) Post(bodyType string, body io.Reader, headers ...http.Header) (*http.Response, error) {
	url, err := n.url()
	if err != nil {
		return nil, err
	}

	req, err := newHalRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", bodyType)

	headers = append([]http.Header{n.sessionHeader}, headers...)
	mergeHeaders(req, headers...)

	return n.HttpClient.Do(req)
}

// Delete performs a DELETE request on the tip of the follow queue.
//
// See GET for a note on how the navigator executes requests.
func (n navigator) Delete(headers ...http.Header) (*http.Response, error) {
	url, err := n.url()
	if err != nil {
		return nil, err
	}

	req, err := newHalRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	headers = append([]http.Header{n.sessionHeader}, headers...)
	mergeHeaders(req, headers...)

	return n.HttpClient.Do(req)
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

func newHalRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/hal+json, application/json")

	return req, nil
}

// getLinks does a GET on a particular URL and try to deserialise it into
// a HAL links collection.
func (n navigator) getLinks(uri string, requestHeader http.Header) (Links, error) {
	req, err := newHalRequest("GET", uri, nil)
	if err != nil {
		return Links{}, err
	}

	mergeHeaders(req, n.sessionHeader, requestHeader)

	res, err := n.HttpClient.Do(req)
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
		return Links{}, fmt.Errorf("Unable to unmarshal '%s': %v", string(body), err)
	}

	return m, nil
}

// getEmbedded does a GET on a particular URL and try to deserialise it into
// a HAL representation and returns the uri for a particular embedded resource
func (n navigator) getEmbedded(uri string, rel string, requestHeader http.Header) (string, error) {
	req, err := newHalRequest("GET", uri, nil)
	if err != nil {
		return "", err
	}

	mergeHeaders(req, n.sessionHeader, requestHeader)

	res, err := n.HttpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error requesting embedded resources: %v", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading request body: %v", err)
	}

	var m Embeds

	if err := json.Unmarshal(body, &m); err != nil {
		return "", fmt.Errorf("Unable to unmarshal '%s': %v", string(body), err)
	}

	link, ok := m.Resources[rel]
	if !ok {
		return "", fmt.Errorf("Request body '%s' doesn't contain embedded resource %s",
			string(body), rel)
	}

	self, err := link.Href("self")
	if err != nil {
		return "", fmt.Errorf("Error extracting self href: %v", err)
	}

	return self, nil
}
