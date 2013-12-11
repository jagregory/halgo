package halgo

type Links struct {
	Items map[string]LinkSet `json:"_links,omitempty"`
	// Curies CurieSet
}

func (l Links) Self(href string) Links {
	return l.Link("self", href)
}

func (l Links) Next(href string) Links {
	return l.Link("next", href)
}

func (l Links) Link(rel, href string) Links {
	return l.Add(rel, Link{Href: href})
}

func (l Links) Add(rel string, links ...Link) Links {
	if l.Items == nil {
		l.Items = make(map[string]LinkSet)
	}

	set, exists := l.Items[rel]

	if exists {
		set = append(set, links...)
	} else {
		set = make([]Link, len(links))
		copy(set, links)
	}

	l.Items[rel] = set

	return l
}

// A Link with a href/URL and a relationship
type Link struct {
	// The "href" property is REQUIRED.
	// Its value is either a URI [RFC3986] or a URI Template [RFC6570].
	// If the value is a URI Template then the Link Object SHOULD have a
	// "templated" attribute whose value is true.
	Href string `json:"href"`

	// The "templated" property is OPTIONAL.
	// Its value is boolean and SHOULD be true when the Link Object's "href"
	// property is a URI Template.
	// Its value SHOULD be considered false if it is undefined or any other
	// value than true.
	Templated bool `json:"templated,omitempty"`

	// The "type" property is OPTIONAL.
	// Its value is a string used as a hint to indicate the media type
	// expected when dereferencing the target resource.
	Type string `json:"type,omitempty"`

	// The "deprecation" property is OPTIONAL.
	// Its presence indicates that the link is to be deprecated (i.e.
	// removed) at a future date.  Its value is a URL that SHOULD provide
	// further information about the deprecation.
	// A client SHOULD provide some notification (for example, by logging a
	// warning message) whenever it traverses over a link that has this
	// property.  The notification SHOULD include the deprecation property's
	// value so that a client manitainer can easily find information about
	// the deprecation.
	Deprecation string `json:"deprecation,omitempty"`

	// The "name" property is OPTIONAL.
	// Its value MAY be used as a secondary key for selecting Link Objects
	// which share the same relation type.
	Name string `json:"name,omitempty"`

	// The "profile" property is OPTIONAL.
	// Its value is a string which is a URI that hints about the profile (as
	// defined by [I-D.wilde-profile-link]) of the target resource.
	Profile string `json:"profile,omitempty"`

	// The "title" property is OPTIONAL.
	// Its value is a string and is intended for labelling the link with a
	// human-readable identifier (as defined by [RFC5988]).
	Title string `json:"title,omitempty"`

	// The "hreflang" property is OPTIONAL.
	// Its value is a string and is intended for indicating the language of
	// the target resource (as defined by [RFC5988]).
	HrefLang string `json:"hreflang,omitempty"`
}
