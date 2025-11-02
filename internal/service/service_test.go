// Package service_test provides comprehensive tests for the service package.
// This test suite ensures the orchestration logic works correctly with mocked dependencies.
package service_test

import (
	"errors"
	"testing"

	"sanitize/internal/interfaces"
	"sanitize/internal/service"
)

// Mock implementations for testing

// mockSanitizer provides a mock implementation of FolderSanitizer
type mockSanitizer struct {
	sanitizeFunc func(string) string
}

func (m *mockSanitizer) SanitizeName(name string) string {
	if m.sanitizeFunc != nil {
		return m.sanitizeFunc(name)
	}
	return name + "_sanitized"
}

// mockWalker provides a mock implementation of DirectoryWalker
type mockWalker struct {
	walkFunc func(string) ([]interfaces.FolderInfo, error)
}

func (m *mockWalker) Walk(rootPath string) ([]interfaces.FolderInfo, error) {
	if m.walkFunc != nil {
		return m.walkFunc(rootPath)
	}
	return []interfaces.FolderInfo{
		{Path: "/test/folder1", Name: "folder1", Depth: 1, Parent: "/test"},
		{Path: "/test/folder2", Name: "folder2", Depth: 1, Parent: "/test"},
	}, nil
}

// mockProcessor provides a mock implementation of FolderProcessor
type mockProcessor struct {
	processFunc func(interfaces.FolderInfo, string, bool) (*interfaces.RenameResult, error)
}

func (m *mockProcessor) ProcessRename(folder interfaces.FolderInfo, newName string, dryRun bool) (*interfaces.RenameResult, error) {
	if m.processFunc != nil {
		return m.processFunc(folder, newName, dryRun)
	}
	return &interfaces.RenameResult{
		Success:    true,
		OldPath:    folder.Path,
		NewPath:    folder.Parent + "/" + newName,
		WasRenamed: folder.Name != newName,
		Error:      nil,
	}, nil
}

// mockReporter provides a mock implementation of ProgressReporter
type mockReporter struct {
	progressCalls []progressCall
	errorCalls    []error
	completeCalls []interfaces.ProcessingSummary
}

type progressCall struct {
	current int
	total   int
	message string
}

func (m *mockReporter) ReportProgress(current, total int, message string) {
	m.progressCalls = append(m.progressCalls, progressCall{current, total, message})
}

func (m *mockReporter) ReportError(err error) {
	m.errorCalls = append(m.errorCalls, err)
}

func (m *mockReporter) ReportComplete(summary interfaces.ProcessingSummary) {
	m.completeCalls = append(m.completeCalls, summary)
}

// Tests

// TestSanitizeService_SanitizeDirectory_Success tests successful sanitization
func TestSanitizeService_SanitizeDirectory_Success(t *testing.T) {
	sanitizer := &mockSanitizer{
		sanitizeFunc: func(name string) string {
			return name + "_clean"
		},
	}

	walker := &mockWalker{
		walkFunc: func(path string) ([]interfaces.FolderInfo, error) {
			return []interfaces.FolderInfo{
				{Path: "/test/folder1", Name: "folder1", Depth: 1, Parent: "/test"},
				{Path: "/test/folder2", Name: "folder2", Depth: 1, Parent: "/test"},
			}, nil
		},
	}

	processor := &mockProcessor{
		processFunc: func(folder interfaces.FolderInfo, newName string, dryRun bool) (*interfaces.RenameResult, error) {
			return &interfaces.RenameResult{
				Success:    true,
				OldPath:    folder.Path,
				NewPath:    folder.Parent + "/" + newName,
				WasRenamed: true,
				Error:      nil,
			}, nil
		},
	}

	reporter := &mockReporter{}

	svc := service.NewSanitizeService(sanitizer, walker, processor, reporter)

	err := svc.SanitizeDirectory("/test", false)
	if err != nil {
		t.Fatalf("SanitizeDirectory() returned error: %v", err)
	}

	// Verify progress reporting was called
	if len(reporter.progressCalls) != 2 {
		t.Errorf("Expected 2 progress calls, got %d", len(reporter.progressCalls))
	}

	// Verify completion reporting
	if len(reporter.completeCalls) != 1 {
		t.Errorf("Expected 1 complete call, got %d", len(reporter.completeCalls))
	}

	summary := reporter.completeCalls[0]
	if summary.TotalFolders != 2 {
		t.Errorf("Expected 2 total folders, got %d", summary.TotalFolders)
	}
	if summary.RenamedCount != 2 {
		t.Errorf("Expected 2 renamed folders, got %d", summary.RenamedCount)
	}
}

// TestSanitizeService_SanitizeDirectory_WalkError tests walker error handling
func TestSanitizeService_SanitizeDirectory_WalkError(t *testing.T) {
	sanitizer := &mockSanitizer{}
	walker := &mockWalker{
		walkFunc: func(path string) ([]interfaces.FolderInfo, error) {
			return nil, errors.New("walk failed")
		},
	}
	processor := &mockProcessor{}
	reporter := &mockReporter{}

	svc := service.NewSanitizeService(sanitizer, walker, processor, reporter)

	err := svc.SanitizeDirectory("/test", false)
	if err == nil {
		t.Error("Expected error when walker fails, but got none")
	}

	// Verify error was reported
	if len(reporter.errorCalls) == 0 {
		t.Error("Expected error to be reported")
	}
}

