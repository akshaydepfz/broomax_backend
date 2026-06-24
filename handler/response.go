package handler

// JSONEnvelope is the standard API response wrapper.
type JSONEnvelope struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}
