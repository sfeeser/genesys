// Package cognition manages external intelligence tiers for the Genesis Engine.
// It enforces the FAST/DEEP split and tier-local temporal invariants.
// Deterministic Status: SEQUENCED (SDK-TRUTHFUL)
package cognition

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"google.golang.org/genai" // Authoritative Modern SDK
)

// Tier identifies the reasoning hemisphere.
type Tier string

const (
	TierFast Tier = "FAST"
	TierDeep Tier = "DEEP"
)

// Config represents the raw determinants from the environment.
type Config struct {
	Tier        Tier
	APIKey      string
	Model       string
	Delay       time.Duration
	FallbackKey string
}

// Client handles communication with a specific Cognitive Tier.
type Client struct {
	cfg    Config
	apiKey string

	// Shared Tier Physics: Enforces Chapter 13.3
	mu       *sync.Mutex
	lastCall *time.Time
}

// NewClient initializes a client with mandatory shared components.
func NewClient(cfg Config, sharedMu *sync.Mutex, sharedClock *time.Time) (*Client, error) {
	if sharedMu == nil || sharedClock == nil {
		return nil, errors.New("cognition: shared tier limiter components cannot be nil")
	}

	c := &Client{
		cfg:      cfg,
		mu:       sharedMu,
		lastCall: sharedClock,
	}

	// Fallback Resolution: Tier-local > Global Fallback
	c.apiKey = cfg.APIKey
	if c.apiKey == "" {
		c.apiKey = cfg.FallbackKey
	}

	return c, nil
}

// waitEnforce implements Slot Reservation with Rollback for sequential tier spacing.
func (c *Client) waitEnforce(ctx context.Context) error {
	if c.cfg.Delay <= 0 {
		return nil
	}

	c.mu.Lock()
	now := time.Now()
	
	var reservedSlot time.Time
	if c.lastCall.IsZero() || now.Sub(*c.lastCall) >= c.cfg.Delay {
		reservedSlot = now
	} else {
		reservedSlot = c.lastCall.Add(c.cfg.Delay)
	}

	previousLastCall := *c.lastCall
	*c.lastCall = reservedSlot
	c.mu.Unlock()

	wait := time.Until(reservedSlot)
	if wait > 0 {
		timer := time.NewTimer(wait)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			c.mu.Lock()
			// Rollback bandwidth reservation if we are still the frontier
			if c.lastCall.Equal(reservedSlot) {
				*c.lastCall = previousLastCall
			}
			c.mu.Unlock()
			return ctx.Err()
		case <-timer.C:
		}
	}

	return nil
}

// Verify asserts connectivity by requesting a deterministic "PONG".
// Enforces Chapter 13.3: Strict Field-Aware Compliance.
func (c *Client) Verify(ctx context.Context) error {
	if c.apiKey == "" {
		return fmt.Errorf("cognition: %s tier has no API key", c.cfg.Tier)
	}

	if err := c.waitEnforce(ctx); err != nil {
		return err
	}

	// SDK ALIGNMENT: Using pkg.go.dev/google.golang.org/genai signatures
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  c.apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return fmt.Errorf("cognition: %s init failed: %w", c.cfg.Tier, err)
	}

	// SDK ALIGNMENT: genai.Text returns []*Content for simple generation
	resp, err := client.Models.GenerateContent(
		ctx,
		c.cfg.Model,
		genai.Text("System check. Respond only with PONG."),
		nil,
	)
	if err != nil {
		return fmt.Errorf("cognition: %s API failure: %w", c.cfg.Tier, err)
	}

	// RESPONSE HARDENING: Defensive extraction of first candidate
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return fmt.Errorf("cognition: %s returned empty response", c.cfg.Tier)
	}

	// FIELD-AWARE EXTRACTION: Accessing the semantic text payload directly
	part := resp.Candidates[0].Content.Parts[0]
	text := strings.TrimSpace(part.Text)

	if text == "" {
		return fmt.Errorf("cognition: %s returned empty text part", c.cfg.Tier)
	}

	if text != "PONG" {
		return fmt.Errorf("cognition: %s integrity failure; expected PONG, got %q", c.cfg.Tier, text)
	}

	return nil
}
