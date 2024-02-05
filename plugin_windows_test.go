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
			name:        "backslash strip prefix",
			target:      "hello",
			srcPath:     `foo/bar/world`,
			stripPrefix: `foo\bar`,
			expected:    "/hello/world",
		},
		{
			name:        "forward slash strip prefix",
			target:      "hello",
			srcPath:     "foo/bar/world",
			stripPrefix: `foo/bar`,
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

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "/path/to/file.txt",
			expected: "path/to/file.txt",
		},
		{
			input:    "C:\\Users\\username\\Documents\\file.doc",
			expected: "C:\\Users\\username\\Documents\\file.doc",
		},
		{
			input:    "relative/path/to/file",
			expected: "relative/path/to/file",
		},
		{
			input:    "file.txt",
			expected: "file.txt",
		},
		{
			input:    "/root/directory/",
			expected: "root/directory/",
		},
		{
			input:    "no_slash",
			expected: "no_slash",
		},
	}

	for _, tc := range tests {
		result := normalizePath(tc.input)
		if result != tc.expected {
			t.Errorf("Expected: %s, Got: %s", tc.expected, result)
		}
	}
}
