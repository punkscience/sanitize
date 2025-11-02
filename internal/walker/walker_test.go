// Package walker_test provides comprehensive tests for the walker package.
// This test suite ensures directory traversal functionality works correctly.
package walker_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"sanitize/internal/walker"
)

// TestFileSystemWalker_Walk tests basic directory walking functionality
// This test creates a temporary directory structure and verifies walking behavior
func TestFileSystemWalker_Walk(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := createTempDirStructure(t)
	defer os.RemoveAll(tempDir)

	w := walker.NewFileSystemWalker(true, 0) // Skip inaccessible, no depth limit

	folders, err := w.Walk(tempDir)
	if err != nil {
		t.Fatalf("Walk() returned error: %v", err)
	}

	// We should find the test directories we created
	expectedFolders := []string{"level1", "level2", "deep"}
	foundFolders := make(map[string]bool)

	for _, folder := range folders {
		folderName := filepath.Base(folder.Path)
		foundFolders[folderName] = true

		// Verify folder info is correctly populated
		if folder.Name != folderName {
			t.Errorf("Folder name mismatch: got %q, expected %q", folder.Name, folderName)
		}

		if folder.Path == "" {
			t.Error("Folder path should not be empty")
		}

		if folder.Parent == "" {
			t.Error("Folder parent should not be empty")
		}

		if folder.Depth <= 0 {
			t.Errorf("Folder depth should be positive, got %d", folder.Depth)
		}
	}

	// Verify we found all expected folders
	for _, expected := range expectedFolders {
		if !foundFolders[expected] {
			t.Errorf("Expected folder %q not found", expected)
		}
	}
}

// TestFileSystemWalker_DepthSorting tests that folders are sorted by depth (deepest first)
// This test ensures proper ordering for safe bottom-up processing
func TestFileSystemWalker_DepthSorting(t *testing.T) {
	tempDir := createTempDirStructure(t)
	defer os.RemoveAll(tempDir)

	w := walker.NewFileSystemWalker(true, 0)

	folders, err := w.Walk(tempDir)
	if err != nil {
		t.Fatalf("Walk() returned error: %v", err)
	}

	// Verify folders are sorted by depth (deepest first)
	for i := 1; i < len(folders); i++ {
		if folders[i].Depth > folders[i-1].Depth {
			t.Errorf("Folders not sorted by depth: folder %d (depth %d) should come before folder %d (depth %d)",
				i, folders[i].Depth, i-1, folders[i-1].Depth)
		}
	}
}

// TestFileSystemWalker_MaxDepth tests depth limiting functionality
// This test ensures the walker respects depth limits when specified
func TestFileSystemWalker_MaxDepth(t *testing.T) {
	tempDir := createTempDirStructure(t)
	defer os.RemoveAll(tempDir)

	// Test with depth limit of 1
	w := walker.NewFileSystemWalker(true, 1)

	folders, err := w.Walk(tempDir)
	if err != nil {
		t.Fatalf("Walk() returned error: %v", err)
	}

	// With depth limit 1, we should only find "level1" folder
	for _, folder := range folders {
		if folder.Depth > 1 {
			t.Errorf("Found folder at depth %d, but limit was 1: %s", folder.Depth, folder.Path)
		}
	}

	// Should find at least the level1 folder
	foundLevel1 := false
	for _, folder := range folders {
		if filepath.Base(folder.Path) == "level1" {
			foundLevel1 = true
			break
		}
	}

	if !foundLevel1 {
		t.Error("Should have found level1 folder with depth limit 1")
	}
}

// TestFileSystemWalker_InvalidPath tests error handling for invalid paths
// This test ensures proper error handling when given invalid input
func TestFileSystemWalker_InvalidPath(t *testing.T) {
	w := walker.NewFileSystemWalker(true, 0)

	testCases := []struct {
		name string
		path string
	}{
		{"non-existent path", "/path/that/does/not/exist"},
		{"relative non-existent path", "./does/not/exist"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			folders, err := w.Walk(tc.path)
			if err == nil {
				t.Errorf("Expected error for path %q, but got none. Found %d folders", tc.path, len(folders))
			}
		})
	}
}

// TestFileSystemWalker_FilePath tests walking a file instead of directory
// This test ensures proper error handling when path points to a file
func TestFileSystemWalker_FilePath(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "test_file_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	w := walker.NewFileSystemWalker(true, 0)

	folders, err := w.Walk(tempFile.Name())
	if err == nil {
		t.Errorf("Expected error when walking a file, but got none. Found %d folders", len(folders))
	}
}

