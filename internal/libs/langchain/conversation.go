package langchain

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
)

// This file is the framework-first memory surface. It wraps langchaingo's
// `memory` package (ConversationWindowBuffer + a pluggable
// schema.ChatMessageHistory store) so the rest of the app gets windowed,
// persisted conversation memory WITHOUT importing langchaingo's memory
// internals directly.
//
// Note on `chains`: langchaingo's chains/prompts packages would also provide
// a turnkey ConversationChain, but they (a) drag in heavy transitive deps
// (sprig, gonja/Jinja2, starlark) the project deliberately avoids, (b) have
// no JSON mode (incompatible with quiz generation), and (c) bypass this
// client's MathError/retry/JSON wrapper. So we use the framework's MEMORY
// classes for history/windowing/persistence and keep the actual model call
// on this client (Generate) — which is where the value of "context memory"
// actually lives.

// LLM exposes the underlying langchaingo model for callers that need it.
func (c *Client) LLM() llms.Model { return c.llm }

// NewWindowMemory builds a langchaingo window-buffer memory backed by the
// supplied history store. windowSize bounds how many recent turns are
// surfaced for a prompt; the store retains full history (the buffer trims
// the read view, and the MySQL history's SetMessages is a no-op so nothing
// is pruned from the database).
func NewWindowMemory(history schema.ChatMessageHistory, windowSize int) schema.Memory {
	if windowSize <= 0 {
		windowSize = 5
	}
	return memory.NewConversationWindowBuffer(windowSize, memory.WithChatHistory(history))
}

// LoadHistoryString returns the windowed prior conversation rendered as a
// plain string ("Human: ...\nAI: ..."), ready to inject into a prompt.
// Empty when there is no prior context.
func LoadHistoryString(ctx context.Context, mem schema.Memory) (string, error) {
	vars, err := mem.LoadMemoryVariables(ctx, map[string]any{})
	if err != nil {
		return "", err
	}
	if v, ok := vars[mem.GetMemoryKey(ctx)].(string); ok {
		return v, nil
	}
	return "", nil
}

// SaveTurn persists one (human input, ai output) exchange through the
// memory's history store.
func SaveTurn(ctx context.Context, mem schema.Memory, input, output string) error {
	return mem.SaveContext(ctx,
		map[string]any{"input": input},
		map[string]any{"output": output},
	)
}
