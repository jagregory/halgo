# halgo

[HAL](http://stateless.co/hal_specification.html) implementation in Go.

> HAL is a simple format that gives a consistent and easy way to hyperlink between resources in your API.

Halgo helps with generating HAL-compliant JSON from Go structs, and
provides a Navigator for walking a HAL-compliant API.

[![GoDoc](https://godoc.org/github.com/jagregory/halgo?status.png)](https://godoc.org/github.com/jagregory/halgo)

## Install

    go get github.com/jagregory/halgo

## Usage

Serialising a resource with HAL links:

```go
import "github.com/jagregory/halgo"

type MyResource struct {
  halgo.Links
  Name string
}

res := MyResource{
  Links: Links{}.
    Self("/orders").
    Next("/orders?page=2").
    Link("ea:find", "/orders{?id}").
    Add("ea:admin", Link{Href: "/admins/2", Title: "Fred"}, Link{Href: "/admins/5", Title: "Kate"}),
  Name: "James",
}

bytes, _ := json.Marshal(res)

fmt.Println(bytes)

// {
//   "_links": {
//     "self": { "href": "/orders" },
//     "next": { "href": "/orders?page=2" },
//     "ea:find": { "href": "/orders{?id}", "templated": true },
//     "ea:admin": [{
//         "href": "/admins/2",
//         "title": "Fred"
//     }, {
//         "href": "/admins/5",
//         "title": "Kate"
//     }]
//   },
//   "Name": "James"
// }
```

Navigating a HAL-compliant API:

```go
res, err := halgo.Navigator("http://example.com").
  Follow("products").
  Followf("page", halgo.P{"n": 10}).
  Get()
```

The following operations can be chained together to navigate a HAL-compliant API:

   * `Follow(rel string)` - Follow the relation `rel`
   * `Followf(rel string, params P)` - Follow the relation `rel` and
     use `params` to evaluate any underlying template.
   * `Extract(rel string)` - Fetch the location of an embedded
     resource named `rel` and navigate to the full representation of
     that resource (i.e., follow its `self` URI).
   * `SetSessionHeader(header string, value string)` - Set a new
     header to all requests in this chain (e.g., `Authorization` header)
   * `AddSessionHeader(header string, value string)` - Add a new
     header to all requests in this chain.
   * `SetRequestHeader(header string, value string)` - Set a new
     header to the operation immediately preceeding this call in the chain.
   * `AddRequestHeader(header string, value string)` - Add a new
     header to the operation immediately preceeding this call in the chain.

These chains can be terminated by the following requests (each of
which returns an `*http.Response` or `error`):

   * `Get(headers ...http.Header)` - Perform a `GET` request (with optional additional headers)
   * `Options(headers ...http.Header)` - Perform an `OPTIONS` request
     (with optional additional headers)
   * `Post(bodyType string, body io.Reader, headers ...http.Header)` - Perform a
     `POST` request with `body` of
	 content type `bodyType` (and optional additional headers)
   * `PostForm(data url.Values, headers ...http.Header)` - Perform a
     `POST` request with form `data` (and optional additional
     headers)
   * `Patch(bodyType string, body io.Reader, headers
	 ...http.Header)` - Perform a `PATCH` request with `body` of
	 content type `bodyType` (and optional additional headers)
   * `Put(bodyType string, body io.Reader, headers
	 ...http.Header)` - Perform a `PUT` request with `body` of
	 content type `bodyType` (and optional additional headers)
   * `Delete(headers ...http.Header)` - Perform a `DELETE` request
     (with optional additional headers)

In addition, following any request that returns a `Location` header
(typically a `Post`), a new navigator instance can be created that is
rooted at the location specified by the `Location` header by calling:

   * `Location(resp *http.Response)` - which returns either a new
	 navigator or an error

Deserialising a resource: 

```go
import "github.com/jagregory/halgo"

type MyResource struct {
  halgo.Links
  Name string
}

data := []byte(`{
  "_links": {
    "self": { "href": "/orders" },
    "next": { "href": "/orders?page=2" },
    "ea:find": { "href": "/orders{?id}", "templated": true },
    "ea:admin": [{
        "href": "/admins/2",
        "title": "Fred"
    }, {
        "href": "/admins/5",
        "title": "Kate"
    }]
  },
  "Name": "James"
}`)

res := MyResource{}
json.Unmarshal(data, &res)

res.Name // "James"
res.Links.Href("self") // "/orders"
res.Links.HrefParams("self", Params{"id": 123}) // "/orders?id=123"
```

## TODO

* Curies
* Embedded resources
