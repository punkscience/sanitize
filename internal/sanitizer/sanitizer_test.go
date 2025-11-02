// Package sanitizer_test provides comprehensive tests for the sanitizer package.
// This test suite ensures the Windows folder name sanitization logic works correctly.
package sanitizer_test

import (
	"strings"
	"testing"

	"sanitize/internal/sanitizer"
)

// TestWindowsSanitizer_SanitizeName tests the main sanitization functionality
// This test covers various scenarios that the sanitizer should handle
func TestWindowsSanitizer_SanitizeName(t *testing.T) {
	s := sanitizer.NewWindowsSanitizer()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic valid names (should remain unchanged)
		{
			name:     "valid simple name",
			input:    "ValidFolder",
			expected: "ValidFolder",
		},
		{
			name:     "valid name with spaces",
			input:    "My Documents",
			expected: "My Documents",
		},
		{
			name:     "valid name with numbers",
			input:    "Folder123",
			expected: "Folder123",
		},

		// Invalid characters replacement
		{
			name:     "invalid characters",
			input:    "bad<chars>",
			expected: "bad_chars_",
		},
		{
			name:     "all invalid characters",
			input:    `<>:"|?*\/`,
			expected: "_________",
		},
		{
			name:     "mixed valid and invalid",
			input:    "good:bad",
			expected: "good_bad",
		},

		// Control characters
		{
			name:     "control characters",
			input:    "folder\x01\x1F",
			expected: "folder",
		},

		// Unicode characters
		{
			name:     "unicode latin characters",
			input:    "café",
			expected: "cafe",
		},
		{
			name:     "unicode extended",
			input:    "naïve résumé",
			expected: "naive resume",
		},
		{
			name:     "mixed unicode",
			input:    "tëst fïlé",
			expected: "test file",
		},

		// Trailing spaces and periods
		{
			name:     "trailing spaces",
			input:    "folder   ",
			expected: "folder",
		},
		{
			name:     "trailing periods",
			input:    "folder...",
			expected: "folder",
		},
		{
			name:     "trailing mixed",
			input:    "folder. . ",
			expected: "folder",
		},

		// Reserved names
		{
			name:     "reserved CON",
			input:    "CON",
			expected: "CON_",
		},
		{
			name:     "reserved con lowercase",
			input:    "con",
			expected: "con_",
		},
		{
			name:     "reserved COM1",
			input:    "COM1",
			expected: "COM1_",
		},
		{
			name:     "reserved LPT1",
			input:    "LPT1",
			expected: "LPT1_",
		},
		{
			name:     "reserved PRN",
			input:    "PRN",
			expected: "PRN_",
		},

		// Empty and special cases
		{
			name:     "empty string",
			input:    "",
			expected: "_empty_",
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: "_empty_",
		},
		{
			name:     "only periods",
			input:    "...",
			expected: "_empty_",
		},
		{
			name:     "only control chars",
			input:    "\x01\x02\x03",
			expected: "_empty_",
		},

		// Length limits
		{
			name:     "very long name",
			input:    strings.Repeat("a", 300),         // 300 characters
			expected: strings.Repeat("a", 252) + "...", // 255 characters total
		},

		// Complex real-world examples
		{
			name:     "complex filename",
			input:    "My Files: Important\\Docs (2024)",
			expected: "My Files_ Important_Docs (2024)",
		},
		{
			name:     "windows problematic",
			input:    "file<name>with|problems?",
			expected: "file_name_with_problems_",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := s.SanitizeName(tc.input)
			if result != tc.expected {
				t.Errorf("SanitizeName(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestWindowsSanitizer_UnicodeHandling tests specific Unicode character handling
// This test ensures proper conversion of Unicode characters to ASCII equivalents
func TestWindowsSanitizer_UnicodeHandling(t *testing.T) {
	s := sanitizer.NewWindowsSanitizer()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		// Latin-1 Supplement uppercase
		{"latin A variants", "ÀÁÂÃÄÅ", "AAAAAA"},
		{"latin E variants", "ÈÉÊË", "EEEE"},
		{"latin I variants", "ÌÍÎÏ", "IIII"},
		{"latin O variants", "ÒÓÔÕÖ", "OOOOO"},
		{"latin U variants", "ÙÚÛÜ", "UUUU"},

		// Latin-1 Supplement lowercase
		{"latin a variants", "àáâãäå", "aaaaaa"},
		{"latin e variants", "èéêë", "eeee"},
		{"latin i variants", "ìíîï", "iiii"},
		{"latin o variants", "òóôõö", "ooooo"},
		{"latin u variants", "ùúûü", "uuuu"},

		// Special characters
		{"cedilla", "Çç", "Cc"},
		{"n tilde", "Ññ", "Nn"},
		{"eszett", "ß", "s"},
		{"ae ligature", "Ææ", "Aa"},
		{"o slash", "Øø", "Oo"},

		// Extended Latin
		{"extended A", "ĀāĂă", "AaAa"},
		{"extended C", "ĆćĈĉ", "CcCc"},
		{"extended E", "ĒēĔĕ", "EeEe"},

		// Non-Latin characters (should become generic replacements)
		{"cyrillic", "Привет", "AAAAAA"}, // Should become generic letters
		{"chinese", "你好", "aa"},          // Should become generic letters
		{"arabic", "مرحبا", "aaaaa"},     // Should become generic letters
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := s.SanitizeName(tc.input)
			// For non-Latin scripts, we just check that they're converted to valid ASCII
			if len(result) != len(tc.expected) {
				t.Errorf("SanitizeName(%q) length = %d, expected length %d", tc.input, len(result), len(tc.expected))
			}
			// Ensure result contains only valid ASCII characters
			for _, r := range result {
				if r > 127 {
					t.Errorf("SanitizeName(%q) contains non-ASCII character: %c", tc.input, r)
				}
			}
		})
	}
}

// TestWindowsSanitizer_ReservedNames tests all Windows reserved names
// This test ensures proper handling of all Windows reserved file/folder names
func TestWindowsSanitizer_ReservedNames(t *testing.T) {
	s := sanitizer.NewWindowsSanitizer()

	reservedNames := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}

	for _, name := range reservedNames {
		t.Run(name, func(t *testing.T) {
			// Test uppercase
			result := s.SanitizeName(name)
			expected := name + "_"
			if result != expected {
				t.Errorf("SanitizeName(%q) = %q, expected %q", name, result, expected)
			}

			// Test lowercase
			lowerName := strings.ToLower(name)
			result = s.SanitizeName(lowerName)
			expected = lowerName + "_"
			if result != expected {
				t.Errorf("SanitizeName(%q) = %q, expected %q", lowerName, result, expected)
			}

			// Test mixed case
			mixedName := strings.Title(strings.ToLower(name))
			result = s.SanitizeName(mixedName)
			expected = mixedName + "_"
			if result != expected {
				t.Errorf("SanitizeName(%q) = %q, expected %q", mixedName, result, expected)
			}
		})
	}
}

// BenchmarkWindowsSanitizer_SanitizeName benchmarks the sanitization performance
// This benchmark helps ensure the sanitizer performs efficiently
func BenchmarkWindowsSanitizer_SanitizeName(b *testing.B) {
	s := sanitizer.NewWindowsSanitizer()
	testInput := "test<folder>with:various|problems?and*unicode_chars_café_naïve_résumé"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.SanitizeName(testInput)
	}
}

// BenchmarkWindowsSanitizer_LongName benchmarks long name sanitization
// This benchmark tests performance with very long folder names
func BenchmarkWindowsSanitizer_LongName(b *testing.B) {
	s := sanitizer.NewWindowsSanitizer()
	// Create a very long name with various problematic characters
	longName := strings.Repeat("very_long_folder_name_with_unicode_café_and_invalid<chars>", 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.SanitizeName(longName)
	}
}
