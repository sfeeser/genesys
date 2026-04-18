// Package telemetry provides real-time observability for the Genesis Engine.
// It emits the Kinetic Signals used by the Canvas (UI) to track Metamorphosis.
// Deterministic Status: SEQUENCED (ASYNCHRONOUS-FANOUT)
package telemetry

import (
	"encoding/json"
	"sync"

	"genesis/internal/identity" // Allowed L1 Import
)

// SignalType categorizes the nature of the telemetry event.
type SignalType string

const (
	SignalTransition SignalType = "transition"
	SignalGate       SignalType = "gate"
	SignalPulse      SignalType = "pulse"
	SignalError      SignalType = "error"
)

// Signal represents a single observability event.
type Signal struct {
	Type      SignalType `json:"type"`
	NodeID    *string    `json:"node_id,omitempty"` // Optional: Enforces Grammar Fidelity
	Payload   string     `json:"payload"`          // Constrained: Enforces Determinism
	Timestamp int64      `json:"timestamp"`
}

// observer represent an isolated, buffered signaling channel.
type observer struct {
	ch   chan []byte
	done chan struct{}
}

// Hub coordinates the emission of signals to multiple observers.
type Hub struct {
	mu        sync.RWMutex
	observers []*observer
	buffer    int // Size of the non-blocking buffer
}

// New initializes a new telemetry hub with asynchronous isolation.
func New(bufferSize int) *Hub {
	return &Hub{
		observers: make([]*observer, 0),
		buffer:    bufferSize,
	}
}

// Register adds an observer and starts its dedicated worker.
// Enforces Chapter 8: Async Isolation.
func (h *Hub) Register(w interface{ Write([]byte) (int, error) }) {
	h.mu.Lock()
	defer h.mu.Unlock()

	obs := &observer{
		ch:   make(chan []byte, h.buffer),
		done: make(chan struct{}),
	}

	// Dedicated worker per observer to prevent cross-stalling
	go func() {
		for data := range obs.ch {
			_, _ = w.Write(data)
		}
		close(obs.done)
	}()

	h.observers = append(h.observers, obs)
}

// Emit broadcasts a signal. Enforces Chapter 5 (Determinism) and 12 (Identity).
func (h *Hub) Emit(st SignalType, id *identity.NodeID, payload string, unix int64) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var idStr *string
	if id != nil {
		s := id.String()
		idStr = &s
	}

	sig := Signal{
		Type:      st,
		NodeID:    idStr,
		Payload:   payload,
		Timestamp: unix,
	}

	data, err := json.Marshal(sig)
	if err != nil {
		// Critical failure: instead of silent drop, we emit a recovery signal
		data = []byte(`{"type":"error","payload":"telemetry: marshal failure"}`)
	}
	data = append(data, '\n')

	for _, obs := range h.observers {
		select {
		case obs.ch <- data:
			// Success: Signal buffered
		default:
			// Non-blocking: If buffer is full, we drop to protect the engine
		}
	}
}

// Pulse emits a system-level heartbeat without identity contamination.
func (h *Hub) Pulse(message string, unix int64) {
	h.Emit(SignalPulse, nil, message, unix)
}
