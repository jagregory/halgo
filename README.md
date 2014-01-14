# halgo

[HAL](http://stateless.co/hal_specification.html) implementation in Go.

> HAL is a simple format that gives a consistent and easy way to hyperlink between resources in your API.

This library helps with serialising and deserialising structures containing embedded links to other resources.

## Install

    go get github.com/jagregory/halgo

## Usage

Serialising a resource with HAL links:

```go
import github.com/jagregory/halgo

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

Deserialising and querying a resource: 

```go
import github.com/jagregory/halgo

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
