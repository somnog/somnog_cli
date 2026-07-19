package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
    Use:   "new [project-name]",
    Short: "Create a new somnog full-stack project",
    Long: `Scaffolds a new somnog project by cloning the template repository
and setting up the development environment.

Example:
  somnog new my-app
  cd my-app
  somnog start`,
    Args: cobra.ExactArgs(1),
    Run:  runNew,
}

var templateRepo string


func init() {
    newCmd.Flags().StringVar(&templateRepo, "template", DefaultTemplateRepo, "Git repository to use as template")
}

func runNew(cmd *cobra.Command, args []string) {
    name := args[0]

    if _, err := os.Stat(name); err == nil {
        fmt.Fprintf(os.Stderr, "Error: directory %q already exists\n", name)
        os.Exit(1)
    }

    fmt.Printf("Creating new somnog project: %s\n\n", name)

    fmt.Println("  Cloning template...")
    clone := exec.Command("git", "clone", "--depth=1", templateRepo, name)
    clone.Stdout = os.Stdout
    clone.Stderr = os.Stderr
    if err := clone.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error cloning template: %v\n", err)
        os.Exit(1)
    }

    os.RemoveAll(name + "/.git")

    fmt.Println("  Initializing git...")
    gitInit := exec.Command("git", "init")
    gitInit.Dir = name
    gitInit.Stdout = os.Stdout
    gitInit.Stderr = os.Stderr
    _ = gitInit.Run()

    fmt.Println("  Installing dependencies...")
    install := exec.Command("pnpm", "install")
    install.Dir = name
    install.Stdout = os.Stdout
    install.Stderr = os.Stderr
    if err := install.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Warning: pnpm install failed: %v\n", err)
    }

    fmt.Println()
    fmt.Printf("Project %q created successfully!\n\n", name)
    fmt.Println("Next steps:")
    fmt.Printf("  cd %s\n", name)
    fmt.Println("  cp .env.example .env")
    fmt.Println("  somnog start")
}