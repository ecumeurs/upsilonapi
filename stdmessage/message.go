// Package stdmessage defines the universal communication envelope for all Upsilon Hub services.
// It ensures that every API response follows a consistent structural contract.
// @spec-link [[api_standard_envelope]]
package stdmessage

import (
	"github.com/google/uuid"
)

// MetaNil is a type that represents an empty map[string]any, will be sent as {}.
type MetaNil map[string]any

// DataNil is a type that represents an empty map[string]any, will be sent as {}.
type DataNil map[string]any

// StandardMessage is the generic polymorphic container for all Upsilon API responses.
// It enforces the 'RequestID', 'Success', and 'Message' fields as per the hub's architectural mandate.
type StandardMessage[T any, M any] struct {
	RequestID string `json:"request_id"`
	Message   string `json:"message"`
	Success   bool   `json:"success"`
	Data      T      `json:"data"`
	Meta      M      `json:"meta"`
}

// NewWithMeta instantiates a new standard message with custom metadata.
// It generates a unique V7 UUID for every message to facilitate trace identification.
func NewWithMeta[T any, M any](message string, success bool, data T, meta M) *StandardMessage[T, M] {
	// 1. Traceability: Generate a time-ordered UUID for the request ID.
	uid, _ := uuid.NewV7()
	
	// 2. Assembly: Map all input fields into the standard envelope structure.
	return &StandardMessage[T, M]{
		RequestID: uid.String(),
		Message:   message,
		Success:   success,
		Data:      data,
		Meta:      meta,
	}
}

// New instantiates a new standard message with default empty metadata.
// This is the standard constructor for most success and error responses.
func New[T any](message string, success bool, data T) *StandardMessage[T, MetaNil] {
	// 1. Traceability: Generate a unique identifier for log correlation.
	uid, _ := uuid.NewV7()
	
	// 2. Assembly: Initialize the message with a persistent MetaNil instance.
	return &StandardMessage[T, MetaNil]{
		RequestID: uid.String(),
		Message:   message,
		Success:   success,
		Data:      data,
		Meta:      MetaNil{},
	}
}
