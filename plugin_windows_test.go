//go:build windows

package main

import (
	"testing"
)

func TestResolveWinKey(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		srcPath     string
		stripPrefix string
		expected    string
	}{
		{
			name:        "target not set",
			target:      "",
			srcPath:     "/foo/bar",
			stripPrefix: "/foo",
			expected:    "/bar",
		},
		{
			name:        "strip prefix not set",
			target:      "/hello",
			srcPath:     "/foo/bar",
			stripPrefix: "",
			expected:    "/hello/foo/bar",
		},
		{
			name:        "everything set",
			target:      "hello",
			srcPath:     "/foo/bar",
			stripPrefix: "/foo",
			expected:    "/hello/bar",
		},
		{
			name:        "backslash src path",
			target:      "hello",
			srcPath:     `foo\bar`,
			stripPrefix: "",
			expected:    "/hello/foo/bar",
		},
		{
			name:        "backslash src path and strip prefix",
			target:      "hello",
			srcPath:     `foo\bar\world`,
			stripPrefix: `foo\bar`,
			expected:    "/hello/world",
		},
		{
			name:        "backslash src path and forward slash strip prefix",
			target:      "hello",
			srcPath:     `foo\bar\world`,
			stripPrefix: "foo/bar",
			expected:    "/hello/world",
		},
		{
			name:        "forward slash src path and backslash strip prefix",
			target:      "hello",
			srcPath:     "foo/bar/world",
			stripPrefix: `foo\bar`,
			expected:    "/hello/world",
		},
	}

	for _, tc := range tests {
		got := resolveKey(tc.target, tc.srcPath, tc.stripPrefix)
		if tc.expected != got {
			t.Fatalf("%s: expected error: %v, got: %v", tc.name, tc.expected, got)
		}
	}
}
