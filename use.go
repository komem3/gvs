package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var useLocalArg bool

var UseCmd = &cobra.Command{
	Use:   "use [version]",
	Short: "Select Go version",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := Use(cmd.Context(), args[0]); err != nil {
			fatal(cmd.Context(), err)
		}
	},
}

func init() {
	UseCmd.Flags().BoolVar(&useLocalArg, "local", false, "use in local")
}

var (
	globalVersionFile = "version"
	localVersionFile  = ".go-version"
)

func Use(_ context.Context, versionStr string) error {
	var (
		baseDir     string
		versionFile string
		err         error
	)
	if useLocalArg {
		baseDir = "."
		versionFile = localVersionFile
	} else {
		baseDir, err = checkInit()
		if err != nil {
			return err
		}
		versionFile = globalVersionFile
	}
	if err := os.WriteFile(filepath.Join(baseDir, versionFile), []byte(strings.TrimLeft(versionStr, "vgo")), 0644); err != nil {
		return err
	}
	return nil
}
