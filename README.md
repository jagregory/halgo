# halgo

[HAL](http://stateless.co/hal_specification.html) implementation in Go.

## Install

    go get github.com/jagregory/halgo

## Usage

    import github.com/jagregory/halgo

Resource:

    type MyResource struct {
      halgo.Links
      Name string
    }

    res := MyResource{
      Links: halgo.NewLinks(
        halgo.Self("abc"),
      ),
      Name:  "James",
    }

JSON:

    {
      "_links": {
        "self": { "href": "/products/123" }
      },
      "Name": "Soap"
    }

## TODO

* Templated flag in hyperlink
* Curies
* Embedded