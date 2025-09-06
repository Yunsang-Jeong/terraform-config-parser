package cmd

import (
	"context"

	"github.com/Yunsang-Jeong/terraform-config-parser/pkg/logger"
	"github.com/Yunsang-Jeong/terraform-config-parser/version"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

var (
	logLevel string
)

var rootCmd = &cobra.Command{
	Use:     "github.com/Yunsang-Jeong/terraform-config-parser",
	Short:   "Parse Terraform configurations from various sources",
	Version: version.GetVersion(),
	Long: `A CLI tool to parse and analyze Terraform configurations from local filesystem 
or remote Git repositories (GitHub/GitLab).`,
	Example: `  # Parse local directory
  terraform-config-parser local terraform-directory
  
  # Parse Git repository
  terraform-config-parser git https://github.com/owner/repo
  
  # Parse specific branch and subdirectory
  terraform-config-parser git https://github.com/owner/repo --branch main --subdir modules/vpc
  
  # Enable debug logging
  terraform-config-parser local . --log-level debug`,
}

func Execute(ctx context.Context) error {
	// Initialize logger
	if err := logger.Init(logLevel); err != nil {
		return err
	}
	defer logger.Sync()

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

func init() {
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", logger.InfoLevel, "Log level (debug, info, error)")

	// Set custom version template for --version flag
	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)
}
