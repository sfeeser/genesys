# Genesis Engine v7.0 

Chapter 1: The Manifesto  
Chapter 2: The SQLite Registry  
Chapter 3: The Metamorphosis Pipeline  
Chapter 4: The Telemetry of Transformation  
Chapter 5: The Surgical Inner Loop  
Chapter 6: The Hexagonal Gates  
Chapter 7: The Canvas  
Chapter 8: The Mechanical Delivery  
Chapter 9: The Agentic Loop  
Chapter 10: The Greenfield Protocol  

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
* **NodeID:** `kind.visibility.module.package.receiver.symbol.arity`.
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
    receiver TEXT NOT NULL CHECK (receiver IN ('none', 'pointer', 'value')),
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



# **Chapter 7: The Canvas (Observational Authority)**

The Canvas is the primary interface for **Observational Authority**. It renders a **Convergence Graph** derived directly from the Registry’s `edges` and `scc_cluster_members` tables. 

### **The "Why" for the Developer**
The Canvas turns "Debugging" into "Observing." Instead of tailing logs, the developer watches the DAG. If the vibration starts to climb and the Tension Lines turn red, the developer knows exactly where the architecture is fighting its own constraints.

## **7.1. The Convergence Graph (SCC Compression)**
To manage the "Cognitive Explosion" of 10,000 nodes, the Canvas uses **SCC Compression**:
* **The Blob:** Tightly coupled cyclic regions (Strongly Connected Components) are visually grouped into singular, translucent "Blobs."
* **The Intra-Cluster View:** Clicking a Blob expands it to reveal the internal nodes and their local cycle logic.
* **The External Boundary:** Edges pointing to standard libraries or external modules are rendered as "Anchors" at the edge of the Canvas, visually grounding the local project to its environment.

## **7.2. Kinetic Telemetry (The Solver's Heartbeat)**
The Canvas uses a physical engine to represent the invisible work of the **Convergence Controller (CRA)**.

### **A. Vibration (Entropy & Thrashing)**
Vibration represents the delta between current code and the **Hexagonal Envelope**.
* **High Amplitude:** The node is failing **Gate A (Physics)** or **Gate D (Compilation)**. The solver is rapidly iterating.
* **Low Amplitude:** The node has passed structural gates and is fine-tuning **Gate C (Behavior)**.
* **Stasis:** The node has hit **Equilibrium** (State 5: Sequenced).

### **B. Pulsing (Iteration Pressure)**
The "Pulse" is the engine's heartbeat. 
* **Rapid Pulse:** High retry density. The CRA is burning tokens/compute to resolve a complex conflict.
* **Flatline:** If a node is not Green but the pulse stops, the engine has **STALLED**.

## **7.3. Tension Lines (Edge Pressure)**
Edges are rendered as elastic "Tension Lines" that communicate the status of **Typed Dependencies**.
* **Type-Mismatched (Red Tension):** The line appears taught and red, pulling the source and target nodes together. This indicates a **Contract-ID** mismatch.
* **Logic-Incompatible (Orange Tension):** The line vibrates, indicating that while the signatures match, the **Behavioral Invariants** (Gate C) are failing across the boundary.
* **Satisfied (Green/Blue):** The line is relaxed and stable.

## **7.4. The Sovereignty Shockwave**
When a node is assigned **Class-0 Authority** or reaches **State 3 (Anchored)**, it emits a visual "Shockwave." 
* This wave ripples through the **Tension Lines**, visually updating the maturity of downstream "Subject" nodes. 
* It serves as a warning: the "Sovereign" node is now a fixed point in the universe; all other nodes must bend their logic to satisfy its contract.

## **7.5. The Gate Overlay**
When the user hovers over a vibrating node, a hexagonal HUD appears, showing the status of the six **Hexagonal Gates**.
* **Gates A-B (Structural):** Pulse at the top.
* **Gates C-D (Functional):** Pulse in the center.
* **Gates E-F (Architectural):** Pulse at the bottom.
* This allows the developer to instantly see *why* a node is stuck: *"It's compiling (Gate D), but it's failing the Behavioral Invariants (Gate C)."*


# **Chapter 8: The Mechanical Delivery (Localhost Sovereignty)**

Genesis is delivered as a local-first service. It is a **Project-Level Daemon** that runs in your workspace, ensuring your code, your registry, and your intent never leave your machine.

## **8.1. The Genesis Daemon (`genesis serve`)**
The engine is a self-contained Go binary. When executed, it launches a high-performance HTTP/2 and WebSocket server on `localhost:8080`.

