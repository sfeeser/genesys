// Path: genesis/cmd/genesis/ping.go
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
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 45*time.Second)
		defer cancel()

		fmt.Printf("Genesis Handshake | Audit ID: %d\n", auditUnix)
		fmt.Println("--------------------------------------------------")

		tiers := []cognition.Tier{cognition.TierFast, cognition.TierDeep, cognition.TierEmbed}
		clocks := []*time.Time{&fastClock, &deepClock, &embedClock}
		mutexes := []*sync.Mutex{&fastMu, &deepMu, &embedMu}

		var failed bool
		for i, tier := range tiers {
			if err := verifyTier(ctx, tier, mutexes[i], clocks[i]); err != nil {
				failed = true
				fmt.Printf("TIER: %-5s | STATUS: ❌ FAIL | %v\n", tier, err)
			} else {
				fmt.Printf("TIER: %-5s | STATUS: ✅ PONG\n", tier)
			}
		}

		fmt.Println("--------------------------------------------------")
		if failed {
			return fmt.Errorf("one or more cognitive tiers failed verification")
		}
		return nil
	},
}

func verifyTier(ctx context.Context, tier cognition.Tier, mu *sync.Mutex, clock *time.Time) error {
	var model, key string

	// Determinant Ingestion
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
	default:
		return fmt.Errorf("unsupported tier: %s", tier)
	}

	// Fallback Resolution
	if key == "" {
		key = os.Getenv("GENESIS_API_KEY")
	}

	// Loud Delay Ingestion (Requirement B)
	delaySec := 0
	if raw := os.Getenv("GENESIS_API_DELAY"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("invalid GENESIS_API_DELAY for tier %s: %w", tier, err)
		}
		delaySec = v
	}

	// Specific Error Reporting (Requirement C)
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
