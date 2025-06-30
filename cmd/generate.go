/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"github.com/dhiaayachi/openmcp/generator"
	"github.com/spf13/cobra"
)

var (
	input   string
	output  string
	baseUrl string
	module  string
)

// MCPServerData holds the variables that will be passed to the template.
type MCPServerData struct {
	PackageName    string
	Port           int
	HandlerName    string
	HandlerType    string
	HandlerMessage string
}

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate an mcp server",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		if input == "" {
			return errors.New("input file is required")
		}
		if output == "" {
			return errors.New("output directory is empty")
		}
		if baseUrl == "" {
			return errors.New("base url is required")
		}
		if module == "" {
			return errors.New("module is required")
		}
		return generator.Generate(input, output, baseUrl, module)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.PersistentFlags().StringVarP(&input, "input", "i", "", "input openapi file")
	generateCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "output directory, where the generated mcp-server is stored")
	generateCmd.PersistentFlags().StringVarP(&baseUrl, "url", "b", "", "base url for openapi server")
	generateCmd.PersistentFlags().StringVarP(&module, "module", "m", "", "module name")
}
