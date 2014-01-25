package halgo

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Interface exposing the core request generating http.Client methods
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
	Head(url string) (*http.Response, error)
	Post(url string, bodyType string, body io.Reader) (*http.Response, error)
	PostForm(url string, data url.Values) (*http.Response, error)
}

type LoggingHttpClient struct {
	HttpClient
}

func (c LoggingHttpClient) Do(req *http.Request) (*http.Response, error) {
	fmt.Printf("%s %s\n", req.Method, req.URL)
	return c.HttpClient.Do(req)
}

func (c LoggingHttpClient) Get(url string) (*http.Response, error) {
	fmt.Printf("GET %s\n", url)
	return c.HttpClient.Get(url)
}

func (c LoggingHttpClient) Head(url string) (*http.Response, error) {
	fmt.Printf("HEAD %s\n", url)
	return c.HttpClient.Head(url)
}

func (c LoggingHttpClient) Post(url string, bodyType string, body io.Reader) (*http.Response, error) {
	fmt.Printf("POST %s\n", url)
	return c.HttpClient.Post(url, bodyType, body)
}

func (c LoggingHttpClient) PostForm(url string, data url.Values) (*http.Response, error) {
	fmt.Printf("POST %s\n", url)
	return c.HttpClient.PostForm(url, data)
}
