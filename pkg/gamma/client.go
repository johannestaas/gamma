package gamma

import (
	"net/http"
)

// makeURL just appends the path to the configured `RootURL`.
func (c *GammaClient) makeURL(path string) string {
	return c.RootURL + path
}

// RegisterTool can be called manually, however it is already by default called if you provide the `WithTool` option.
// It just takes the tool name and the `ToolFunc` tool function
func (c *GammaClient) RegisterTool(toolName string, tool ToolFunc) {
	c.ToolRegistry[toolName] = tool
}

func NewGammaClient(opts ...Option) *GammaClient {
	client := &GammaClient{
		RootURL: defaultRootURL,
		Model:   defaultModel,
		HTTPClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
		ToolRegistry: make(map[string]ToolFunc),
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}
