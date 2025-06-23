/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/dhiaayachi/openmcp/generator"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
)

var (
	input  string
	output string
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
		loader := &openapi3.Loader{Context: cmd.Context(), IsExternalRefsAllowed: true}
		doc, err := loader.LoadFromFile(input)
		if err != nil {
			return fmt.Errorf("loading input file '%s' failed: %v", input, err)
		}
		// Validate document
		err = doc.Validate(cmd.Context())
		if err != nil {
			return fmt.Errorf("validate openapi spec failed: %v", err)
		}
		return generator.Generate(input, output)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.PersistentFlags().StringVarP(&input, "input", "i", "", "input openapi file")
	generateCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "output directory, where the generated mcp-server is stored")
}
