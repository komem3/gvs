package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize gvs",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := Initialize(); err != nil {
			fatal(cmd.Context(), err)
		}
		fmt.Printf(`Initialize Success.
Add gvs to PATH

export PATH="$HOME/.gvs/bin:$PATH"

And, select global Go version

gvs use 1.22
`)
	},
}

const gvsDir = ".gvs"

func checkInit() (dir string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dir = filepath.Join(home, gvsDir)
	if _, err := os.Stat(dir); err != nil {
		return "", err
	}
	return dir, nil
}

func Initialize() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(home, gvsDir)
	if err := os.Mkdir(dir, 0755); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "bin"), 0755); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "versions"), 0755); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	if err := createScript(dir, "go"); err != nil {
		return fmt.Errorf("create go script: %w", err)
	}
	if err := createScript(dir, "gofmt"); err != nil {
		return fmt.Errorf("create go script: %w", err)
	}
	return nil
}

func createScript(dir, command string) error {
	if err := os.WriteFile(filepath.Join(dir, "bin", command), []byte("#!/bin/bash\ngvs run "+command+" -- \"$@\"\n"), 0744); err != nil {
		return err
	}
	return nil
}
