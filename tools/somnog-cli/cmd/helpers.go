package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// findProjectRoot walks up directories to find the somnog project root.
// It identifies the root by looking for a grit.config.ts file.
func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "grit.config.ts")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding project
			fmt.Fprintln(os.Stderr, "Error: not inside a somnog project directory.")
			fmt.Fprintln(os.Stderr, "Make sure you run this command from within your project.")
			os.Exit(1)
		}
		dir = parent
	}
}

// splitLines splits a string into lines.
func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

// contains is a case-sensitive substring check.
func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}

// trimSpace trims leading/trailing whitespace from a string.
func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

// toSnake converts PascalCase to snake_case.
// e.g. "InvoiceItem" → "invoice_item"
func toSnake(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, []rune(strings.ToLower(string(r)))...)
	}
	return string(result)
}

// toPlural creates a naive plural for resource names.
// e.g. "Invoice" → "invoices", "Category" → "categories"
func toPlural(s string) string {
	lower := strings.ToLower(s)
	if strings.HasSuffix(lower, "y") {
		return lower[:len(lower)-1] + "ies"
	}
	if strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "x") ||
		strings.HasSuffix(lower, "z") || strings.HasSuffix(lower, "ch") ||
		strings.HasSuffix(lower, "sh") {
		return lower + "es"
	}
	return lower + "s"
}

// toLower lowercases the entire string.
func toLower(s string) string {
	return strings.ToLower(s)
}

// toCamel converts PascalCase to camelCase.
// e.g. "InvoiceItem" → "invoiceItem"
func toCamel(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
