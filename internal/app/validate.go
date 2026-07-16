/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package app

import (
	"fmt"
	"git-vault/internal/config"

	"github.com/spf13/cobra"
)

func ValidateCmd() *cobra.Command {
	var validateCmd = &cobra.Command{
		Use:   "validate",
		Short: "Validate the Git Vault configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := config.Load()

			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "✓ Configuration is valid")
			return nil
		},
	}

	return validateCmd
}
