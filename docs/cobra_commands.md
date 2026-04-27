### **2. Comprehensive Command Profiles**

These are the standalone **Cobra CLI commands** exposed to users. Each command serves as a thin entrypoint that invokes the corresponding function in the Core Processing stages. In the future, when an `evoke:` directive is added to the SpecBook, many of these can be bypassed for fully automated, internal execution.

#### **`init`**
- **Why it is needed:** To establish the **Physical Authority (L2)** of the project by creating a persistent memory store (the genome). Without it, Genesis has no foundation to store knowledge about the codebase.
- **What it does:** Creates the hidden `.genesis/` directory and materializes an empty `genome.db` SQLite database with the required relational and vector similarity search (VSS) schemas.
- **How it does it:** Executes the `Migrate()` path in the Registry, enabling WAL mode and foreign key constraints.
- **Dependencies:** `internal/registry`
- **Potentially Destructive?** **Yes** — will overwrite an existing genome if `--force` is used.

#### **`enrich`**
- **Why it is needed:** To grant the engine **Sight and Intelligence** by turning raw code into a semantically rich, searchable map.
- **What it does:** Acts as a non-destructive initialization and reconciliation step. It discovers new or changed nodes via AST scan and enriches each node with accurate semantic data.
- **How it does it:** Performs a delta scan of the codebase and only processes nodes that are missing or have changed. For each node:
  - If a matching entry exists in `scaffolding.yaml`, it uses the high-quality `business_purpose`, responsibility statement, and 3072-dimensional embedding from the scaffolding.
  - Otherwise, it sends the node to the FAST LLM to generate a business-purpose summary.
  - Computes the 3072-dimensional embedding and hydrates the `semantic_records` table in the genome.
- **Dependencies:** `internal/scanner`, `internal/cognition`, `internal/registry`
- **Potentially Destructive?** **No**. It only appends and updates metadata.

#### **`verify`**
- **Why it is needed:** To detect **Genome Drift** — changes made by humans (or previous stages) outside the expected Genesis flow.
- **What it does:** Compares the stored `logic_hash` and fingerprint of each node against the current physical AST to ensure the engine’s internal truth matches reality.
- **How it does it:** Generates a logic hash based solely on AST structure (ignoring whitespace, comments, and cosmetic changes) and flags any behavioral differences. After `hydrate`, it also confirms that new nodes were properly reconciled by enrich.
- **Dependencies:** `internal/scanner`, `internal/registry`
- **Potentially Destructive?** **No**. This is a read-only audit.

#### **`graph`**
- **Why it is needed:** To calculate the **Blast Radius** of any potential change before it occurs.
- **What it does:** Maps dependencies for a given node — both "who it calls" and "who calls it" (inverted dependencies).
- **How it does it:** Performs a full AST walk to build a call graph by tracing imports and function calls.
- **Dependencies:** `internal/scanner`, `internal/identity`
- **Potentially Destructive?** **No**.

#### **`locate`** (alias: `loc`)
- **Why it is needed:** To enable precise natural language discovery at different levels of granularity — package, file, or node.
- **What it does:** Performs semantic search against the genome and returns the most relevant results based on user intent. Supports scoping to package, file, or node level.
- **How it does it:** Vectorizes the natural language query and performs local cosine similarity search. Returns matching items along with their Blast Radius.
- **Flags:**
  - `--package` — Scope search to package level
  - `--gofile` — Scope search to Go file level
  - `--node` — Scope search to node level (default)
- **Dependencies:** `internal/access`, `internal/index`, `internal/registry`
- **Potentially Destructive?** **No**.

#### **`ping`**
- **Why it is needed:** To verify that the engine’s "brain" (LLM and embedding providers) is reachable and properly configured.
- **What it does:** Validates connectivity and API key validity for the FAST, DEEP, and EMBED models.
- **How it does it:** Sends minimal test requests to each configured provider through the cognition client.
- **Dependencies:** `internal/cognition`
- **Potentially Destructive?** **No**.

