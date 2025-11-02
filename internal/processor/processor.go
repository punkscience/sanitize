// Package processor handles the actual folder renaming operations.
// This implementation follows the Single Responsibility Principle by focusing solely on rename processing.
package processor

import (
	"fmt"
	"os"
	"path/filepath"

	"sanitize/internal/interfaces"
)

// FileSystemProcessor implements the FolderProcessor interface for file system operations
// This struct handles the complexity of folder renaming with collision detection
type FileSystemProcessor struct {
	// maxCollisionRetries limits how many collision resolution attempts to make
	maxCollisionRetries int
}

// NewFileSystemProcessor creates a new instance of FileSystemProcessor with default settings
// This constructor allows for configuration of processing behavior
func NewFileSystemProcessor(maxCollisionRetries int) interfaces.FolderProcessor {
	if maxCollisionRetries <= 0 {
		maxCollisionRetries = 1000 // Default safety limit
	}

	return &FileSystemProcessor{
		maxCollisionRetries: maxCollisionRetries,
	}
}

// ProcessRename handles renaming a single folder with collision detection and error recovery
// This method implements the FolderProcessor interface with comprehensive error handling
func (fsp *FileSystemProcessor) ProcessRename(folder interfaces.FolderInfo, newName string, dryRun bool) (*interfaces.RenameResult, error) {
	// Initialize the result structure
	result := &interfaces.RenameResult{
		Success:    false,
		OldPath:    folder.Path,
		WasRenamed: false,
		Error:      nil,
	}

	// Check if renaming is actually needed
	if newName == folder.Name {
		result.Success = true
		result.NewPath = folder.Path
		result.WasRenamed = false
		return result, nil
	}

	// Construct the target path
	newPath := filepath.Join(folder.Parent, newName)

	// Handle potential name collisions
	finalPath, err := fsp.resolveNameCollision(newPath, newName)
	if err != nil {
		result.Error = fmt.Errorf("failed to resolve name collision: %w", err)
		return result, nil // Return result with error, don't fail the operation
	}

	result.NewPath = finalPath
	result.WasRenamed = true

	// If dry run mode, simulate the operation
	if dryRun {
		result.Success = true
		return result, nil
	}

	// Perform the actual rename operation
	err = fsp.performRename(folder.Path, finalPath)
	if err != nil {
		result.Error = fmt.Errorf("rename operation failed: %w", err)
		return result, nil // Return result with error, don't fail the operation
	}

	result.Success = true
	return result, nil
}

// resolveNameCollision handles naming conflicts by finding an available name
// This method ensures that rename operations don't overwrite existing folders
func (fsp *FileSystemProcessor) resolveNameCollision(targetPath, baseName string) (string, error) {
	// Check if the target path is already available
	if !fsp.pathExists(targetPath) {
		return targetPath, nil
	}

	// Extract directory and file extension if any
	dir := filepath.Dir(targetPath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := baseName
	if ext != "" {
		nameWithoutExt = baseName[:len(baseName)-len(ext)]
	}

	// Try numbered variations until we find an available name
	for counter := 1; counter <= fsp.maxCollisionRetries; counter++ {
		var candidateName string
		if ext != "" {
			candidateName = fmt.Sprintf("%s_%d%s", nameWithoutExt, counter, ext)
		} else {
			candidateName = fmt.Sprintf("%s_%d", nameWithoutExt, counter)
		}

		candidatePath := filepath.Join(dir, candidateName)
		if !fsp.pathExists(candidatePath) {
			return candidatePath, nil
		}
	}

	// If we exhausted all retries, use a timestamp-based fallback
	fallbackName := fmt.Sprintf("%s_conflict", baseName)
	return filepath.Join(dir, fallbackName), nil
}

// pathExists checks if a path exists in the file system
// This method provides safe existence checking with proper error handling
func (fsp *FileSystemProcessor) pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// performRename executes the actual file system rename operation
// This method handles the low-level rename with proper error context
func (fsp *FileSystemProcessor) performRename(oldPath, newPath string) error {
	// Attempt the rename operation
	err := os.Rename(oldPath, newPath)
	if err != nil {
		// Provide more context about the failure
		return fmt.Errorf("failed to rename '%s' to '%s': %w", oldPath, newPath, err)
	}

	return nil
}
