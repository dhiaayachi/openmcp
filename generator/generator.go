package generator

import (
	"fmt"
	"log"
	"os"
	"text/template"
)

// ServerData holds the variables that will be passed to the template.
type ServerData struct {
	PackageName   string
	MCPPort       int
	MCPServerDesc string
	Tools         []ToolData
}

type ToolData struct {
	HandlerName string
	TargetURL   string
	Desc        string
	Method      string
}

func Generate(input string, output string) error {
	data := ServerData{
		PackageName:   "my_mcp_server",
		MCPPort:       8080,
		MCPServerDesc: "test server",
		Tools: []ToolData{
			{
				HandlerName: "MCPHandler",
				TargetURL:   "http://localhost:9000/api/endpoint",
				Desc:        "Some tool",
				Method:      "GET",
			},
		},
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(output, 0755); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	// Parse the template file.
	tmpl, err := template.ParseFiles("templates/mcp.tmpl")
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}

	// Create the output file for the generated server code.
	outputFile, err := os.Create(output + "/main.go")
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer outputFile.Close() // Ensure the file is closed

	log.Printf("Generating server code to %s/main.go...", data.PackageName)
	// Execute the template, writing the output to the specified file.
	if err := tmpl.Execute(outputFile, data); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	// Parse the template file.
	tmplMod, err := template.ParseFiles("templates/go.mod.tmpl")
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}
	// Create the output file for the generated server code.
	outputModFile, err := os.Create(output + "/go.mod")
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer outputModFile.Close() // Ensure the file is closed

	log.Printf("Generating server code to %s/go.mod...", data.PackageName)
	// Execute the template, writing the output to the specified file.
	if err := tmplMod.Execute(outputModFile, data); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	tmplSum, err := template.ParseFiles("templates/go.sum.tmpl")
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}
	// Create the output file for the generated server code.
	outputSumFile, err := os.Create(output + "/go.sum")
	if err != nil {
		return fmt.Errorf("error creating go.sum file: %v", err)
	}
	defer outputSumFile.Close() // Ensure the file is closed

	log.Printf("Generating server code to %s/go.sum...", data.PackageName)
	// Execute the template, writing the output to the specified file.
	if err := tmplSum.Execute(outputSumFile, data); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	log.Println("Server code generated successfully!")
	log.Printf("Now, navigate to the '%s' directory, run 'go mod init %s' and then 'go mod tidy' to get dependencies.", data.PackageName, data.PackageName)
	log.Println("You can then run your server with: go run main.go")
	return nil
}
