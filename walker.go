package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// FolderInfo represents information about a folder that needs to be processed
type FolderInfo struct {
	Path     string
	Name     string
	Depth    int
	Parent   string
}

// sanitizeDirectoryTree walks the directory tree and sanitizes folder names from bottom to top
func sanitizeDirectoryTree(rootPath string, dryRun bool, verbose bool) error {
	// First, collect all directories and their information
	folders, err := collectDirectories(rootPath)
	if err != nil {
		return fmt.Errorf("failed to collect directories: %w", err)
	}

	if verbose {
		fmt.Printf("Found %d directories to process\n", len(folders))
	}

	// Sort folders by depth (deepest first for bottom-up processing)
	sort.Slice(folders, func(i, j int) bool {
		if folders[i].Depth != folders[j].Depth {
			return folders[i].Depth > folders[j].Depth // Deeper folders first
		}
		return folders[i].Path < folders[j].Path // Stable sort by path
	})

	// Process each folder from deepest to shallowest
	renamedCount := 0
	for _, folder := range folders {
		renamed, err := processFolderRename(folder, dryRun, verbose)
		if err != nil {
			return fmt.Errorf("failed to process folder %s: %w", folder.Path, err)
		}
		if renamed {
			renamedCount++
		}
	}

	if verbose || renamedCount > 0 {
		fmt.Printf("Processed %d directories, renamed %d\n", len(folders), renamedCount)
	}

	return nil
}

// collectDirectories recursively collects all directories in the tree
func collectDirectories(rootPath string) ([]FolderInfo, error) {
	var folders []FolderInfo

	// First try the standard filepath.Walk approach
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip directories we can't access but try to collect their info anyway
			if os.IsPermission(err) {
				fmt.Printf("Warning: Permission denied accessing %s\n", path)
				return filepath.SkipDir
			}
			
			// For other errors (like malformed names), try to extract folder info from the path
			fmt.Printf("Warning: Error accessing %s: %v\n", path, err)
			
			// Try to add this problematic directory to our list for processing
			if path != rootPath {
				depth := getPathDepth(path, rootPath)
				parent := filepath.Dir(path)
				name := filepath.Base(path)

				folders = append(folders, FolderInfo{
					Path:   path,
					Name:   name,
					Depth:  depth,
					Parent: parent,
				})
			}
			
			return filepath.SkipDir // Skip this directory's children but continue
		}

		// Only process directories, skip the root directory
		if info.IsDir() && path != rootPath {
			depth := getPathDepth(path, rootPath)
			parent := filepath.Dir(path)
			name := filepath.Base(path)

			folders = append(folders, FolderInfo{
				Path:   path,
				Name:   name,
				Depth:  depth,
				Parent: parent,
			})
		}

		return nil
	})

	// If filepath.Walk had issues but we still found some folders, continue
	// The error might be from encountering problematic directories, which we've already handled
	if err != nil && len(folders) == 0 {
		return folders, fmt.Errorf("failed to collect directories: %w", err)
	}

	return folders, nil
}

// getPathDepth calculates the depth of a path relative to the root
func getPathDepth(path, root string) int {
	relPath, err := filepath.Rel(root, path)
	if err != nil {
		return 0
	}

	if relPath == "." {
		return 0
	}

	// Count the number of separators
	depth := 0
	for _, char := range relPath {
		if char == filepath.Separator {
			depth++
		}
	}
	return depth + 1 // Add 1 because depth is separator count + 1
}

// processFolderRename handles the renaming of a single folder
func processFolderRename(folder FolderInfo, dryRun bool, verbose bool) (bool, error) {
	sanitizedName := sanitizeFolderName(folder.Name)

	// If the name doesn't need to change, skip it
	if sanitizedName == folder.Name {
		if verbose {
			fmt.Printf("✓ %s (no changes needed)\n", folder.Path)
		}
		return false, nil
	}

	// Construct the new path
	newPath := filepath.Join(folder.Parent, sanitizedName)

	// Check if the target already exists
	if _, err := os.Stat(newPath); err == nil {
		// Target exists, need to handle collision
		newPath = handleNameCollision(newPath, sanitizedName)
	}

	if dryRun {
		fmt.Printf("Would rename: %s → %s\n", folder.Path, newPath)
		return true, nil
	}

	// Perform the rename - handle cases where source might be problematic
	err := os.Rename(folder.Path, newPath)
	if err != nil {
		// If rename fails, it might be due to the source path being malformed
		// Log the error but don't fail the entire process
		fmt.Printf("Warning: Failed to rename %s to %s: %v\n", folder.Path, newPath, err)
		return false, nil // Return false for renamed but nil for error to continue processing
	}

	if verbose {
		fmt.Printf("Renamed: %s → %s\n", folder.Path, newPath)
	} else {
		fmt.Printf("Renamed: %s → %s\n", folder.Name, sanitizedName)
	}

	return true, nil
}

// handleNameCollision generates a unique name when the target already exists
func handleNameCollision(basePath, baseName string) string {
	dir := filepath.Dir(basePath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := baseName
	if ext != "" {
		nameWithoutExt = baseName[:len(baseName)-len(ext)]
	}

	counter := 1
	for {
		var newName string
		if ext != "" {
			newName = fmt.Sprintf("%s_%d%s", nameWithoutExt, counter, ext)
		} else {
			newName = fmt.Sprintf("%s_%d", nameWithoutExt, counter)
		}

		newPath := filepath.Join(dir, newName)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}

		counter++
		if counter > 1000 { // Prevent infinite loop
			// Use timestamp as fallback
			newName = fmt.Sprintf("%s_conflict", baseName)
			return filepath.Join(dir, newName)
		}
	}
}