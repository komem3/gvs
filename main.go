package main

import (
	"context"
	"log"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, loggerOutKey{}, log.New(os.Stdout, "[gvs] ", 0))
	ctx = context.WithValue(ctx, loggerErrKey{}, log.New(os.Stderr, "[gvs] ", 0))

	rootCmd := &cobra.Command{Use: "gvs"}
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "output debug log")

	rootCmd.AddCommand(DownloadCmd)
	rootCmd.AddCommand(InitCmd)
	rootCmd.AddCommand(RunCmd)
	rootCmd.AddCommand(UseCmd)
	rootCmd.AddCommand(VersionsCmd)
	rootCmd.AddCommand(InstallCmd)
	rootCmd.ExecuteContext(ctx)
}
