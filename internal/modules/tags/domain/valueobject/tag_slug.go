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
	slug = Slugify(name.String())

	return TagSlug{
		value: slug,
	}, nil
}

func (t TagSlug) String() string {
	return t.value
}

func Slugify(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), " ", "-")
}
