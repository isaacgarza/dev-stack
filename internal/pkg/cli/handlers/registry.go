package handlers

import (
	"fmt"
)

// Registry manages all command handlers
type Registry struct {
	handlers map[string]CommandHandler
}

// NewRegistry creates a new handler registry
func NewRegistry() *Registry {
	registry := &Registry{
		handlers: make(map[string]CommandHandler),
	}

	// Register all handlers
	registry.registerDefaultHandlers()

	return registry
}

// RegisterHandler registers a command handler
func (r *Registry) RegisterHandler(name string, handler CommandHandler) {
	r.handlers[name] = handler
}

// GetHandler retrieves a command handler by name
func (r *Registry) GetHandler(name string) (CommandHandler, error) {
	handler, exists := r.handlers[name]
	if !exists {
		return nil, fmt.Errorf("handler not found for command: %s", name)
	}
	return handler, nil
}

// GetAllHandlers returns all registered handlers
func (r *Registry) GetAllHandlers() map[string]CommandHandler {
	return r.handlers
}

// HasHandler checks if a handler exists for the given command
func (r *Registry) HasHandler(name string) bool {
	_, exists := r.handlers[name]
	return exists
}

// registerDefaultHandlers registers all default command handlers
func (r *Registry) registerDefaultHandlers() {
	// Core service management commands
	r.RegisterHandler("up", NewUpHandler())
	r.RegisterHandler("down", NewDownHandler())
	r.RegisterHandler("status", NewStatusHandler())

	// Additional handlers can be registered here as they are implemented:
	// r.RegisterHandler("logs", NewLogsHandler())
	// r.RegisterHandler("restart", NewRestartHandler())
	// r.RegisterHandler("monitor", NewMonitorHandler())
	// r.RegisterHandler("doctor", NewDoctorHandler())
	// r.RegisterHandler("exec", NewExecHandler())
	// r.RegisterHandler("connect", NewConnectHandler())
	// r.RegisterHandler("backup", NewBackupHandler())
	// r.RegisterHandler("restore", NewRestoreHandler())
	// r.RegisterHandler("cleanup", NewCleanupHandler())
	// r.RegisterHandler("scale", NewScaleHandler())
	// r.RegisterHandler("init", NewInitHandler())
	// r.RegisterHandler("docs", NewDocsHandler())
	// r.RegisterHandler("validate", NewValidateHandler())
	// r.RegisterHandler("version", NewVersionHandler())
	// r.RegisterHandler("generate", NewGenerateHandler())
}
