package cmd

import (
	"fmt"

	"github.com/Yunsang-Jeong/terraform-config-parser/version"
	"github.com/spf13/cobra"
)

var (
	versionLong bool
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long: `Display version information for the terraform-config-parser CLI.
	
The --long flag shows additional build information including commit hash, build time, and Go version.`,
	Example: `  # Show version
  terraform-config-parser version
  
  # Show detailed version info  
  terraform-config-parser version --long`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionLong {
			fmt.Println(version.GetFullVersion())
		} else {
			fmt.Println(version.GetVersion())
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&versionLong, "long", "l", false, "Show detailed version information")
}
