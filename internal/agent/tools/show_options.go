package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/utils"
)

var showOptionsTool = BaseTool{
	name: ToolShowOptions,
	description: `Show interactive option buttons to users for collecting their selection input.

## Use Cases
Use this tool when you need users to select from predefined options, for example:
- Gender selection (Male/Female)
- Yes/No questions
- Multiple choice questions (symptom types, pain levels, etc.)

## Parameters
- question: The question to ask
- options: List of options, each containing label (display text) and value
- multi_select: Whether to allow multiple selections (default false)

## Usage Example
{
    "question": "Please select your gender",
    "options": [
        {"label": "A. Male", "value": "male"},
        {"label": "B. Female", "value": "female"}
    ],
    "multi_select": false
}

## Notes
- After calling this tool, the frontend will display option buttons
- When user clicks, the selected value will be sent as user's reply
- You need to wait for the user's next message to get the selection result`,
	schema: utils.GenerateSchema[ShowOptionsInput](),
}

// ShowOptionsTool implements a tool for showing interactive options to users
type ShowOptionsTool struct {
	BaseTool
}

// ShowOptionsInput defines the input parameters for show_options tool
type ShowOptionsInput struct {
	Question    string           `json:"question" jsonschema:"The question to ask the user"`
	Options     []ShowOptionItem `json:"options" jsonschema:"List of selectable options"`
	MultiSelect bool             `json:"multi_select" jsonschema:"Whether to allow multiple selections"`
}

// ShowOptionItem represents a single option
type ShowOptionItem struct {
	Label string `json:"label" jsonschema:"Display text for the option"`
	Value string `json:"value" jsonschema:"Value to return when selected"`
}

// NewShowOptionsTool creates a new show_options tool instance
func NewShowOptionsTool() *ShowOptionsTool {
	return &ShowOptionsTool{
		BaseTool: showOptionsTool,
	}
}

// Execute executes the show_options tool
func (t *ShowOptionsTool) Execute(ctx context.Context, args json.RawMessage) (*types.ToolResult, error) {
	// Parse args from json.RawMessage
	var input ShowOptionsInput
	if err := json.Unmarshal(args, &input); err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse args: %v", err),
		}, err
	}

	// Validate input
	if input.Question == "" {
		return &types.ToolResult{
			Success: false,
			Error:   "question is required",
		}, nil
	}

	if len(input.Options) == 0 {
		return &types.ToolResult{
			Success: false,
			Error:   "at least one option is required",
		}, nil
	}

	// Format output for AI
	output := fmt.Sprintf("Options displayed to user:\nQuestion: %s\n", input.Question)
	output += "Options:\n"
	for _, opt := range input.Options {
		output += fmt.Sprintf("  - %s (value: %s)\n", opt.Label, opt.Value)
	}
	if input.MultiSelect {
		output += "\n(Multiple selection allowed)"
	}
	output += "\n\nPlease wait for user selection before continuing."

	// Prepare structured data for event emission
	optionsData := make([]map[string]string, len(input.Options))
	for i, opt := range input.Options {
		optionsData[i] = map[string]string{
			"label": opt.Label,
			"value": opt.Value,
		}
	}

	return &types.ToolResult{
		Success: true,
		Output:  output,
		Data: map[string]interface{}{
			"question":            input.Question,
			"options":             optionsData,
			"multi_select":        input.MultiSelect,
			"display_type":        "quick_reply",
			"emit_quick_reply":    true, // Flag for stream handler to emit quick_reply event
			"requires_user_input": true, // Flag to stop agent and wait for user input
		},
	}, nil
}
