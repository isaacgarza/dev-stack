package base

import (
	"fmt"

	"github.com/isaacgarza/dev-stack/internal/pkg/cli/handlers/completion"
	"github.com/isaacgarza/dev-stack/internal/pkg/cli/handlers/core"
	"github.com/isaacgarza/dev-stack/internal/pkg/cli/handlers/doctor"
	inithandler "github.com/isaacgarza/dev-stack/internal/pkg/cli/handlers/init"
	"github.com/isaacgarza/dev-stack/internal/pkg/cli/handlers/services"
	"github.com/isaacgarza/dev-stack/internal/pkg/cli/types"
)

// Registry manages command handlers
type Registry struct {
	handlers map[string]types.CommandHandler
}

// NewRegistry creates a new handler registry
func NewRegistry() *Registry {
	registry := &Registry{
		handlers: make(map[string]types.CommandHandler),
	}
	registry.registerDefaultHandlers()
	return registry
}

// RegisterHandler registers a handler for a command
func (r *Registry) RegisterHandler(name string, handler types.CommandHandler) {
	r.handlers[name] = handler
}

// GetHandler retrieves a handler by name
func (r *Registry) GetHandler(name string) (types.CommandHandler, error) {
	handler, exists := r.handlers[name]
	if !exists {
		return nil, fmt.Errorf("handler not found for command: %s", name)
	}
	return handler, nil
}

// GetAllHandlers returns all registered handlers
func (r *Registry) GetAllHandlers() map[string]types.CommandHandler {
	return r.handlers
}

// HasHandler checks if a handler exists for the given command
func (r *Registry) HasHandler(name string) bool {
	_, exists := r.handlers[name]
	return exists
}

// registerDefaultHandlers registers all default command handlers
func (r *Registry) registerDefaultHandlers() {
	r.RegisterHandler("up", core.NewUpHandler())
	r.RegisterHandler("down", core.NewDownHandler())
	r.RegisterHandler("restart", core.NewRestartHandler())
	r.RegisterHandler("status", core.NewStatusHandler())
	r.RegisterHandler("deps", services.NewDepsHandler())
	r.RegisterHandler("conflicts", services.NewConflictsHandler())
	r.RegisterHandler("services", services.NewServicesHandler())
	r.RegisterHandler("init", inithandler.NewInitHandler())
	r.RegisterHandler("doctor", doctor.NewDoctorHandler())
	r.RegisterHandler("completion", completion.NewCompletionHandler())
}