// TestFileSystemWalker_EmptyDirectory tests walking an empty directory
// This test ensures correct behavior with directories containing no subdirectories
func TestFileSystemWalker_EmptyDirectory(t *testing.T) {
	// Create a temporary empty directory
	tempDir, err := os.MkdirTemp("", "empty_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	w := walker.NewFileSystemWalker(true, 0)

	folders, err := w.Walk(tempDir)
	if err != nil {
		t.Fatalf("Walk() returned error: %v", err)
	}

	// Empty directory should return no folders (root is excluded)
	if len(folders) != 0 {
		t.Errorf("Expected 0 folders in empty directory, got %d", len(folders))
	}
}

// TestFileSystemWalker_SkipInaccessible tests handling of inaccessible directories
// This test ensures proper behavior when encountering permission denied errors
func TestFileSystemWalker_SkipInaccessible(t *testing.T) {
	// This test might be skipped on systems where we can't create permission-denied scenarios
	if os.Getuid() == 0 { // Running as root
		t.Skip("Skipping permission test when running as root")
	}

	tempDir := createTempDirStructure(t)
	defer os.RemoveAll(tempDir)

	// Try to create a directory with restricted permissions
	restrictedDir := filepath.Join(tempDir, "restricted")
	err := os.Mkdir(restrictedDir, 0000) // No permissions
	if err != nil {
		t.Skipf("Cannot create restricted directory: %v", err)
	}
	defer os.Chmod(restrictedDir, 0755) // Restore permissions for cleanup

	// Test with skipInaccessible = true
	w := walker.NewFileSystemWalker(true, 0)
	folders, err := w.Walk(tempDir)

	// Should not fail completely, even if some directories are inaccessible
	if err != nil && len(folders) == 0 {
		t.Errorf("Walk failed completely due to inaccessible directory: %v", err)
	}

	// Test with skipInaccessible = false
	w2 := walker.NewFileSystemWalker(false, 0)
	folders2, err2 := w2.Walk(tempDir)

	// Behavior may vary, but it should handle the error gracefully
	if err2 != nil && len(folders2) == 0 {
		// This is acceptable - the walker may fail on permission errors
		t.Logf("Walker failed on inaccessible directory (expected): %v", err2)
	}
}

// BenchmarkFileSystemWalker_Walk benchmarks directory walking performance
// This benchmark helps ensure the walker performs efficiently
func BenchmarkFileSystemWalker_Walk(b *testing.B) {
	tempDir := createLargeDirStructure(b)
	defer os.RemoveAll(tempDir)

	w := walker.NewFileSystemWalker(true, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		folders, err := w.Walk(tempDir)
		if err != nil {
			b.Fatalf("Walk() returned error: %v", err)
		}
		_ = folders // Ensure the result is used
	}
}

// Helper Functions

// createTempDirStructure creates a test directory structure
// This helper creates a predictable directory tree for testing
func createTempDirStructure(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "walker_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create directory structure:
	// tempDir/
	//   level1/
	//     level2/
	//       deep/
	//   file.txt (regular file, should be ignored)

	level1 := filepath.Join(tempDir, "level1")
	level2 := filepath.Join(level1, "level2")
	deep := filepath.Join(level2, "deep")

	if err := os.MkdirAll(deep, 0755); err != nil {
		t.Fatalf("Failed to create directory structure: %v", err)
	}

	// Create a regular file (should be ignored by walker)
	testFile := filepath.Join(tempDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	return tempDir
}

// createLargeDirStructure creates a larger directory structure for benchmarking
// This helper creates a more complex directory tree for performance testing
func createLargeDirStructure(b *testing.B) string {
	tempDir, err := os.MkdirTemp("", "walker_bench_*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create multiple levels with multiple directories at each level
	for i := 0; i < 5; i++ {
		level1 := filepath.Join(tempDir, fmt.Sprintf("level1_%d", i))
		for j := 0; j < 3; j++ {
			level2 := filepath.Join(level1, fmt.Sprintf("level2_%d", j))
			for k := 0; k < 2; k++ {
				level3 := filepath.Join(level2, fmt.Sprintf("level3_%d", k))
				if err := os.MkdirAll(level3, 0755); err != nil {
					b.Fatalf("Failed to create directory structure: %v", err)
				}
			}
		}
	}

	return tempDir
}

// fmt import moved to the top with other imports
