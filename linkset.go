package halgo

import "encoding/json"

// LinkSet is represents a set of HAL links. Deserialisable from a single
// JSON hash, or a collection of links.
type LinkSet []Link

func (l LinkSet) MarshalJSON() ([]byte, error) {
	if len(l) == 1 {
		return json.Marshal(l[0])
	}

	other := make([]Link, len(l))
	copy(other, l)

	return json.Marshal(other)
}

func (l *LinkSet) UnmarshalJSON(d []byte) error {
	single := Link{}
	err := json.Unmarshal(d, &single)
	if err == nil {
		*l = []Link{single}
		return nil
	}

	if _, ok := err.(*json.UnmarshalTypeError); !ok {
		return err
	}

	multiple := []Link{}
	err = json.Unmarshal(d, &multiple)

	if err == nil {
		*l = multiple
		return nil
	}

	return err
}
