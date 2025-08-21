package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// ===============================
// WILDCARD STRIP PREFIX TESTS
// ===============================

func TestStripWildcardPrefix(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		pattern     string
		expected    string
		expectError bool
		errorMsg    string
	}{
		// Mandatory test cases from spec
		{
			name:     "Single wildcard - */",
			path:     "/harness/artifacts/9f2c1b/module/app.zip",
			pattern:  "/harness/artifacts/*/",
			expected: "module/app.zip",
		},
		{
			name:     "Double wildcard - */*/",
			path:     "/harness/artifacts/hash/nightly/lib.zip",
			pattern:  "/harness/artifacts/*/*/",
			expected: "lib.zip",
		},
		{
			name:     "Triple wildcard - */*/*/",
			path:     "/harness/artifacts/a/b/c/doc.zip",
			pattern:  "/harness/artifacts/*/*/*/",
			expected: "doc.zip",
		},
		{
			name:     "Any depth - **/",
			path:     "/harness/artifacts/x/y/z/file.zip",
			pattern:  "/harness/artifacts/**/",
			expected: "file.zip",
		},
		{
			name:     "No match - path unchanged",
			path:     "/different/path/file.zip",
			pattern:  "/harness/artifacts/*/",
			expected: "/different/path/file.zip",
		},
		{
			name:        "Empty key error",
			path:        "/harness/artifacts/build/app.zip",
			pattern:     "/harness/artifacts/build/app.zip",
			expectError: true,
			errorMsg:    "removes entire path",
		},
		{
			name:     "Pattern with trailing content",
			path:     "/harness/artifacts/build123/services/app.zip",
			pattern:  "/harness/artifacts/*/services/",
			expected: "app.zip",
		},
		{
			name:     "Question mark wildcard",
			path:     "/harness/artifacts/build1/app.zip",
			pattern:  "/harness/artifacts/build?/",
			expected: "app.zip",
		},

		// Additional edge cases
		{
			name:     "Literal prefix (no wildcards)",
			path:     "/harness/artifacts/build/app.zip",
			pattern:  "/harness/artifacts/",
			expected: "build/app.zip",
		},
		{
			name:     "Root wildcard",
			path:     "/build/app.zip",
			pattern:  "/*/",
			expected: "app.zip",
		},
		{
			name:     "Complex nested structure",
			path:     "/harness/artifacts/build-123/services/auth/v1.2/auth-service.zip",
			pattern:  "/harness/artifacts/*/services/*/",
			expected: "v1.2/auth-service.zip",
		},
		{
			name:     "Double asterisk with additional path",
			path:     "/harness/artifacts/very/deep/nested/structure/file.zip",
			pattern:  "/harness/artifacts/**/structure/",
			expected: "file.zip",
		},
		{
			name:     "Multiple question marks",
			path:     "/harness/artifacts/build123/app.zip",
			pattern:  "/harness/artifacts/build???/",
			expected: "app.zip",
		},
		{
			name:     "Empty pattern returns original path",
			path:     "/harness/artifacts/build/app.zip",
			pattern:  "",
			expected: "/harness/artifacts/build/app.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := stripWildcardPrefix(tt.path, tt.pattern)

			if tt.expectError {
				if err == nil {
					t.Errorf("stripWildcardPrefix(%q, %q) expected error, got nil", tt.path, tt.pattern)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("stripWildcardPrefix(%q, %q) error = %v, want error containing %q", tt.path, tt.pattern, err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("stripWildcardPrefix(%q, %q) unexpected error: %v", tt.path, tt.pattern, err)
				} else if result != tt.expected {
					t.Errorf("stripWildcardPrefix(%q, %q) = %q, want %q", tt.path, tt.pattern, result, tt.expected)
				}
			}
		})
	}
}

func TestPatternToRegex(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		testPath string
		matches  bool
	}{
		{
			name:     "Single wildcard matches one segment",
			pattern:  filepath.ToSlash("/harness/artifacts/*/"),
			testPath: filepath.ToSlash("/harness/artifacts/build123/"),
			matches:  true,
		},
		{
			name:     "Single wildcard with exact match",
			pattern:  "/harness/artifacts/*/",
			testPath: "/harness/artifacts/build123/",
			matches:  true,
		},
		{
			name:     "Double wildcard matches any depth",
			pattern:  "/harness/artifacts/**/",
			testPath: "/harness/artifacts/build123/module1/deep/",
			matches:  true,
		},
		{
			name:     "Question mark matches single character",
			pattern:  "/harness/artifacts/build?/",
			testPath: "/harness/artifacts/build1/",
			matches:  true,
		},
		{
			name:     "Question mark doesn't match multiple characters",
			pattern:  "/harness/artifacts/build?/",
			testPath: "/harness/artifacts/build123/",
			matches:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := patternToRegex(tt.pattern)
			if err != nil {
				t.Fatalf("patternToRegex(%q) error: %v", tt.pattern, err)
			}

			matches := re.MatchString(tt.testPath)
			if matches != tt.matches {
				t.Errorf("patternToRegex(%q).MatchString(%q) = %v, want %v", tt.pattern, tt.testPath, matches, tt.matches)
			}
		})
	}
}

