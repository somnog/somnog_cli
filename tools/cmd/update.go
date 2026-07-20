package cmd

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "strings"

    "github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
    Use:     "update",
    Aliases: []string{"self-update"},
    Short:   "Update the somnog CLI to the latest version",
    Long: `Check GitHub for the latest somnog CLI release and update if a newer
version is available.

If Go is installed, 'go install' is used for a fast update. Otherwise the
binary is downloaded directly from the GitHub release.

Examples:
  somnog update
  somnog update --from-release`,
    RunE: func(cmd *cobra.Command, args []string) error {
        printLogo()
        repo := "somnog/somnog_cli"

        fmt.Println(colorPurple(fmt.Sprintf("  Checking for updates (current: v%s)...", Version)))
        fmt.Println()

        resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo))
        if err != nil {
            return fmt.Errorf("checking latest version: %w", err)
        }
        defer resp.Body.Close()

        var release struct {
            TagName string `json:"tag_name"`
        }
        if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
            return fmt.Errorf("parsing release info: %w", err)
        }

        latest := strings.TrimPrefix(release.TagName, "v")
        current := strings.TrimPrefix(Version, "v")

        if latest == current {
            fmt.Println(colorGreen(fmt.Sprintf("  ✓ Already on the latest version (v%s). Nothing to do.", current)))
            fmt.Println()
            return nil
        }

        fmt.Println(colorGray(fmt.Sprintf("  → New version available: v%s → v%s", current, latest)))
        fmt.Println()

        useRelease, _ := cmd.Flags().GetBool("from-release")

        if !useRelease {
            if _, err := exec.LookPath("go"); err != nil {
                fmt.Println(colorGray("  → Go toolchain not on PATH — downloading binary directly."))
                useRelease = true
            }
        }

        if !useRelease {
            binPath, err := os.Executable()
            if err != nil {
                return fmt.Errorf("finding current binary: %w", err)
            }
            binPath, err = filepath.EvalSymlinks(binPath)
            if err != nil {
                return fmt.Errorf("resolving binary path: %w", err)
            }
            installDir := filepath.Dir(binPath)

            target := fmt.Sprintf("github.com/somnog/somnog_cli/tools@v%s", latest)
            fmt.Println(colorGray(fmt.Sprintf("  → Running: go install %s", target)))

            c := exec.Command("go", "install", target)
            c.Env = append(os.Environ(), "GOBIN="+installDir)
            c.Stdout = os.Stdout
            c.Stderr = os.Stderr
            if err := c.Run(); err != nil {
                fmt.Println(colorYellow("  ! go install failed — falling back to binary download."))
                useRelease = true
            } else {
                fmt.Println()
                fmt.Println(colorGreen(fmt.Sprintf("  ✓ Updated to v%s", latest)))
                fmt.Println(colorGray("  Run 'somnog version' to verify."))
                fmt.Println()
                return nil
            }
        }

        if useRelease {
            if err := downloadBinary(repo, release.TagName, latest); err != nil {
                return err
            }
        }

        return nil
    },
}

func init() {
    updateCmd.Flags().Bool("from-release", false, "Skip 'go install' and download the binary directly from the GitHub release")
}

func downloadBinary(repo, tag, version string) error {
    osName := runtime.GOOS
    arch := runtime.GOARCH
    suffix := ""
    if osName == "windows" {
        suffix = ".exe"
    }

    url := fmt.Sprintf("https://github.com/%s/releases/download/%s/somnog-%s-%s%s",
        repo, tag, osName, arch, suffix)

    fmt.Println(colorGray(fmt.Sprintf("  → Downloading binary for %s/%s...", osName, arch)))

    binPath, err := os.Executable()
    if err != nil {
        return fmt.Errorf("finding current binary: %w", err)
    }
    binPath, err = filepath.EvalSymlinks(binPath)
    if err != nil {
        return fmt.Errorf("resolving binary path: %w", err)
    }

    tmpFile, err := os.CreateTemp("", "somnog-update-*")
    if err != nil {
        return fmt.Errorf("creating temp file: %w", err)
    }
    defer os.Remove(tmpFile.Name())

    dlResp, err := http.Get(url)
    if err != nil {
        return fmt.Errorf("downloading binary: %w", err)
    }
    defer dlResp.Body.Close()

    if dlResp.StatusCode != 200 {
        return fmt.Errorf("download failed: HTTP %d — no binary found for %s/%s at %s",
            dlResp.StatusCode, osName, arch, url)
    }

    if _, err := io.Copy(tmpFile, dlResp.Body); err != nil {
        return fmt.Errorf("writing binary: %w", err)
    }
    tmpFile.Close()

    if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
        return fmt.Errorf("setting permissions: %w", err)
    }

    installDir := filepath.Dir(binPath)
    destPath := filepath.Join(installDir, BinaryName+suffix)

    if err := os.Rename(tmpFile.Name(), destPath); err != nil {
        if osName == "windows" {
            return fmt.Errorf("installing binary: %w", err)
        }
        fmt.Println(colorGray("  → Need sudo to write to " + installDir))
        mvCmd := exec.Command("sudo", "mv", tmpFile.Name(), destPath)
        mvCmd.Stdout = os.Stdout
        mvCmd.Stderr = os.Stderr
        if err := mvCmd.Run(); err != nil {
            return fmt.Errorf("installing binary (sudo): %w", err)
        }
    }

    fmt.Println()
    fmt.Println(colorGreen(fmt.Sprintf("  ✓ Updated to v%s", version)))
    fmt.Println(colorGray("  Run 'somnog version' to verify."))
    fmt.Println()
    return nil
}