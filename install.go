package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"
)

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "install tools by global Go version",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		if err := Install(ctx, args); err != nil {
			if errors.Is(err, ErrNotFoundGlobalVersion) {
				fatal(ctx, fmt.Errorf("no specify version. Run `nvs use`"))
			}
			fatal(ctx, err)
		}
	},
}

func Install(ctx context.Context, args []string) error {
	baseDir, err := checkInit()
	if err != nil {
		return err
	}
	v, err := os.ReadFile(filepath.Join(baseDir, globalVersionFile))
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotFoundGlobalVersion
		}
		return err
	}
	parsedVersion, err := parseVersionString(string(v))
	if err != nil {
		return err
	}
	goBasePath, err := findLocalVersion(baseDir, parsedVersion)
	if err != nil {
		return err
	}

	commandArgs := slices.Concat([]string{"install"}, args)
	infof(ctx, "use %s", goBasePath)
	cmd := exec.CommandContext(ctx, filepath.Join(baseDir, "versions", goBasePath, "bin", "go"), commandArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}
