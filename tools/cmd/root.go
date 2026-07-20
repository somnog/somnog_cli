package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   BinaryName,
    Short: "Somnog CLI — full-stack project management tool",
    Long: `Somnog CLI helps you manage your full-stack monorepo.
It can start your development servers, generate new resources,
list API routes, run migrations, and more.`,
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    generateCmd.Aliases = []string{"g"}

    rootCmd.AddCommand(
        versionCmd,
        newCmd,
        startCmd,
        generateCmd,
        removeCmd,
        routesCmd,
        migrateCmd,
        seedCmd,
        deployCmd,
        updateCmd,
        downCmd,
        upCmd,
    )
}

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print the somnog CLI version",
    Run: func(cmd *cobra.Command, args []string) {
        printLogo()
        fmt.Fprintf(os.Stdout, "  %s version %s\n\n", BinaryName, Version)
    },
}