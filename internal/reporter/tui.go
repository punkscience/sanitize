// Package reporter provides TUI (Terminal User Interface) implementation using Bubble Tea.
// This implementation provides an interactive progress display for better user experience.
package reporter

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"sanitize/internal/interfaces"
)

// TUIReporter implements the ProgressReporter interface using Bubble Tea
// This struct provides an interactive terminal UI for progress reporting
type TUIReporter struct {
	program  *tea.Program
	model    *tuiModel
	complete chan interfaces.ProcessingSummary
	dryRun   bool
}

// tuiModel represents the Bubble Tea model for the TUI
// This struct maintains the state of the interactive display
type tuiModel struct {
	current     int
	total       int
	message     string
	errors      []string
	complete    bool
	summary     interfaces.ProcessingSummary
	dryRun      bool
	showErrors  bool
	windowWidth int
}

// progressMsg represents a progress update message
type progressMsg struct {
	current int
	total   int
	message string
}

// errorMsg represents an error message
type errorMsg struct {
	err error
}

// completeMsg represents completion with summary
type completeMsg struct {
	summary interfaces.ProcessingSummary
}

// NewTUIReporter creates a new TUI progress reporter using Bubble Tea
// This constructor initializes the interactive terminal interface
func NewTUIReporter(dryRun bool) interfaces.ProgressReporter {
	model := &tuiModel{
		dryRun:      dryRun,
		errors:      make([]string, 0),
		windowWidth: 80, // Default width
	}

	program := tea.NewProgram(model, tea.WithAltScreen())

	return &TUIReporter{
		program:  program,
		model:    model,
		complete: make(chan interfaces.ProcessingSummary),
		dryRun:   dryRun,
	}
}

// ReportProgress sends progress updates to the TUI
// This method updates the progress display in real-time
func (tr *TUIReporter) ReportProgress(current, total int, message string) {
	if tr.program != nil {
		tr.program.Send(progressMsg{
			current: current,
			total:   total,
			message: message,
		})
	}
}

// ReportError sends error information to the TUI
// This method adds errors to the display list
func (tr *TUIReporter) ReportError(err error) {
	if tr.program != nil {
		tr.program.Send(errorMsg{err: err})
	}
}

// ReportComplete signals completion and shows the summary
// This method finalizes the TUI display with results
func (tr *TUIReporter) ReportComplete(summary interfaces.ProcessingSummary) {
	if tr.program != nil {
		tr.program.Send(completeMsg{summary: summary})
		// Give the program a moment to process the message
		tr.program.Quit()
	}
}

// Bubble Tea Model Methods

// Init initializes the Bubble Tea model
func (m *tuiModel) Init() tea.Cmd {
	return nil
}

// Update handles Bubble Tea messages and updates the model
func (m *tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		return m, nil

	case progressMsg:
		m.current = msg.current
		m.total = msg.total
		m.message = msg.message
		return m, nil

	case errorMsg:
		m.errors = append(m.errors, msg.err.Error())
		return m, nil

	case completeMsg:
		m.complete = true
		m.summary = msg.summary
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "e":
			m.showErrors = !m.showErrors
			return m, nil
		}
	}

	return m, nil
}

// View renders the TUI display
func (m *tuiModel) View() string {
	var b strings.Builder

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("63")).
		Padding(0, 1)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39"))

	progressStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("40"))

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	// Title
	title := "🔧 Folder Name Sanitizer"
	if m.dryRun {
		title += " (DRY RUN)"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	if m.complete {
		// Show completion summary
		b.WriteString(headerStyle.Render("✅ Processing Complete"))
		b.WriteString("\n\n")

		b.WriteString(fmt.Sprintf("📁 Total folders found: %d\n", m.summary.TotalFolders))
		b.WriteString(fmt.Sprintf("⚡ Folders processed: %d\n", m.summary.ProcessedCount))
		b.WriteString(fmt.Sprintf("✏️  Folders renamed: %d\n", m.summary.RenamedCount))
		b.WriteString(fmt.Sprintf("⏭️  Folders skipped: %d\n", m.summary.SkippedCount))

		if m.summary.ErrorCount > 0 {
			b.WriteString(errorStyle.Render(fmt.Sprintf("❌ Errors encountered: %d", m.summary.ErrorCount)))
			b.WriteString("\n")
		}

		b.WriteString(fmt.Sprintf("⏱️  Time elapsed: %s\n", m.summary.ElapsedTime))

		if m.summary.RenamedCount > 0 {
			if m.dryRun {
				b.WriteString("\n")
				b.WriteString(infoStyle.Render(fmt.Sprintf("💡 %d folders would be renamed. Run without --dry-run to apply changes.", m.summary.RenamedCount)))
			} else {
				b.WriteString("\n")
				b.WriteString(progressStyle.Render(fmt.Sprintf("🎉 Successfully sanitized %d folder names!", m.summary.RenamedCount)))
			}
		} else if m.summary.TotalFolders > 0 {
			b.WriteString("\n")
			b.WriteString(infoStyle.Render("✨ All folder names are already compatible."))
		}

		if len(m.errors) > 0 {
			b.WriteString("\n\n")
			b.WriteString(infoStyle.Render("Press 'e' to toggle error details, 'q' to quit"))
		} else {
			b.WriteString("\n\n")
			b.WriteString(infoStyle.Render("Press 'q' to quit"))
		}

	} else {
		// Show progress
		if m.total > 0 {
			percentage := float64(m.current) / float64(m.total) * 100
			progressBar := m.createProgressBar(percentage)

			b.WriteString(headerStyle.Render("Processing Folders"))
			b.WriteString("\n\n")
			b.WriteString(progressStyle.Render(progressBar))
			b.WriteString("\n")
			b.WriteString(fmt.Sprintf("Progress: %d/%d (%.1f%%)", m.current, m.total, percentage))
			b.WriteString("\n\n")
		}

		if m.message != "" {
			b.WriteString("Current: ")
			b.WriteString(infoStyle.Render(m.message))
			b.WriteString("\n")
		}

		if len(m.errors) > 0 {
			b.WriteString("\n")
			b.WriteString(errorStyle.Render(fmt.Sprintf("⚠️  %d errors encountered", len(m.errors))))
		}

		b.WriteString("\n\n")
		b.WriteString(infoStyle.Render("Press 'q' to quit"))
	}

	// Show errors if requested
	if m.showErrors && len(m.errors) > 0 {
		b.WriteString("\n\n")
		b.WriteString(headerStyle.Render("Error Details:"))
		b.WriteString("\n")
		for i, err := range m.errors {
			if i >= 10 { // Limit to 10 errors to avoid overwhelming the display
				b.WriteString(errorStyle.Render(fmt.Sprintf("... and %d more errors", len(m.errors)-10)))
				break
			}
			b.WriteString(errorStyle.Render(fmt.Sprintf("• %s", err)))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// createProgressBar creates a visual progress bar
func (m *tuiModel) createProgressBar(percentage float64) string {
	width := m.windowWidth - 20 // Leave space for other content
	if width < 20 {
		width = 20
	}
	if width > 60 {
		width = 60
	}

	filled := int(percentage / 100 * float64(width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	return fmt.Sprintf("▕%s▏", bar)
}
