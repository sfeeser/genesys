# **Chapter 1: The Genesis Manifesto (v7.0 - RELEASE)**

## **1.1. The Genesis Value Proposition**
Genesis is a **Bounded Synthesis System** that transforms architectural intent into a deterministic engineering discipline. It operates on the principle that the **Specification is the Permanent Asset (The Soul)** and the code is a transient, generated liability. By utilizing a local-first **Registry Engine**, Genesis enforces architectural physics and type-safety through a **Virtual Loom**, ensuring every materialization is verified, stable, and reproducible.

## **1.2. The Authority Stack**
To resolve architectural drift, Genesis enforces a three-tier hierarchy:
1.  **The Specbook (Normative Authority):** The immutable source of truth for architectural intent.
2.  **The Registry Engine (Physical Authority):** A local-first, relational database (`.genesis/genome.db`) that enforces the physical state, graph structure, and convergence metadata.
3.  **The Canvas (Observational Authority):** A kinetic, read-only visualization of the Registry Engine's state.

## **1.3. The DNA Registry: Structural & Semantic**
The Registry Engine replaces flat-file manifests with a relational model to manage complexity at scale:
* **The Structural Genome:** Persists the **Identity Quad**, typed dependency edges, and **SCC (Strongly Connected Component)** clusters. It acts as the "Hard Physics" layer.
* **The Semantic Index:** A non-authoritative, advisory vector index stored within the engine. It enables heuristic **Search-by-Intent** via local inference. **Note:** All index selections are advisory and must be validated by the Structural Genome and the Acceptance Envelope.
* **The Audit Export:** The engine generates a **Canonical Audit Export** (YAML/JSON) to provide a deterministic, human-readable trail for version control and peer review.

## **1.4. The Identity Quad: Environment-Stable Determinism**
Every node is anchored by four immutable dimensions:
* **NodeID:** `kind.visibility.module.package.receiver_shape.symbol.arity`.
* **C-ID (Contract):** Canonical signature and generic constraints.
* **L-ID (Logic):** SHA-256 of the normalized, order-independent AST.
* **D-ID (Dependency/Environment):** A recursive digest of the transitive dependency graph, `go.sum` hash, toolchain version, and explicit build context (**GOOS, GOARCH, tags, and flags**).

## **1.5. The Convergence Controller (CRA 3.0)**
The CRA is a **Bounded Optimization Solver** that treats SCCs as atomic mutation units.
* **Atomic SCC Mutation:** Within a cycle, the CRA performs a coordinated multi-node resolution. Partial updates are prohibited; the SCC must converge as a batch.
* **Authority Partitioning:** The solver is strictly bound by node authority. It cannot mutate Class-0 (Immutable) nodes to solve a conflict; it must warp negotiable nodes or suggest topology changes.

## **1.6. The Agentic Development Pipeline**
Genesis follows a strictly governed, transactional execution flow:
1.  **DRAFT:** Heuristic discovery via Semantic Index.
2.  **GRAPH:** Deterministic "Blast Radius" calculation via relational queries.
3.  **PLAN:** A **CRA-governed optimization cycle** generating code in a VFS sandbox.
4.  **APPLY:** A **Transactional Batch Surgery** that splices code and updates the Registry. If any gate fails, the transaction rolls back.

## **1.7. The Hexagonal Acceptance Envelope**
All synthesis must pass the **Hexagonal Gates**: **Gate A** (Physics), **Gate B** (Identity Coherence), **Gate C** (Behavioral Invariants), **Gate D** (Compilation), **Gate E** (Canonical Replay), and **Gate F** (Cost/Complexity).

## **1.8. The Kinetic Canvas & Convergence Graph**
The Canvas renders a **Convergence Graph** derived from the Registry Engine’s dependency data.
* **SCC Compression:** Tightly coupled cyclic regions are visually collapsed into "Blobs" to represent atomic mutation units.
* **Kinetic Telemetry:** Observational signals—**Pulse** (iteration pressure) and **Vibration** (instability)—allow the developer to monitor the "Virtual Loom" without interrupting the solver.


