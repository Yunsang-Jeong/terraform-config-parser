package cmd

import (
	"fmt"
	"log"
	"terraform-config-parser/pkg/parser"
	"terraform-config-parser/pkg/source"

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

		// Create local source
		src := source.NewLocalSource(path, source.SourceConfig{
			SubDir: localSubDir,
		})

		// Execute parsing
		if err := parseAndOutput(src); err != nil {
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
	// Fetch source
	fs, rootPath, err := src.Fetch()
	if err != nil {
		return fmt.Errorf("failed to fetch source: %w", err)
	}
	defer src.Cleanup()

	// Parse Terraform configuration
	p := parser.NewParser(fs)
	tfconfig, err := p.ParseTerraformWorkspace(rootPath)
	if err != nil {
		return fmt.Errorf("failed to parse Terraform workspace: %w", err)
	}

	// Generate summary
	summary, err := tfconfig.Summary(true)
	if err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	fmt.Println(string(summary))
	return nil
}
