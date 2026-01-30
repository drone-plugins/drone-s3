package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsDir(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "testdir")
	testFile := filepath.Join(tmpDir, "testfile.txt")

	// Create a test directory
	err := os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a test file
	file, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	tests := []struct {
		name        string
		source      string
		matches     []string
		expectError bool
		expectSkip  bool
		errorContains string
	}{
		{
			name:        "file should not error",
			source:      testFile,
			matches:     []string{testFile},
			expectError: false,
			expectSkip:  false,
		},
		{
			name:        "directory without glob should error", 
			source:      testDir,
			matches:     []string{testDir},
			expectError: true,
			expectSkip:  false,
			errorContains: "specified without glob pattern",
		},
		{
			name:        "directory with glob pattern should skip",
			source:      testDir,
			matches:     []string{testDir + "/file1.txt", testDir + "/file2.txt"},
			expectError: false,
			expectSkip:  true,
		},
		{
			name:        "non-existent path should skip",
			source:      "/non/existent/path", 
			matches:     []string{},
			expectError: false,
			expectSkip:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := isDir(tc.source, tc.matches)
			
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err == errSkip {
					t.Errorf("Expected fatal error but got skip error")
				} else if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain '%s', but got: %v", tc.errorContains, err)
				}
			} else if tc.expectSkip {
				if err != errSkip {
					t.Errorf("Expected skip error but got: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}