#### **`hydrate`** (alias: `hyd`)
- **Why it is needed:** To turn the abstract Desired State into a real, compilable Go codebase skeleton.
- **What it does:** Generates the directory structure, packages, and `.go` files with fully resolved signatures and imports, but leaves function bodies hollow.
- **How it does it:** Reads `scaffolding.yaml`, processes packages/files in hydration_order (core → leaf), writes the files, and triggers a targeted `enrich` pass on the new code so that high-quality intent from scaffolding is properly merged into the genome.
- **Dependencies:** `internal/scaffolding`, `internal/scanner`, `internal/registry`
- **Potentially Destructive?** **Yes**. Writes real `.go` files to disk.

## Core Processing Stages
**Stage 1: The Anchor (init)**  
- **Purpose:** To establish the persistent "Memory" of the project.  
- **What it does:** It materializes the physical infrastructure required for Genesis to function.  
- **Mechanism:** Creates the hidden `.genesis/` directory and bootstraps the `genome.db` SQLite database. This database stores both relational data (AST structures) and semantic data (vector embeddings).  
- **Destructive?** **Yes**. It will overwrite an existing genome if explicitly forced.

**Stage 2: The Sight (enrich)**  
- **Purpose:** To give the engine intelligence and semantic understanding of the existing codebase.  
- **What it does:** This is a non-destructive initialization and reconciliation step. It populates the genome with new or changed nodes and enriches each node with accurate semantic data.  
- **Mechanism:**  
  1. **AST Scan**: Identifies every function, struct, interface, and other relevant nodes.  
  2. **Delta Detection**: Only processes nodes that are new or have changed.  
  3. **Semantic Enrichment**: If a matching entry exists in `scaffolding.yaml`, it uses the high-quality responsibility and embedding from scaffolding; otherwise, it uses the FAST LLM to generate a business-purpose description.  
  4. **Embedding**: Computes a 3072-dimensional vector and stores it in the genome.  
- **Destructive?** **No**. It only hydrates and updates metadata.

**Stage 3: The Verify**  
- **Purpose:** To ensure the engine’s internal "Truth" matches physical reality and to detect Genome Drift.  
- **What it does:** It compares the stored genome against the current codebase to identify any unauthorized changes.  
- **Mechanism:** Calculates a Logic Hash based on AST structure for every node and compares it to the stored value. Ignores cosmetic changes. After Stage 8 (The Skeleton), it confirms that new nodes were correctly reconciled by enrich.  
- **Destructive?** **No**.

**Stage 4: Structural Mapping (graph)**  
- **Purpose:** To calculate the Blast Radius of any potential change.  
- **What it does:** It builds a complete map of interdependencies between code components.  
- **Mechanism:** Performs a full AST walk to capture dependencies ("who calls whom") and inverted dependencies ("who is called by whom").  
- **Destructive?** **No**.

**Stage 5: Discovery (locate)**  
- **Purpose:** To navigate the codebase using natural language intent.  
- **What it does:** Finds specific logic based on semantic queries and returns the most relevant nodes along with their Blast Radius.  
- **Mechanism:** Vectorizes the user’s query and performs local cosine similarity search against the 3072-dimensional embeddings.  
- **Destructive?** **No**.

**Stage 6: Gatekeeper**  
- **Purpose:** To act as a quality gate — determining whether the Specbook is clear, complete, and actionable enough to proceed.  
- **What it does:** Evaluates the `genesis.yaml` Specbook and decides whether to PASS or FAIL.  
- **Mechanism:** Uses the DEEP LLM to assess clarity, completeness, consistency, and feasibility against Go best practices and architectural constraints.  
- **Output:** PASS or FAIL + detailed feedback (including improvement prompts on failure).  
- **Destructive?** **No**.

---
---

**Complete Specification – Stages 7–9: The Intelligent Coding Agent**

These stages are the only place in the entire Genesis Engine where the **coding agent** (DEEP LLM + tool-calling loop) is permitted to operate. All prior stages (1–6) remain purely deterministic and mechanical.

### Core Principles for Stages 7–9

