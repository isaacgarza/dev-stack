package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// DepsHandler handles the deps command
type DepsHandler struct{}

// NewDepsHandler creates a new deps handler
func NewDepsHandler() *DepsHandler {
	return &DepsHandler{}
}

// Handle executes the deps command
func (h *DepsHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	if len(args) == 0 {
		return fmt.Errorf("service name required")
	}

	serviceName := args[0]
	return h.showDependencyTree(serviceName)
}

// showDependencyTree shows the dependency tree for a service
func (h *DepsHandler) showDependencyTree(serviceName string) error {
	// Load all service dependencies
	serviceMap, err := h.loadAllServiceDependencies()
	if err != nil {
		return fmt.Errorf("failed to load service dependencies: %w", err)
	}

	// Check if service exists
	if _, exists := serviceMap[serviceName]; !exists {
		return fmt.Errorf("service '%s' not found", serviceName)
	}

	fmt.Printf("🔗 Dependency tree for %s:\n", serviceName)
	
	// Show direct dependencies
	deps := serviceMap[serviceName]
	if len(deps) == 0 {
		fmt.Println("  No dependencies")
		return nil
	}

	// Build and display dependency tree
	visited := make(map[string]bool)
	h.printDependencyTree(serviceName, serviceMap, visited, 0)

	// Show resolved order
	resolvedOrder, err := h.resolveDependencyOrder([]string{serviceName}, serviceMap)
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	fmt.Printf("\n📋 Start order: %s\n", strings.Join(resolvedOrder, " → "))

	return nil
}

// printDependencyTree recursively prints the dependency tree
func (h *DepsHandler) printDependencyTree(serviceName string, serviceMap map[string][]string, visited map[string]bool, depth int) {
	if visited[serviceName] {
		return
	}
	visited[serviceName] = true

	indent := strings.Repeat("  ", depth)
	deps := serviceMap[serviceName]
	
	if depth == 0 {
		fmt.Printf("%s%s\n", indent, serviceName)
	}

	for _, dep := range deps {
		fmt.Printf("%s├── %s\n", indent, dep)
		h.printDependencyTree(dep, serviceMap, visited, depth+1)
	}
}

// loadAllServiceDependencies loads dependencies for all services
func (h *DepsHandler) loadAllServiceDependencies() (map[string][]string, error) {
	servicesPath := "internal/config/services"
	serviceMap := make(map[string][]string)
	
	categories := []string{"database", "cache", "messaging", "observability", "cloud"}
	for _, category := range categories {
		categoryPath := filepath.Join(servicesPath, category)
		if _, err := os.Stat(categoryPath); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(categoryPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
				continue
			}

			serviceName := strings.TrimSuffix(entry.Name(), ".yaml")
			serviceFile := filepath.Join(categoryPath, entry.Name())
			
			data, err := os.ReadFile(serviceFile)
			if err != nil {
				continue
			}

			var serviceData map[string]interface{}
			if err := yaml.Unmarshal(data, &serviceData); err != nil {
				continue
			}

			var dependencies []string
			if deps, exists := serviceData["dependencies"]; exists {
				if depsMap, ok := deps.(map[string]interface{}); ok {
					if required, exists := depsMap["required"]; exists {
						if reqList, ok := required.([]interface{}); ok {
							for _, req := range reqList {
								if reqStr, ok := req.(string); ok {
									dependencies = append(dependencies, reqStr)
								}
							}
						}
					}
				}
			}
			serviceMap[serviceName] = dependencies
		}
	}

	return serviceMap, nil
}

// resolveDependencyOrder resolves dependencies and returns ordered list
func (h *DepsHandler) resolveDependencyOrder(selectedServices []string, serviceMap map[string][]string) ([]string, error) {
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	var result []string

	var visit func(string) error
	visit = func(serviceName string) error {
		if visiting[serviceName] {
			return fmt.Errorf("circular dependency detected involving service: %s", serviceName)
		}
		if visited[serviceName] {
			return nil
		}

		visiting[serviceName] = true
		for _, dep := range serviceMap[serviceName] {
			if err := visit(dep); err != nil {
				return err
			}
		}
		visiting[serviceName] = false
		visited[serviceName] = true
		result = append(result, serviceName)
		return nil
	}

	for _, service := range selectedServices {
		if err := visit(service); err != nil {
			return selectedServices, err
		}
	}

	return result, nil
}

// ValidateArgs validates the command arguments
func (h *DepsHandler) ValidateArgs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("service name required")
	}
	return nil
}

// GetRequiredFlags returns required flags for this command
func (h *DepsHandler) GetRequiredFlags() []string {
	return []string{}
}
