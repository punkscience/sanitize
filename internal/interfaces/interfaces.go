// Package interfaces defines the contracts used throughout the sanitize application.
// This follows the Interface Segregation Principle by defining focused, specific interfaces.
package interfaces

// FolderSanitizer defines the contract for sanitizing folder names
// This interface follows the Single Responsibility Principle - it only handles name sanitization
type FolderSanitizer interface {
	// SanitizeName takes a folder name and returns a sanitized version that is Windows-compatible
	SanitizeName(name string) string
}

// DirectoryWalker defines the contract for walking directory trees
// This interface abstracts the directory traversal logic
type DirectoryWalker interface {
	// Walk traverses the directory tree and returns folder information
	Walk(rootPath string) ([]FolderInfo, error)
}

// FolderProcessor defines the contract for processing folder renames
// This interface handles the actual renaming operations
type FolderProcessor interface {
	// ProcessRename handles renaming a single folder with collision detection
	ProcessRename(folder FolderInfo, newName string, dryRun bool) (*RenameResult, error)
}

// ProgressReporter defines the contract for reporting progress during operations
// This interface allows for different UI implementations (CLI, TUI, etc.)
type ProgressReporter interface {
	// ReportProgress sends progress updates during processing
	ReportProgress(current, total int, message string)
	// ReportError sends error information
	ReportError(err error)
	// ReportComplete signals that processing is finished
	ReportComplete(summary ProcessingSummary)
}

// FolderInfo represents information about a folder to be processed
// This struct encapsulates all necessary folder metadata
type FolderInfo struct {
	Path   string // Full path to the folder
	Name   string // Current folder name
	Depth  int    // Depth level from root (for ordering)
	Parent string // Parent directory path
}

// RenameResult contains the outcome of a rename operation
// This struct provides detailed information about what happened during rename
type RenameResult struct {
	Success    bool   // Whether the rename was successful
	OldPath    string // Original path
	NewPath    string // New path after rename
	WasRenamed bool   // Whether the folder actually needed renaming
	Error      error  // Any error that occurred
}

// ProcessingSummary contains statistics about the entire processing operation
// This struct provides a complete overview of what was accomplished
type ProcessingSummary struct {
	TotalFolders   int    // Total number of folders found
	ProcessedCount int    // Number of folders processed
	RenamedCount   int    // Number of folders actually renamed
	ErrorCount     int    // Number of errors encountered
	SkippedCount   int    // Number of folders skipped
	ElapsedTime    string // Time taken for the operation
}
