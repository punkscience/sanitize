// Package service provides the main orchestration logic for the sanitize application.
// This implementation follows the Dependency Inversion Principle by depending on interfaces rather than concrete implementations.
package service

import (
	"fmt"
	"time"

	"sanitize/internal/interfaces"
)

// SanitizeService orchestrates the folder sanitization process
// This struct demonstrates the Open/Closed Principle - it's open for extension via interface implementations
type SanitizeService struct {
	sanitizer interfaces.FolderSanitizer
	walker    interfaces.DirectoryWalker
	processor interfaces.FolderProcessor
	reporter  interfaces.ProgressReporter
}

// NewSanitizeService creates a new instance of SanitizeService with the provided dependencies
// This constructor follows the Dependency Injection pattern for better testability and flexibility
func NewSanitizeService(
	sanitizer interfaces.FolderSanitizer,
	walker interfaces.DirectoryWalker,
	processor interfaces.FolderProcessor,
	reporter interfaces.ProgressReporter,
) *SanitizeService {
	return &SanitizeService{
		sanitizer: sanitizer,
		walker:    walker,
		processor: processor,
		reporter:  reporter,
	}
}

// SanitizeDirectory performs the complete folder sanitization process
// This method coordinates all the different components to achieve the business goal
func (ss *SanitizeService) SanitizeDirectory(rootPath string, dryRun bool) error {
	startTime := time.Now()

	// Step 1: Walk the directory tree to collect folder information
	folders, err := ss.walker.Walk(rootPath)
	if err != nil {
		ss.reporter.ReportError(fmt.Errorf("failed to walk directory tree: %w", err))
		return err
	}

	// Initialize processing statistics
	totalFolders := len(folders)
	processedCount := 0
	renamedCount := 0
	errorCount := 0
	skippedCount := 0

	// Step 2: Process each folder for sanitization
	for i, folder := range folders {
		// Report progress
		progressMsg := fmt.Sprintf("Processing: %s", folder.Name)
		ss.reporter.ReportProgress(i+1, totalFolders, progressMsg)

		// Sanitize the folder name
		sanitizedName := ss.sanitizer.SanitizeName(folder.Name)

		// Process the rename operation
		result, err := ss.processor.ProcessRename(folder, sanitizedName, dryRun)
		processedCount++

		if err != nil {
			ss.reporter.ReportError(fmt.Errorf("failed to process folder %s: %w", folder.Path, err))
			errorCount++
			continue
		}

		// Handle the result
		if result.Error != nil {
			ss.reporter.ReportError(fmt.Errorf("rename error for %s: %w", folder.Path, result.Error))
			errorCount++
		} else if result.WasRenamed && result.Success {
			renamedCount++
		} else if !result.WasRenamed {
			skippedCount++
		}
	}

	// Step 3: Generate and report the final summary
	elapsedTime := time.Since(startTime)
	summary := interfaces.ProcessingSummary{
		TotalFolders:   totalFolders,
		ProcessedCount: processedCount,
		RenamedCount:   renamedCount,
		ErrorCount:     errorCount,
		SkippedCount:   skippedCount,
		ElapsedTime:    elapsedTime.String(),
	}

	ss.reporter.ReportComplete(summary)

	// Return error if there were critical issues
	if errorCount > 0 && renamedCount == 0 {
		return fmt.Errorf("sanitization completed with %d errors and no successful renames", errorCount)
	}

	return nil
}
