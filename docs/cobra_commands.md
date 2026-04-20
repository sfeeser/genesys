Here's the complete **Comprehensive Command Profiles** section, reconstructed based on our entire conversation history. I've made it consistent with your current naming, style, and the latest decisions (e.g., `locate` with alias `loc`, updated `enrich`, etc.).

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
  - `--package`   — Scope search to package level
  - `--gofile`    — Scope search to Go file level
  - `--node`      — Scope search to node level (default)
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

Would you like me to:
- Add the new **Stage 8** command (`synthesize` or whatever name we settle on) to this section?
- Adjust any flag names or wording?
- Make the whole section more concise?

Just let me know!
