package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)


func findProjectRoot() string {
    dir, err := os.Getwd()
    if err != nil {
        return "."
    }
    for {
        if _, err := os.Stat(filepath.Join(dir, ProjectConfigFile)); err == nil {
            return dir
        }
        parent := filepath.Dir(dir)
        if parent == dir {
            fmt.Fprintln(os.Stderr, "Error: not inside a somnog project directory.")
            fmt.Fprintln(os.Stderr, "Make sure you run this command from within your project.")
            os.Exit(1)
        }
        dir = parent
    }
}


func splitLines(s string) []string {
    return strings.Split(s, "\n")
}

func contains(s, sub string) bool {
    return strings.Contains(s, sub)
}

func trimSpace(s string) string {
    return strings.TrimSpace(s)
}

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

func toLower(s string) string {
    return strings.ToLower(s)
}

func toCamel(s string) string {
    if len(s) == 0 {
        return s
    }
    return strings.ToLower(s[:1]) + s[1:]
}