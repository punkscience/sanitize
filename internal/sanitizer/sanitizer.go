// Package sanitizer provides Windows-compatible folder name sanitization.
// This implementation follows the Single Responsibility Principle by focusing solely on name sanitization logic.
package sanitizer

import (
	"regexp"
	"strings"
	"unicode"

	"sanitize/internal/interfaces"
)

// WindowsSanitizer implements the FolderSanitizer interface for Windows compatibility
// This struct encapsulates all the rules and logic for Windows folder name sanitization
type WindowsSanitizer struct {
	// invalidChars contains characters that are not allowed in Windows folder names
	invalidChars []rune
	// reservedNames contains case-insensitive reserved names in Windows
	reservedNames map[string]bool
	// controlCharsRegex matches ASCII control characters (0-31)
	controlCharsRegex *regexp.Regexp
	// maxNameLength defines the maximum allowed folder name length
	maxNameLength int
}

// NewWindowsSanitizer creates a new instance of WindowsSanitizer with default Windows rules
// This constructor initializes all the Windows-specific rules and constraints
func NewWindowsSanitizer() interfaces.FolderSanitizer {
	return &WindowsSanitizer{
		invalidChars: []rune{'<', '>', ':', '"', '|', '?', '*', '\\', '/'},
		reservedNames: map[string]bool{
			"CON": true, "PRN": true, "AUX": true, "NUL": true,
			"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
			"COM6": true, "COM7": true, "COM8": true, "COM9": true,
			"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
			"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
		},
		controlCharsRegex: regexp.MustCompile(`[\x00-\x1F]`),
		maxNameLength:     255,
	}
}

// SanitizeName sanitizes a folder name according to Windows naming rules
// This method implements the FolderSanitizer interface and ensures Windows compatibility
func (ws *WindowsSanitizer) SanitizeName(name string) string {
	// Handle empty input
	if name == "" {
		return "_empty_"
	}

	// Remove control characters (ASCII 0-31)
	name = ws.controlCharsRegex.ReplaceAllString(name, "")

	// Process each character for validity
	name = ws.processCharacters(name)

	// Apply Windows-specific rules
	name = ws.applyWindowsRules(name)

	return name
}

// processCharacters handles character-by-character processing for Unicode and invalid characters
// This method converts Unicode to ASCII and replaces invalid characters
func (ws *WindowsSanitizer) processCharacters(name string) string {
	// Convert to runes for proper Unicode handling
	runes := []rune(name)
	sanitized := make([]rune, 0, len(runes))

	for _, r := range runes {
		// Check if it's an invalid character
		if ws.containsRune(ws.invalidChars, r) {
			sanitized = append(sanitized, '_')
		} else if r > 127 { // Non-ASCII character
			// Convert Unicode to closest ASCII equivalent
			ascii := ws.unicodeToASCII(r)
			if ascii != 0 {
				sanitized = append(sanitized, ascii)
			} else {
				sanitized = append(sanitized, '_')
			}
		} else {
			sanitized = append(sanitized, r)
		}
	}

	return string(sanitized)
}

// applyWindowsRules applies Windows-specific naming rules
// This method handles trimming, reserved names, and length limits
func (ws *WindowsSanitizer) applyWindowsRules(name string) string {
	// Remove leading/trailing spaces
	name = strings.TrimSpace(name)

	// If empty after trimming, use placeholder
	if name == "" {
		return "_empty_"
	}

	// Remove trailing periods and spaces (Windows doesn't allow this)
	name = strings.TrimRight(name, ". ")

	// If empty after trimming periods/spaces, use placeholder
	if name == "" {
		return "_empty_"
	}

	// Check for reserved names (case insensitive)
	upperName := strings.ToUpper(name)
	if ws.reservedNames[upperName] {
		name = name + "_"
	}

	// Handle length limit
	if len(name) > ws.maxNameLength {
		name = name[:ws.maxNameLength-3] + "..."
	}

	// Final check - if result contains only spaces, replace with placeholder
	if strings.TrimSpace(name) == "" {
		return "_empty_"
	}

	return name
}

