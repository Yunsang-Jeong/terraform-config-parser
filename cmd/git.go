package cmd

import (
	"log"
	"terraform-config-parser/pkg/source"

	"github.com/spf13/cobra"
)

var (
	gitBranch string
	gitSubDir string
)

var gitCmd = &cobra.Command{
	Use:   "git <url>",
	Short: "Parse Terraform configurations from Git repository",
	Long: `Parse Terraform configurations from a remote Git repository.
Supports both GitHub and GitLab repositories with HTTPS and SSH URLs.
Uses your system's Git configuration for authentication (SSH keys, credential helpers, etc.).`,
	Example: `  # Parse default branch
  terraform-config-parser git https://github.com/owner/repo
  
  # Parse specific branch
  terraform-config-parser git https://github.com/owner/repo --branch develop
  
  # Parse subdirectory in specific branch
  terraform-config-parser git https://github.com/owner/repo --branch main --subdir modules/vpc
  
  # SSH URL support (uses your SSH keys automatically)
  terraform-config-parser git git@github.com:owner/repo.git
  
  # Private repositories work with your existing Git credentials`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]

		// Create git source (uses system Git configuration)
		src := source.NewGitSource(url, source.SourceConfig{
			Branch: gitBranch,
			SubDir: gitSubDir,
		})

		// Execute parsing
		if err := parseAndOutput(src); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(gitCmd)

	gitCmd.Flags().StringVarP(&gitBranch, "branch", "b", "", "Git branch or tag to use (default: repository default branch)")
	gitCmd.Flags().StringVar(&gitSubDir, "subdir", "", "Subdirectory within the repository")
}
