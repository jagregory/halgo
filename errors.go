package halgo

import "fmt"

// LinkNotFoundError is returned when a link with the specified relation
// couldn't be found in the links collection.
type LinkNotFoundError struct {
	rel string
}

func (err LinkNotFoundError) Error() string {
	return fmt.Sprintf("Response didn't contain link with relation: %s", err.rel)
}

// InvalidUrlError is returned when a link contains a malformed or invalid
// url.
type InvalidUrlError struct {
	url string
}

func (err InvalidUrlError) Error() string {
	return fmt.Sprintf("Invalid URL: %s", err.url)
}
