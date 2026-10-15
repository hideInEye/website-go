package schema

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
)

func TestPostPatchPreservesOmittedFields(t *testing.T) {
	var input PostInput
	if err := json.Unmarshal([]byte(`{"slug":"renamed-post"}`), &input); err != nil {
		t.Fatal(err)
	}
	if !input.Valid(false) {
		t.Fatal("valid partial update was rejected")
	}
	if input.Status.Present || input.TagIDs.Present {
		t.Fatal("omitted fields must not reset status or tags")
	}
}

func TestPostPatchCanClearNullableFields(t *testing.T) {
	var input PostInput
	body := `{"coverUrl":null,"seoTitle":null,"seoDescription":null,"tagIds":[]}`
	if err := json.Unmarshal([]byte(body), &input); err != nil {
		t.Fatal(err)
	}
	if !input.Valid(false) {
		t.Fatal("explicit clearing of nullable fields was rejected")
	}
	fields := map[string]OptionalString{
		"coverUrl":       input.CoverURL,
		"seoTitle":       input.SEOTitle,
		"seoDescription": input.SEODescription,
	}
	for name, field := range fields {
		if !field.Present || field.Value != nil {
			t.Errorf("%s must retain explicit null", name)
		}
	}
	if !input.TagIDs.Present || input.TagIDs.Value == nil || len(input.TagIDs.Value) != 0 {
		t.Fatal("empty tagIds must explicitly clear tags")
	}
}

func TestPostInputValidation(t *testing.T) {
	var input PostInput
	body := `{"title":"标题","excerpt":"摘要","content":"# 内容","categoryId":"cat-1"}`
	if err := json.Unmarshal([]byte(body), &input); err != nil {
		t.Fatal(err)
	}
	if !input.Valid(true) {
		t.Fatal("valid article rejected")
	}

	for _, body := range []string{
		`{"title":null}`,
		`{"status":"OTHER"}`,
		`{"coverUrl":"javascript:alert(1)"}`,
		`{"title":""}`,
		`{"categoryId":null}`,
	} {
		t.Run(body, func(t *testing.T) {
			var input PostInput
			if err := json.Unmarshal([]byte(body), &input); err == nil && input.Valid(false) {
				t.Error("invalid patch accepted")
			}
		})
	}
	if (PostInput{}).Valid(true) {
		t.Error("empty create must be rejected")
	}
	if !(PostInput{}).Valid(false) {
		t.Error("empty patch must not introduce required fields")
	}
	var nullTags PostInput
	if err := json.Unmarshal([]byte(`{"tagIds":[null]}`), &nullTags); err == nil {
		t.Error("null tag array elements must not silently become empty strings")
	}
}

func TestPostQueryValidation(t *testing.T) {
	tests := []struct {
		query string
		valid bool
	}{
		{query: "", valid: true},
		{query: "page=2&pageSize=100&category=code&tag=go&q=%20hello%20", valid: true},
		{query: "page=0"},
		{query: "page="},
		{query: "page=-1"},
		{query: "pageSize=101"},
		{query: "page=NaN"},
		{query: "page=1&page=2"},
		{query: "unexpected=value"},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			values, err := url.ParseQuery(test.query)
			if err != nil {
				t.Fatal(err)
			}
			query, err := ParsePostQuery(values)
			if valid := err == nil; valid != test.valid {
				t.Errorf("query validity = %v, want %v", valid, test.valid)
			}
			if test.query == "" && (query.Page != 1 || query.PageSize != 10) {
				t.Error("default pagination changed")
			}
			if err == nil && strings.Contains(test.query, "q=") && query.Search != "hello" {
				t.Error("search string is not trimmed")
			}
		})
	}
}
