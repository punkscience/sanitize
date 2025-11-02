// Package walker provides directory tree traversal functionality.
// This implementation follows the Single Responsibility Principle by focusing solely on directory walking logic.
package walker

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"sanitize/internal/interfaces"
)

// FileSystemWalker implements the DirectoryWalker interface for file system traversal
// This struct handles the complexity of walking directory trees safely
type FileSystemWalker struct {
	// skipInaccessible determines whether to skip directories that can't be accessed
	skipInaccessible bool
	// maxDepth limits how deep the walker will traverse (0 = unlimited)
	maxDepth int
}

// NewFileSystemWalker creates a new instance of FileSystemWalker with default settings
// This constructor allows for configuration of walker behavior
func NewFileSystemWalker(skipInaccessible bool, maxDepth int) interfaces.DirectoryWalker {
	return &FileSystemWalker{
		skipInaccessible: skipInaccessible,
		maxDepth:         maxDepth,
	}
}

// Walk traverses the directory tree and returns folder information sorted by depth
// This method implements the DirectoryWalker interface with proper error handling
func (fsw *FileSystemWalker) Walk(rootPath string) ([]interfaces.FolderInfo, error) {
	// Validate the root path exists and is accessible
	if err := fsw.validateRootPath(rootPath); err != nil {
		return nil, fmt.Errorf("invalid root path: %w", err)
	}

	// Collect all directories using filepath.Walk
	folders, err := fsw.collectDirectories(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to collect directories: %w", err)
	}

	// Sort folders by depth (deepest first) for safe bottom-up processing
	fsw.sortFoldersByDepth(folders)

	return folders, nil
}

// validateRootPath ensures the root path exists and is a directory
// This method provides early validation to prevent unnecessary processing
func (fsw *FileSystemWalker) validateRootPath(rootPath string) error {
	// Convert to absolute path for consistency
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return fmt.Errorf("unable to resolve absolute path: %w", err)
	}

	// Check if path exists and is accessible
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path not accessible: %w", err)
	}

	// Ensure it's a directory
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory")
	}

	return nil
}

// collectDirectories recursively collects all directories in the tree
// This method handles errors gracefully and maintains a complete directory list
func (fsw *FileSystemWalker) collectDirectories(rootPath string) ([]interfaces.FolderInfo, error) {
	var folders []interfaces.FolderInfo
	var collectErrors []error

	// Use filepath.Walk for comprehensive directory traversal
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		return fsw.processWalkPath(path, info, err, rootPath, &folders, &collectErrors)
	})

	// If we encountered errors but still have folders, continue with warnings
	if len(collectErrors) > 0 {
		// Log warnings about inaccessible directories
		for _, collectErr := range collectErrors {
			// In a real implementation, this might use a proper logger
			fmt.Printf("Warning: %v\n", collectErr)
		}
	}

	// Return error only if we couldn't collect any folders and had a critical error
	if err != nil && len(folders) == 0 {
		return folders, fmt.Errorf("critical error during directory walk: %w", err)
	}

	return folders, nil
}

// processWalkPath handles each path encountered during directory traversal
// This method implements the logic for each filepath.Walk callback
func (fsw *FileSystemWalker) processWalkPath(path string, info os.FileInfo, err error, rootPath string, folders *[]interfaces.FolderInfo, collectErrors *[]error) error {
	// Handle path access errors
	if err != nil {
		if fsw.skipInaccessible && os.IsPermission(err) {
			*collectErrors = append(*collectErrors, fmt.Errorf("permission denied: %s", path))
			return filepath.SkipDir
		}

		// For problematic paths, try to extract folder info anyway
		if path != rootPath {
			folderInfo := fsw.extractFolderInfoFromPath(path, rootPath)
			*folders = append(*folders, folderInfo)
			*collectErrors = append(*collectErrors, fmt.Errorf("error accessing %s: %w", path, err))
		}

		return filepath.SkipDir
	}

	// Process directories (skip the root directory itself)
	if info.IsDir() && path != rootPath {
		depth := fsw.calculateDepth(path, rootPath)

		// Check depth limit if specified
		if fsw.maxDepth > 0 && depth > fsw.maxDepth {
			return filepath.SkipDir
		}

		folderInfo := interfaces.FolderInfo{
			Path:   path,
			Name:   filepath.Base(path),
			Depth:  depth,
			Parent: filepath.Dir(path),
		}

		*folders = append(*folders, folderInfo)
	}

	return nil
}

// extractFolderInfoFromPath creates FolderInfo from a problematic path
// This method helps recover folder information even when path access fails
func (fsw *FileSystemWalker) extractFolderInfoFromPath(path, rootPath string) interfaces.FolderInfo {
	return interfaces.FolderInfo{
		Path:   path,
		Name:   filepath.Base(path),
		Depth:  fsw.calculateDepth(path, rootPath),
		Parent: filepath.Dir(path),
	}
}

// calculateDepth determines the depth of a path relative to the root
// This method provides consistent depth calculation for sorting
func (fsw *FileSystemWalker) calculateDepth(path, root string) int {
	relPath, err := filepath.Rel(root, path)
	if err != nil {
		return 0
	}

	if relPath == "." {
		return 0
	}

	// Count the number of path separators
	depth := 0
	for _, char := range relPath {
		if char == filepath.Separator {
			depth++
		}
	}
	return depth + 1 // Add 1 because depth is separator count + 1
}

// sortFoldersByDepth sorts folders by depth (deepest first) for bottom-up processing
// This method ensures safe processing order to avoid path conflicts during renames
func (fsw *FileSystemWalker) sortFoldersByDepth(folders []interfaces.FolderInfo) {
	sort.Slice(folders, func(i, j int) bool {
		// Primary sort: deeper folders first
		if folders[i].Depth != folders[j].Depth {
			return folders[i].Depth > folders[j].Depth
		}
		// Secondary sort: stable sort by path for consistency
		return folders[i].Path < folders[j].Path
	})
}
