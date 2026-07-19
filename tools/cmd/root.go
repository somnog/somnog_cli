package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

const Version = "1.0.0"

var rootCmd = &cobra.Command{
    Use:   "somnog",
    Short: "Somnog CLI — full-stack project management tool",
    Long: `Somnog CLI helps you manage your full-stack monorepo.
It can start your development servers, generate new resources,
list API routes, and more.`,
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    rootCmd.AddCommand(
        versionCmd,
        newCmd,
        startCmd,
        generateCmd,
        routesCmd,
        migrateCmd,
        seedCmd,
    )
}

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print the somnog CLI version",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Fprintf(os.Stdout, "somnog version %s\n", Version)
    },
}