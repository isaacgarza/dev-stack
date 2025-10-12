package base

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/isaacgarza/dev-stack/internal/pkg/cli/types"
)

// MockCommandHandler implements the CommandHandler interface for testing
type MockCommandHandler struct {
	name          string
	requiredFlags []string
	validateError error
	handleError   error
}

func (m *MockCommandHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *types.BaseCommand) error {
	return m.handleError
}

func (m *MockCommandHandler) ValidateArgs(args []string) error {
	return m.validateError
}

func (m *MockCommandHandler) GetRequiredFlags() []string {
	return m.requiredFlags
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	assert.NotNil(t, registry)
	assert.NotNil(t, registry.handlers)

	// Check that default handlers are registered
	expectedHandlers := []string{
		"up", "down", "restart", "status",
		"deps", "conflicts", "services", "init",
	}

	for _, handlerName := range expectedHandlers {
		assert.True(t, registry.HasHandler(handlerName), "Handler %s should be registered", handlerName)
	}
}

func TestRegistry_RegisterHandler(t *testing.T) {
	registry := NewRegistry()
	mockHandler := &MockCommandHandler{name: "test"}

	registry.RegisterHandler("test", mockHandler)

	assert.True(t, registry.HasHandler("test"))

	handler, err := registry.GetHandler("test")
	assert.NoError(t, err)
	assert.Equal(t, mockHandler, handler)
}

func TestRegistry_GetHandler(t *testing.T) {
	registry := NewRegistry()
	mockHandler := &MockCommandHandler{name: "test"}
	registry.RegisterHandler("test", mockHandler)

	t.Run("existing handler", func(t *testing.T) {
		handler, err := registry.GetHandler("test")
		assert.NoError(t, err)
		assert.Equal(t, mockHandler, handler)
	})

	t.Run("non-existing handler", func(t *testing.T) {
		handler, err := registry.GetHandler("nonexistent")
		assert.Error(t, err)
		assert.Nil(t, handler)
		assert.Contains(t, err.Error(), "handler not found for command: nonexistent")
	})
}

func TestRegistry_GetAllHandlers(t *testing.T) {
	registry := NewRegistry()
	mockHandler1 := &MockCommandHandler{name: "test1"}
	mockHandler2 := &MockCommandHandler{name: "test2"}

	registry.RegisterHandler("test1", mockHandler1)
	registry.RegisterHandler("test2", mockHandler2)

	allHandlers := registry.GetAllHandlers()

	assert.Contains(t, allHandlers, "test1")
	assert.Contains(t, allHandlers, "test2")
	assert.Equal(t, mockHandler1, allHandlers["test1"])
	assert.Equal(t, mockHandler2, allHandlers["test2"])

	// Should also contain default handlers
	assert.Contains(t, allHandlers, "up")
	assert.Contains(t, allHandlers, "down")
	assert.Contains(t, allHandlers, "init")
}

func TestRegistry_HasHandler(t *testing.T) {
	registry := NewRegistry()
	mockHandler := &MockCommandHandler{name: "test"}

	t.Run("handler exists", func(t *testing.T) {
		registry.RegisterHandler("test", mockHandler)
		assert.True(t, registry.HasHandler("test"))
	})

	t.Run("handler does not exist", func(t *testing.T) {
		assert.False(t, registry.HasHandler("nonexistent"))
	})

	t.Run("default handlers exist", func(t *testing.T) {
		assert.True(t, registry.HasHandler("up"))
		assert.True(t, registry.HasHandler("down"))
		assert.True(t, registry.HasHandler("restart"))
		assert.True(t, registry.HasHandler("status"))
		assert.True(t, registry.HasHandler("deps"))
		assert.True(t, registry.HasHandler("conflicts"))
		assert.True(t, registry.HasHandler("services"))
		assert.True(t, registry.HasHandler("init"))
	})
}

func TestRegistry_DefaultHandlers(t *testing.T) {
	registry := NewRegistry()

	// Test that all default handlers are properly registered and not nil
	defaultHandlers := []string{
		"up", "down", "restart", "status",
		"deps", "conflicts", "services", "init",
	}

	for _, handlerName := range defaultHandlers {
		t.Run(handlerName+" handler", func(t *testing.T) {
			assert.True(t, registry.HasHandler(handlerName))

			handler, err := registry.GetHandler(handlerName)
			assert.NoError(t, err)
			assert.NotNil(t, handler)

			// Verify it implements the CommandHandler interface
			assert.Implements(t, (*types.CommandHandler)(nil), handler)
		})
	}
}

func TestRegistry_HandlerOverride(t *testing.T) {
	registry := NewRegistry()

	// Get original handler
	originalHandler, err := registry.GetHandler("up")
	assert.NoError(t, err)
	assert.NotNil(t, originalHandler)

	// Override with mock handler
	mockHandler := &MockCommandHandler{name: "mock-up"}
	registry.RegisterHandler("up", mockHandler)

	// Verify override worked
	newHandler, err := registry.GetHandler("up")
	assert.NoError(t, err)
	assert.Equal(t, mockHandler, newHandler)
	assert.NotEqual(t, originalHandler, newHandler)
}

func TestRegistry_EmptyHandlerName(t *testing.T) {
	registry := NewRegistry()
	mockHandler := &MockCommandHandler{name: "empty"}

	// Register handler with empty name
	registry.RegisterHandler("", mockHandler)

	// Should be able to retrieve it
	handler, err := registry.GetHandler("")
	assert.NoError(t, err)
	assert.Equal(t, mockHandler, handler)
	assert.True(t, registry.HasHandler(""))
}

func TestRegistry_NilHandler(t *testing.T) {
	registry := NewRegistry()

	// Register nil handler
	registry.RegisterHandler("nil-handler", nil)

	// Should be able to retrieve nil
	handler, err := registry.GetHandler("nil-handler")
	assert.NoError(t, err)
	assert.Nil(t, handler)
	assert.True(t, registry.HasHandler("nil-handler"))
}

func TestMockCommandHandler(t *testing.T) {
	t.Run("mock handler implementation", func(t *testing.T) {
		mockHandler := &MockCommandHandler{
			name:          "test",
			requiredFlags: []string{"flag1", "flag2"},
			validateError: assert.AnError,
			handleError:   assert.AnError,
		}

		// Test ValidateArgs
		err := mockHandler.ValidateArgs([]string{"arg1"})
		assert.Equal(t, assert.AnError, err)

		// Test GetRequiredFlags
		flags := mockHandler.GetRequiredFlags()
		assert.Equal(t, []string{"flag1", "flag2"}, flags)

		// Test Handle
		err = mockHandler.Handle(context.Background(), nil, nil, nil)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("mock handler without errors", func(t *testing.T) {
		mockHandler := &MockCommandHandler{
			name:          "test",
			requiredFlags: []string{},
		}

		err := mockHandler.ValidateArgs([]string{})
		assert.NoError(t, err)

		err = mockHandler.Handle(context.Background(), nil, nil, nil)
		assert.NoError(t, err)

		flags := mockHandler.GetRequiredFlags()
		assert.Empty(t, flags)
	})
}