- The coding agent operates **exclusively via MCP-style tool calls**. It never receives giant context dumps.
- Every response must end with a === GENESIS CONTINUATION DIRECTIVE === block (the state forwarder).
- The Genesis orchestrator treats this directive as the single source of truth for “what happens next.”
- Node Biography + Scaffold Graph + Continuation Directive together provide ongoing knowledge of the project state.
- Every decision is made with **on-demand context** pulled via the seven tools listed below.
- The **Node Biography** (Stateful History & Telemetry) is recorded for every node so the agent can see its own past actions and avoid endless looping or regressions.
- Every materialization step (write to code) is followed **immediately** by `enrich` so the Genome stays perfectly synchronized with physical AST.
- The **Scaffold Graph** is the single authoritative blueprint from the moment Stage 7 runs.
- SCCs are treated as **atomic mutation units** — the agent must handle the entire blob together.
- **Coreness** (`x-y`) is the primary architectural importance metric:  
  - `x` = number of inverse dependencies (fan-in / blast radius)  
  - `y` = number of dependencies (fan-out)  
  This value is stored on every scaffold node and returned by the blast-radius tool.

**MCP-like Tool-Use Philosophy**  
Real senior engineers ask questions when they need information. The coding agent must do the same. Tool-use provides human-like reasoning, efficiency, better judgment, and reduced hallucination.

**Tools the Coding Agent May Call** (these are the **only** interface to the world):

1. `get_node_history(node_id)` — full growth_history + white_blood_cell_attacks + synthesis_outcome_history
2. `get_original_spec(node_id)` — exact SpecBook text this node was born from
3. `get_related_nodes(node_id, limit)` — summary of neighboring nodes (with coreness)
4. `get_project_health()` — current overall project consistency score (0–100)
5. `get_previous_llm_reasoning(node_id)` — what the agent said in prior attempts on this node
6. `get_blast_radius(node_id)` — fresh blast radius including coreness `x-y`, risk level, and affected SCCs
7. `get_node_code(node_id)` — current implementation on disk (if any)

**Node Biography Fields** (stored per node in Registry + Scaffold Graph):

- `growth_history` — array of stage numbers in chronological order (e.g. `[7, 8, 9, 8, 9]`)
- `white_blood_cell_attacks` — array of regression events: `{stage, reason, timestamp, cycle}`
- `synthesis_attempts` — integer count of Stage 9 attempts on this node
- `synthesis_outcome_history` — array of `{outcome, reason, timestamp}`
- `fingerprint`, `logic_hash`, `dependencies`
- `project_health_score` (at time of last action)
- `coreness` — stored as string `"x-y"`

### The Scaffold Graph (Authoritative Blueprint)

- Built purely from the SpecBook (plus any existing code if present).
- Lives permanently in the Registry (dedicated tables: `scaffold_nodes`, `scaffold_edges`, `scaffold_revisions`, `scaffold_scc`).
- Contains only three levels of granularity: package → file → symbol (functions, structs, interfaces, methods — no local variables or tiny helpers).
- Stores coreness, responsibility statements, maturity, authority_class, and embeddings.
- Is versioned via graph revisions exactly like the Genome.
- Serves as the map the coding agent follows for all materialization and mutation decisions.

### The Canvas (Observational Layer Only)

- Purely read-only kinetic visualization served on `localhost:8080`.
- Renders the Scaffold Graph + Genome state with maturity colors, vibration, tension lines, SCC blobs, sovereignty shockwaves, etc.
- Highlights the current node being processed and shows real-time biography events.
- Does **not** store or edit code — it only observes the results of the Surgical Inner Loop.

### Stage 7 – Scaffolding (Graph Construction – No Physical Write Yet)

**Trigger:** `genesis scaffold` or first run of `genesis gen`.

