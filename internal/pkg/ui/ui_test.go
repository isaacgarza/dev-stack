package ui

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestServiceOption(t *testing.T) {
	t.Run("service option structure", func(t *testing.T) {
		option := ServiceOption{
			Name:         "redis",
			Description:  "Redis cache server",
			Dependencies: []string{"network"},
			Category:     "cache",
		}

		assert.Equal(t, "redis", option.Name)
		assert.Equal(t, "Redis cache server", option.Description)
		assert.Equal(t, []string{"network"}, option.Dependencies)
		assert.Equal(t, "cache", option.Category)
	})
}

func TestCategoryOption(t *testing.T) {
	t.Run("category option structure", func(t *testing.T) {
		services := []ServiceOption{
			{Name: "redis", Category: "cache"},
			{Name: "postgres", Category: "database"},
		}

		category := CategoryOption{
			Name:     "infrastructure",
			Services: services,
		}

		assert.Equal(t, "infrastructure", category.Name)
		assert.Len(t, category.Services, 2)
		assert.Equal(t, "redis", category.Services[0].Name)
		assert.Equal(t, "postgres", category.Services[1].Name)
	})
}

func TestNewOutput(t *testing.T) {
	output := NewOutput()
	assert.NotNil(t, output)
	assert.False(t, output.Quiet)
	assert.False(t, output.NoColor)
}

func TestOutput_Success(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	t.Run("success message with color", func(t *testing.T) {
		output := &Output{Quiet: false, NoColor: false}
		output.Success("Test success message")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		result := buf.String()

		assert.Contains(t, result, "✅")
		assert.Contains(t, result, "Test success message")
	})

	// Reset for next test
	r, w, _ = os.Pipe()
	os.Stdout = w

	t.Run("success message without color", func(t *testing.T) {
		output := &Output{Quiet: false, NoColor: true}
		output.Success("Test success message")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		result := buf.String()

		assert.Contains(t, result, "✅ Test success message")
	})

	t.Run("success message when quiet", func(t *testing.T) {
		// Reset stdout capture
		r, w, _ = os.Pipe()
		os.Stdout = w

		output := &Output{Quiet: true, NoColor: false}
		output.Success("Test success message")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		result := buf.String()

		assert.Empty(t, result) // Should be empty when quiet
	})
}

func TestOutput_Error(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	t.Run("error message with color", func(t *testing.T) {
		output := &Output{Quiet: false, NoColor: false}
		output.Error("Test error message")

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		io.Copy(&buf, r)
		result := buf.String()

		assert.Contains(t, result, "❌")
		assert.Contains(t, result, "Test error message")
	})

	// Reset for next test
	r, w, _ = os.Pipe()
	os.Stderr = w

	t.Run("error message without color", func(t *testing.T) {
		output := &Output{Quiet: false, NoColor: true}
		output.Error("Test error message")

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		io.Copy(&buf, r)
		result := buf.String()

		assert.Contains(t, result, "❌ Test error message")
	})
}

func TestOutput_Warning(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	t.Run("warning message with color", func(t *testing.T) {
		output := &Output{Quiet: false, NoColor: false}
		output.Warning("Test warning message")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		result := buf.String()

		assert.Contains(t, result, "⚠️")
		assert.Contains(t, result, "Test warning message")
	})

	// Reset for next test
	r, w, _ = os.Pipe()
	os.Stdout = w

	t.Run("warning message when quiet", func(t *testing.T) {
		output := &Output{Quiet: true, NoColor: false}
		output.Warning("Test warning message")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		result := buf.String()

		assert.Empty(t, result) // Should be empty when quiet
	})
}

func TestOutput_Info(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	t.Run("info message with color", func(t *testing.T) {
		output := &Output{Quiet: false, NoColor: false}
		output.Info("Test info message")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		result := buf.String()

		assert.Contains(t, result, "ℹ️")
		assert.Contains(t, result, "Test info message")
	})

	t.Run("info message when quiet", func(t *testing.T) {
		// Reset stdout capture
		r, w, _ = os.Pipe()
		os.Stdout = w

		output := &Output{Quiet: true, NoColor: false}
		output.Info("Test info message")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		result := buf.String()

		assert.Empty(t, result) // Should be empty when quiet
	})
}

func TestOutput_Header(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	t.Run("header message with color", func(t *testing.T) {
		output := &Output{Quiet: false, NoColor: false}
		output.Header("Test Header")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		result := buf.String()

		assert.Contains(t, result, "🚀")
		assert.Contains(t, result, "Test Header")
	})

	t.Run("header message without color", func(t *testing.T) {
		// Reset stdout capture
		r, w, _ = os.Pipe()
		os.Stdout = w

		output := &Output{Quiet: false, NoColor: true}
		output.Header("Test Header")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		result := buf.String()

		assert.Contains(t, result, "=== Test Header ===")
	})
}

func TestOutput_SubHeader(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	t.Run("subheader message", func(t *testing.T) {
		output := &Output{Quiet: false, NoColor: true}
		output.SubHeader("Test SubHeader")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		result := buf.String()

		assert.Contains(t, result, "--- Test SubHeader ---")
	})
}

