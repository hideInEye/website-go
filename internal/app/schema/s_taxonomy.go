package schema

type TaxonomyInput struct {
	Name        OptionalString `json:"name"`
	Slug        OptionalString `json:"slug"`
	Description OptionalString `json:"description"`
}

func (input TaxonomyInput) Valid(create bool) bool {
	return input.Name.Valid(1, 80, create, false) &&
		input.Slug.Valid(1, 120, create, false) &&
		input.Description.Valid(0, 500, false, true)
}

type CategoryWithCount struct {
	Category
	PostCount int64 `json:"postCount"`
}

type TagWithCount struct {
	Tag
	PostCount int64 `json:"postCount"`
}
