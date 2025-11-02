package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var rootPath string
	var dryRun bool
	var verbose bool

	flag.StringVar(&rootPath, "path", ".", "Root path to sanitize (default: current directory)")
	flag.BoolVar(&dryRun, "dry-run", false, "Show what would be renamed without making changes")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.Parse()

	// Convert to absolute path
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		log.Fatalf("Error resolving path: %v", err)
	}

	// Verify the path exists and is a directory
	info, err := os.Stat(absPath)
	if err != nil {
		log.Fatalf("Error accessing path %s: %v", absPath, err)
	}
	if !info.IsDir() {
		log.Fatalf("Path %s is not a directory", absPath)
	}

	if verbose {
		fmt.Printf("Starting sanitization of directory tree: %s\n", absPath)
		if dryRun {
			fmt.Println("DRY RUN MODE: No changes will be made")
		}
	}

	// Start the sanitization process
	err = sanitizeDirectoryTree(absPath, dryRun, verbose)
	if err != nil {
		log.Fatalf("Error during sanitization: %v", err)
	}

	if verbose {
		fmt.Println("Sanitization completed successfully")
	}
}