// ===============================
// INTEGRATION TESTS - ResolveKey
// ===============================

func TestResolveKeyWithWildcards(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		srcPath     string
		stripPrefix string
		expected    string
	}{
		{
			name:        "Informatica's scenario - single wildcard",
			target:      "deployment",
			srcPath:     "/harness/artifacts/build-123/module1/app.zip",
			stripPrefix: "/harness/artifacts/*/",
			expected:    "/deployment/module1/app.zip",
		},
		{
			name:        "Deep nested with double wildcard",
			target:      "releases",
			srcPath:     "/harness/artifacts/build-456/services/auth/v1.0/auth-service.zip",
			stripPrefix: "/harness/artifacts/**/services/",
			expected:    "/releases/auth/v1.0/auth-service.zip",
		},
		{
			name:        "Question mark pattern",
			target:      "upload",
			srcPath:     "/harness/artifacts/build1/app.zip",
			stripPrefix: "/harness/artifacts/build?/",
			expected:    "/upload/app.zip",
		},
		{
			name:        "No wildcard - literal prefix",
			target:      "backup",
			srcPath:     "/harness/artifacts/build123/lib.zip",
			stripPrefix: "/harness/artifacts/",
			expected:    "/backup/build123/lib.zip",
		},
		{
			name:        "Pattern doesn't match - path unchanged",
			target:      "fallback",
			srcPath:     "/different/location/file.zip",
			stripPrefix: "/harness/artifacts/*/",
			expected:    "/fallback/different/location/file.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveKey(tt.target, tt.srcPath, tt.stripPrefix)
			if result != tt.expected {
				t.Errorf("resolveKey(%q, %q, %q) = %q, want %q",
					tt.target, tt.srcPath, tt.stripPrefix, result, tt.expected)
			}
		})
	}
}

// ===============================
// ERROR HANDLING TESTS
// ===============================

func TestWildcardErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Pattern doesn't start with slash",
			pattern: "harness/artifacts/*/",
			wantErr: true,
			errMsg:  "must start with '/'",
		},
		{
			name:    "Pattern too long (>256 chars)",
			pattern: "/" + strings.Repeat("very-long-directory-name/", 15) + "*/", // >256 chars
			wantErr: true,
			errMsg:  "too long",
		},
		{
			name:    "Too many wildcards (>20)",
			pattern: "/" + strings.Repeat("*/", 21),
			wantErr: true,
			errMsg:  "too many wildcards",
		},
		{
			name:    "Empty segment",
			pattern: "/harness//artifacts/*/",
			wantErr: true,
			errMsg:  "empty segment",
		},
		{
			name:    "Invalid ** usage",
			pattern: "/harness/**artifacts/*/",
			wantErr: true,
			errMsg:  "standalone directory segment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStripPrefix(tt.pattern)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateStripPrefix(%q) expected error, got nil", tt.pattern)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateStripPrefix(%q) error = %v, want error containing %q", tt.pattern, err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateStripPrefix(%q) unexpected error: %v", tt.pattern, err)
				}
			}
		})
	}
}

// ===============================
// WINDOWS NORMALIZATION TESTS
// ===============================

func TestWindowsBackslashInputs(t *testing.T) {
	t.Run("validate single-backslash anchored pattern accepted", func(t *testing.T) {
		if err := validateStripPrefix(`\harness\artifacts\*/`); err != nil {
			t.Fatalf("validateStripPrefix backslash pattern unexpected error: %v", err)
		}
	})

	t.Run("reject UNC double-backslash pattern (empty segment)", func(t *testing.T) {
		if err := validateStripPrefix(`\\harness\\artifacts\\*/`); err == nil {
			t.Fatalf("expected error for UNC-style pattern, got nil")
		}
	})

	t.Run("reject drive-letter pattern", func(t *testing.T) {
		if err := validateStripPrefix(`C:\\harness\\artifacts\\*/`); err == nil {
			t.Fatalf("expected error for drive-letter pattern, got nil")
		}
	})
}

func TestExecStyleNormalizationWithWindowsPatterns(t *testing.T) {
	// Ensure Windows-style strip_prefix works on forward-slash paths after normalization
	patternWin := `\harness\artifacts\*/`
	if err := validateStripPrefix(patternWin); err != nil {
		t.Fatalf("validateStripPrefix(%q) error: %v", patternWin, err)
	}
	path := "/harness/artifacts/abc123/module/app.zip"
	stripped, err := stripWildcardPrefix(path, patternWin)
	if err != nil {
		t.Fatalf("stripWildcardPrefix error: %v", err)
	}
	if want := "module/app.zip"; stripped != want {
		t.Fatalf("stripped=%q want %q", stripped, want)
	}
}

// ===============================
// PERFORMANCE BENCHMARK
// ===============================

func BenchmarkStripWildcardPrefix(b *testing.B) {
	path := "/harness/artifacts/build-12345/services/auth/v1.2.3/auth-service.zip"
	pattern := "/harness/artifacts/*/services/*/"

	// Pre-compile pattern (this would happen once in real usage)
	_, err := patternToRegex(pattern)
	if err != nil {
		b.Fatalf("Failed to compile pattern: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = stripWildcardPrefix(path, pattern)
	}
}
