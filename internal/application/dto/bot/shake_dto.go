package bot

type ShakeReq struct {
	ProviderBotName string `json:"provider"`
	UID             int64  `json:"-"`
}

// ShakeRes is the result of an AI connection handshake.
//
// It always travels in a Success envelope (mstatus=200): a handshake is
// best-effort by definition, so a failure to reach the LLM is reported via
// Shaked=false + Reason rather than an error envelope. That way the screen
// that triggered the ping (typically the home screen) never shows the user
// a scary error just because the handshake could not run.
type ShakeRes struct {
	// Shaked is true when the LLM connection is primed — either a probe
	// just succeeded, or a recent one is still fresh (see Cached).
	Shaked bool `json:"shaked"`

	// Cached is true when this response was served from a still-fresh
	// previous handshake without contacting the LLM again.
	Cached bool `json:"cached"`

	// Provider is the bot provider that serves AI requests, e.g.
	// "langchain". Empty when the bot adapter is disabled.
	Provider string `json:"provider,omitempty"`

	// Model is the resolved model id reported by the provider. May be empty.
	Model string `json:"model,omitempty"`

	// LatencyMs is the round-trip time of the probe in milliseconds. 0 for a
	// cached or disabled result.
	LatencyMs int64 `json:"latency_ms"`

	// Reason carries a machine-readable explanation when Shaked=false
	// ("bot_disabled", "shake_failed"). Empty on success.
	Reason string `json:"reason,omitempty"`

	Content string `json:"content"`
}
