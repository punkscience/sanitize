// Package reporter provides progress reporting implementations.
// This package offers both CLI and TUI (Bubble Tea) reporting capabilities.
package reporter

import (
	"fmt"

	"sanitize/internal/interfaces"
)

// CLIReporter implements the ProgressReporter interface for command-line output
// This struct provides simple text-based progress reporting
type CLIReporter struct {
	verbose bool
	dryRun  bool
}

// NewCLIReporter creates a new CLI progress reporter
// This constructor configures the reporter for different output modes
func NewCLIReporter(verbose, dryRun bool) interfaces.ProgressReporter {
	return &CLIReporter{
		verbose: verbose,
		dryRun:  dryRun,
	}
}

// ReportProgress sends progress updates to the console
// This method provides real-time feedback during processing
func (cr *CLIReporter) ReportProgress(current, total int, message string) {
	if cr.verbose {
		fmt.Printf("[%d/%d] %s\n", current, total, message)
	}
}

// ReportError sends error information to the console
// This method ensures errors are visible to the user
func (cr *CLIReporter) ReportError(err error) {
	fmt.Printf("Error: %v\n", err)
}

// ReportComplete signals that processing is finished with a summary
// This method provides a comprehensive overview of the operation results
func (cr *CLIReporter) ReportComplete(summary interfaces.ProcessingSummary) {
	if cr.dryRun {
		fmt.Println("\n=== DRY RUN SUMMARY ===")
		fmt.Println("No changes were made to the file system")
	} else {
		fmt.Println("\n=== PROCESSING SUMMARY ===")
	}

	fmt.Printf("Total folders found: %d\n", summary.TotalFolders)
	fmt.Printf("Folders processed: %d\n", summary.ProcessedCount)
	fmt.Printf("Folders renamed: %d\n", summary.RenamedCount)
	fmt.Printf("Folders skipped: %d\n", summary.SkippedCount)

	if summary.ErrorCount > 0 {
		fmt.Printf("Errors encountered: %d\n", summary.ErrorCount)
	}

	fmt.Printf("Time elapsed: %s\n", summary.ElapsedTime)

	if summary.RenamedCount > 0 {
		if cr.dryRun {
			fmt.Printf("\n%d folders would be renamed. Run without --dry-run to apply changes.\n", summary.RenamedCount)
		} else {
			fmt.Printf("\nSuccessfully sanitized %d folder names.\n", summary.RenamedCount)
		}
	} else if summary.TotalFolders > 0 {
		fmt.Println("\nAll folder names are already compatible.")
	}
}