# **Chapter 2: The DNA Registry (Final Hardening)**

## **2.1. Registry Physical Architecture**
The Registry is a single SQLite 3 database (`.genesis/genome.db`). 

### **The Canonical Audit Export (Law)**
Any operation that commits a change to the SQLite DB must trigger a deterministic, sorted YAML export to `genome.yaml`. This ensures that **Git** remains the auditor of the binary engine's state.

## **2.2. The Hardened Schema**

### **A. Metadata (The Environment Singleton)**
We have tightened the singleton to include full build provenance and session verification.

```sql
CREATE TABLE metadata (
    singleton INTEGER PRIMARY KEY CHECK (singleton = 1),
    go_version TEXT NOT NULL,
    goos TEXT NOT NULL,
    goarch TEXT NOT NULL,
    build_tags_json TEXT NOT NULL,
    build_flags_json TEXT NOT NULL,
    cgo_enabled INTEGER NOT NULL CHECK (cgo_enabled IN (0,1)),
    go_sum_hash TEXT NOT NULL,
    module_graph_hash TEXT NOT NULL,
    workspace_mode TEXT NOT NULL,
    last_sequence_hash TEXT NOT NULL,
    schema_version TEXT NOT NULL DEFAULT 'v7'
);
```

### **B. Nodes (State-Aware Integrity)**
We have componentized the NodeID and added state-based nullability rules to be enforced by Chapter 3 logic.

