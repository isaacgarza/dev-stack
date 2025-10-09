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

// ConflictsHandler handles the conflicts command
type ConflictsHandler struct{}

// NewConflictsHandler creates a new conflicts handler
func NewConflictsHandler() *ConflictsHandler {
	return &ConflictsHandler{}
}

// Handle executes the conflicts command
func (h *ConflictsHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string, base *BaseCommand) error {
	if len(args) < 2 {
		return fmt.Errorf("at least two service names required")
	}

	return h.checkConflicts(args)
}

// checkConflicts checks for conflicts between specified services
func (h *ConflictsHandler) checkConflicts(services []string) error {
	// Load conflict information for all services
	serviceConflicts, err := h.loadServiceConflicts()
	if err != nil {
		return fmt.Errorf("failed to load service conflicts: %w", err)
	}

	fmt.Printf("🔍 Checking conflicts for: %s\n", strings.Join(services, ", "))

	conflicts := h.findConflicts(services, serviceConflicts)
	
	if len(conflicts) == 0 {
		fmt.Println("✅ No conflicts detected")
		return nil
	}

	fmt.Println("⚠️  Conflicts detected:")
	for _, conflict := range conflicts {
		fmt.Printf("  %s\n", conflict)
	}

	return fmt.Errorf("conflicts found between services")
}

// loadServiceConflicts loads conflict information for all services
func (h *ConflictsHandler) loadServiceConflicts() (map[string][]string, error) {
	servicesPath := "internal/config/services"
	serviceConflicts := make(map[string][]string)
	
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

			var conflictsList []string
			if deps, exists := serviceData["dependencies"]; exists {
				if depsMap, ok := deps.(map[string]interface{}); ok {
					if conflictsData, exists := depsMap["conflicts"]; exists {
						if conflictsArray, ok := conflictsData.([]interface{}); ok {
							for _, conflict := range conflictsArray {
								if conflictStr, ok := conflict.(string); ok {
									conflictsList = append(conflictsList, conflictStr)
								}
							}
						}
					}
				}
			}
			serviceConflicts[serviceName] = conflictsList
		}
	}

	return serviceConflicts, nil
}

// findConflicts finds conflicts between services
func (h *ConflictsHandler) findConflicts(services []string, serviceConflicts map[string][]string) []string {
	var conflicts []string

	// Check for conflicts between selected services
	for i, service1 := range services {
		for j, service2 := range services {
			if i >= j {
				continue
			}
			
			// Check if service1 conflicts with service2
			for _, conflict := range serviceConflicts[service1] {
				if conflict == service2 {
					conflicts = append(conflicts, fmt.Sprintf("%s conflicts with %s", service1, service2))
				}
			}
			
			// Check if service2 conflicts with service1
			for _, conflict := range serviceConflicts[service2] {
				if conflict == service1 {
					conflicts = append(conflicts, fmt.Sprintf("%s conflicts with %s", service2, service1))
				}
			}
		}
	}

	return conflicts
}

// ValidateArgs validates the command arguments
func (h *ConflictsHandler) ValidateArgs(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("at least two service names required")
	}
	return nil
}

// GetRequiredFlags returns required flags for this command
func (h *ConflictsHandler) GetRequiredFlags() []string {
	return []string{}
}
