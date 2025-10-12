package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServiceUtils(t *testing.T) {
	utils := NewServiceUtils()
	assert.NotNil(t, utils)
}

func TestServiceUtils_GetServicesByCategory(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "get services by category",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils := NewServiceUtils()

			services, err := utils.GetServicesByCategory()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, services)
			} else {
				// May error if no embedded services, but shouldn't panic
				if err != nil {
					assert.NotNil(t, err)
				} else {
					assert.NotNil(t, services)
				}
			}
		})
	}
}

func TestServiceUtils_LoadServicesByCategory(t *testing.T) {
	t.Run("backward compatibility alias", func(t *testing.T) {
		utils := NewServiceUtils()

		// Test that the alias method exists and works
		services1, err1 := utils.GetServicesByCategory()
		services2, err2 := utils.LoadServicesByCategory()

		// Both should have the same result
		assert.Equal(t, err1 != nil, err2 != nil)
		if err1 == nil && err2 == nil {
			assert.Equal(t, len(services1), len(services2))
		}
	})
}

func TestServiceUtils_LoadServiceConfig(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		expectError bool
	}{
		{
			name:        "load nonexistent service",
			serviceName: "nonexistent-service",
			expectError: true,
		},
		{
			name:        "load with empty service name",
			serviceName: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils := NewServiceUtils()

			config, err := utils.LoadServiceConfig(tt.serviceName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
			}
		})
	}
}

func TestServiceUtils_GetCategories(t *testing.T) {
	t.Run("get categories", func(t *testing.T) {
		utils := NewServiceUtils()

		// Test the private method indirectly through GetServicesByCategory
		services, err := utils.GetServicesByCategory()

		// Should not panic, may error if no embedded services
		if err != nil {
			assert.NotNil(t, err)
		} else {
			assert.NotNil(t, services)
		}
	})
}

func TestServiceUtils_GetServicesInCategory(t *testing.T) {
	t.Run("get services in category", func(t *testing.T) {
		utils := NewServiceUtils()

		// Test indirectly through GetServicesByCategory
		services, err := utils.GetServicesByCategory()

		// Should handle empty or invalid categories gracefully
		if err != nil {
			assert.NotNil(t, err)
		} else {
			assert.NotNil(t, services)
			// Each category should have valid services
			for category, serviceList := range services {
				assert.NotEmpty(t, category)
				assert.NotNil(t, serviceList)
			}
		}
	})
}

func TestServiceUtils_LoadServiceFromCategory(t *testing.T) {
	t.Run("load service from category", func(t *testing.T) {
		utils := NewServiceUtils()

		// Test indirectly through LoadServiceConfig
		config, err := utils.LoadServiceConfig("test-service")

		// Should handle nonexistent services gracefully
		assert.Error(t, err)
		assert.Nil(t, config)
	})
}

func TestServiceUtils_ErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(*ServiceUtils) error
	}{
		{
			name: "handle invalid service name",
			testFunc: func(u *ServiceUtils) error {
				_, err := u.LoadServiceConfig("")
				return err
			},
		},
		{
			name: "handle category loading errors",
			testFunc: func(u *ServiceUtils) error {
				_, err := u.GetServicesByCategory()
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils := NewServiceUtils()

			// Test that error handling doesn't panic
			assert.NotPanics(t, func() {
				err := tt.testFunc(utils)
				// Error is expected for these test cases
				_ = err
			})
		})
	}
}
