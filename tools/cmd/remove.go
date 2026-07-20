package cmd

import (
    "bufio"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/spf13/cobra"
)

var removeForce bool

var removeCmd = &cobra.Command{
    Use:     "remove",
    Aliases: []string{"rm"},
    Short:   "Remove components from your project",
}

var removeResourceCmd = &cobra.Command{
    Use:   "resource <Name>",
    Short: "Remove a previously generated resource",
    Long: `Delete all generated files for a resource and reverse the marker-based
injections in routes.go, types/index.ts, schemas/index.ts, and resources/index.ts.

Example:
  somnog remove resource Invoice
  somnog remove resource Invoice --force`,
    Args: cobra.ExactArgs(1),
    Run:  runRemoveResource,
}

func init() {
    removeResourceCmd.Flags().BoolVar(&removeForce, "force", false, "Skip confirmation prompt")
    removeCmd.AddCommand(removeResourceCmd)
}

func runRemoveResource(cmd *cobra.Command, args []string) {
    printLogo()
    name := args[0]

    if len(name) == 0 || name[0] < 'A' || name[0] > 'Z' {
        fmt.Fprintln(os.Stderr, colorRed("Error: resource name must start with an uppercase letter (e.g., Invoice)"))
        os.Exit(1)
    }

    root := findProjectRoot()
    snake := toSnake(name)
    camel := toCamel(name)
    plural := toPlural(snake)

    if !removeForce {
        fmt.Printf(colorYellow("  ⚠  This will remove all files and injections for resource %q.\n"), name)
        fmt.Print("  Continue? [y/N]: ")
        reader := bufio.NewReader(os.Stdin)
        answer, _ := reader.ReadString('\n')
        answer = strings.TrimSpace(strings.ToLower(answer))
        if answer != "y" && answer != "yes" {
            fmt.Println(colorGray("  Cancelled."))
            return
        }
    }

    fmt.Println()
    fmt.Println(colorPurple(fmt.Sprintf("  Removing resource: %s", name)))
    fmt.Println()

    filesToDelete := []string{
        filepath.Join(root, ModelsPath, snake+".go"),
        filepath.Join(root, ServicesPath, snake+"_service.go"),
        filepath.Join(root, HandlersPath, snake+"_handler.go"),
        filepath.Join(root, TSTypesPath, snake+".ts"),
        filepath.Join(root, TSSchemasPath, snake+".ts"),
        filepath.Join(root, AdminResPath, snake+".ts"),
    }

    deleted := 0
    for _, f := range filesToDelete {
        if err := os.Remove(f); err == nil {
            fmt.Printf("  %s %s\n", colorRed("deleted"), colorGray(relPath(root, f)))
            deleted++
        }
    }

    // --- types/index.ts ---
    typesIndex := filepath.Join(root, TSTypesPath, "index.ts")
    removeLineFromFile(typesIndex, fmt.Sprintf("export * from './%s'", snake))
    fmt.Printf("  %s %s\n", colorCyan("cleaned"), colorGray(relPath(root, typesIndex)))

    // --- schemas/index.ts ---
    schemasIndex := filepath.Join(root, TSSchemasPath, "index.ts")
    removeLineFromFile(schemasIndex, fmt.Sprintf("export * from './%s'", snake))
    fmt.Printf("  %s %s\n", colorCyan("cleaned"), colorGray(relPath(root, schemasIndex)))

    // --- resources/index.ts ---
    adminIndex := filepath.Join(root, AdminResPath, "index.ts")
    removeLineContainingFromFile(adminIndex, fmt.Sprintf("%sResource", camel))
    removeLineContainingFromFile(adminIndex, fmt.Sprintf("%sResource", name))
    fmt.Printf("  %s %s\n", colorCyan("cleaned"), colorGray(relPath(root, adminIndex)))

    // --- routes.go ---
    routesFile := filepath.Join(root, APIRoutesPath)
    if data, err := os.ReadFile(routesFile); err == nil {
        content := string(data)

        content = removeLineContaining(content, fmt.Sprintf("New%sHandler", name))

        for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
            content = removeLineContaining(content, fmt.Sprintf(`protected.%s("/%s`, method, plural))
        }

        for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
            content = removeLineContaining(content, fmt.Sprintf(`admin.%s("/%s`, method, plural))
        }

        os.WriteFile(routesFile, []byte(content), 0644)
        fmt.Printf("  %s %s\n", colorCyan("cleaned"), colorGray(relPath(root, routesFile)))
    }

    fmt.Println()
    fmt.Printf(colorGreen("  ✓ Removed resource %q (%d files deleted).\n"), name, deleted)
    fmt.Println()
}

// ── file mutation helpers ────────────────────────────────────────────────────

func removeLine(content, line string) string {
    lines := strings.Split(content, "\n")
    var result []string
    for _, l := range lines {
        if strings.TrimSpace(l) != strings.TrimSpace(line) {
            result = append(result, l)
        }
    }
    return strings.Join(result, "\n")
}

func removeLineContaining(content, substr string) string {
    lines := strings.Split(content, "\n")
    var result []string
    for _, l := range lines {
        if !strings.Contains(l, substr) {
            result = append(result, l)
        }
    }
    return strings.Join(result, "\n")
}

func removeLineFromFile(path, line string) {
    data, err := os.ReadFile(path)
    if err != nil {
        return
    }
    content := removeLine(string(data), line)
    os.WriteFile(path, []byte(content), 0644)
}

func removeLineContainingFromFile(path, substr string) {
    data, err := os.ReadFile(path)
    if err != nil {
        return
    }
    content := removeLineContaining(string(data), substr)
    os.WriteFile(path, []byte(content), 0644)
}