* **The Concurrent Registry (WAL Mode):** The server operates the `.genesis/genome.db` in **Write-Ahead Logging (WAL)** mode. This allows the **CRA Solver** to execute heavy write-transactions for mutation worksets without blocking the **Canvas UI** from reading the current graph state.
* **Virtual Loom:** It initializes an in-memory VFS (Virtual File System) to stage "Hydration" (State 4) code before it is committed to the physical disk.
* **Toolchain Bridge:** The server manages a throttled worker pool for the `go` compiler and `test` runners, ensuring that background "Gate" checks do not starve the local UI thread.

## **8.2. The Event Sequencer (UI Propagation)**
To ensure the Canvas provides a diagnostic-grade view of the solver, the server acts as a **State Aggregator**. 

* **Ordered Guarantees:** The server utilizes an internal event queue to ensure that "Gate A" passes are never rendered after a "Gate B" failure for the same node. The UI always reflects a linear, logical progression of the **Metamorphosis Pipeline**.
* **Event Coalescing & Debouncing:** During high-velocity solver cycles, the server coalesces rapid updates into **60Hz UI Frames**. This prevents WebSocket saturation and ensures the "Vibration" and "Pulse" on the Canvas remain fluid and meaningful.

## **8.3. The WebSocket Pulse (Real-time Diagnostics)**
The link between the **Registry Engine** and the **Canvas** is a bi-directional WebSocket stream.
* **The Telemetry Stream:** Streams node maturity changes, gate status, and "Kinetic Tension" levels.
* **The Control Backchannel:** Allows the user to "Intervene" via the browser, pausing the solver or manually anchoring a node's contract.

## **8.4. The Lifecycle of a Request**
1.  **Intent:** You run `genesis draft "Add health-check endpoint"` in your terminal.
2.  **Ingestion:** The Go daemon identifies the target package via the SQLite index.
3.  **Visualization:** The Localhost UI instantly centers the Canvas on the affected **SCC Cluster**, which begins to **Pulse** in white.
4.  **Verification:** As you approve the plan, the server initiates the **Metamorphosis Pipeline**. You watch the nodes **Vibrate** in yellow as the gates are tested.
5.  **Materialization:** Upon successful **Sequencing**, the server performs the **Atomic Swap** to disk and updates the **Canonical YAML Export**.

## **8.5. Security & Privacy Model**
* **Zero-Exfiltration:** No code is transmitted to any cloud. If an external LLM provider is used, only the **Surgery Fragment** and its **Contract Context** are sent.
* **Localhost Binding:** The server binds strictly to `127.0.0.1`. It is inaccessible from the external network, preserving the "Sovereignty" of the project.


# **Chapter 9: The Agentic Loop (The CRA Solver)**

The **Convergence Controller (CRA)** is the reasoning engine of Genesis. It manages the iterations within the **Metamorphosis Pipeline**, specifically during the **PLAN** and **HYDRATING** stages. Its goal is to reach an equilibrium where the code satisfies both the **Specbook's Intent** and the **Registry's Physics**.

## **9.1. The Bounded Reasoning Cycle**
The CRA operates within a **Closed-Loop Feedback System**. Every suggestion must pass through a "Formal Filter" before it is considered for the **Mutation Workset**.
1.  **Proposal:** The agent suggests a mutation based on the **Gene**.
2.  **Simulation:** The mutation is staged in the **VFS**.
3.  **Audit:** The **Hexagonal Envelope** (Gates A & B) evaluates the physics.
4.  **Termination:** If equilibrium is not reached within $N$ cycles, the CRA triggers a **Controlled Exit**.

## **9.2. Sovereignty-Driven Resolution**
The solver uses **Authority Classes** to decide which node must "warp" during a conflict.
* **Class 0 (Immutable):** Fixed points. The CRA must solve around these.
* **Class 2 (Negotiable):** Plastic nodes. The CRA prioritizes mutating these to satisfy constraints.
* **Topology Shifts:** If two Class-0 nodes are incompatible, the CRA suggests an **Adapter** or **Shim** rather than violating a contract.

## **9.3. The Friendly Panic: Failure Taxonomy**
The CRA is designed to be a "Trust Surface." It distinguishes between **Intentional Contradictions** and **Systemic Violations**.

### **A. Specbook Panic (UNSAT)**
* **Nature:** The solver behaved correctly, but the user's requirements are logically impossible under Go's type physics.
* **Canvas UX:** The affected nodes turn **Static Orange**. The **Minimal Conflict Set (MCS)** is highlighted with "Tension Lines" showing exactly where the logic breaks.
* **Response:** The system offers "Relaxation Paths" (e.g., *"Interface X expects a String, but Struct Y provides an Int. Would you like to update the Interface?"*).

