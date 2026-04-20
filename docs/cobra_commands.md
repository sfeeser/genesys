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

---

### Core Processing Stages

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
- **Mechanism:** Calculates a Logic Hash based on AST structure for every node and compares it to the stored value. Ignores cosmetic changes. After `hydrate`, it confirms that new nodes were correctly reconciled by enrich.  
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

**Stage 7: Scaffolding**  
- **Purpose:** Define the "Desired State" — the complete structural scaffold and high-level organization of the project that later stages will hydrate with implementation code.  
- **What it does:** This is the engine’s non-destructive planning phase. It analyzes the current codebase, ingests the Specbook, and produces a validated blueprint.  
- **Mechanism:** Ingests the Specbook, validates architectural constraints, calculates a dependency graph, and writes `scaffolding.yaml`. For each package and file, it generates a clear responsibility statement and 3072-dimensional embedding.  
- **Structure defined:**  
  - A list of **feature/domain packages** under `internal/`, ordered from core/central features toward supporting/leaf features.  
  - For each package: a list of `.go` files, ordered from core logic toward supporting files.  
  - Use `/internal/` for all core domain and business logic (protected from external imports).  
  - Use `/cmd/<appname>/` as the ultimate leaf that only wires dependencies and starts the application (contains no business logic).  
  - Every package and every `.go` file must have:  
    - A **clear, one-sentence responsibility statement**  
    - A **3072-dimensional vector embedding** (generated so that local agents can semantically search and select the most appropriate package or file when adding or moving nodes in later stages)  
  - **Dependency graph** must be calculated and validated. Acyclic dependencies are **strongly preferred**. Small, well-justified cycles are allowed when they are resolved cleanly with interfaces defined in the core.  
  - Avoid deep directory nesting — favor shallow, cohesive packages.  
  - **Refactor strategy:** Prefer changes in leaf packages first. Core packages may be extended or refactored when doing so leads to cleaner, more idiomatic code, reduces duplication, or better expresses business rules. Any core changes must be reflected back into the scaffolding.yaml to keep the Desired State in sync.  
- **Destructive?** **No** (only affects the output file; previous version is archived).

**Stage 8: Hydration**  
- **Purpose:** To turn the hollow skeleton produced by hydration into fully implemented, working code while safely handling necessary core extensions.  
- **What it does:** This stage takes the hydrated skeleton and intelligently fills in the function bodies, adding or extending core packages when required by leaf functionality.  
- **Mechanism:**  
  - Reads the Desired State from `scaffolding.yaml` and the current genome state.  
  - Processes nodes in **hydration_order** (core → leaf).  
  - For each node:  
    - Generates high-quality implementation code using the DEEP LLM, guided by the responsibility statement and surrounding context.  
    - If a leaf requirement reveals a missing abstraction or service in the core, the stage proposes the necessary core extension.  
    - Core extensions are applied first, followed by a re-hydration of affected leaf nodes if needed.  
  - After code is written, it triggers a mandatory `verify` pass to detect any drift and confirm consistency.  
  - Updates the genome with new `logic_hash`, `maturity` ("implemented"), and updated dependencies.  
- **Edge Case Handling:**  
  - **Core Extension Feedback Loop:** Core-to-leaf ordering is treated as a strong guideline, not an absolute rule. When leaf functionality requires new core behavior, core extensions are allowed. However, to prevent infinite oscillation, each synthesis run tracks a `synthesis_iteration` counter and a `core_extension_count`. Automatic core extensions are limited to a maximum of 2 per synthesis session. After the second extension, further core changes require user confirmation or a manual re-scaffolding. If no measurable progress (new nodes materialized or logic_hash stabilized) occurs after 3 full cycles, the leaf node is implemented with whatever core services are currently available and flagged as “partial_implementation”.  
  - **Enrich Overwrite Cycle:** The original DEEP intent is preserved. Each node maintains both `original_responsibility` and `original_business_purpose` (never overwritten) as well as `current_responsibility` and `current_business_purpose` fields. This allows later stages to judge how well the implemented code matches the original intent.  
  - **Verify False-Positive Drift Loop:** Verify behavior is modulated by node maturity level. Nodes in “Synthesizing” state have relaxed drift detection. Verify only flags changes as unauthorized drift when the node is in “Anchored” or “Implemented” maturity.  
  - **Scaffolding.yaml Desync Loop:** `scaffolding.yaml` is treated as a living document. When core extensions or new files are added during synthesis, new entries are appended to `scaffolding.yaml` with state tracking fields (`state`, `last_modified_by`, `synthesis_version`).  
  - **Hydrate ↔ Synthesis Ordering Deadlock:** When Synthesis needs to create a brand-new file or package that did not exist in the original scaffolding, it first appends the new node to `scaffolding.yaml` with `state: "planned"`, then triggers a targeted (partial) `hydrate` only for the new nodes. Previously hydrated nodes retain their existing maturity state.  
- **Destructive?** **Yes**. It performs physical writes to source code files and updates the genome.
