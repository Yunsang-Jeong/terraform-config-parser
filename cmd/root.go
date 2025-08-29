package cmd

import (
	"context"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "terraform-config-parser",
	Short: "Parse Terraform configurations from various sources",
	Long: `A CLI tool to parse and analyze Terraform configurations from local filesystem 
or remote Git repositories (GitHub/GitLab).`,
	Example: `  # Parse local directory
  terraform-config-parser local terraform-directory
  
  # Parse Git repository
  terraform-config-parser git https://github.com/owner/repo
  
  # Parse specific branch and subdirectory
  terraform-config-parser git https://github.com/owner/repo --branch main --subdir modules/vpc`,
}

func Execute(ctx context.Context) error {
	// Remove help for root command
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	// Remove shell completion
	rootCmd.CompletionOptions = cobra.CompletionOptions{
		DisableDefaultCmd:   true,
		DisableNoDescFlag:   true,
		DisableDescriptions: true,
		HiddenDefaultCmd:    true,
	}

	return fang.Execute(ctx, rootCmd)
}
