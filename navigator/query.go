package navigator

import "github.com/jtacoma/uritemplates"

type Params map[string]interface{}

// Find the href of a link by its relationship. Returns
// "" if a link doesn't exist.
func (l Links) Href(rel string) (string, error) {
	return l.HrefParams(rel, nil)
}

// Find the href of a link by its relationship, expanding any URI Template
// parameters with params. Returns "" if a link doesn't exist.
func (l Links) HrefParams(rel string, params map[string]interface{}) (string, error) {
	for relf, links := range l {
		if relf == rel {
			link := links[0] // TODO: handle multiple here
			template, err := uritemplates.Parse(link.Href)
			if err != nil {
				return "", err
			}

			return template.Expand(params)
		}
	}

	return "", nil
}