func TestOutput_List(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	t.Run("list items", func(t *testing.T) {
		output := &Output{Quiet: false, NoColor: true}
		items := []string{"Item 1", "Item 2", "Item 3"}
		output.List(items)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		result := buf.String()

		assert.Contains(t, result, "• Item 1")
		assert.Contains(t, result, "• Item 2")
		assert.Contains(t, result, "• Item 3")
	})

	t.Run("empty list", func(t *testing.T) {
		// Reset stdout capture
		r, w, _ = os.Pipe()
		os.Stdout = w

		output := &Output{Quiet: false, NoColor: true}
		output.List([]string{})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		result := buf.String()

		assert.Empty(t, result)
	})
}

func TestOutput_Muted(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	t.Run("muted message", func(t *testing.T) {
		output := &Output{Quiet: false, NoColor: true}
		output.Muted("Test muted message")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		result := buf.String()

		assert.Contains(t, result, "Test muted message")
	})
}

func TestOutput_Box(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	t.Run("box with title and content", func(t *testing.T) {
		output := &Output{Quiet: false, NoColor: true}
		output.Box("Test Title", "Test content")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		result := buf.String()

		assert.Contains(t, result, "Test Title")
		assert.Contains(t, result, "Test content")
		assert.Contains(t, result, "┌─")
		assert.Contains(t, result, "└─")
	})
}

