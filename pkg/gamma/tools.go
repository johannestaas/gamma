package gamma

import (
	"encoding/json"
	"log/slog"
)

// ToolResult provides the result and an optional error, mapping what go functions usually return into a single
// struct so it can be easily JSON serializable.
type ToolResult struct {
	Result any
	Error  error
}

// ToolFunc is a function that takes a map of args single parameter and returns a `ToolResult`, just any and an
// optional error.
type ToolFunc func(args map[string]any) ToolResult

// ToolFuncDef defines a tool function with a name, optional description and optional parameters.
type ToolFuncDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
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

// CallableFunctionToolDef implements the base `CallableToolDef` interface so it can be provided as a tool with the
// `WithTool` option.
type CallableFunctionToolDef struct {
	Name        string
	Description string
	Callable    ToolFunc
	Parameters  map[string]any
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
