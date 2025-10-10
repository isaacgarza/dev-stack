package display

import (
	"fmt"
	"io"
	"strings"
)

// CreateFormatter creates a formatter based on the specified format
func CreateFormatter(format string, writer io.Writer) (Formatter, error) {
	switch strings.ToLower(format) {
	case "table", "":
		return NewTableFormatter(writer), nil
	case "json":
		return NewJSONFormatter(writer), nil
	case "yaml", "yml":
		return NewYAMLFormatter(writer), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// GetSupportedFormats returns a list of supported output formats
func GetSupportedFormats() []string {
	return []string{"table", "json", "yaml"}
}
