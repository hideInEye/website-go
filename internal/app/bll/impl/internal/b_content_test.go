package internal

import (
	"strings"
	"testing"

	"go-gin/internal/app/schema"
)

func TestSlugNormalizationMatchesExistingBlog(t *testing.T) {
	tests := map[string]string{
		" Hello, World! ": "hello-world",
		"代码 笔记":           "代码-笔记",
		"Ｃｏｄｅ":            "code",
		"Café Notes":      "cafe-notes",
		"---":             "",
	}
	for input, want := range tests {
		if got := normalizePostSlug(input); got != want {
			t.Errorf("normalizePostSlug(%q) = %q, want %q", input, got, want)
		}
	}
	if len([]rune(normalizePostSlug(strings.Repeat("中", 200)))) != 170 {
		t.Error("slug maximum length changed")
	}
}

func TestPublicCommentAuthorAndProfileDoNotExposeCredentials(t *testing.T) {
	user := schema.User{
		ID:           "user",
		Username:     "reader",
		Email:        "reader@example.test",
		PasswordHash: "must-not-leak",
		Role:         "ADMIN",
	}
	byline := commentUser(user)
	for _, key := range []string{"id", "email", "role", "passwordHash"} {
		if _, exists := byline[key]; exists {
			t.Errorf("public author includes %s", key)
		}
	}
	if _, exists := profileObject(user)["passwordHash"]; exists {
		t.Error("private profile contains password hash")
	}
}
