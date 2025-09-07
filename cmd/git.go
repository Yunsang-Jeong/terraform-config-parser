package cmd

import (
	"log"

	"github.com/Yunsang-Jeong/terraform-config-parser/pkg/logger"
	"github.com/Yunsang-Jeong/terraform-config-parser/pkg/source"

	"github.com/spf13/cobra"
)

var (
	gitRef    string
	gitSubDir string
)

var gitCmd = &cobra.Command{
	Use:   "git <url>",
	Short: "Parse Terraform configurations from Git repository",
	Long: `Parse Terraform configurations from a remote Git repository.
Supports both GitHub and GitLab repositories with HTTPS and SSH URLs.
Uses your system's Git configuration for authentication (SSH keys, credential helpers, etc.).

The --ref parameter accepts:
- Branch names: main, develop, feature/xyz
- Tag names: v1.0.0, 1.2.3, release-1.0
- Commit hashes: abc123def, a1b2c3d4e5f6...`,
	Example: `  # Parse default branch
  terraform-config-parser git https://github.com/owner/repo
  
  # Parse specific branch
  terraform-config-parser git https://github.com/owner/repo --ref develop
  
  # Parse specific tag
  terraform-config-parser git https://github.com/owner/repo --ref v1.0.0
  
  # Parse specific commit
  terraform-config-parser git https://github.com/owner/repo --ref abc123def
  
  # Parse subdirectory in specific reference
  terraform-config-parser git https://github.com/owner/repo --ref main --subdir modules/vpc
  
  # SSH URL support (uses your SSH keys automatically)
  terraform-config-parser git git@github.com:owner/repo.git
  
  # Private repositories work with your existing Git credentials`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]

		logger.InfoKV("Processing git repository", "url", url, "ref", gitRef, "subdir", gitSubDir)

		src := source.NewGitSource(url, source.SourceConfig{
			Ref:    gitRef,
			SubDir: gitSubDir,
		})

		if err := parseAndOutput(src); err != nil {
			logger.ErrorKV("Failed to parse and output git source", "url", url, "ref", gitRef, "subdir", gitSubDir, "error", err)
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(gitCmd)

	gitCmd.Flags().StringVarP(&gitRef, "ref", "r", "", "Git reference to use: branch name, tag name, or commit hash (default: repository default branch)")
	gitCmd.Flags().StringVar(&gitSubDir, "subdir", "", "Subdirectory within the repository")
}
