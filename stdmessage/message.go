package stdmessage

import (
	"github.com/google/uuid"
)

// MetaNil is a type that represents an empty map[string]any, will be sent as {}
type MetaNil map[string]any

// DataNil is a type that represents an empty map[string]any, will be sent as {}
type DataNil map[string]any

// @spec-link [[api_standard_envelope]]
type StandardMessage[T any, M any] struct {
	RequestID string `json:"request_id"`
	Message   string `json:"message"`
	Success   bool   `json:"success"`
	Data      T      `json:"data"`
	Meta      M      `json:"meta"`
}

// NewWithMeta: instantiate a new standard message with meta
func NewWithMeta[T any, M any](message string, success bool, data T, meta M) *StandardMessage[T, M] {
	uid, _ := uuid.NewV7()
	return &StandardMessage[T, M]{
		RequestID: uid.String(),
		Message:   message,
		Success:   success,
		Data:      data,
		Meta:      meta,
	}
}

// New: instantiate a new standard message
func New[T any](message string, success bool, data T) *StandardMessage[T, MetaNil] {
	uid, _ := uuid.NewV7()
	return &StandardMessage[T, MetaNil]{
		RequestID: uid.String(),
		Message:   message,
		Success:   success,
		Data:      data,
		Meta:      MetaNil{},
	}
}
