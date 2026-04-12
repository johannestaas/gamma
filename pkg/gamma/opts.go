package gamma

import (
	"net/http"
	"time"
)

const defaultRootURL = "http://localhost:11434"
const defaultModel = "gemma4"
const defaultHTTPTimeout = 60.0 * time.Second

// Option provides a number of utility functions that you can use to modify your Gamma client.
type Option func(*GammaClient)

// WithTimeout sets the HTTP client's timeout. If you update the HTTP client after, you
// lose this option.
func WithTimeout(d time.Duration) Option {
	return func(c *GammaClient) {
		c.HTTPClient.Timeout = d
	}
}

// WithHTTPClient will override the http client if you want to do any custom testing.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *GammaClient) {
		c.HTTPClient = httpClient
	}
}

// WithModel defaults to a specific LLM model.
func WithModel(model string) Option {
	return func(c *GammaClient) {
		c.Model = model
	}
}

// WithRootURL defaults to localhost, but you can use any remote service.
func WithRootURL(rootURL string) Option {
	return func(c *GammaClient) {
		c.RootURL = rootURL
	}
}

// WithTool takes a callable tool definition `CallableToolDef` and adds it to the list
// of the client's tools, and then registers the callable function/tool so that the LLM may
// invoke it.
func WithTool(ctd CallableToolDef) Option {
	return func(c *GammaClient) {
		c.ToolDefs = append(c.ToolDefs, ctd.ToToolDef())
		c.RegisterTool(ctd.GetName(), ctd.GetCallable())
	}
}
