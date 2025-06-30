package generator

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	openapi_parser "github.com/dhiaayachi/openmcp/openapi-parser"
	"github.com/getkin/kin-openapi/openapi3"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"unicode"
)

//go:embed templates/mcp-conf.json.tmpl
var McpConfTemplate string

//go:embed templates/mcp.tmpl
var serverTemplate string

// templateData holds all the information needed to render the server template.
type templateData struct {
	PackageName string
	Url         string
	Endpoints   []apiEndpoint
}

// apiEndpoint represents a single API operation to be exposed as an MCP tool.
type apiEndpoint struct {
	ToolName        string
	HandlerFuncName string
	Description     string
	Path            string
	Method          string
	Parameters      []toolParameter
}

// toolParameter represents an input parameter for an MCP tool.
type toolParameter struct {
	Name        string
	JSONName    string
	TypeName    string
	Description string
	Required    bool
}

type serverConfigJson struct {
	McpServerName string
}

func Generate(input string, outputDir string, baseUrl string, module string) error {

	log.Println("Starting OpenAPI to MCP Server Generator...")
	openapiFilePath := input

	// 2. Load and parse the OpenAPI 3 specification.
	spec, err := openapi_parser.ToOpenAPIV3(openapiFilePath)
	if err != nil {
		log.Fatalf("Failed to load OpenAPI spec: %v", err)
	}

	// Validate the loaded spec
	if err := spec.Validate(context.Background()); err != nil {
		log.Fatalf("Invalid OpenAPI spec: %v", err)
	}

	log.Printf("Successfully loaded and validated OpenAPI spec: '%s'", spec.Info.Title)

	// 3. Prepare the data structure for the template.
	data := templateData{
		PackageName: "main",
		Url:         baseUrl,
	}

	// 4. Iterate through paths and operations to build endpoint data.
	for path, pathItem := range spec.Paths.Map() {
		for method, operation := range pathItem.Operations() {
			if operation.Summary == "" {
				log.Printf("Skipping operation %s %s because it's missing an 'operationId'", method, path)
				continue
			}

			reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
			toolName := reg.ReplaceAllString(operation.Summary, "")
			toolName = toCamelCase(strings.ReplaceAll(toolName, " ", "_"))
			endpoint := apiEndpoint{
				ToolName:        toolName,
				HandlerFuncName: toCamelCase("handle_" + strings.ReplaceAll(toolName, " ", "_")),
				Description:     strings.ReplaceAll(operation.Summary, "\"", "\\\""),
				Path:            path,
				Method:          method,
			}

			for _, param := range operation.Parameters {
				p := param.Value
				schema, err := goTypeFromSchema(p.Schema.Value)
				if err != nil {
					log.Printf("Failed to convert schema to Go type: %v", err)
					continue
				}
				endpoint.Parameters = append(endpoint.Parameters, toolParameter{
					Name:        toCamelCase(p.Name),
					JSONName:    p.Name,
					TypeName:    schema,
					Description: strings.ReplaceAll(p.Description, "\"", "\\\""),
					Required:    p.Required,
				})
			}

			// Handle request body as parameters
			if operation.RequestBody != nil && operation.RequestBody.Value != nil {
				content := operation.RequestBody.Value.Content
				if jsonContent, ok := content["application/json"]; ok && jsonContent.Schema != nil {
					if jsonContent.Schema.Value.Type.Includes("object") {
						for propName, propSchema := range jsonContent.Schema.Value.Properties {
							schema, err := goTypeFromSchema(propSchema.Value)
							if err != nil {
								log.Printf("Failed to convert schema to Go type: %v", err)
								continue
							}
							endpoint.Parameters = append(endpoint.Parameters, toolParameter{
								Name:        toCamelCase(propName),
								JSONName:    propName,
								TypeName:    schema,
								Description: strings.ReplaceAll(propSchema.Value.Description, "\"", "\\\""),
								Required:    isRequired(propName, jsonContent.Schema.Value.Required),
							})
						}
					}
				}
			}

			data.Endpoints = append(data.Endpoints, endpoint)
			log.Printf("Discovered endpoint: %s -> %s", endpoint.ToolName, endpoint.HandlerFuncName)
		}
	}

	// 5. Create and parse the template, including the custom function.
	funcMap := template.FuncMap{
		"ToUpperFirst": func(s string) string {
			if s == "" {
				return ""
			}
			r := []rune(s)
			r[0] = unicode.ToUpper(r[0])
			return string(r)
		},
	}

	// 6. Create and parse the template.
	tmpl, err := template.New("mcp-server").Funcs(funcMap).Parse(serverTemplate)
	if err != nil {
		log.Printf("Failed to parse template: %v", err)
		return err
	}

	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		log.Printf("Failed to create output directory: %v", err)
		return err
	}

	fileInfo, err := os.Stat(outputDir)
	if os.IsNotExist(err) {
		log.Fatalf("Error: Directory '%s' does not exist.", outputDir)
	}
	if err != nil {
		log.Fatalf("Error accessing path '%s': %v", outputDir, err)
	}
	if !fileInfo.IsDir() {
		log.Fatalf("Error: The provided path '%s' is a file, not a directory.", outputDir)
	}

	// --- 2. Check if the Directory is Empty ---
	// os.ReadDir reads the directory and returns a slice of its entries.
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		log.Fatalf("Error reading directory contents: %v", err)
	}

	// If the slice of entries is empty, there's nothing to do.
	if len(entries) != 0 {
		// --- 3. Ask for User Confirmation ---
		// If the directory is not empty, inform the user and ask for confirmation.
		fmt.Printf("Warning: The directory '%s' is not empty.\n", outputDir)
		fmt.Print("Are you sure you want to delete all of its contents? (y/N): ")

		// Create a new reader to get input from the user's terminal.
		reader := bufio.NewReader(os.Stdin)
		// Read user input until they press Enter.
		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error reading your response: %v", err)
		}

		// Normalize the response to lowercase and remove leading/trailing whitespace.
		response = strings.ToLower(strings.TrimSpace(response))

		// --- 4. Process User's Decision ---
		// If the user did not explicitly type "y" or "yes", cancel the operation.
		if response == "y" || response == "yes" {
			// If the user confirmed, proceed with deletion.
			fmt.Println("Proceeding with deletion...")
			for _, entry := range entries {
				// Construct the full path for each file/subdirectory.
				fullPath := filepath.Join(outputDir, entry.Name())
				// os.RemoveAll can delete both files and non-empty directories.
				if err := os.RemoveAll(fullPath); err != nil {
					// If an error occurs, print it to stderr but continue trying
					// to delete the other files.
					fmt.Fprintf(os.Stderr, "Failed to delete '%s': %v\n", fullPath, err)
				} else {
					fmt.Printf("Deleted: %s\n", fullPath)
				}
			}
		}
	}
	// 7. Generate the outputDir file.
	outputFile, err := os.Create(outputDir + "/main.go")
	if err != nil {
		log.Printf("Failed to create outputDir file: %v", err)
		return err
	}
	defer outputFile.Close()

	// 8. Execute the template and write to the file.
	if err := tmpl.Execute(outputFile, data); err != nil {
		log.Printf("Failed to execute template: %v", err)
		return err
	}

	tmplMcpConf, err := template.New("mcp-conf").Parse(McpConfTemplate)
	if err != nil {
		log.Printf("Failed to parse template: %v", err)
		return err
	}
	// 7. Generate the outputDir file.
	outputFileMcpConf, err := os.Create(outputDir + "/mcp.json")
	if err != nil {
		log.Printf("Failed to create mcp conf outputDir file: %v", err)
		return err
	}
	defer outputFileMcpConf.Close()

	dataMCPConf := serverConfigJson{
		McpServerName: toCamelCase(strings.ReplaceAll(spec.Info.Title, " ", "_")),
	}

	// 8. Execute the template and write to the file.
	if err := tmplMcpConf.Execute(outputFileMcpConf, dataMCPConf); err != nil {
		log.Printf("Failed to execute mcp conf template: %v", err)
		return err
	}

	goExecutablePath, err := exec.LookPath("go")
	if err != nil {
		log.Printf("Failed to find go binary: %v", err)
		return err
	}

	command := exec.Command(goExecutablePath, "mod", "init", module)
	command.Dir = outputDir
	err = command.Run()
	if err != nil {
		log.Printf("Failed to execute go mod init: %v", err)
		return err
	}
	command = exec.Command(goExecutablePath, "mod", "tidy")
	command.Dir = outputDir
	err = command.Run()
	if err != nil {
		log.Printf("Failed to execute tidy: %v", err)
		return err
	}

	log.Println("Successfully generated!")
	return nil

}

// toCamelCase converts a snake_case or kebab-case string to camelCase.
func toCamelCase(s string) string {
	var result string
	upperNext := true
	for _, r := range s {
		if r == '_' || r == '-' {
			upperNext = true
			continue
		}
		if upperNext {
			result += string(unicode.ToUpper(r))
			upperNext = false
		} else {
			result += string(r)
		}
	}
	return result
}

// goTypeFromSchema converts an OpenAPI schema type to a Go type.
func goTypeFromSchema(schema *openapi3.Schema) (string, error) {
	if schema.Type.Includes("string") {
		return "String", nil
	}
	if schema.Type.Includes("integer") {
		return "Number", nil
	}
	if schema.Type.Includes("number") {
		return "Number", nil
	}
	if schema.Type.Includes("boolean") {
		return "Boolean", nil
	}
	if schema.Type.Includes("array") {
		return "Array", nil
	}
	if schema.Type.Includes("object") {
		return "Object", nil
	}
	return "", fmt.Errorf("unsupported schema type: %s", schema.Type)
}

// isRequired checks if a property name is in the list of required properties.
func isRequired(propName string, required []string) bool {
	for _, req := range required {
		if req == propName {
			return true
		}
	}
	return false
}
