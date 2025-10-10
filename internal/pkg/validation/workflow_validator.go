package validation

import (
	"fmt"

	"github.com/isaacgarza/dev-stack/internal/pkg/config"
)

// WorkflowValidator validates workflows and profiles
type WorkflowValidator struct {
	config *config.CommandConfig
}

// NewWorkflowValidator creates a new workflow validator
func NewWorkflowValidator(config *config.CommandConfig) *WorkflowValidator {
	return &WorkflowValidator{
		config: config,
	}
}

// ValidateWorkflows validates workflow definitions
func (v *WorkflowValidator) ValidateWorkflows(result *ValidationResult) {
	for workflowName, workflow := range v.config.Workflows {
		prefix := "workflows." + workflowName

		if workflow.Name == "" {
			AddError(result, "workflows", prefix+".name", "Workflow name is required", "MISSING_WORKFLOW_NAME", "medium", "Add name to workflow "+workflowName)
		}

		if workflow.Description == "" {
			AddWarning(result, "workflows", prefix+".description", "Workflow description is recommended", "MISSING_WORKFLOW_DESCRIPTION", "Add description to workflow "+workflowName)
		}

		if len(workflow.Steps) == 0 {
			AddError(result, "workflows", prefix+".steps", "Workflow has no steps", "EMPTY_WORKFLOW", "medium", "Add steps to workflow "+workflowName)
		}

		for i, step := range workflow.Steps {
			stepPrefix := fmt.Sprintf("%s.steps[%d]", prefix, i)

			if step.Command == "" {
				AddError(result, "workflows", stepPrefix+".command", "Workflow step command is required", "MISSING_STEP_COMMAND", "medium", "Add command to workflow step")
			}

			if step.Description == "" {
				AddWarning(result, "workflows", stepPrefix+".description", "Workflow step description is recommended", "MISSING_STEP_DESCRIPTION", "Add description to workflow step")
			}
		}
	}
}

// ValidateProfiles validates profile definitions
func (v *WorkflowValidator) ValidateProfiles(result *ValidationResult) {
	for profileName, profile := range v.config.Profiles {
		prefix := "profiles." + profileName

		if profile.Name == "" {
			AddError(result, "profiles", prefix+".name", "Profile name is required", "MISSING_PROFILE_NAME", "medium", "Add name to profile "+profileName)
		}

		if profile.Description == "" {
			AddWarning(result, "profiles", prefix+".description", "Profile description is recommended", "MISSING_PROFILE_DESCRIPTION", "Add description to profile "+profileName)
		}

		if len(profile.Services) == 0 {
			AddError(result, "profiles", prefix+".services", "Profile has no services", "EMPTY_PROFILE", "medium", "Add services to profile "+profileName)
		}
	}
}