**Exact Flow:**
1. Coding agent receives **minimal initial context**: SpecBook + current Registry state (empty or partial in greenfield) + any existing Scaffold Graph.
2. Agent uses tools (`get_original_spec`, `get_project_health`, `get_related_nodes`) as needed to analyze requirements.
3. Agent proposes and builds the complete Scaffold Graph (package/file/symbol layout, responsibility statements, initial edges, topological order, and SCC detection).
4. Agent writes the new Scaffold Graph revision into the Registry.
5. Agent exports/updates the human-readable `scaffolding.yaml`.
6. Records initial `growth_history` entries for all new scaffold nodes.

**Exit Condition:** Scaffold Graph is now the single source of truth. **No files have been written to disk yet.** After writing the Scaffold Graph, the agent outputs its Continuation Directive with next_action: "hydrate" or next_action: "pause_for_human".

### 15.4 Stage 8 – The Skeleton (Agent-Driven Hollow Materialization + Genome Bootstrap)

**Trigger:** `genesis hydrate` or automatically after successful Stage 7.

**Exact Sequential Flow:**
1. Coding agent is handed **only** the current Scaffold Graph revision.
2. Agent walks the graph in topological order (core packages → leaf packages, treating SCCs as atomic units).
3. For each node in the graph:
   - Agent prepares the hollow code (package declaration, imports, structs, interfaces, function signatures, and `// TODO: implement per responsibility`).
   - Agent calls the **Surgical Inner Loop**:
     - Code is first staged in the **VFS**.
     - Surgeon runs **Gate A** (Physics / AST + go/types check).
     - On success → atomic rename from VFS to real disk file.
   - **Immediately after** successful write:
     - Agent calls `enrich` (Stage 2) to scan the new physical AST and populate the Genome.
     - Maturity is set to `hollow` in both Scaffold Graph and Genome.
     - Canvas updates in real time (node turns translucent white + pulsing).
   - Records entry in `growth_history` and `synthesis_outcome_history`.
4. If any write fails Gate A, records a `white_blood_cell_attack` and retries (max 2 attempts) or pauses with visible error on Canvas.

**Exit Condition:** All scaffold nodes exist as hollow but compilable `.go` files on disk and fully populated entries in the Genome at maturity `hollow`.

### Stage 9 – Synthesis (Intelligent Implementation Loop)

**Trigger:** `genesis synthesize` or `genesis gen` (after skeleton is complete).

**Exact Loop Flow:**
1. Coding agent receives current Scaffold Graph revision + updated Genome state.
2. Agent walks the graph in topological order (SCC blobs treated as atomic).
3. For each node that is not yet at maturity `implemented`:
   - Agent **must use tools** to gather exactly what it needs before reasoning (see 15.0 tool list).
   - Agent checks `synthesis_attempts` and `white_blood_cell_attacks` to avoid loops.
   - Agent reasons and generates the real implementation logic.
   - Agent calls the **Surgical Inner Loop** for a **targeted DST splice** into the existing file (via VFS).
   - On success:
     - Agent calls `enrich` (or targeted `verify + enrich`) to update Genome with new `logic_hash`, maturity = `implemented`, and dependencies.
     - Records success in `synthesis_outcome_history`.
     - Canvas updates (yellow vibrating → green solid).
4. **Loop-prevention rule:** After 3 failed attempts on the same node or SCC, the agent records a final `white_blood_cell_attack`, pauses, and surfaces an UNSAT condition with the minimal conflict set (using `get_blast_radius` and `get_related_nodes`).
5. After every synthesis attempt (success or failure): the agent must output a Continuation Directive telling the orchestrator exactly what to do next (continue, retry, pause, raise UNSAT, propose graph delta, etc.).
    - State Forwarder Mechanics (Orchestrator Side)
       - The Go engine listens for the === GENESIS CONTINUATION DIRECTIVE === marker.
       - It validates and stores the directive.
       - It immediately prepares the next AI call using the directive’s instructions.
       - This creates a clean, deterministic hand-off between stateless prompts while giving the AI full agency over its own workflow.

**Termination Conditions:**
- All nodes in the current Scaffold Graph revision reach `implemented` maturity, or
- Agent explicitly raises UNSAT with full evidence from tools and biography.
