/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package app

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func VersionCmd() *cobra.Command {
	var verbose bool

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the Git Vault version",
		Run: func(cmd *cobra.Command, args []string) {
			if !verbose {
				fmt.Printf("Git Vault version: %s\n", Version)
				return
			}
			fmt.Printf("git-vault version %s\n", Version)
			fmt.Printf("	commit: %s\n", Commit)
			fmt.Printf("	build date: %s\n", BuildDate)
			fmt.Printf("	go version: %s\n", runtime.Version())
			fmt.Printf("	platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}

	versionCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show detailed build information")

	return versionCmd
}
