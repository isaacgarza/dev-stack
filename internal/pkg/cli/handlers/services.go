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

// ServicesHandler handles the services command
type ServicesHandler struct{}

// NewServicesHandler creates a new services handler
func NewServicesHandler() *ServicesHandler {
	return &ServicesHandler{}
}

// Handle executes the services command
func (h *ServicesHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	category, _ := cmd.Flags().GetString("category")

	if category != "" {
		return h.listServicesByCategory(category)
	}

	return h.listAllServices()
}

// listAllServices lists all services grouped by category
func (h *ServicesHandler) listAllServices() error {
	servicesByCategory, err := NewServiceUtils().LoadServicesByCategory()
	if err != nil {
		return fmt.Errorf("failed to load services: %w", err)
	}

	fmt.Println("📦 Available Services:")
	
	for category, services := range servicesByCategory {
		fmt.Printf("\n  %s:\n", strings.Title(category))
		for _, service := range services {
			fmt.Printf("    %-15s - %s", service.Name, service.Description)
			if len(service.Dependencies) > 0 {
				fmt.Printf(" (requires: %s)", strings.Join(service.Dependencies, ", "))
			}
			fmt.Println()
		}
	}

	return nil
}

// listServicesByCategory lists services in a specific category
func (h *ServicesHandler) listServicesByCategory(category string) error {
	servicesByCategory, err := NewServiceUtils().LoadServicesByCategory()
	if err != nil {
		return fmt.Errorf("failed to load services: %w", err)
	}

	services, exists := servicesByCategory[category]
	if !exists {
		return fmt.Errorf("category '%s' not found", category)
	}

	fmt.Printf("📦 %s Services:\n", strings.Title(category))
	for _, service := range services {
		fmt.Printf("  %-15s - %s", service.Name, service.Description)
		if len(service.Dependencies) > 0 {
			fmt.Printf(" (requires: %s)", strings.Join(service.Dependencies, ", "))
		}
		fmt.Println()
	}

	return nil
}

// loadServicesByCategory loads services organized by category
func (h *ServicesHandler) loadServicesByCategory() (map[string][]struct {
	name         string
	description  string
	dependencies []string
}, error) {
	servicesPath := "internal/config/services"
	categories := []string{"database", "cache", "messaging", "observability", "cloud"}
	
	servicesByCategory := make(map[string][]struct {
		name         string
		description  string
		dependencies []string
	})

	for _, category := range categories {
		categoryPath := filepath.Join(servicesPath, category)
		if _, err := os.Stat(categoryPath); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(categoryPath)
		if err != nil {
			continue
		}

		var categoryServices []struct {
			name         string
			description  string
			dependencies []string
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

			description, _ := serviceData["description"].(string)
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

			categoryServices = append(categoryServices, struct {
				name         string
				description  string
				dependencies []string
			}{
				name:         serviceName,
				description:  description,
				dependencies: dependencies,
			})
		}

		if len(categoryServices) > 0 {
			servicesByCategory[category] = categoryServices
		}
	}

	return servicesByCategory, nil
}

// ValidateArgs validates the command arguments
func (h *ServicesHandler) ValidateArgs(args []string) error {
	return nil
}

// GetRequiredFlags returns required flags for this command
func (h *ServicesHandler) GetRequiredFlags() []string {
	return []string{}
}
