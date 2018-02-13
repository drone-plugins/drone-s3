package main

import "testing"

func TestMatcher(t *testing.T) {
	match := "/data/css/file.cgz"

	stringMap := make(map[string]string)
	stringMap[".*(css|cgz)$"] = "text/css"

	want := "text/css"
	got := matcher(match, stringMap)

	if got != want {
		t.Errorf("matcher(%s, %v)\n\tgot: %q, want: %q", match, stringMap, got, want)
	}
}

func TestMatcherEmpty(t *testing.T) {
	match := "/data/css/file.cgz"

	stringMap := make(map[string]string)
	stringMap[""] = "text/css"

	want := "text/css"
	got := matcher(match, stringMap)

	if got != want {
		t.Errorf("matcher(%s, %v)\n\tgot: %q, want: %q", match, stringMap, got, want)
	}
}
