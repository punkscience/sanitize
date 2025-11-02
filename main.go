// Package main provides the entry point for the sanitize CLI application.
// This implementation uses Cobra for command-line interface and follows SOLID principles.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"sanitize/internal/interfaces"
	"sanitize/internal/processor"
	"sanitize/internal/reporter"
	"sanitize/internal/sanitizer"
	"sanitize/internal/service"
	"sanitize/internal/walker"
)

// CLI flags
var (
	rootPath string
	dryRun   bool
	verbose  bool
	tui      bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sanitize",
	Short: "Sanitize folder names for Windows compatibility",
	Long: `Sanitize recursively walks a folder tree and renames directories to be compatible 
with Windows naming conventions.

Features:
- Removes invalid Windows characters: < > : " | ? * \ /
- Removes control characters (ASCII 0-31)
- Trims trailing spaces and periods
- Handles Windows reserved names (CON, PRN, AUX, NUL, COM1-COM9, LPT1-LPT9)
- Converts Unicode/non-ASCII characters to closest ASCII equivalents
- Enforces 255-character length limit
- Handles name collisions by appending numbers
- Dry-run mode to preview changes
- Verbose output for detailed progress`,
	RunE: runSanitize,
}

// runSanitize executes the main sanitization logic
// This function orchestrates all the components following the Dependency Injection pattern
func runSanitize(cmd *cobra.Command, args []string) error {
	// Convert to absolute path for consistency
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return fmt.Errorf("error resolving path: %w", err)
	}

	// Validate the path exists and is a directory
	if err := validatePath(absPath); err != nil {
		return err
	}

	// Create the dependency chain following SOLID principles
	folderSanitizer := sanitizer.NewWindowsSanitizer()
	directoryWalker := walker.NewFileSystemWalker(true, 0) // Skip inaccessible, no depth limit
	folderProcessor := processor.NewFileSystemProcessor(1000)

	// Create the appropriate reporter based on flags
	var progressReporter interfaces.ProgressReporter
	if tui {
		progressReporter = reporter.NewTUIReporter(dryRun)
	} else {
		progressReporter = reporter.NewCLIReporter(verbose, dryRun)
	}

	// Create the main service with all dependencies injected
	sanitizeService := service.NewSanitizeService(
		folderSanitizer,
		directoryWalker,
		folderProcessor,
		progressReporter,
	)

	// Report the start of processing
	if verbose {
		fmt.Printf("Starting sanitization of directory tree: %s\n", absPath)
		if dryRun {
			fmt.Println("DRY RUN MODE: No changes will be made")
		}
	}

	// Execute the sanitization process
	err = sanitizeService.SanitizeDirectory(absPath, dryRun)
	if err != nil {
		return fmt.Errorf("error during sanitization: %w", err)
	}

	return nil
}

// validatePath ensures the provided path exists and is a directory
// This function provides early validation to prevent unnecessary processing
func validatePath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("error accessing path %s: %w", path, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path %s is not a directory", path)
	}

	return nil
}

// init initializes the CLI flags and configuration
// This function sets up the Cobra command structure
func init() {
	// Define command flags with appropriate defaults and help text
	rootCmd.Flags().StringVarP(&rootPath, "path", "p", ".", "Root path to sanitize")
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Show what would be renamed without making changes")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().BoolVarP(&tui, "tui", "t", false, "Use Terminal UI (Bubble Tea) for interactive progress")
}

// main is the entry point of the application
// This function follows Go best practices for CLI applications
func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
