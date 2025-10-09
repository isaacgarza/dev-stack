package handlers

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/isaacgarza/dev-stack/internal/config"
	"gopkg.in/yaml.v3"
)

// ServiceUtils provides shared utilities for service operations
type ServiceUtils struct{}

// NewServiceUtils creates a new service utilities instance
func NewServiceUtils() *ServiceUtils {
	return &ServiceUtils{}
}

// LoadServicesByCategory loads services organized by category from embedded filesystem
func (u *ServiceUtils) LoadServicesByCategory() (map[string][]ServiceInfo, error) {
	categories := []string{"database", "cache", "messaging", "observability", "cloud"}
	servicesByCategory := make(map[string][]ServiceInfo)

	for _, category := range categories {
		categoryPath := filepath.Join("services", category)
		
		entries, err := config.EmbeddedServicesFS.ReadDir(categoryPath)
		if err != nil {
			continue // Category doesn't exist, skip
		}

		var categoryServices []ServiceInfo

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
				continue
			}

			serviceName := strings.TrimSuffix(entry.Name(), ".yaml")
			serviceFile := filepath.Join(categoryPath, entry.Name())
			
			data, err := config.EmbeddedServicesFS.ReadFile(serviceFile)
			if err != nil {
				continue
			}

			var serviceData map[string]interface{}
			if err := yaml.Unmarshal(data, &serviceData); err != nil {
				continue
			}

			serviceInfo := ServiceInfo{
				Name:        serviceName,
				Category:    category,
				Description: getString(serviceData, "description"),
				UsageNotes:  getString(serviceData, "usage_notes"),
			}

			// Extract dependencies
			if deps, exists := serviceData["dependencies"]; exists {
				if depsMap, ok := deps.(map[string]interface{}); ok {
					if required, exists := depsMap["required"]; exists {
						serviceInfo.Dependencies = getStringSlice(required)
					}
				}
			}

			// Extract metadata
			serviceInfo.Options = getStringSlice(serviceData["options"])
			serviceInfo.Examples = getStringSlice(serviceData["examples"])
			serviceInfo.Links = getStringSlice(serviceData["links"])

			categoryServices = append(categoryServices, serviceInfo)
		}

		if len(categoryServices) > 0 {
			servicesByCategory[category] = categoryServices
		}
	}

	return servicesByCategory, nil
}

// LoadServiceConfig loads a service configuration from embedded filesystem
func (u *ServiceUtils) LoadServiceConfig(serviceName string) (*ServiceConfig, error) {
	categories := []string{"database", "cache", "messaging", "observability", "cloud"}
	
	for _, category := range categories {
		servicePath := fmt.Sprintf("services/%s/%s.yaml", category, serviceName)
		data, err := config.EmbeddedServicesFS.ReadFile(servicePath)
		if err != nil {
			continue // Try next category
		}

		var serviceConfig ServiceConfig
		if err := yaml.Unmarshal(data, &serviceConfig); err != nil {
			return nil, fmt.Errorf("failed to parse service config for %s: %w", serviceName, err)
		}

		return &serviceConfig, nil
	}
	
	return nil, fmt.Errorf("service %s not found in any category", serviceName)
}

// LoadAllServiceDependencies loads dependencies for all services
func (u *ServiceUtils) LoadAllServiceDependencies() (map[string][]string, error) {
	serviceMap := make(map[string][]string)
	categories := []string{"database", "cache", "messaging", "observability", "cloud"}
	
	for _, category := range categories {
		categoryPath := filepath.Join("services", category)
		
		entries, err := config.EmbeddedServicesFS.ReadDir(categoryPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
				continue
			}

			serviceName := strings.TrimSuffix(entry.Name(), ".yaml")
			serviceFile := filepath.Join(categoryPath, entry.Name())
			
			data, err := config.EmbeddedServicesFS.ReadFile(serviceFile)
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

// ResolveDependencies resolves service dependencies and returns ordered list
func (u *ServiceUtils) ResolveDependencies(selectedServices []string) ([]string, error) {
	serviceMap, err := u.LoadAllServiceDependencies()
	if err != nil {
		return selectedServices, err
	}

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

// ServiceInfo represents service information
type ServiceInfo struct {
	Name         string
	Description  string
	Category     string
	Dependencies []string
	Options      []string
	Examples     []string
	UsageNotes   string
	Links        []string
}

// Helper functions
func getString(data map[string]interface{}, key string) string {
	if val, exists := data[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getStringSlice(val interface{}) []string {
	if val == nil {
		return nil
	}
	
	if slice, ok := val.([]interface{}); ok {
		var result []string
		for _, item := range slice {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	
	return nil
}
