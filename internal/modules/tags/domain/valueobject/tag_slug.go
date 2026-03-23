package valueobject

import "strings"

type TagSlug struct {
	value string
}

func NewTagSlug(slug string) (TagSlug, error) {
	name, err := NewTagName(slug)
	if err != nil {
		return TagSlug{}, err
	}
	slug = strings.ToLower(name.String())

	return TagSlug{
		value: slug,
	}, nil
}

func (t TagSlug) String() string {
	return t.value
}