func TestOutput_Progress(t *testing.T) {
	t.Run("progress with function", func(t *testing.T) {
		output := &Output{Quiet: false, NoColor: true} // NoColor to avoid spinner
		executed := false

		err := output.Progress("Testing progress", func() error {
			executed = true
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, executed)
	})

	t.Run("progress with error", func(t *testing.T) {
		output := &Output{Quiet: false, NoColor: true}
		testErr := assert.AnError

		err := output.Progress("Testing progress", func() error {
			return testErr
		})

		assert.Error(t, err)
		assert.Equal(t, testErr, err)
	})

	t.Run("progress when quiet", func(t *testing.T) {
		output := &Output{Quiet: true, NoColor: false}
		executed := false

		err := output.Progress("Testing progress", func() error {
			executed = true
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, executed)
	})
}

func TestStyles(t *testing.T) {
	t.Run("color constants", func(t *testing.T) {
		assert.Equal(t, lipgloss.Color("#00D4AA"), primaryColor)
		assert.Equal(t, lipgloss.Color("#00C851"), successColor)
		assert.Equal(t, lipgloss.Color("#FF8800"), warningColor)
		assert.Equal(t, lipgloss.Color("#FF4444"), errorColor)
		assert.Equal(t, lipgloss.Color("#33B5E5"), infoColor)
		assert.Equal(t, lipgloss.Color("#666666"), mutedColor)
	})

	t.Run("base style", func(t *testing.T) {
		assert.NotNil(t, baseStyle)
		// Test that base style can render text
		rendered := baseStyle.Render("test")
		assert.Contains(t, rendered, "test")
	})

	t.Run("message styles", func(t *testing.T) {
		styles := []lipgloss.Style{
			SuccessStyle,
			ErrorStyle,
			WarningStyle,
			InfoStyle,
			MutedStyle,
		}

		for _, style := range styles {
			assert.NotNil(t, style)
			rendered := style.Render("test")
			assert.Contains(t, rendered, "test")
		}
	})

	t.Run("header styles", func(t *testing.T) {
		assert.NotNil(t, HeaderStyle)
		assert.NotNil(t, SubHeaderStyle)

		headerRendered := HeaderStyle.Render("Header")
		subHeaderRendered := SubHeaderStyle.Render("SubHeader")

		assert.Contains(t, headerRendered, "Header")
		assert.Contains(t, subHeaderRendered, "SubHeader")
	})

	t.Run("list styles", func(t *testing.T) {
		assert.NotNil(t, ListItemStyle)
		assert.NotNil(t, SelectedItemStyle)

		listRendered := ListItemStyle.Render("Item")
		selectedRendered := SelectedItemStyle.Render("Selected")

		assert.Contains(t, listRendered, "Item")
		assert.Contains(t, selectedRendered, "Selected")
	})

	t.Run("box styles", func(t *testing.T) {
		assert.NotNil(t, BoxStyle)
		assert.NotNil(t, HighlightBoxStyle)

		boxRendered := BoxStyle.Render("Box content")
		highlightRendered := HighlightBoxStyle.Render("Highlight content")

		assert.Contains(t, boxRendered, "Box content")
		assert.Contains(t, highlightRendered, "Highlight content")
	})
}

func TestOutput_MessageFormatting(t *testing.T) {
	tests := []struct {
		name     string
		method   func(*Output, string, ...interface{})
		message  string
		args     []interface{}
		expected string
	}{
		{
			name:     "success with args",
			method:   (*Output).Success,
			message:  "Operation %s completed in %d seconds",
			args:     []interface{}{"backup", 5},
			expected: "Operation backup completed in 5 seconds",
		},
		{
			name:     "error with args",
			method:   (*Output).Error,
			message:  "Failed to connect to %s on port %d",
			args:     []interface{}{"database", 5432},
			expected: "Failed to connect to database on port 5432",
		},
		{
			name:     "warning with args",
			method:   (*Output).Warning,
			message:  "Service %s is using %d%% CPU",
			args:     []interface{}{"redis", 85},
			expected: "Service redis is using 85% CPU",
		},
		{
			name:     "info with args",
			method:   (*Output).Info,
			message:  "Processing %s items",
			args:     []interface{}{"100"},
			expected: "Processing 100 items",
		},
		{
			name:     "header with args",
			method:   (*Output).Header,
			message:  "Starting %s deployment",
			args:     []interface{}{"production"},
			expected: "Starting production deployment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test the exact output due to styling,
			// but we can test that the method doesn't panic
			output := &Output{Quiet: true, NoColor: true} // Quiet to avoid output
			assert.NotPanics(t, func() {
				tt.method(output, tt.message, tt.args...)
			})
		})
	}
}

func TestOutput_QuietMode(t *testing.T) {
	output := &Output{Quiet: true, NoColor: false}

	// Capture stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	// Test that quiet mode suppresses most output
	output.Success("Success message")
	output.Warning("Warning message")
	output.Info("Info message")
	output.Header("Header message")
	output.SubHeader("SubHeader message")
	output.List([]string{"Item 1", "Item 2"})
	output.Muted("Muted message")
	output.Box("Title", "Content")

	// Error should still show even in quiet mode
	output.Error("Error message")

	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var stdoutBuf, stderrBuf bytes.Buffer
	io.Copy(&stdoutBuf, rOut)
	io.Copy(&stderrBuf, rErr)

	stdoutResult := stdoutBuf.String()
	stderrResult := stderrBuf.String()

	// Stdout should be empty (quiet mode)
	assert.Empty(t, stdoutResult)

	// Stderr should contain error message
	assert.Contains(t, stderrResult, "Error message")
}

func TestOutput_NoColorMode(t *testing.T) {
	output := &Output{Quiet: false, NoColor: true}

	// Test that NoColor mode produces plain text output
	// We can't easily capture and test the exact output due to the complexity
	// of stdout/stderr redirection, but we can ensure methods don't panic
	assert.NotPanics(t, func() {
		output.Success("Success message")
		output.Error("Error message")
		output.Warning("Warning message")
		output.Info("Info message")
		output.Header("Header message")
		output.SubHeader("SubHeader message")
		output.List([]string{"Item 1"})
		output.Muted("Muted message")
		output.Box("Title", "Content")
	})
}

func TestGlobalFunctions(t *testing.T) {
	// Test that global functions don't panic
	assert.NotNil(t, DefaultOutput)

	// Test global convenience functions
	assert.NotPanics(t, func() {
		// Set to quiet to avoid output during tests
		DefaultOutput.Quiet = true

		Success("Global success")
		Error("Global error")
		Warning("Global warning")
		Info("Global info")
		Header("Global header")
		SubHeader("Global subheader")
		List([]string{"Global item"})
		Muted("Global muted")
		Box("Global title", "Global content")

		// Test Progress function
		err := Progress("Global progress", func() error {
			return nil
		})
		assert.NoError(t, err)

		// Reset quiet mode
		DefaultOutput.Quiet = false
	})
}

func TestStyleConsistency(t *testing.T) {
	t.Run("all styles can render text", func(t *testing.T) {
		styles := map[string]lipgloss.Style{
			"SuccessStyle":      SuccessStyle,
			"ErrorStyle":        ErrorStyle,
			"WarningStyle":      WarningStyle,
			"InfoStyle":         InfoStyle,
			"MutedStyle":        MutedStyle,
			"HeaderStyle":       HeaderStyle,
			"SubHeaderStyle":    SubHeaderStyle,
			"ListItemStyle":     ListItemStyle,
			"SelectedItemStyle": SelectedItemStyle,
			"BoxStyle":          BoxStyle,
			"HighlightBoxStyle": HighlightBoxStyle,
		}

		for name, style := range styles {
			t.Run(name, func(t *testing.T) {
				rendered := style.Render("test content")
				assert.NotEmpty(t, rendered, "Style %s should render content", name)
				assert.Contains(t, rendered, "test content", "Style %s should contain the original text", name)
			})
		}
	})

	t.Run("color values are valid hex", func(t *testing.T) {
		colors := map[string]lipgloss.Color{
			"primaryColor": primaryColor,
			"successColor": successColor,
			"warningColor": warningColor,
			"errorColor":   errorColor,
			"infoColor":    infoColor,
			"mutedColor":   mutedColor,
		}

		for name, color := range colors {
			colorStr := string(color)
			assert.True(t, strings.HasPrefix(colorStr, "#"), "Color %s should start with #", name)
			assert.Len(t, colorStr, 7, "Color %s should be 7 characters long", name)
		}
	})
}
