package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/yhlooo/kubectl-cache/pkg/commands/options"
	"github.com/yhlooo/kubectl-cache/pkg/version"
)

const versionTemplate = `Version:   {{ .Version }}
GitCommit: {{ .GitCommit }}
GoVersion: {{ .GoVersion }}
Arch:      {{ .Arch }}
OS:        {{ .OS }}
`

// NewVersionCommandWithOptions 基于选项创建 version 子命令
func NewVersionCommandWithOptions(opts *options.VersionOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Validate(); err != nil {
				return err
			}

			info := version.GetVersion()

			switch opts.OutputFormat {
			case "yaml":
				raw, err := yaml.Marshal(info)
				if err != nil {
					return err
				}
				fmt.Println(string(raw))
			case "json":
				raw, err := json.Marshal(info)
				if err != nil {
					return err
				}
				fmt.Println(string(raw))
			default:
				tpl, err := template.New("version").Parse(versionTemplate)
				if err != nil {
					return err
				}
				return tpl.Execute(os.Stdout, info)
			}

			return nil
		},
	}

	// 将选项绑定到命令行
	opts.AddPFlags(cmd.Flags())

	return cmd
}
