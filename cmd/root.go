package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "terraform-config-parser",
	Short: "Parse Terraform configurations from various sources",
	Long: `A CLI tool to parse and analyze Terraform configurations from local filesystem 
or remote Git repositories (GitHub/GitLab).`,
	Example: `  # Parse local directory
  terraform-config-parser local ./terraform
  
  # Parse Git repository
  terraform-config-parser git https://github.com/owner/repo
  
  # Parse specific branch and subdirectory
  terraform-config-parser git https://github.com/owner/repo --branch main --subdir modules/vpc`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
