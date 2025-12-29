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

func TestResolveSource(t *testing.T) {
	tests := []struct {
		sourceDir   string
		source      string
		stripPrefix string
		expected    string
	}{
		// Test case 1
		{
			sourceDir:   "/home/user/documents",
			source:      "/home/user/documents/file.txt",
			stripPrefix: "output-",
			expected:    "output-file.txt",
		},
		// Test case 2
		{
			sourceDir:   "assets",
			source:      "assets/images/logo.png",
			stripPrefix: "",
			expected:    "images/logo.png",
		},
		// Test case 3
		{
			sourceDir:   "/var/www/html",
			source:      "/var/www/html/pages/index.html",
			stripPrefix: "web",
			expected:    "webpages/index.html",
		},
		// Test case 4
		{
			sourceDir:   "dist",
			source:      "dist/js/app.js",
			stripPrefix: "public",
			expected:    "publicjs/app.js",
		},
	}

	for _, tc := range tests {
		result := resolveSource(tc.sourceDir, tc.source, tc.stripPrefix)
		if result != tc.expected {
			t.Errorf("Expected: %s, Got: %s", tc.expected, result)
		}
	}
}

