package cmd

import (
	"fmt"
	"log"

	"github.com/Yunsang-Jeong/terraform-config-parser/pkg/logger"
	"github.com/Yunsang-Jeong/terraform-config-parser/pkg/parser"
	"github.com/Yunsang-Jeong/terraform-config-parser/pkg/source"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
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

		logger.Info("Processing local directory",
			zap.String("path", path),
			zap.String("subdir", localSubDir))

		// Create local source
		src := source.NewLocalSource(path, source.SourceConfig{
			SubDir: localSubDir,
		})

		// Execute parsing
		if err := parseAndOutput(src); err != nil {
			logger.Error("Failed to parse and output local source", zap.Error(err))
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(localCmd)

	localCmd.Flags().StringVar(&localSubDir, "subdir", "", "Subdirectory within the target path")
}

// parseAndOutput is a common function used by both local and git commands
func parseAndOutput(src source.Source) error {
	logger.Info("Starting terraform configuration parsing")

	// Fetch source
	logger.Debug("Fetching source")
	fs, rootPath, err := src.Fetch()
	if err != nil {
		logger.Error("Failed to fetch source", zap.Error(err))
		return fmt.Errorf("failed to fetch source: %w", err)
	}
	logger.Debug("Successfully fetched source", zap.String("root_path", rootPath))
	defer src.Cleanup()

	// Parse Terraform configuration
	logger.Debug("Creating parser and parsing terraform workspace")
	p := parser.NewParser(fs)
	tfconfig, err := p.ParseTerraformWorkspace(rootPath)
	if err != nil {
		logger.Error("Failed to parse terraform workspace", zap.String("root_path", rootPath), zap.Error(err))
		return fmt.Errorf("failed to parse Terraform workspace: %w", err)
	}

	// Generate summary
	logger.Debug("Generating terraform configuration summary")
	summary, err := tfconfig.Summary(true)
	if err != nil {
		logger.Error("Failed to generate summary", zap.Error(err))
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	logger.Info("Successfully completed terraform configuration parsing")
	fmt.Println(string(summary))
	return nil
}
