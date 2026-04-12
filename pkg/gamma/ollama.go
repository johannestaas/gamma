package gamma

import "net/http"

// Message is the base message like in ChatML format.
// It's expected to be streaming.
type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	ToolName  string     `json:"tool_name,omitempty"`
	Thinking  string     `json:"thinking,omitempty"`
}

// ChatRequest is sent out as a POST request to Ollama and provides the specified model, the messages, and the tools
// available in the client.
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []ToolDef `json:"tools,omitempty"`
	Stream   bool      `json:"stream"`
}

// ChatResponseChunk is what ollama will provide back when streaming.
type ChatResponseChunk struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

// StreamChunk is what the Gamma framework provides through the channels opened by chatting.
type StreamChunk struct {
	// Raw model content as text.
	Content string
	// Raw model thoughts as text.
	Thinking string
	// Whether the model is done spitting out tokens.
	Done bool
	// Whether the model used a tool with this message.
	UsedTool bool
	// Whether there was any error, such as HTTP response errors or json unmarshaling errors.
	Err error
}

// GammaClient is the root structure you interface with to start chatting with a model.
type GammaClient struct {
	// defaults to http://localhost:11434 , override with `WiithRootURL`
	RootURL string
	// defaults to gemma4, override with `WithModel`
	Model string
	// You _can_ override the HTTP client if you want to do something funny or specific testing. It just makes POST
	// requests.
	HTTPClient *http.Client
	// You likely don't want to mess with these by hand. Use the `WithTool` option.
	ToolDefs []ToolDef
	// This registers all tools by name and points at a golang callable, however I recommend using the `WithTool`
	// option rather than modifying this manually.
	ToolRegistry map[string]ToolFunc
}
