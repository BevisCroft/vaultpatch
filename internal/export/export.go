package export

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Format represents the output serialization format.
type Format string

const (
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
	FormatEnv  Format = "env"
)

// Exporter writes secret maps to files or stdout in a chosen format.
type Exporter struct {
	format Format
}

// New returns an Exporter for the given format string.
// Returns an error if the format is not recognised.
func New(format string) (*Exporter, error) {
	f := Format(strings.ToLower(format))
	switch f {
	case FormatJSON, FormatYAML, FormatEnv:
		return &Exporter{format: f}, nil
	default:
		return nil, fmt.Errorf("export: unsupported format %q (choose json, yaml, env)", format)
	}
}

// Write serializes secrets and writes them to path.
// If path is "-" the output goes to stdout.
func (e *Exporter) Write(secrets map[string]string, path string) error {
	data, err := e.marshal(secrets)
	if err != nil {
		return err
	}
	if path == "-" {
		_, err = fmt.Print(string(data))
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("export: create directory: %w", err)
	}
	return os.WriteFile(path, data, 0o600)
}

func (e *Exporter) marshal(secrets map[string]string) ([]byte, error) {
	switch e.format {
	case FormatJSON:
		return json.MarshalIndent(secrets, "", "  ")
	case FormatYAML:
		return yaml.Marshal(secrets)
	case FormatEnv:
		return marshalEnv(secrets), nil
	default:
		return nil, fmt.Errorf("export: unknown format %q", e.format)
	}
}

func marshalEnv(secrets map[string]string) []byte {
	var sb strings.Builder
	for k, v := range secrets {
		fmt.Fprintf(&sb, "%s=%q\n", strings.ToUpper(k), v)
	}
	return []byte(sb.String())
}
