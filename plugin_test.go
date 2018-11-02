package main

import "testing"

func TestGlobMatcher(t *testing.T) {
	var tests = []struct {
		patterns map[string]string
		path     string
		want     string
	}{
		{map[string]string{"": "text/css"}, "data/css/file.cgz", ""},
		{map[string]string{"*": "text/css"}, "data/css/file.cgz", "text/css"},
		{map[string]string{".cgz": "text/css"}, "data/css/file.cgz", ""},
		{map[string]string{"*.cgz": "text/css"}, "data/css/file.cgz", "text/css"},
		{map[string]string{"*.cgz": "text/css"}, "data/css/file.tgz", ""},
		{map[string]string{"data*.cgz": "text/css"}, "data/css/file.cgz", "text/css"},
		{map[string]string{"*css*.cgz": "text/css"}, "data/css/file.cgz", "text/css"},
	}

	for _, test := range tests {
		got := globMatch(test.path, test.patterns)
		if got != test.want {
			t.Errorf("matcher(%s, %v)\n\tgot: %q, want: %q", test.path, test.patterns, got, test.want)
		}
	}
}