// TestSanitizeService_SanitizeDirectory_ProcessingErrors tests handling of processing errors
func TestSanitizeService_SanitizeDirectory_ProcessingErrors(t *testing.T) {
	sanitizer := &mockSanitizer{}

	walker := &mockWalker{
		walkFunc: func(path string) ([]interfaces.FolderInfo, error) {
			return []interfaces.FolderInfo{
				{Path: "/test/folder1", Name: "folder1", Depth: 1, Parent: "/test"},
				{Path: "/test/folder2", Name: "folder2", Depth: 1, Parent: "/test"},
			}, nil
		},
	}

	processor := &mockProcessor{
		processFunc: func(folder interfaces.FolderInfo, newName string, dryRun bool) (*interfaces.RenameResult, error) {
			if folder.Name == "folder1" {
				return nil, errors.New("processing failed")
			}
			return &interfaces.RenameResult{
				Success:    true,
				OldPath:    folder.Path,
				NewPath:    folder.Parent + "/" + newName,
				WasRenamed: true,
				Error:      nil,
			}, nil
		},
	}

	reporter := &mockReporter{}

	svc := service.NewSanitizeService(sanitizer, walker, processor, reporter)

	err := svc.SanitizeDirectory("/test", false)
	if err != nil {
		t.Fatalf("SanitizeDirectory() returned error: %v", err)
	}

	// Verify error was reported
	if len(reporter.errorCalls) == 0 {
		t.Error("Expected processing error to be reported")
	}

	// Verify completion summary reflects the error
	if len(reporter.completeCalls) != 1 {
		t.Fatalf("Expected 1 complete call, got %d", len(reporter.completeCalls))
	}

	summary := reporter.completeCalls[0]
	if summary.ErrorCount != 1 {
		t.Errorf("Expected 1 error count, got %d", summary.ErrorCount)
	}
	if summary.RenamedCount != 1 {
		t.Errorf("Expected 1 renamed count, got %d", summary.RenamedCount)
	}
}

// TestSanitizeService_SanitizeDirectory_DryRun tests dry run mode
func TestSanitizeService_SanitizeDirectory_DryRun(t *testing.T) {
	sanitizer := &mockSanitizer{}
	walker := &mockWalker{}

	dryRunProcessor := &mockProcessor{
		processFunc: func(folder interfaces.FolderInfo, newName string, dryRun bool) (*interfaces.RenameResult, error) {
			if !dryRun {
				t.Error("Expected dry run mode to be passed to processor")
			}
			return &interfaces.RenameResult{
				Success:    true,
				OldPath:    folder.Path,
				NewPath:    folder.Parent + "/" + newName,
				WasRenamed: true,
				Error:      nil,
			}, nil
		},
	}

	reporter := &mockReporter{}

	svc := service.NewSanitizeService(sanitizer, walker, dryRunProcessor, reporter)

	err := svc.SanitizeDirectory("/test", true) // Dry run mode
	if err != nil {
		t.Fatalf("SanitizeDirectory() returned error: %v", err)
	}
}

// TestSanitizeService_SanitizeDirectory_NoChangesNeeded tests when no folders need renaming
func TestSanitizeService_SanitizeDirectory_NoChangesNeeded(t *testing.T) {
	sanitizer := &mockSanitizer{
		sanitizeFunc: func(name string) string {
			return name // No changes
		},
	}

	walker := &mockWalker{}

	processor := &mockProcessor{
		processFunc: func(folder interfaces.FolderInfo, newName string, dryRun bool) (*interfaces.RenameResult, error) {
			return &interfaces.RenameResult{
				Success:    true,
				OldPath:    folder.Path,
				NewPath:    folder.Path,
				WasRenamed: false, // No rename needed
				Error:      nil,
			}, nil
		},
	}

	reporter := &mockReporter{}

	svc := service.NewSanitizeService(sanitizer, walker, processor, reporter)

	err := svc.SanitizeDirectory("/test", false)
	if err != nil {
		t.Fatalf("SanitizeDirectory() returned error: %v", err)
	}

	// Verify completion summary reflects no renames
	if len(reporter.completeCalls) != 1 {
		t.Fatalf("Expected 1 complete call, got %d", len(reporter.completeCalls))
	}

	summary := reporter.completeCalls[0]
	if summary.RenamedCount != 0 {
		t.Errorf("Expected 0 renamed count, got %d", summary.RenamedCount)
	}
	if summary.SkippedCount != 2 {
		t.Errorf("Expected 2 skipped count, got %d", summary.SkippedCount)
	}
}

// TestSanitizeService_SanitizeDirectory_EmptyDirectory tests handling of empty directories
func TestSanitizeService_SanitizeDirectory_EmptyDirectory(t *testing.T) {
	sanitizer := &mockSanitizer{}

	walker := &mockWalker{
		walkFunc: func(path string) ([]interfaces.FolderInfo, error) {
			return []interfaces.FolderInfo{}, nil // Empty directory
		},
	}

	processor := &mockProcessor{}
	reporter := &mockReporter{}

	svc := service.NewSanitizeService(sanitizer, walker, processor, reporter)

	err := svc.SanitizeDirectory("/empty", false)
	if err != nil {
		t.Fatalf("SanitizeDirectory() returned error: %v", err)
	}

	// Verify completion summary for empty directory
	if len(reporter.completeCalls) != 1 {
		t.Fatalf("Expected 1 complete call, got %d", len(reporter.completeCalls))
	}

	summary := reporter.completeCalls[0]
	if summary.TotalFolders != 0 {
		t.Errorf("Expected 0 total folders, got %d", summary.TotalFolders)
	}
}
