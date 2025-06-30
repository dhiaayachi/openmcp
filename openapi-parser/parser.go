package openapi_parser

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/invopop/yaml"
)

// ToOpenAPIV3 takes a filename for an OpenAPI definition in either JSON or YAML format,
// determines its version, and converts it to OpenAPI v3.
func ToOpenAPIV3(filename string) (*openapi3.T, error) {
	// Read the file content.
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filename, err)
	}

	// First, unmarshal into a generic map to check the version.
	var versionFinder map[string]interface{}
	if err := yaml.Unmarshal(data, &versionFinder); err != nil {
		return nil, fmt.Errorf("error unmarshaling YAML/JSON to find version: %w", err)
	}

	// Check for the "swagger" or "openapi" version key.
	if swaggerVersion, ok := versionFinder["swagger"]; ok {
		if versionStr, isString := swaggerVersion.(string); isString && versionStr == "2.0" {
			return convertV2ToV3(data)
		}
		return nil, fmt.Errorf("unsupported Swagger version: %v", swaggerVersion)

	} else if openAPIVersion, ok := versionFinder["openapi"]; ok {
		if versionStr, isString := openAPIVersion.(string); isString {
			if versionStr[:1] == "3" {
				return parseV3(data)
			}
			return nil, fmt.Errorf("unsupported OpenAPI version: %s", versionStr)
		}
		return nil, fmt.Errorf("invalid openapi version format: %T", openAPIVersion)
	}

	return nil, fmt.Errorf("unable to determine OpenAPI/Swagger version")
}

// convertV2ToV3 handles the conversion of a Swagger/OpenAPI v2 definition to OpenAPI v3.
func convertV2ToV3(data []byte) (*openapi3.T, error) {
	var v2Doc openapi2.T
	// Unmarshal into a v2 document struct.
	if err := yaml.Unmarshal(data, &v2Doc); err != nil {
		return nil, fmt.Errorf("error unmarshaling OpenAPI v2 spec: %w", err)
	}

	// Use the kin-openapi library to convert from v2 to v3.
	v3Doc, err := openapi2conv.ToV3(&v2Doc)
	if err != nil {
		return nil, fmt.Errorf("error converting OpenAPI v2 to v3: %w", err)
	}

	return v3Doc, nil
}

// parseV3 handles parsing an already compliant OpenAPI v3 definition.
func parseV3(data []byte) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	// The loader's LoadFromData method can handle both JSON and YAML.
	v3Doc, err := loader.LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("error loading OpenAPI v3 spec: %w", err)
	}

	// Validate the loaded document.
	if err = v3Doc.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("OpenAPI v3 spec is invalid: %w", err)
	}

	return v3Doc, nil
}
