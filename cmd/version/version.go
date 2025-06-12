package version

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/lburgazzoli/gira/internal/version"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display version, commit, and build date information for gira.`,
	RunE:  runVersion,
}

type VersionInfo struct {
	Version string `json:"version" yaml:"version"`
	Commit  string `json:"commit" yaml:"commit"`
	Date    string `json:"date" yaml:"date"`
}

func runVersion(cmd *cobra.Command, args []string) error {
	versionInfo := VersionInfo{
		Version: version.GetVersion(),
		Commit:  version.GetCommit(),
		Date:    version.GetDate(),
	}

	return outputResult(cmd, versionInfo)
}

func outputResult(cmd *cobra.Command, result VersionInfo) error {
	outputFormat, _ := cmd.Root().PersistentFlags().GetString("output")

	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		return encoder.Encode(result)
	case "table":
		return outputTable(result)
	case "":
		return outputPlain(result)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

func outputTable(versionInfo VersionInfo) error {
	table := tablewriter.NewTable(os.Stdout)
	table.Options(tablewriter.WithRendition(
		tw.Rendition{
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenColumns: tw.Off,
				},
			},
		},
	))

	table.Header("Field", "Value")
	table.Append([]string{"Version", versionInfo.Version})
	table.Append([]string{"Commit", versionInfo.Commit})
	table.Append([]string{"Date", versionInfo.Date})

	return table.Render()
}

func outputPlain(versionInfo VersionInfo) error {
	fmt.Printf("version : %s\n", versionInfo.Version)
	fmt.Printf("commit  : %s\n", versionInfo.Commit)
	fmt.Printf("built   : %s\n", versionInfo.Date)
	return nil
}
