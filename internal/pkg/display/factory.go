package display

import (
	"fmt"
	"io"
	"strings"
)

// Factory implements FormatterFactory interface
type Factory struct{}

// NewFactory creates a new formatter factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateFormatter creates a formatter based on the specified format
func (f *Factory) CreateFormatter(format string, writer io.Writer) (Formatter, error) {
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
func (f *Factory) GetSupportedFormats() []string {
	return []string{"table", "json", "yaml"}
}