### **B. Engine Panic (Invariant Violation)**
* **Nature:** The system has detected a violation of its own internal guarantees (e.g., a Replay Mismatch, an SCC Atomicity failure, or a D-ID drift).
* **Canvas UX:** The entire Canvas **Freezes and Desaturates**. A "Hard Lock" icon appears on the Registry Engine status.
* **Response:** The solver halts immediately. The system provides a **Reproducibility Hook** and a snapshot of the last valid Registry state. This is a "Stop the Line" moment for the machine.

## **9.4. The Cost Function (Architectural Parsimony)**
To prevent "Over-Engineering," the CRA applies a **Complexity Tax**:
* **New Node Penalty:** Discourages creating unnecessary abstractions.
* **Signature Mutation Penalty:** Favors logic changes over breaking API contracts.
* **Reuse Reward:** Incentivizes the agent to utilize existing nodes in the **Genome**.

## **9.5. SCC Batch Reasoning**
For nodes within an **SCC**, the CRA provides the agent with the **Entire Cluster Context**. The agent must provide a valid mutation for the entire cycle simultaneously. On the Canvas, this is visualized as the **SCC Swarm** vibrating until all members of the cycle pass the **Hexagonal Gates** in the same transaction.


# **Chapter 10: The Unidirectional User Interface (The Greenfield Protocol)**

The Genesis UI is the **Control Surface** for the Project. It is designed to minimize cognitive load by separating **Architectural Intent** from **Surgical Execution**. It follows a "Fire and Observe" model, where the user defines the destination and the engine navigates the physical terrain of the code.

## **10.1. The Greenfield Protocol: Spawning Life from Intent**
To bootstrap or evolve a project, the user provides the **Soul** (Vision), the **Physics** (Specbook), and a target.

```bash
./saayn genesis -v vision.md -s specbook.yaml --target ./my-app
```

## **10.2. The 5-State Genome State Machine**
The UI reflects the **Metamorphosis Pipeline** state, enforcing strict guardrails to prevent "vibe-coding" drift.

| Genome State | Action & Focus | Agent Guardrails |
| :--- | :--- | :--- |
| **1. Conceptual** | Defines the "Gallery" (Purpose). | **No Code Allowed.** Agent must reject Go logic. |
| **2. Hollow** | Generates the "Canvas" (Stubs). | **Zero-Logic Rule.** No `if`, `for`, or assignments. |
| **3. Anchored** | **The Signature Lock.** | **Contract Freeze.** Registry forbids signature changes. |
| **4. Hydrating** | The **Surgical Phase** (Logic). | **Isolation.** Cannot call nodes still in State 1. |
| **5. Sequenced** | **Equilibrium** (Materialized). | **Hash Locking.** D-ID/L-ID committed to SQLite. |

## **10.3. The Mechanical Trace (CLI Observability)**
The CLI provides a high-fidelity "Nervous System" trace. It isn't just logging; it's proving the **Hexagonal Gates** are opening and closing.

```plaintext
🧬 PHASE 0: CONTEXTUAL INGESTION
--------------------------------------------------------------------------------
📄 Vision:   'Distributed Worker' intent identified in vision.md
📜 Physics:  12 Nodes identified in specbook.yaml
🏗️ Build Order: [model] -> [registry] -> [worker] -> [main]

🌱 MATERIALIZING GENOME (SURGICAL INNER LOOP ACTIVE)
--------------------------------------------------------------------------------
[03/12] PROCESSING: internal/worker/worker.go

    🔬 DRAFTING: Initializing node 'saayn.Worker.Start'...
    
    ⚖️  GATE A (PHYSICS): Walking via go/ast...
       ├─ Syntax Check... ✅
       └─ Interface Check (Worker)... ✅

    🧠 GATE B (IDENTITY): Checking against Signature Lock...
       └─ ✅ Match: func(context.Context) error

    🔧 REMEDIATION (Iteration 1/3):
       ├─ 🚩 FINDING: "Logic uses time.Sleep. Vision requires context-aware cancellation."
       ├─ Applying AST Patch...
       └─ Re-verifying Physics... ✅

    💾 COMMIT: Writing to Registry (Sequenced)...
       └─ ✅ Logic Hash: d4e5f6 | D-ID: locked-env-v1
```

## **10.4. The Trust Surface: Handling Panics**
The UI distinguishes between the developer's "Human Errors" and the engine's "Systemic Failures."

