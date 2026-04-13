package gamma

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

// ToolResult provides the result and an optional error, mapping what go functions usually return into a single
// struct so it can be easily JSON serializable.
type ToolResult struct {
	Result any    `json:"result"`
	Error  string `json:"error,omitempty"`
}

// ErrorToToolError takes an error and turns it into a `ToolResult` with that error, usable by the LLM.
func ErrorToToolError(e error) ToolResult {
	return ToolResult{
		Error: e.Error(),
	}
}

// StringToToolError takes an error string and generates a `ToolResult` with that error, usable by the LLM.
func StringToToolError(s string) ToolResult {
	return ToolResult{
		Error: s,
	}
}

// GetOptionalArg returns a pointer to the value for the tool call argument if it was provided, or a default value.
// It errors out if the type coercion fails.
func GetOptionalArg[T any](args map[string]any, key string, defaultValue T) (*T, bool, error) {
	val, ok := args[key]
	if !ok {
		return &defaultValue, false, nil
	}

	typed, ok := val.(T)
	if !ok {
		return nil, false, fmt.Errorf("arg %q has wrong type: got %T", key, val)
	}

	return &typed, true, nil
}

// GetRequiredArg returns a required argument or errors if the type coercion fails.
func GetRequiredArg[T any](args map[string]any, key string) (*T, error) {
	val, ok := args[key]
	if !ok {
		return nil, fmt.Errorf("tool args did not have key %q", key)
	}

	typed, ok := val.(T)
	if !ok {
		return nil, fmt.Errorf("arg %q has wrong type: got %T", key, val)
	}

	return &typed, nil
}

// ToolFunc is a function that takes a map of args single parameter and returns a `ToolResult`, just any and an
// optional error.
type ToolFunc func(args map[string]any) ToolResult

// ToolFuncDef defines a tool function with a name, optional description and optional parameters.
type ToolFuncDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  ToolParameters `json:"parameters"`
}

// ToolDef is the structure that ollama and most tool-using LLMs expect to see.
type ToolDef struct {
	Type     string      `json:"type"`
	Function ToolFuncDef `json:"function"`
}

// ToolCallFunction is what the LLM will provide you with to tell you to call a function, which also specifies the
// argument values as keyword arguments for a `map[string]any`.
type ToolCallFunction struct {
	Index     int            `json:"index,omitempty"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// ToolCall is the root structure the LLM will provide to tell you to call a specific tool.
type ToolCall struct {
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

// ToolCallResult will be provided back to the LLM to show it what the tool returned. It maps a `ToolCall` provided
// by the LLM to a `ToolResult`, provided by the binary.
type ToolCallResult struct {
	Call   ToolCall
	Result ToolResult
}

// ToMessageContent will marshal the tool result into JSON and return the string for the LLM.
func (tcr *ToolCallResult) ToMessageContent() string {
	data, err := json.Marshal(tcr.Result)
	if err != nil {
		slog.Error("could not marshal tool call result", "tool_result", tcr.Result)
		panic(err)
	}
	slog.Debug("returning ToMessageContent from Result", "tool_result", tcr.Result)
	return string(data)
}

// ToMessage turns a tool call result into a `Message` so that it can be added to the messages array, the ChatML
// format most LLM expect.
func (tcr *ToolCallResult) ToMessage() Message {
	return Message{
		Role:     "tool",
		ToolName: tcr.Call.Function.Name,
		Content:  tcr.ToMessageContent(),
	}
}

// CallableToolDef specifies an interface that any struct can implement to be used with the `OllamaClient`'s
// `WithTool` option.
type CallableToolDef interface {
	GetName() string
	ToToolDef() ToolDef
	GetCallable() ToolFunc
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// NamedProperty is an explicitly named argument to a tool function property. You must provide the LLM with details
// about how to use your tools, and each function argument.
type NamedProperty struct {
	Name        string
	Type        string
	Description string
	Required    bool
}

type ToolParameters struct {
	Type                 string              `json:"type"`
	Properties           map[string]Property `json:"properties"`
	Required             []string            `json:"required"`
	AdditionalProperties bool                `json:"additionalProperties"`
}

// NewToolParameters provides the specification to the LLM for the tool's function's parameters.
func NewToolParameters(namedProps ...NamedProperty) ToolParameters {
	props := make(map[string]Property)
	required := make([]string, 0)
	for _, p := range namedProps {
		if p.Required {
			required = append(required, p.Name)
		}
		t := ""
		switch p.Type {
		case "string", "boolean", "number", "array", "object":
			t = p.Type
		case "integer", "float", "float32", "float64":
			t = "number"
		case "bool":
			t = "boolean"
		default:
			// Panic might be a bit extreme, but I'd rather panic early if you aren't explicitly setting up the right
			// API for the LLM.
			panic(fmt.Sprintf("cant handle type %v %q, it must be a valid JSON type like 'number' or 'string'", p, p.Type))
		}
		props[p.Name] = Property{
			Type:        t,
			Description: p.Description,
		}
	}
	return ToolParameters{
		Type:                 "object",
		Properties:           props,
		Required:             required,
		AdditionalProperties: false,
	}
}

// CallableFunctionToolDef implements the base `CallableToolDef` interface so it can be provided as a tool with the
// `WithTool` option.
type CallableFunctionToolDef struct {
	Name        string
	Description string
	Callable    ToolFunc
	Parameters  ToolParameters
}

// ToToolDef provides the tool definition when the client tracks a new tool.
func (td CallableFunctionToolDef) ToToolDef() ToolDef {
	return ToolDef{
		Type: "function",
		Function: ToolFuncDef{
			Name:        td.Name,
			Description: td.Description,
			Parameters:  td.Parameters,
		},
	}
}

// GetName just returns the string name of the function.
func (td CallableFunctionToolDef) GetName() string {
	return td.Name
}

// GetCallable just returns the callable raw function to be used as a tool.
func (td CallableFunctionToolDef) GetCallable() ToolFunc {
	return td.Callable
}
