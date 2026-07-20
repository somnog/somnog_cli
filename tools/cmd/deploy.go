package cmd

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "github.com/spf13/cobra"
)

var (
    deployHost    string
    deployPort    string
    deployKey     string
    deployDomain  string
    deployAppPort string
)

var deployCmd = &cobra.Command{
    Use:   "deploy",
    Short: "Deploy the application to a remote server via SSH",
    Long: `Build the Go API binary (linux/amd64), upload via SCP, and restart
the systemd service on the remote host.

Configuration can come from flags or environment variables:
  DEPLOY_HOST   — SSH host (e.g. user@server.com)
  DEPLOY_KEY_FILE — path to SSH private key
  DEPLOY_DOMAIN — domain for Caddy reverse proxy

Examples:
  somnog deploy --host user@server.com --domain myapp.com
  somnog deploy --host root@192.168.1.1 --key ~/.ssh/id_rsa`,
    RunE: func(cmd *cobra.Command, args []string) error {
        printLogo()

        if deployHost == "" {
            deployHost = os.Getenv("DEPLOY_HOST")
        }
        if deployKey == "" {
            deployKey = os.Getenv("DEPLOY_KEY_FILE")
        }
        if deployDomain == "" {
            deployDomain = os.Getenv("DEPLOY_DOMAIN")
        }
        if deployHost == "" {
            return fmt.Errorf("--host is required (or set DEPLOY_HOST env var)")
        }

        root := findProjectRoot()
        apiDir := filepath.Join(root, "apps", "api")
        appName := filepath.Base(root)
        binaryName := appName + "-server"
        remotePath := fmt.Sprintf("/opt/%s/%s", appName, binaryName)

        fmt.Println(colorPurple(fmt.Sprintf("  Deploying %s → %s", appName, deployHost)))
        fmt.Println()

        // 1. Build linux/amd64 binary
        fmt.Println(colorGray("  [1/3] Building linux/amd64 binary..."))
        binOut := filepath.Join(apiDir, binaryName)
        build := exec.Command("go", "build", "-ldflags=-s -w", "-o", binOut, "./cmd/server/...")
        build.Dir = apiDir
        build.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=linux", "GOARCH=amd64")
        build.Stdout = os.Stdout
        build.Stderr = os.Stderr
        if err := build.Run(); err != nil {
            return fmt.Errorf("build failed: %w", err)
        }
        defer os.Remove(binOut)

        // 2. SCP binary to server
        fmt.Println(colorGray("  [2/3] Uploading binary..."))
        scpArgs := buildScpArgs(deployPort, deployKey)
        scpArgs = append(scpArgs, binOut, deployHost+":"+remotePath)
        scp := exec.Command("scp", scpArgs...)
        scp.Stdout = os.Stdout
        scp.Stderr = os.Stderr
        if err := scp.Run(); err != nil {
            return fmt.Errorf("upload failed: %w", err)
        }

        // 3. Restart systemd service
        fmt.Println(colorGray("  [3/3] Restarting service..."))
        remoteCmd := fmt.Sprintf(
            "chmod +x %s && sudo systemctl restart %s || echo 'Tip: create a systemd service for %s'",
            remotePath, appName, appName,
        )
        sshArgs := buildSshArgs(deployHost, deployPort, deployKey)
        sshArgs = append(sshArgs, remoteCmd)
        ssh := exec.Command("ssh", sshArgs...)
        ssh.Stdout = os.Stdout
        ssh.Stderr = os.Stderr
        if err := ssh.Run(); err != nil {
            return fmt.Errorf("restart failed: %w", err)
        }

        fmt.Println()
        fmt.Println(colorGreen("  ✓ Deployment successful!"))
        if deployDomain != "" {
            fmt.Println(colorCyan(fmt.Sprintf("  Live at: https://%s", deployDomain)))
        }
        fmt.Println()
        return nil
    },
}

func init() {
    deployCmd.Flags().StringVar(&deployHost, "host", "", "SSH host (e.g. user@server.com) — or DEPLOY_HOST env var")
    deployCmd.Flags().StringVar(&deployPort, "port", "22", "SSH port")
    deployCmd.Flags().StringVar(&deployKey, "key", "", "Path to SSH private key — or DEPLOY_KEY_FILE env var")
    deployCmd.Flags().StringVar(&deployDomain, "domain", "", "Domain for Caddy reverse proxy — or DEPLOY_DOMAIN env var")
    deployCmd.Flags().StringVar(&deployAppPort, "app-port", "8080", "Port the Go API listens on")
}

func buildSshArgs(host, port, key string) []string {
    args := []string{"-o", "StrictHostKeyChecking=no"}
    if port != "" && port != "22" {
        args = append(args, "-p", port)
    }
    if key != "" {
        args = append(args, "-i", key)
    }
    _ = strings.TrimSpace(host)
    args = append(args, host)
    return args
}

func buildScpArgs(port, key string) []string {
    args := []string{"-o", "StrictHostKeyChecking=no"}
    if port != "" && port != "22" {
        args = append(args, "-P", port)
    }
    if key != "" {
        args = append(args, "-i", key)
    }
    return args
}