### **A. Specbook Panic (The UNSAT Wall)**
* **Scenario:** The vision.md and specbook.yaml are logically incompatible.
* **UI Response:** "Your constraints cannot be satisfied as written." The CLI highlights the **Minimal Conflict Set** (MCS).
* **Tone:** Collaborative. "We hit a wall. Here is where the physics breaks."

### **B. Engine Panic (Invariant Violation)**
* **Scenario:** The engine detected a "Physics Break" (e.g., L-ID mismatch or SCC atomicity failure).
* **UI Response:** **HARD STOP.** The Canvas desaturates. The Registry locks.
* **Tone:** Serious. "The engine detected a violation of its own guarantees. Snapshotting state for recovery."

## **10.5. Short-Term Memory Preservation**
Genesis treats the **Registry Engine** and the **CLI Log** as its persistent memory. Because every node is anchored by a **Logic Hash**, the agent never "forgets" where it is. If the process is killed, pointing SAAYN back at the `.genesis/genome.db` allows it to resume exactly one hash after the last successful sequence.


# **Chapter 11: The Architectural Preflight (Gate 0)**

Chapter 11 defines the **Mandatory Preflight Protocol** that must be satisfied before the Genesis Engine is permitted to execute. This chapter establishes a "Sanity Firewall" between **Architectural Design** and **Code Materialization**.

### **11.1. The Law of Preflighted Execution**
Genesis executes preflighted designs. It does not perform open-ended architectural negotiation at runtime. 
> **Law:** Genesis executes. It does not debate. If the map is broken, the engine does not start.

### **11.2. The Design-Time Handshake (Auditor Prompt)**
Users must "Bless" the `specbook.yaml` using this standardized audit before invoking the engine:
> "You are the Genesis Auditor. Perform a binary PASS/FAIL audit of this `specbook.yaml` against the Laws of Dependency Physics:
> 1. **Topology Integrity:** Analyze `allowed_imports`. **FAIL** if any package imports a higher layer or if `internal` imports `mcp`.
> 2. **Sibling Isolation:** **FAIL** if `surgeon`, `audit`, or `auditlog` import each other.
> 3. **Grammar Normalization:** **FAIL** if `node_id` grammar uses `receiver_shape`. Required: `kind.visibility.module.package.receiver.symbol.arity`.
> 4. **SCC Compliance:** **FAIL** if a cycle spans multiple layers. **PASS** only for intra-package recursion."


### **The "Gate 0" Auditor Prompt (Normative)**
Users are encouraged to run this prompt against their Specbook before invoking the engine.
> "You are the Genesis Architectural Auditor. Perform a binary PASS/FAIL audit of this `specbook.yaml` against the Laws of Dependency Physics:
> 1. **Topology Integrity:** Identify forbidden dependency cycles and illegal upward dependencies (e.g., a utility importing an orchestrator). 
>    * *Note: Intentional SCC-eligible cycles are permitted only if they do not violate layer boundaries or authority classes.*
> 2. **Identity Grammar:** Verify `node_id` follows: `kind.visibility.module.package.receiver.symbol.arity`.
> 3. **Stability Flow:** High-stability nodes (`authority_class: 0`) must never depend on lower-stability nodes (`authority_class: 2`).
> 4. **Boundary Law:** Verify no internal logic package imports the `mcp` (Model Context Protocol) layer.
> 
> Return **# PREFLIGHT STATUS: PASS** or a list of **Structural Violations**."

## **11.3. The Genesis Runtime Gate (Gate 0)**
When `./saayn genesis` is executed, the engine performs its own internal, automated version of the Preflight. This is a **Hard Guardrail** to protect the integrity of the SQLite Registry.

1.  **Ingestion:** The engine parses the `specbook.yaml` into an in-memory graph.
2.  **Validation:** It runs a **Cycle Detection Algorithm**.
    * **FAIL:** If a cycle crosses a **Layer Boundary** (e.g., utility → main → utility).
    * **PASS (SCC-Mapped):** If a cycle is contained within a single **SCC-eligible package** or defined mutation unit.
3.  **The Hard Stop:** If any check fails, Genesis exits immediately with a **Specbook Panic**. It will not create files, modify the Registry, or burn tokens.

## **11.4. Failure Handling: The Specbook Panic**
If Preflight fails, the system provides a **Minimal Conflict Set (MCS)**. 
* **Visual Feedback:** The **Canvas** highlights offending edges in **Static Orange**.
* **Recovery:** Genesis is blocked. The Architect must correct the `specbook.yaml` and re-run the Preflight. There is no override flag.

