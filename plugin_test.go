package main

import (
	"encoding/json"
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

func TestBuildObjectURL(t *testing.T) {
	tests := []struct {
		name   string
		plugin Plugin
		target string
		want   string
	}{
		{
			name:   "default AWS, no region specified",
			plugin: Plugin{Bucket: "my-bucket"},
			target: "builds/app.zip",
			want:   "https://s3.amazonaws.com/my-bucket/builds/app.zip",
		},
		{
			name:   "default AWS, us-east-1 uses legacy global endpoint",
			plugin: Plugin{Bucket: "my-bucket", Region: "us-east-1"},
			target: "builds/app.zip",
			want:   "https://s3.amazonaws.com/my-bucket/builds/app.zip",
		},
		{
			name:   "AWS opt-in region uses region-qualified endpoint",
			plugin: Plugin{Bucket: "my-bucket", Region: "ap-east-1"},
			target: "builds/app.zip",
			want:   "https://s3.ap-east-1.amazonaws.com/my-bucket/builds/app.zip",
		},
		{
			name:   "custom endpoint with path style (e.g. MinIO)",
			plugin: Plugin{Bucket: "my-bucket", Endpoint: "minio.internal:9000", PathStyle: true},
			target: "builds/app.zip",
			want:   "https://minio.internal:9000/my-bucket/builds/app.zip",
		},
		{
			name:   "custom endpoint with virtual-hosted style",
			plugin: Plugin{Bucket: "my-bucket", Endpoint: "https://nyc3.digitaloceanspaces.com", PathStyle: false},
			target: "builds/app.zip",
			want:   "https://my-bucket.nyc3.digitaloceanspaces.com/builds/app.zip",
		},
		{
			name:   "target with leading slash is normalized",
			plugin: Plugin{Bucket: "my-bucket"},
			target: "/builds/app.zip",
			want:   "https://s3.amazonaws.com/my-bucket/builds/app.zip",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.plugin.buildObjectURL(tc.target)
			if got != tc.want {
				t.Errorf("buildObjectURL() = %q; want %q", got, tc.want)
			}
		})
	}
}

func TestWriteArtifactFile(t *testing.T) {
	t.Run("writes fileUpload/v1 JSON when PLUGIN_ARTIFACT_FILE is set", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "artifact")
		t.Setenv("PLUGIN_ARTIFACT_FILE", tmpFile)

		entries := []fileArtifactEntry{
			{
				Name:       "app.zip",
				URL:        "s3://my-bucket/builds/app.zip",
				FilePath:   "builds/app.zip",
				BucketName: "my-bucket",
				Region:     "us-east-1",
				Digest:     "abc123",
			},
		}
		writeArtifactFile(entries)

		data, err := os.ReadFile(tmpFile)
		if err != nil {
			t.Fatalf("Expected artifact file to exist: %v", err)
		}

		var got artifactFile
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Expected valid JSON: %v", err)
		}
		if got.Kind != "fileUpload/v1" {
			t.Errorf("Expected kind=fileUpload/v1, got %s", got.Kind)
		}
		if len(got.Data.FileArtifacts) != 1 {
			t.Fatalf("Expected 1 file artifact, got %d", len(got.Data.FileArtifacts))
		}
		f := got.Data.FileArtifacts[0]
		if f.Name != "app.zip" {
			t.Errorf("Unexpected name: %s", f.Name)
		}
		if f.URL != "s3://my-bucket/builds/app.zip" {
			t.Errorf("Unexpected URL: %s", f.URL)
		}
		if f.FilePath != "builds/app.zip" {
			t.Errorf("Unexpected filePath: %s", f.FilePath)
		}
		if f.BucketName != "my-bucket" {
			t.Errorf("Unexpected bucketName: %s", f.BucketName)
		}
		if f.Region != "us-east-1" {
			t.Errorf("Unexpected region: %s", f.Region)
		}
		if f.Digest != "abc123" {
			t.Errorf("Unexpected digest: %s", f.Digest)
		}
	})

	t.Run("no-op when PLUGIN_ARTIFACT_FILE is not set", func(t *testing.T) {
		t.Setenv("PLUGIN_ARTIFACT_FILE", "")
		// Should not panic or write anything
		writeArtifactFile([]fileArtifactEntry{{Name: "f", URL: "s3://b/f"}})
	})
}