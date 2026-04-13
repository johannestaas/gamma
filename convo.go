package gamma

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// handleToolCalls takes a message from the LLM, and for each tool call, runs the function and tracks the response.
func (c *GammaClient) handleToolCalls(msg Message) (results []ToolCallResult) {
	for _, tc := range msg.ToolCalls {
		handler, ok := c.ToolRegistry[tc.Function.Name]
		var result ToolCallResult
		if !ok {
			result = ToolCallResult{
				Call: tc,
				Result: ToolResult{
					Result: nil,
					Error:  fmt.Sprintf("missing tool: %q", tc.Function.Name),
				},
			}
			slog.Warn("failed to find function", "function_name", tc.Function.Name)
		} else {
			result = ToolCallResult{
				Call:   tc,
				Result: handler(tc.Function.Arguments),
			}
			slog.Debug("ran function and got result",
				"function_name", tc.Function.Name,
				"function_result", result,
			)
		}
		results = append(results, result)
	}
	return
}

type StreamingConversation interface {
	Ask(context.Context, string) <-chan StreamChunk
}

// Convo is the internal structure that tracks a conversation with an LLM through the Ollama streaming chat API.
type Convo struct {
	// If you want to modify anything, it's likely `Messages` to change what an assistant or user said, or mess with
	// the system prompt after the fact.
	Messages []Message
	client   *GammaClient
	url      string
}

// NewConvo takes an optional system prompt and returns a `Convo` that you can use to `Ask` questions. This is invoked from the GammaClient itself.
// This makes no HTTP calls yet.
func (c *GammaClient) NewConvo(systemPrompt string) *Convo {
	var msgs []Message
	if systemPrompt != "" {
		msgs = append(msgs, Message{
			Role:    "system",
			Content: systemPrompt,
		})
	}
	return &Convo{
		Messages: msgs,
		client:   c,
		url:      c.makeURL("/api/chat"),
	}
}

// AddMessage just appends a message to the end of the current array.
func (convo *Convo) AddMessage(message Message) {
	convo.Messages = append(convo.Messages, message)
}

// Ask is the main `Convo` interface that begins the POST request and streaming of tokens.
// It returns a channel where it will send each `StreamChunk` as they come in so you can stream tokens visually.
func (convo *Convo) Ask(ctx context.Context, prompt string) <-chan StreamChunk {
	out := make(chan StreamChunk)
	convo.Messages = append(convo.Messages, Message{
		Role:    "user",
		Content: prompt,
	})

	go func() {
		defer close(out)

		reqBody := ChatRequest{
			Model:    convo.client.Model,
			Messages: convo.Messages,
			Tools:    convo.client.ToolDefs,
			Stream:   true,
		}

		data, err := json.Marshal(reqBody)
		slog.Debug(fmt.Sprintf("Conversation request body: %+v\n", reqBody))
		if err != nil {
			out <- StreamChunk{Err: err}
			return
		}

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			convo.url,
			bytes.NewReader(data),
		)
		if err != nil {
			out <- StreamChunk{Err: err}
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := convo.client.HTTPClient.Do(req)
		if err != nil {
			out <- StreamChunk{Err: err}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			out <- StreamChunk{Err: fmt.Errorf("unexpected status: %s", resp.Status)}
			return
		}

		scanner := bufio.NewScanner(resp.Body)

		// Might need to raise max token line size for larger JSON chunks
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			var chunk ChatResponseChunk
			if err := json.Unmarshal(scanner.Bytes(), &chunk); err != nil {
				out <- StreamChunk{Err: fmt.Errorf("decode chunk: %w", err)}
				return
			}

			results := convo.client.handleToolCalls(chunk.Message)

			usedTool := len(results) > 0

			for _, r := range results {
				newMsg := r.ToMessage()
				convo.AddMessage(newMsg)
			}

			out <- StreamChunk{
				Content:  chunk.Message.Content,
				Thinking: chunk.Message.Thinking,
				Done:     chunk.Done,
				UsedTool: usedTool,
			}

			if chunk.Done {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			out <- StreamChunk{Err: err}
			return
		}
	}()

	return out
}
