### Node Biography / Data Structure (Final Consolidated Version)

Each node in the genome shall maintain the following rich, stateful biography:

**Core Identity**
- `uuid` — Unique identifier for the node
- `public_id` — Human-readable identifier (e.g. `scanner.ScanFile[scan.go]`)
- `uip` — Unique Import Path (e.g. `internal/scanner.ScanFile`)
- `file_path` — Full relative path to the file
- `package_name` — Name of the package

**Lifecycle & State**
- `genesis_state` — Current numeric state (1–9)
- `maturity` — Current maturity level: `"conceptual"`, `"hollow"`, `"anchored"`, `"synthesizing"`, `"implemented"`, `"partial_implementation"`, `"broken"`
- `synthesis_order` — Integer used for deterministic core-to-leaf processing
- `synthesis_priority` — Computed score for ordering (higher = more core = processed earlier)

**Intent Preservation (Critical)**
- `original_responsibility` — One-sentence responsibility from scaffolding (never overwritten)
- `original_business_purpose` — High-quality description from DEEP planning (never overwritten)
- `original_spec_snippet` — Excerpt from the SpecBook that this node was created to fulfill

**Current State**
- `current_responsibility` — May be updated over time
- `current_business_purpose` — May be updated over time
- `last_llm_reasoning` — The reasoning from the most recent Synthesis decision on this node

**History & Telemetry (Stateful Biography)**
- `growth_history` — Array of stage numbers in chronological order (e.g. `[1, 2, 3, 7, 8, 9, 8, 9]`)
- `white_blood_cell_attacks` — Array of objects recording every inversion/regression:
  ```json
  {
    "stage": 9,
    "reason": "Missing core abstraction 'PaymentValidator' required by leaf node 'ProcessOrder'",
    "timestamp": "2026-04-21T09:12:45Z",
    "cycle": 2
  }
  ```
- `synthesis_attempts` — Integer count of how many times Stage 9 has attempted to implement this node
- `synthesis_outcome_history` — Array of outcomes with reasons (your requested tweak):
  ```json
  [
    {
      "outcome": "success",
      "reason": "Implementation matched responsibility statement"
    },
    {
      "outcome": "partial_implementation",
      "reason": "Missing core service 'CurrencyConverter'"
    },
    {
      "outcome": "failed",
      "reason": "Circular dependency detected after core extension"
    }
  ]
  ```

**Technical Data**
- `fingerprint` — AST-derived signature
- `logic_hash` — Hash based on AST structure
- `dependencies` — List of nodes this node depends on

**Project Context**
- `project_health_score` — Integer 0–100 representing overall project consistency at time of last decision
- `related_node_summary` — Short summary of the most relevant neighboring nodes and their current maturity

---

This version now gives the deciding LLM a **true historical biography** of each node — not just the last outcome, but the full sequence of successes, partials, failures, and the reasons behind every regression.

Would you like me to:
- Add this as a formal **"Node Biography Schema"** section in the specbook?
- Update Stage 9 (Synthesis) to explicitly describe how it uses this full history?
- Or integrate it into the existing stages we already have?

Let me know your preference and we'll keep moving forward.
