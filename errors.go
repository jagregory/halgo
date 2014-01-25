package halgo

import "fmt"

type LinkNotFoundError struct {
	rel string
}

func (err LinkNotFoundError) Error() string {
	return fmt.Sprintf("Response didn't contain link with relation: %s", err.rel)
}
