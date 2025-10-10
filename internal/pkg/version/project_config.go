package version

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// ProjectConfigManager manages project-specific version configurations
type ProjectConfigManager struct {
	configDir string
}

// NewProjectConfigManager creates a new project configuration manager
func NewProjectConfigManager(configDir string) *ProjectConfigManager {
	return &ProjectConfigManager{
		configDir: configDir,
	}
}

// GetProjectConfig gets the version configuration for a project
func (p *ProjectConfigManager) GetProjectConfig(projectPath string) (*ProjectVersionConfig, error) {
	configsPath := filepath.Join(p.configDir, "project_configs.json")

	data, err := os.ReadFile(configsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, NewVersionError(ErrProjectConfig, "no project configuration found", nil)
		}
		return nil, err
	}

	var configs []ProjectVersionConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		absPath = projectPath
	}

	for _, config := range configs {
		if config.ProjectPath == absPath {
			return &config, nil
		}
	}

	return nil, NewVersionError(ErrProjectConfig, "project configuration not found", nil)
}

// SetProjectConfig sets the version configuration for a project
func (p *ProjectConfigManager) SetProjectConfig(config ProjectVersionConfig) error {
	configsPath := filepath.Join(p.configDir, "project_configs.json")

	absPath, err := filepath.Abs(config.ProjectPath)
	if err != nil {
		absPath = config.ProjectPath
	}
	config.ProjectPath = absPath
	config.LastUsed = time.Now()

	var configs []ProjectVersionConfig
	data, err := os.ReadFile(configsPath)
	if err == nil {
		_ = json.Unmarshal(data, &configs)
	}

	found := false
	for i := range configs {
		if configs[i].ProjectPath == config.ProjectPath {
			configs[i] = config
			found = true
			break
		}
	}

	if !found {
		configs = append(configs, config)
	}

	if err := os.MkdirAll(p.configDir, 0755); err != nil {
		return err
	}

	data, err = json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configsPath, data, 0644)
}

// ListProjectConfigs lists all project configurations
func (p *ProjectConfigManager) ListProjectConfigs() ([]ProjectVersionConfig, error) {
	configsPath := filepath.Join(p.configDir, "project_configs.json")

	data, err := os.ReadFile(configsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []ProjectVersionConfig{}, nil
		}
		return nil, err
	}

	var configs []ProjectVersionConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, err
	}

	return configs, nil
}
