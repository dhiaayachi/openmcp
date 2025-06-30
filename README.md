![OpenAPI to Model Context Protocol Server Generator](docs/logo.png)
# OpenAPI to Model Context Protocol Server Generator

OpenMCP is a command-line tool that generates a server implementing the Model Context Protocol from an OpenAPI v2 or v3 specification. It takes a JSON or YAML file as input and produces a ready-to-run Go application that exposes the API endpoints as tools.

## Getting Started

### Usage

To generate a server, use the `generate` command:

```bash
openmcp generate -i <input-file> -o <output-directory> -b <base-url> -m <module-name>
```

**Arguments:**

* `-i`, `--input`: Path to the OpenAPI specification file (JSON or YAML).
* `-o`, `--output`: Directory to store the generated server.
* `-b`, `--url`: Base URL for the OpenAPI server.
* `-m`, `--module`: Go module name for the generated project.

**Example:**

```bash
openmcp generate -i examples/simple.json -o ./generated-server -b "http://localhost:8080" -m "github.com/example/generated-server"
```

## Examples

The `examples` directory contains sample OpenAPI specifications:

* `simple.json`: A basic OpenAPI v3 specification.