```sql
CREATE TABLE nodes (
    node_id TEXT PRIMARY KEY,
    
    -- COMPONENTIZED IDENTITY
    kind TEXT NOT NULL,
    visibility TEXT NOT NULL,
    module_path TEXT NOT NULL,
    package_path TEXT NOT NULL,
    receiver_shape TEXT NOT NULL, -- 'none', 'pointer', or 'value'
    symbol_name TEXT NOT NULL,
    arity INTEGER NOT NULL,

    -- THE IDENTITY QUAD
    contract_id TEXT,         -- Required for maturity >= 'anchored'
    canonical_contract TEXT,  -- Required for maturity >= 'anchored'
    logic_hash TEXT,          -- Required for maturity == 'sequenced'
    dependency_hash TEXT,     -- Required for maturity == 'sequenced'

    maturity TEXT NOT NULL CHECK (maturity IN ('draft', 'hollow', 'anchored', 'hydrated', 'sequenced')),
    authority_class INTEGER NOT NULL CHECK (authority_class IN (0,1,2)),
    gene TEXT,
    business_purpose TEXT,
    last_audit_timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### **C. The Graph & SCC Model (Revision-Synced)**
We have unified `graph_revision` and `graph_hash` to eliminate redundancy and ensured SCC membership is explicitly revision-scoped.

```sql
CREATE TABLE graph_revisions (
    graph_hash TEXT PRIMARY KEY, -- SHA-256 of the total edge set
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE edges (
    source_node_id TEXT NOT NULL,
    target_node_id TEXT,           -- Internal
    target_external_symbol TEXT,   -- External
    edge_kind TEXT NOT NULL,
    graph_hash TEXT NOT NULL,
    PRIMARY KEY (source_node_id, target_node_id, target_external_symbol, edge_kind, graph_hash),
    FOREIGN KEY (source_node_id) REFERENCES nodes(node_id),
    FOREIGN KEY (graph_hash) REFERENCES graph_revisions(graph_hash)
);

CREATE TABLE scc_cluster_defs (
    cluster_id TEXT NOT NULL,
    graph_hash TEXT NOT NULL,
    authority_partitioned INTEGER NOT NULL,
    node_count INTEGER NOT NULL,
    PRIMARY KEY (cluster_id, graph_hash),
    FOREIGN KEY (graph_hash) REFERENCES graph_revisions(graph_hash)
);

CREATE TABLE scc_cluster_members (
    cluster_id TEXT NOT NULL,
    graph_hash TEXT NOT NULL,
    node_id TEXT NOT NULL,
    PRIMARY KEY (cluster_id, graph_hash, node_id),
    FOREIGN KEY (cluster_id, graph_hash) REFERENCES scc_cluster_defs(cluster_id, graph_hash),
    FOREIGN KEY (node_id) REFERENCES nodes(node_id)
);
```

### **D. Semantic Provenance (The Record Bridge)**
Vectors are no longer "floating." They are bound to a specific node AND a specific inference profile.

```sql
CREATE TABLE inference_profiles (
    profile_id TEXT PRIMARY KEY,
    model_name TEXT NOT NULL,
    model_revision TEXT NOT NULL,
    dimensions INTEGER NOT NULL,
    distance_metric TEXT NOT NULL,
    summary_schema_version TEXT NOT NULL,
    summary_prompt_hash TEXT NOT NULL,
    chunking_policy TEXT NOT NULL,
    normalization_policy TEXT NOT NULL
);

CREATE TABLE semantic_records (
    record_id INTEGER PRIMARY KEY,
    node_id TEXT NOT NULL,
    profile_id TEXT NOT NULL,
    summary_hash TEXT NOT NULL,
    is_stale INTEGER NOT NULL CHECK (is_stale IN (0,1)),
    FOREIGN KEY (node_id) REFERENCES nodes(node_id),
    FOREIGN KEY (profile_id) REFERENCES inference_profiles(profile_id)
);

CREATE VIRTUAL TABLE semantic_index USING vss0(
    record_id INTEGER,
    vector(3072)
);
```

## **2.3. Transactional Enforcement (The Solver Workspace)**
To guarantee SCC atomicity and environment safety:

1.  **Environment Sentinel:** Before any mutation transaction, the engine verifies the current `runtime_context` against the `metadata` singleton. If the OS, Architecture, or Tags differ, the transaction is forbidden.
2.  **The Mutation Workset:** The CRA does not update `nodes` directly. It writes to a `mutation_worksets` table.
    * **The Finalization Gate:** A trigger on the workset verifies that if any node in an SCC is being mutated, **all** members of that SCC (for the current `graph_hash`) must be present in the workset with a valid state transition.
    * **Atomic Flush:** Only upon 100% SCC completeness and **Hexagonal Gate** approval is the workset flushed to the `nodes` table.


# **Chapter 3: The Metamorphosis Pipeline**

The pipeline is a **Unidirectional State Machine**. While the CRA may iterate within a state, a node only moves forward once it clears the specific **Hexagonal Gates** associated with that transition.

## **3.1. State 1: Conceptual (The Gallery)**
* **Definition:** The node exists only as a **PublicID** and a **Gene** (Spec) in the `nodes` table.
* **Goal:** Define the "Business Purpose" and boundary.
* **Exit Gate:** Semantic Indexing. The node must be "Enriched" and searchable in the `semantic_index`.
* **Registry State:** `maturity = 'draft'`.

## **3.2. State 2: Hollow (The Canvas)**
* **Definition:** The **C-ID (Contract)** is generated. The node has a Go signature but no body (it returns a default or panics).
* **Goal:** Establish the "Skeleton" of the package.
* **The Virtual Loom:** All Hollow nodes are staged in the **VFS (Virtual File System)**.
* **Exit Gate (Gate A):** The VFS package must pass `go/types` check. All interfaces must be satisfied by these stubs.
* **Registry State:** `maturity = 'hollow'`.

## **3.3. State 3: Anchored (The Contract)**
* **Definition:** The **Signature Lock**. The `contract_id` is hashed and written to the Registry.
* **Goal:** Freeze the API so callers can be safely synthesized.
* **The Sovereignty Check:** If a node is marked `authority_class = 0`, its Contract is frozen here. Any attempt to change the signature during later stages results in a **CRA Termination (UNSAT)**.
* **Exit Gate (Gate B):** Identity Quad coherence.
* **Registry State:** `maturity = 'anchored'`.

## **3.4. State 4: Hydrating (The Surgery)**
* **Definition:** The Logic is injected. The "Agentic Loop" (governed by the CRA) writes the actual Go code.
* **The Mutation Workset:** Logic is written to the `mutation_worksets` table, not the physical disk.
* **SCC Synchronization:** If a node is part of an SCC, the entire cluster must reach the end of "Hydration" simultaneously.
* **Exit Gates (Gate C & D):** Node-local tests must pass, and the package must compile.
* **Registry State:** `maturity = 'hydrated'`.

## **3.5. State 5: Sequenced (Equilibrium)**
* **Definition:** The **L-ID (Logic Hash)** and **D-ID (Dependency Hash)** are calculated and locked.
* **Goal:** Persistence and Materialization. 
* **The Atomic Flush:** The Registry Engine moves the data from the `mutation_workset` to the `nodes` table and triggers the **Batch AST Surgeon** to write the code to the physical disk.
* **Exit Gate (Gate E & F):** Canonical Replay check. The materialized code must generate the exact same Logic Hash as the Workset.
* **Registry State:** `maturity = 'sequenced'`.

### **The "Not-My-First-Rodeo" Conflict Handling**

| Scenario | Pipeline Response |
| :--- | :--- |
| **Signature Drift** | Node is demoted to **Hollow**. Callers are flagged for re-reconciliation. |
| **Logic Failure** | Node stays in **Hydrating**. CRA attempts a "Topology Change" or "Adapter Synthesis." |
| **Environment Mismatch** | Pipeline Pauses. **D-ID** is invalidated. Re-Sequencing required. |


### **The "Market" Differentiator**
Most AI tools try to go from **State 1 to State 5** in one jump. That’s why they break. Genesis forces a "Pause" at **State 3 (Anchored)**. By locking the contracts before writing the logic, we ensure that the "Physics" of the system is solved before we ever spend a single token on the "Brain."


# **Chapter 4: The Telemetry of Transformation**

While the Registry Engine enforces the "Hard Physics" of the project, the **Kinetic Canvas** provides the "Nervous System" feedback. This chapter defines how the visual layer interprets the five states of the Metamorphosis Pipeline to provide real-time, diagnostic telemetry.

## **4.1. Visual State Mapping (The Maturity Spectrum)**
The Canvas does not render static boxes; it renders **Maturity Velocity**. Each node's visual signature is a direct reflection of its `maturity` column in the SQLite `nodes` table.

| Pipeline State | Canvas Representation | Diagnostic Meaning |
| :--- | :--- | :--- |
| **1: Conceptual** | **Ghost Node (Dashed White)** | **Intent Only.** The node exists in the `nodes` table as a spec but has no physical footprint in the VFS. |
| **2: Hollow** | **Translucent White (Pulsing)** | **Skeleton Phase.** The node has a C-ID (Contract) and exists as a stub. Pulse frequency reflects iteration pressure. |
| **3: Anchored** | **Blue Halo (Static)** | **Signature Lock.** The API is frozen. This node is now a fixed constraint for all surrounding logic. |
| **4: Hydrating** | **Yellow Core (Vibrating)** | **Active Synthesis.** The CRA is injecting logic and running the **Hexagonal Gates** (C, D). |
| **5: Sequenced** | **Green Solid (Static)** | **Equilibrium.** L-ID and D-ID are locked. The node is materialized to disk. |

## **4.2. Kinetic Telemetry & Solver Feedback**
To prevent the "Black Box" problem common in AI tools, the Canvas uses kinetic physics to reveal what the **Convergence Controller (CRA)** is doing under the hood.

### **A. The Sovereignty Shockwave**
When a node transitions from **Hollow** to **Anchored**, it emits a visual "ripple" across its dependency edges. This informs the developer that a contract has been locked and any "Subject" nodes downstream must now warp their logic to satisfy this new sovereign constraint.

### **B. The SCC Swarm (Atomic Vibration)**
When the engine identifies a **Strongly Connected Component**, the affected nodes visually clump into a **Blob**. 
* **Coordinated Vibration:** During the **Hydrating** state, the entire Blob vibrates in unison. 
* **The Deadlock Signal:** If the Vibration Amplitude increases without a change in State, it signifies a **Logic Deadlock**, indicating the CRA is thrashing between incompatible constraints within the cycle.


### **C. Tension Lines (Constraint Mapping)**
Edges between nodes are rendered as "Tension Lines" that change color based on the **Acceptance Envelope** status:
* **Red (High Tension):** Gate A or B failure (Type/Contract mismatch).
* **Orange (Medium Tension):** Gate C or D failure (Logic/Compilation mismatch).
* **Blue/Green (Locked):** Constraints satisfied.

## **4.3. User Interpretation Model**
The Canvas is designed for "At-a-Glance" troubleshooting:
* **Blobs:** Tightly coupled regions that must evolve together.
* **Vibration:** Active work or instability. If it doesn't stop, the spec is likely **UNSAT**.
* **Pulse:** The "Heartbeat" of the solver. If the pulse stops but the node is not Green, the process has **STALLED**.


# **Chapter 5: The Surgical Inner Loop**

The Surgical Inner Loop is the execution phase of the **APPLY** stage. It is responsible for taking the "Approved Logic" from the **Mutation Workset** (State 4) and physically weaving it into the source files (State 5).

### **The "Why" for the Reliability Engineer**
Legacy AI tools often treat codebases like text files, leading to missing brackets or "hallucinated" imports. By operating exclusively at the **Syntax Tree** level, Genesis makes it physically impossible to produce a file that is not syntactically valid. The "Surgery" isn't a text-replace; it's a **Biological Graft**.

## **5.1. The Toolchain: AST vs. DST**
Genesis utilizes a dual-syntax-tree approach to ensure that generated code is indistinguishable from high-quality human code.

* **`go/ast` (The X-Ray):** Used during the **GRAPH** and **PLAN** stages to perform high-speed type checking and dependency tracing. It is the "Scientific" view of the code.
* **`dave/dst` (The Scalpel):** Used during the **APPLY** stage. Unlike the standard library AST, the **Decorated Syntax Tree (DST)** preserves "decorations" (comments, line breaks, and grouping). This ensures that when Genesis edits a function, it doesn't "sanitize" away the human-written documentation around it.


## **5.2. The Splicing Protocol**
The Surgeon follows a strict "Targeted Replacement" strategy rather than a full-file rewrite.

1.  **Node Localization:** The Surgeon uses the **NodeID** components from the Registry to locate the exact byte-offset of the target symbol (function, struct, or interface) in the physical file.
2.  **Constraint Verification:** Before cutting, the Surgeon re-verifies the `logic_hash` of the existing code on disk. If the disk has changed since the **PLAN** stage (Unauthorized Drift), the surgery is aborted to prevent a collision.
3.  **The DST Merge:** The new logic is parsed into a DST fragment and "stitched" into the target file's tree. 
    * **Import Management:** The Surgeon automatically reconciles imports. It adds missing packages and removes unused ones, respecting existing aliases.
4.  **Formatting Alignment:** The Surgeon applies a `gofmt`-compliant pass to the modified DST, ensuring that the new code matches the project's visual style.

## **5.3. The Atomic Materialization (VFS to Disk)**
To prevent "Torn State" (where a file is partially written before a crash), the Surgeon uses an **Atomic Swap** pattern:

* **Step 1:** The modified DST is rendered to a temporary buffer in the **VFS**.
* **Step 2:** The **Hexagonal Envelope** runs **Gate D** (Compilation) on the buffer.
* **Step 3:** If valid, the buffer is written to a `.tmp` file on the physical disk.
* **Step 4:** A filesystem `rename()` call replaces the original file with the `.tmp` file. This is a nearly instantaneous, atomic operation at the OS level.

## **5.4. Collision-Safe Scoping**
When the CRA generates code for State 4 (Hydrating), the Surgeon provides a "Safe Scope" map:
* **Shadowing Protection:** If the agent tries to name a local variable `err` but a variable named `err` is already dominant in that scope, the Surgeon forces a rename (e.g., `err2`) during the DST stitch to preserve logical intent without shadowing.
* **Receiver Stability:** If a method is being added to a struct, the Surgeon ensures it uses the same receiver name (e.g., `(f *Fighter)`) as the existing method set.


# **Chapter 6: The Hexagonal Gates**

The Hexagonal Gates are a set of six mandatory verification layers that every mutation workset must clear before moving from **State 4 (Hydrating)** to **State 5 (Sequenced)**.

### **The "Fail-Fast" Design**
The gates are ordered by **Computational Cost**. Gate A (Fast AST check) happens thousands of times a minute. Gate D (Full Compilation) and Gate F (Complexity) only happen once a candidate has been "Hydrated." This ensures the engine doesn't waste expensive "Deep LLM" or "Full Build" time on code that can't even pass a basic type-check.

## **6.1. Gate A: Physics (Structural Integrity)**
* **Mechanism:** `go/ast` + `go/types`.
* **The Check:** Does the code parse? Are types consistent? If the node is a struct, does it implement the interfaces it claims to? 
* **Failure Mode:** "Syntax Error" or "Interface Mismatch." The solver is sent back to rethink the signature.

## **6.2. Gate B: Identity (The Quad Anchor)**
* **Mechanism:** Registry Comparator.
* **The Check:** Does the generated node match the **Identity Quad** (NodeID, C-ID, L-ID, D-ID) authorized in the **PLAN** stage? 
* **The Authorized Delta:** This gate allows for "Approved Drift"—if the CRA intended to change a signature, Gate B verifies the new signature matches the **Specbook** exactly.
* **Failure Mode:** "Identity Drift." Usually indicates the agent hallucinated a signature change it wasn't authorized to make.

## **6.3. Gate C: Behavioral (Invariant Verification)**
* **Mechanism:** Go Test Runner + Sandbox.
* **The Check:**
    * **C1 (Local):** Do the node’s internal table-driven tests pass?
    * **C2 (Global):** Do the package-level invariants remain intact?
* **The Sandbox:** Tests are run in a **Virtual File System (VFS)** to prevent side effects on the actual repository.
* **Failure Mode:** "Logic Regression." The solver must rewrite the body to satisfy the behavioral contract.

## **6.4. Gate D: Genomic (Full-System Compilation)**
* **Mechanism:** `go build` + `go vet`.
* **The Check:** Does the entire package (including all callers and dependencies) still compile with the new node in place? 
* **The SCC Batch:** For cyclic dependencies, this gate is cleared only when the **entire cluster** is ready.
* **Failure Mode:** "Integration Failure." Often reveals "Ghost Dependencies" or circular imports that weren't caught in the local AST scan.

## **6.5. Gate E: Replay (Deterministic Stability)**
* **Mechanism:** Canonical Re-Parser.
* **The Check:** If the engine takes the materialized code and re-calculates its **Logic Hash (L-ID)**, does it match the hash stored in the **Mutation Workset**?
* **The Normalizer:** This gate ignores whitespace, comment formatting, and import ordering to ensure only **Semantic Determinism** is measured.
* **Failure Mode:** "Nondeterministic Materialization." Indicates the **Surgical Inner Loop** introduced an unintentional change.

## **6.6. Gate F: Cost (Architectural Fitness)**
* **Mechanism:** Complexity Heuristics.
* **The Check:** Did the synthesis stay within the "Complexity Budget"?
    * Is the **Cyclomatic Complexity** too high?
    * Did the solver introduce too many **Adapter/Shim** nodes?
    * Is the **Fan-out** (dependency count) excessive?
* **Failure Mode:** "Architectural Bloat." The CRA is forced to find a simpler solution or escalate to the user for a manual refactor.



























