package schema

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"
)

// OptionalString distinguishes omitted PATCH fields from explicit null values.
type OptionalString struct {
	Present bool
	Value   *string
}

func (s *OptionalString) UnmarshalJSON(data []byte) error {
	s.Present = true
	return json.Unmarshal(data, &s.Value)
}

func (s OptionalString) Valid(minimum, maximum int, required, nullable bool) bool {
	if !s.Present {
		return !required
	}
	if s.Value == nil {
		return nullable
	}
	length := utf8.RuneCountInString(*s.Value)
	return length >= minimum && (maximum == 0 || length <= maximum)
}

func (s OptionalString) Text(fallback string) string {
	if s.Value == nil {
		return fallback
	}
	return *s.Value
}

type OptionalStrings struct {
	Present bool
	Value   []string
}

func (s *OptionalStrings) UnmarshalJSON(data []byte) error {
	s.Present = true
	if string(data) == "null" {
		return nil
	}
	var values []*string
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	s.Value = make([]string, len(values))
	for index, value := range values {
		if value == nil {
			return fmt.Errorf("array entries must be strings")
		}
		s.Value[index] = *value
	}
	return nil
}

func ValidContentURL(value *string) bool {
	if value == nil {
		return true
	}
	u, err := url.Parse(*value)
	return err == nil &&
		(u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "ftp") &&
		u.Hostname() != "" && !strings.ContainsAny(*value, "\r\n\t ")
}

type PostQuery struct {
	Page     int
	PageSize int
	Category string
	Tag      string
	Search   string
}

func ParsePostQuery(values url.Values) (PostQuery, error) {
	query := PostQuery{Page: 1, PageSize: 10}
	for key, entries := range values {
		if len(entries) != 1 {
			return query, fmt.Errorf("分页或筛选参数不正确")
		}
		switch key {
		case "page", "pageSize", "category", "tag", "q":
		default:
			return query, fmt.Errorf("分页或筛选参数不正确")
		}
	}
	for key, target := range map[string]*int{"page": &query.Page, "pageSize": &query.PageSize} {
		if entries, exists := values[key]; exists {
			number, err := strconv.Atoi(entries[0])
			if err != nil || number < 1 || (key == "pageSize" && number > 100) {
				return query, fmt.Errorf("分页参数不正确")
			}
			*target = number
		}
	}
	query.Category, query.Tag, query.Search = values.Get("category"), values.Get("tag"), strings.TrimSpace(values.Get("q"))
	if utf8.RuneCountInString(query.Category) > 120 ||
		utf8.RuneCountInString(query.Tag) > 120 ||
		utf8.RuneCountInString(query.Search) > 100 ||
		query.Page > math.MaxInt/query.PageSize {
		return query, fmt.Errorf("筛选参数过长或页码过大")
	}
	return query, nil
}

type PostInput struct {
	Title          OptionalString  `json:"title"`
	Slug           OptionalString  `json:"slug"`
	Excerpt        OptionalString  `json:"excerpt"`
	Content        OptionalString  `json:"content"`
	CoverURL       OptionalString  `json:"coverUrl"`
	Status         OptionalString  `json:"status"`
	SEOTitle       OptionalString  `json:"seoTitle"`
	SEODescription OptionalString  `json:"seoDescription"`
	CategoryID     OptionalString  `json:"categoryId"`
	TagIDs         OptionalStrings `json:"tagIds"`
}

func (input PostInput) Valid(create bool) bool {
	validStatus := !input.Status.Present ||
		(input.Status.Value != nil &&
			(*input.Status.Value == "DRAFT" || *input.Status.Value == "PUBLISHED" || *input.Status.Value == "ARCHIVED"))
	return input.Title.Valid(1, 255, create, false) &&
		input.Slug.Valid(0, 191, false, true) &&
		input.Excerpt.Valid(1, 500, create, false) &&
		input.Content.Valid(1, 0, create, false) &&
		input.CategoryID.Valid(0, 0, create, false) &&
		input.CoverURL.Valid(0, 2048, false, true) &&
		ValidContentURL(input.CoverURL.Value) &&
		input.SEOTitle.Valid(0, 255, false, true) &&
		input.SEODescription.Valid(0, 500, false, true) && validStatus
}

type PostChanges struct {
	Values      map[string]interface{}
	TagIDs      []string
	ReplaceTags bool
}
