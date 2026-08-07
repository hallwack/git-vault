/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package app

import (
	"fmt"

	"git-vault/internal/session"

	"github.com/spf13/cobra"
)

func LockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lock",
		Short: "Lock the repository by clearing the cached master key",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !session.IsActive() {
				fmt.Println("Git Vault is already locked")
				return nil
			}

			if err := session.Clear(); err != nil {
				return err
			}

			fmt.Println("Git Vault has been locked")

			return nil
		},
	}
}
