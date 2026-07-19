package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start all development servers (API + Web + Admin)",
	Long: `Starts all development servers in parallel using pnpm turbo.
This runs the Go API with hot-reload (air), the Next.js web app,
and the Next.js admin panel concurrently.`,
	Run: func(cmd *cobra.Command, args []string) {
		runInProject("pnpm", "--parallel", "--filter", "./apps/*", "--if-present", "run", "dev")
	},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		runInProject("go", "run", "./cmd/migrate/...")
	},
}

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Run database seeders",
	Run: func(cmd *cobra.Command, args []string) {
		runInProject("go", "run", "./cmd/seed/...")
	},
}

var routesCmd = &cobra.Command{
	Use:   "routes",
	Short: "List all registered API routes",
	Run: func(cmd *cobra.Command, args []string) {
		// Print routes by parsing the routes.go file statically
		root := findProjectRoot()
		routesFile := root + "/apps/api/internal/routes/routes.go"
		data, err := os.ReadFile(routesFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading routes file: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Registered API routes:")
		fmt.Println("──────────────────────────────────────────────")
		lines := splitLines(string(data))
		methods := []string{`r.GET(`, `r.POST(`, `r.PUT(`, `r.PATCH(`, `r.DELETE(`}
		for _, line := range lines {
			for _, m := range methods {
				if contains(line, m) {
					trimmed := trimSpace(line)
					fmt.Printf("  %s\n", trimmed)
				}
			}
		}
	},
}

// runInProject runs a command in the project root directory and
// forwards stdin/stdout/stderr and OS signals to the child process.
func runInProject(name string, args ...string) {
	root := findProjectRoot()

	c := exec.Command(name, args...)
	c.Dir = root
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigs {
			if c.Process != nil {
				_ = c.Process.Signal(sig)
			}
		}
	}()

	if err := c.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
