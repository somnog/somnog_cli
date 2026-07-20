package cmd

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/spf13/cobra"
)

const maintenanceFile = ".maintenance"

var downCmd = &cobra.Command{
    Use:   "down",
    Short: "Put the application in maintenance mode (503 all requests)",
    Long: `Creates a .maintenance file at the project root. Your API middleware
should check for this file and return 503 Service Unavailable when present.

Run 'somnog up' to bring the application back online.`,
    Run: func(cmd *cobra.Command, args []string) {
        printLogo()
        root := findProjectRoot()
        path := filepath.Join(root, maintenanceFile)

        if _, err := os.Stat(path); err == nil {
            fmt.Println(colorYellow("  Application is already in maintenance mode."))
            return
        }

        if err := os.WriteFile(path, []byte("maintenance"), 0644); err != nil {
            fmt.Fprintf(os.Stderr, colorRed("Error: %v\n"), err)
            os.Exit(1)
        }

        fmt.Println(colorYellow("  ⚠  Application is now in MAINTENANCE MODE."))
        fmt.Println(colorGray("  All requests will receive 503 Service Unavailable."))
        fmt.Println(colorGray("  Run 'somnog up' to bring it back online."))
        fmt.Println()
    },
}

var upCmd = &cobra.Command{
    Use:   "up",
    Short: "Bring the application back online",
    Long:  "Removes the .maintenance file, allowing normal request handling to resume.",
    Run: func(cmd *cobra.Command, args []string) {
        printLogo()
        root := findProjectRoot()
        path := filepath.Join(root, maintenanceFile)

        if err := os.Remove(path); err != nil {
            if os.IsNotExist(err) {
                fmt.Println(colorGray("  Application is not in maintenance mode."))
                return
            }
            fmt.Fprintf(os.Stderr, colorRed("Error: %v\n"), err)
            os.Exit(1)
        }

        fmt.Println(colorGreen("  ✓ Application is back ONLINE!"))
        fmt.Println(colorGray("  Normal request handling has resumed."))
        fmt.Println()
    },
}