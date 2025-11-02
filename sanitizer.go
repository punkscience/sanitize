package main

import (
	"regexp"
	"strings"
	"unicode"
)

// Windows invalid characters for folder names
var invalidChars = []rune{'<', '>', ':', '"', '|', '?', '*', '\\', '/'}

// Windows reserved names (case insensitive)
var reservedNames = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
	"COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
	"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
}

// Control characters regex (ASCII 0-31)
var controlCharsRegex = regexp.MustCompile(`[\x00-\x1F]`)

// sanitizeFolderName sanitizes a folder name according to Windows rules
func sanitizeFolderName(name string) string {
	if name == "" {
		return "_empty_"
	}

	// Remove control characters (ASCII 0-31)
	name = controlCharsRegex.ReplaceAllString(name, "")

	// Convert to runes for proper Unicode handling
	runes := []rune(name)
	sanitized := make([]rune, 0, len(runes))

	// Process each character
	for _, r := range runes {
		// Check if it's an invalid character
		if containsRune(invalidChars, r) {
			sanitized = append(sanitized, '_')
		} else if r > 127 { // Non-ASCII character
			// Convert Unicode to closest ASCII equivalent
			ascii := unicodeToASCII(r)
			if ascii != 0 {
				sanitized = append(sanitized, ascii)
			} else {
				sanitized = append(sanitized, '_')
			}
		} else {
			sanitized = append(sanitized, r)
		}
	}

	result := string(sanitized)

	// Remove leading/trailing spaces
	result = strings.TrimSpace(result)

	// If empty after trimming, use placeholder
	if result == "" {
		result = "_empty_"
	}

	// Remove trailing periods and spaces (Windows doesn't allow this)
	result = strings.TrimRight(result, ". ")

	// If empty after trimming periods/spaces, use placeholder
	if result == "" {
		result = "_empty_"
	}

	// Check for reserved names (case insensitive)
	upperResult := strings.ToUpper(result)
	if reservedNames[upperResult] {
		result = result + "_"
	}

	// Handle length limit (255 characters)
	if len(result) > 255 {
		result = result[:252] + "..."
	}

	// Final check - if result contains only spaces, replace with placeholder
	if strings.TrimSpace(result) == "" {
		result = "_empty_"
	}

	return result
}

// containsRune checks if a slice of runes contains a specific rune
func containsRune(slice []rune, r rune) bool {
	for _, item := range slice {
		if item == r {
			return true
		}
	}
	return false
}

// unicodeToASCII converts Unicode characters to their closest ASCII equivalents
func unicodeToASCII(r rune) rune {
	// Common Unicode to ASCII mappings
	switch {
	case r >= 'À' && r <= 'Ý': // Latin-1 Supplement uppercase
		return unicodeLatinToASCII(r)
	case r >= 'à' && r <= 'ÿ': // Latin-1 Supplement lowercase
		return unicodeLatinToASCII(r)
	case r >= 0x0100 && r <= 0x017F: // Latin Extended-A
		return unicodeExtendedLatinToASCII(r)
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
func unicodeLatinToASCII(r rune) rune {
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
func unicodeExtendedLatinToASCII(r rune) rune {
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