package utils

import (
	"fmt"
	"strings"

	"github.com/isaacgarza/dev-stack/internal/config"
	"github.com/isaacgarza/dev-stack/internal/pkg/cli/types"
	"github.com/isaacgarza/dev-stack/internal/pkg/constants"
	"gopkg.in/yaml.v3"
)

// ServiceUtils provides shared utilities for service operations
type ServiceUtils struct{}

// NewServiceUtils creates a new service utilities instance
func NewServiceUtils() *ServiceUtils {
	return &ServiceUtils{}
}

// GetServicesByCategory loads services organized by category
func (u *ServiceUtils) GetServicesByCategory() (map[string][]types.ServiceInfo, error) {
	categories, err := u.getCategories()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]types.ServiceInfo)
	for _, category := range categories {
		services, err := u.getServicesInCategory(category)
		if err != nil {
			continue
		}
		if len(services) > 0 {
			result[category] = services
		}
	}

	return result, nil
}

// LoadServicesByCategory is an alias for backward compatibility
func (u *ServiceUtils) LoadServicesByCategory() (map[string][]types.ServiceInfo, error) {
	return u.GetServicesByCategory()
}

// LoadServiceConfig loads a service configuration
func (u *ServiceUtils) LoadServiceConfig(serviceName string) (*types.ServiceConfig, error) {
	categories, err := u.getCategories()
	if err != nil {
		return nil, err
	}

	for _, category := range categories {
		config, err := u.loadServiceFromCategory(category, serviceName)
		if err == nil {
			return config, nil
		}
	}

	return nil, fmt.Errorf("service %s not found", serviceName)
}

// LoadAllServiceDependencies loads dependencies for all services
func (u *ServiceUtils) LoadAllServiceDependencies() (map[string][]string, error) {
	categories, err := u.getCategories()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	for _, category := range categories {
		deps, err := u.getDependenciesInCategory(category)
		if err != nil {
			continue
		}
		for service, serviceDeps := range deps {
			result[service] = serviceDeps
		}
	}

	return result, nil
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
			return fmt.Errorf("circular dependency detected: %s", serviceName)
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

// Helper methods
func (u *ServiceUtils) getCategories() ([]string, error) {
	entries, err := config.EmbeddedServicesFS.ReadDir("services")
	if err != nil {
		return nil, fmt.Errorf("failed to read services directory: %w", err)
	}

	var categories []string
	for _, entry := range entries {
		if entry.IsDir() {
			categories = append(categories, entry.Name())
		}
	}
	return categories, nil
}

func (u *ServiceUtils) getServicesInCategory(category string) ([]types.ServiceInfo, error) {
	categoryPath := fmt.Sprintf("services/%s", category)
	entries, err := config.EmbeddedServicesFS.ReadDir(categoryPath)
	if err != nil {
		return nil, err
	}

	var services []types.ServiceInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), constants.ServiceConfigExtension) {
			continue
		}

		serviceName := strings.TrimSuffix(entry.Name(), constants.ServiceConfigExtension)
		serviceInfo, err := u.parseServiceInfo(categoryPath, entry.Name(), serviceName, category)
		if err != nil {
			continue
		}
		services = append(services, serviceInfo)
	}

	return services, nil
}

func (u *ServiceUtils) parseServiceInfo(categoryPath, fileName, serviceName, category string) (types.ServiceInfo, error) {
	serviceFile := fmt.Sprintf("%s/%s", categoryPath, fileName)
	data, err := config.EmbeddedServicesFS.ReadFile(serviceFile)
	if err != nil {
		return types.ServiceInfo{}, err
	}

	var serviceData map[string]interface{}
	if err := yaml.Unmarshal(data, &serviceData); err != nil {
		return types.ServiceInfo{}, err
	}

	return types.ServiceInfo{
		Name:         serviceName,
		Category:     category,
		Description:  getString(serviceData, "description"),
		UsageNotes:   getString(serviceData, "usage_notes"),
		Dependencies: getDependencies(serviceData),
		Options:      getStringSlice(serviceData["options"]),
		Examples:     getStringSlice(serviceData["examples"]),
		Links:        getStringSlice(serviceData["links"]),
	}, nil
}

func (u *ServiceUtils) loadServiceFromCategory(category, serviceName string) (*types.ServiceConfig, error) {
	servicePath := fmt.Sprintf("services/%s/%s%s", category, serviceName, constants.ServiceConfigExtension)
	data, err := config.EmbeddedServicesFS.ReadFile(servicePath)
	if err != nil {
		return nil, err
	}

	var serviceConfig types.ServiceConfig
	if err := yaml.Unmarshal(data, &serviceConfig); err != nil {
		return nil, fmt.Errorf("failed to parse service config for %s: %w", serviceName, err)
	}

	return &serviceConfig, nil
}

func (u *ServiceUtils) getDependenciesInCategory(category string) (map[string][]string, error) {
	categoryPath := fmt.Sprintf("services/%s", category)
	entries, err := config.EmbeddedServicesFS.ReadDir(categoryPath)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), constants.ServiceConfigExtension) {
			continue
		}

		serviceName := strings.TrimSuffix(entry.Name(), constants.ServiceConfigExtension)
		deps, err := u.parseServiceDependencies(categoryPath, entry.Name())
		if err != nil {
			continue
		}
		result[serviceName] = deps
	}

	return result, nil
}

func (u *ServiceUtils) parseServiceDependencies(categoryPath, fileName string) ([]string, error) {
	serviceFile := fmt.Sprintf("%s/%s", categoryPath, fileName)
	data, err := config.EmbeddedServicesFS.ReadFile(serviceFile)
	if err != nil {
		return nil, err
	}

	var serviceData map[string]interface{}
	if err := yaml.Unmarshal(data, &serviceData); err != nil {
		return nil, err
	}

	return getDependencies(serviceData), nil
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

func getDependencies(serviceData map[string]interface{}) []string {
	deps, exists := serviceData["dependencies"]
	if !exists {
		return nil
	}

	depsMap, ok := deps.(map[string]interface{})
	if !ok {
		return nil
	}

	required, exists := depsMap["required"]
	if !exists {
		return nil
	}

	return getStringSlice(required)
}