// containsRune checks if a slice of runes contains a specific rune
// This helper method provides efficient rune searching
func (ws *WindowsSanitizer) containsRune(slice []rune, r rune) bool {
	for _, item := range slice {
		if item == r {
			return true
		}
	}
	return false
}

// unicodeToASCII converts Unicode characters to their closest ASCII equivalents
// This method provides comprehensive Unicode to ASCII mapping
func (ws *WindowsSanitizer) unicodeToASCII(r rune) rune {
	// Common Unicode to ASCII mappings
	switch {
	case r >= 'À' && r <= 'Ý': // Latin-1 Supplement uppercase
		return ws.unicodeLatinToASCII(r)
	case r >= 'à' && r <= 'ÿ': // Latin-1 Supplement lowercase
		return ws.unicodeLatinToASCII(r)
	case r >= 0x0100 && r <= 0x017F: // Latin Extended-A
		return ws.unicodeExtendedLatinToASCII(r)
	case unicode.IsLetter(r):
		// For other letters, try to find base form
		if unicode.IsUpper(r) {
			return 'A'
		}
		return 'a'
	case unicode.IsDigit(r):
		return '0'
	case unicode.IsSpace(r):
		return ' '
	case unicode.IsPunct(r):
		return '_'
	default:
		return 0 // Will be replaced with underscore
	}
}

// unicodeLatinToASCII handles Latin-1 Supplement characters
// This method provides specific mappings for common Latin characters
func (ws *WindowsSanitizer) unicodeLatinToASCII(r rune) rune {
	replacements := map[rune]rune{
		'À': 'A', 'Á': 'A', 'Â': 'A', 'Ã': 'A', 'Ä': 'A', 'Å': 'A', 'Æ': 'A',
		'Ç': 'C', 'È': 'E', 'É': 'E', 'Ê': 'E', 'Ë': 'E', 'Ì': 'I', 'Í': 'I',
		'Î': 'I', 'Ï': 'I', 'Ð': 'D', 'Ñ': 'N', 'Ò': 'O', 'Ó': 'O', 'Ô': 'O',
		'Õ': 'O', 'Ö': 'O', 'Ø': 'O', 'Ù': 'U', 'Ú': 'U', 'Û': 'U', 'Ü': 'U',
		'Ý': 'Y', 'Þ': 'T', 'ß': 's', 'à': 'a', 'á': 'a', 'â': 'a', 'ã': 'a',
		'ä': 'a', 'å': 'a', 'æ': 'a', 'ç': 'c', 'è': 'e', 'é': 'e', 'ê': 'e',
		'ë': 'e', 'ì': 'i', 'í': 'i', 'î': 'i', 'ï': 'i', 'ð': 'd', 'ñ': 'n',
		'ò': 'o', 'ó': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o', 'ø': 'o', 'ù': 'u',
		'ú': 'u', 'û': 'u', 'ü': 'u', 'ý': 'y', 'þ': 't', 'ÿ': 'y',
	}

	if ascii, exists := replacements[r]; exists {
		return ascii
	}
	return 0
}

// unicodeExtendedLatinToASCII handles Latin Extended-A characters
// This method provides mappings for extended Latin character sets
func (ws *WindowsSanitizer) unicodeExtendedLatinToASCII(r rune) rune {
	// Simplified mapping for common extended Latin characters
	switch {
	case r >= 0x0100 && r <= 0x0105: // Ā ā Ă ă Ą ą
		if r%2 == 0 {
			return 'A'
		}
		return 'a'
	case r >= 0x0106 && r <= 0x010D: // Ć ć Ĉ ĉ Ċ ċ Č č
		if r%2 == 0 {
			return 'C'
		}
		return 'c'
	case r >= 0x010E && r <= 0x0111: // Ď ď Đ đ
		if r%2 == 0 {
			return 'D'
		}
		return 'd'
	case r >= 0x0112 && r <= 0x011B: // Ē ē Ĕ ĕ Ė ė Ę ę Ě ě
		if r%2 == 0 {
			return 'E'
		}
		return 'e'
	default:
		// For other extended Latin, return base ASCII
		if unicode.IsUpper(r) {
			return 'A'
		}
		return 'a'
	}
}
