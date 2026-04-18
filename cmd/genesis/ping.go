// Package main handles the Cobra CLI surface for the Genesis Engine.
// Path: genesis/cmd/genesis/ping.go
// Deterministic Status: SEQUENCED (CERTIFIED)
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"genesis/internal/cognition"
	"github.com/spf13/cobra"
)

// Shared Tier State (Chapter 13.3)
// Mutexes and clocks are persisted at the package level to survive command execution
// if called repeatedly within the same process.
var (
	fastMu     sync.Mutex
	fastClock  time.Time
	deepMu     sync.Mutex
	deepClock  time.Time
	embedMu    sync.Mutex
	embedClock time.Time
)

func init() {
	rootCmd.AddCommand(pingCmd)
}

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Verifies connectivity for the Cognitive Triad (FAST, DEEP, EMBED)",
	Long:  `Executes a transactional PONG handshake with all configured AI hemispheres.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// CHAPTER 14.3: Derive context from shell-managed root context
		ctx, cancel := context.WithTimeout(cmd.Context(), 45*time.Second)
		defer cancel()

		fmt.Printf("Genesis Handshake | Audit ID: %d\n", auditUnix)
		fmt.Println("--------------------------------------------------")

		type tierCheck struct {
			tier  cognition.Tier
			mu    *sync.Mutex
			clock *time.Time
		}

		checks := []tierCheck{
			{cognition.TierFast, &fastMu, &fastClock},
			{cognition.TierDeep, &deepMu, &deepClock},
			{cognition.TierEmbed, &embedMu, &embedClock},
		}

		var failed bool
		for _, check := range checks {
			if err := verifyTier(ctx, check.tier, check.mu, check.clock); err != nil {
				failed = true
				fmt.Printf("TIER: %-5s | STATUS: ❌ FAIL | %v\n", check.tier, err)
			} else {
				fmt.Printf("TIER: %-5s | STATUS: ✅ PONG\n", check.tier)
			}
		}

		fmt.Println("--------------------------------------------------")
		if failed {
			return fmt.Errorf("one or more cognitive tiers failed verification")
		}
		return nil
	},
}

// verifyTier hydrates a client using the Precedence Law (Flags > Env > Defaults).
func verifyTier(ctx context.Context, tier cognition.Tier, mu *sync.Mutex, clock *time.Time) error {
	var model, key string

	// 1. Determinant Ingestion
	switch tier {
	case cognition.TierFast:
		model = os.Getenv("GENESIS_FAST_MODEL")
		key = os.Getenv("GENESIS_FAST_API_KEY")
	case cognition.TierDeep:
		model = os.Getenv("GENESIS_DEEP_MODEL")
		key = os.Getenv("GENESIS_DEEP_API_KEY")
	case cognition.TierEmbed:
		model = os.Getenv("GENESIS_EMBED_MODEL")
		key = os.Getenv("GENESIS_EMBED_API_KEY")
	}

	// 2. Fallback Resolution
	if key == "" {
		key = os.Getenv("GENESIS_API_KEY")
	}

	// 3. Loud Delay Ingestion (Requirement: Fail on malformed input)
	delaySec := 0
	if raw := os.Getenv("GENESIS_API_DELAY"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("invalid GENESIS_API_DELAY for tier %s: %w", tier, err)
		}
		delaySec = v
	}

	// 4. Specific Error Reporting (Requirement: Tier-aware error messages)
	if model == "" {
		return fmt.Errorf("missing model determinant for tier %s", tier)
	}
	if key == "" {
		return fmt.Errorf("missing API key determinant for tier %s", tier)
	}

	cfg := cognition.Config{
		Tier:   tier,
		APIKey: key,
		Model:  model,
		Delay:  time.Duration(delaySec) * time.Second,
	}

	client, err := cognition.NewClient(cfg, mu, clock)
	if err != nil {
		return err
	}

	return client.Verify(ctx)
}
