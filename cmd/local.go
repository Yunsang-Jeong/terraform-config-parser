package cmd

import (
	"fmt"
	"log"

	"github.com/Yunsang-Jeong/terraform-config-parser/pkg/logger"
	"github.com/Yunsang-Jeong/terraform-config-parser/pkg/parser"
	"github.com/Yunsang-Jeong/terraform-config-parser/pkg/source"

	"github.com/spf13/cobra"
)

var (
	localSubDir string
)

var localCmd = &cobra.Command{
	Use:   "local <path>",
	Short: "Parse Terraform configurations from local filesystem",
	Long: `Parse Terraform configurations from a local directory.
You can specify a subdirectory within the target path.`,
	Example: `  # Parse current directory
  terraform-config-parser local .
  
  # Parse specific directory
  terraform-config-parser local /path/to/terraform
  
  # Parse subdirectory
  terraform-config-parser local ./terraform --subdir modules/vpc`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]

		logger.InfoKV("Processing local directory", "path", path, "subdir", localSubDir)

		src := source.NewLocalSource(path, source.SourceConfig{
			SubDir: localSubDir,
		})

		if err := parseAndOutput(src); err != nil {
			logger.ErrorKV("Failed to parse and output local source", "path", path, "subdir", localSubDir, "error", err)
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(localCmd)

	localCmd.Flags().StringVar(&localSubDir, "subdir", "", "Subdirectory within the target path")
}

func parseAndOutput(src source.Source) error {
	logger.InfoKV("Starting terraform configuration parsing")

	logger.DebugKV("Fetching source")
	fs, rootPath, err := src.Fetch()
	if err != nil {
		return fmt.Errorf("failed to fetch source: %w", err)
	}
	logger.DebugKV("Successfully fetched source", "root_path", rootPath)
	defer src.Cleanup()

	logger.DebugKV("Creating parser and parsing terraform workspace")
	p := parser.NewParser(fs, parser.Simple)
	tfconfig, err := p.ParseTerraformWorkspace(rootPath)
	if err != nil {
		return fmt.Errorf("failed to parse Terraform workspace: %w", err)
	}

	logger.DebugKV("Generating terraform configuration summary")
	summary, err := tfconfig.Summary(true)
	if err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	logger.InfoKV("Successfully completed terraform configuration parsing")
	fmt.Println(string(summary))
	return nil
}
