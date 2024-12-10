package v1

import "context"

// PluginInfo provides general information about the plugin.
type PluginInfo struct {
	// Name is the unique name of the plugin.
	Name string
	// Version is the version of the plugin.
	Version string
	// Description provides a brief explanation of the plugin's purpose.
	Description string
}

// Plugin defines the interface for registering
// and retrieving plugin information.
type Plugin interface {
	// GetInfo returns the metadata of the plugin,
	// such as name, version, and description.
	GetInfo() (PluginInfo, error)
}

// Component represents the configuration
// or resource required to execute an action or check.
type Component interface{}

// Action represents a specific operation
// to be executed by the plugin.
type Action interface{}

// Check represents a validation or
// state-checking operation.
type Check interface{}

// ActionService defines methods for handling actions within the plugin.
type ActionService interface {

	/*
		GetAction parses the input data and returns a structured Action.
		Input:
		  - data: map[string]interface{} containing action configuration.
		Output:
		  - Action: structured representation of the action.
		  - error: if the input data is invalid or incomplete.
	*/
	GetAction(data map[string]interface{}) (Action, error)

	/*
		ValidateYAMLAction validates the provided Action.
		Input:
		  - ctx: context for managing request lifecycle.
		  - action: the Action to validate.
		Output:
		  - error: if the Action fails validation.
	*/
	ValidateYAMLAction(ctx context.Context, action Action) error

	/*
		ExecAction executes the specified Action.
		Input:
		  - ctx: context for managing request lifecycle.
		  - component: the Component required to execute the Action.
		  - action: the Action to execute.
		Output:
		  - error: if the execution fails.
	*/
	ExecAction(ctx context.Context, component Component, action Action) error

	/*
		GetActionDesc provides a textual description of the Action.
		This can be used for logging or debugging purposes.
		Input:
		  - ctx: context for managing request lifecycle.
		  - action: the Action to describe.
		Output:
		  - string: a human-readable description of the Action.
	*/
	GetActionDesc(ctx context.Context, action Action) string

	/*
		GetCheck parses the input data and returns a structured Check.
		Input:
		  - data: map[string]interface{} containing check configuration.
		Output:
		  - Check: structured representation of the check.
		  - error: if the input data is invalid or incomplete.
	*/
	GetCheck(data map[string]interface{}) (Check, error)

	/*
		ValidateYAMLCheck validates the provided Check.
		Input:
		  - ctx: context for managing request lifecycle.
		  - check: the Check to validate.
		Output:
		  - error: if the Check fails validation.
	*/
	ValidateYAMLCheck(ctx context.Context, check Check) error

	/*
		ExecCheck performs the specified Check and returns its result.
		Input:
		  - ctx: context for managing request lifecycle.
		  - component: the Component to evaluate against.
		  - check: the Check to perform.
		Output:
		  - bool: true if the Check passes, false otherwise.
		  - error: if the execution fails.
	*/
	ExecCheck(ctx context.Context, component Component, check Check) (bool, error)

	/*
		GetCheckDesc provides a textual description of the Check.
		This can be used for logging or debugging purposes.
		Input:
		  - ctx: context for managing request lifecycle.
		  - action: the Check to describe.
		Output:
		  - string: a human-readable description of the Check.
	*/
	GetCheckDesc(ctx context.Context, action Action) string
}

// ComponentService defines methods for handling components within the plugin.
type ComponentService interface {

	/*
		GetComponent parses the input data and returns a structured Component.
		Input:
		  - data: map[string]interface{} containing component configuration.
		Output:
		  - Component: structured representation of the component.
		  - error: if the input data is invalid or incomplete.
	*/
	GetComponent(data map[string]interface{}) (Component, error)

	/*
		ValidateYAMLComponent validates the provided Component.
		Input:
		  - component: the Component to validate.
		Output:
		  - error: if the Component fails validation.
	*/
	ValidateYAMLComponent(component Component) error

	/*
		GetComponentDesc provides a textual description of the Component.
		This can be used for logging or debugging purposes.
		Input:
		  - ctx: context for managing request lifecycle.
		  - component: the Component to describe.
		Output:
		  - string: a human-readable description of the Component.
	*/
	GetComponentDesc(ctx context.Context, component Component) string
}

// Executor is a composite interface that combines
// Plugin, ComponentService, and ActionService.
type Executor interface {
	ComponentService
	ActionService
	Plugin
}
