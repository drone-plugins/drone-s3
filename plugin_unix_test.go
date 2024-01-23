//go:build aix || darwin || dragonfly || freebsd || (js && wasm) || linux || netbsd || openbsd || solaris

package main

import (
	"testing"
)

func TestResolveUnixKey(t *testing.T) {
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
	}

	for _, tc := range tests {
		got := resolveKey(tc.target, tc.srcPath, tc.stripPrefix)
		if tc.expected != got {
			t.Fatalf("%s: expected error: %v, got: %v", tc.name, tc.expected, got)
		}
	}
}

func TestResolveDir(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "example-string",
			expected: "example-string",
		},
		{
			input:    "/path/to/file",
			expected: "path/to/file",
		},
		{
			input:    "12345",
			expected: "12345",
		},
		{
			input:    "/root/directory",
			expected: "root/directory",
		},
		{
			input:    "no_slash",
			expected: "no_slash",
		},
	}

	for _, tc := range tests {
		result := resolveDir(tc.input)
		if result != tc.expected {
			t.Errorf("Expected: %s, Got: %s", tc.expected, result)
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