## **11.5. Rationale: Separation of Concerns**
| Domain | Responsibility | Focus |
| :--- | :--- | :--- |
| **Preflight (Design)** | Validation, Critique, Normalization. | Thinking/Debate. |
| **Genesis (Execution)** | Deterministic Materialization. | Doing/Building. |


The "Internal Constitution" has been promoted. By moving this to **Chapter 12**, we establish it as the final destination of the Manifesto—the **Self-Referential Blueprint**. 

We are now defining the **Genesis Genome**. This is the specific package topology required to build the Genesis Engine itself. If the engine can materialize its own source code using these laws, the Greenfield Protocol is proven.

***

### **Critic’s Brief: Technical Architecture Review**
**Role:** Lead Systems Architect / Dependency Engineer.
**Objective:** Validate the "Bootstrap Topology" for Genesis.
**Focus Areas:**
1.  **DAG Sequence:** Does the Build Order (Topological Sort) ensure that leaf nodes are materialized before orchestrators?
2.  **Package Completeness:** Does the updated set cover all mechanical requirements (SQLite, VFS, AST, WebSockets)?
3.  **Self-Hosting Integrity:** Is the hierarchy robust enough for Genesis to "Sequencing" itself?

***

## **Chapter 12: Package Topology & Dependency Law**

Chapter 12 is the **Internal Constitution**. It defines the allowed structure of the Genesis Engine and provides the ground truth for the Preflight Gate.

### **12.1. Canonical allowed_imports (Closed World)**
To satisfy the Topology Laws, the engine is restricted to the following explicit import graph:

| Package | Layer | Allowed Imports (Downstream Only) |
| :--- | :--- | :--- |
| `internal/identity` | L1 | [] (Root) |
| `internal/registry` | L2 | `internal/identity` |
| `internal/spec` | L3 | `internal/identity` |
| `internal/scanner` | L4 | `internal/identity`, `internal/spec` |
| `internal/staging` | L5 | `internal/identity`, `internal/registry`, `internal/spec` |
| **`internal/surgeon`** | L6 | `internal/identity`, `internal/registry`, `internal/scanner`, `internal/staging` |
| **`internal/audit`** | L7 | `internal/identity`, `internal/registry`, `internal/spec`, `internal/scanner`, `internal/staging` |
| **`internal/auditlog`** | L8 | `internal/identity`, `internal/registry` |
| `internal/metamorphosis`| L9 | `internal/identity`, `internal/registry`, `internal/spec`, `internal/staging`, `internal/surgeon`, `internal/audit`, `internal/auditlog` |
| `internal/orchestrator` | L10 | `internal/identity`, `internal/registry`, `internal/spec`, `internal/metamorphosis` |
| `internal/telemetry` | L11 | `internal/identity`, `internal/registry`, `internal/auditlog`, `internal/orchestrator` |
| `internal/mcp` | L12 | `internal/identity`, `internal/orchestrator` |
| `cmd/saayn` | L13 | `internal/orchestrator`, `internal/mcp` |

### **12.2. Deterministic Build Order**
The DAG Sequence for self-materialization is:
1. `identity` 2. `registry` 3. `spec` 4. `scanner` 5. `staging` 6. `surgeon` & `audit` (Siblings) 7. `auditlog` 8. `metamorphosis` 9. `orchestrator` 10. `telemetry` 11. `mcp` 12. `cmd/saayn` (Apex).

### **12.3. Corrected Grammar Standard**
* **Field:** `receiver`
* **Values:** `none | value | pointer`
* **Rule:** All references to `receiver_shape` in the Registry Schema (Chapter 2) and Metadata are hereby deprecated and replaced with `receiver`.

## **12.4. Forbidden Patterns (Architectural Crimes)**
* **The Circular Feedback:** `metamorphosis` → `orchestrator`. (The state machine must be a servant to the orchestrator).
* **Boundary Bleed:** Any `internal` package importing `telemetry` or `mcp`. (Transport and UI are observers, not participants).
* **Registry Bypass:** Any package performing direct file I/O without going through `staging` or `registry`.

## **12.5. Rationale: The Self-Sustaining Loop**
By defining Genesis using this topology, we allow the engine to treat itself as a **Specbook**. 
* **The Test:** If we update the `identity` package, the Orchestrator will see the drift, trigger a **Sovereignty Shockwave**, and force a re-reconciliation of every layer above it. 
* **The Proof:** If the engine can successfully "Surround-Hydrate" its own `metamorphosis` package without breaking the DAG, the **Greenfield Protocol** is verified.

