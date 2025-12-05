# Relia OSS Implementation Plan — MVP

## Strategic Positioning
**Relia** is the "ESLint for cloud costs" — a developer-first tool that catches infrastructure waste in the pull request.

**Core Promise**: "See the cost of your code before you merge."

**Target Persona**: "Jordan", the Series B Tech Lead who is tired of monthly "why is the bill so high" meetings.

**MVP Scope (Tier 1 Focus)**:
- **Local-First**: Beautiful TUI (Rich) + Python.
- **CI-Native**: GitHub Action Integration.
- **AWS-First**: EC2, RDS, and S3 focus.

---

## Epic Map (Scope)

### Epic 0: Foundation & Tooling
**Goal**: Repository structure, CI/CD pipeline, and "Hello World" CLI experience.
- **Story 0.1**: Repository Scaffold (Poetry, Typer, Rich).
- **Story 0.2**: Testing & Quality (Pytest, Black, Isort, Mypy).
- **Story 0.3**: CI/CD (GitHub Actions for tests).
- **Story 0.4**: Mock TUI (Prototype the "Wow" output).
- **Story 0.5**: Dockerfile (Official image based on python:slim).

### Epic 1: Core Parsing Engine (The "Eyes")
**Goal**: Accurately read infrastructure state from code.
- **Story 1.1**: Terraform Parser (Static `.tf` parsing via `python-hcl2`).
- **Story 1.2**: Resource Normalization (Convert HCL dicts to standard `Resource` objects).
- **Story 1.3**: Variable Resolution (Basic variable handling, warn on complex dynamic blocks).
- **Story 1.4**: Diff Detection (Identify Add vs Modify vs Delete).
- **Story 1.5**: Terraform Plan JSON Fallback (Optional "deep mode" using `terraform show -json` for 100% accuracy).

### Epic 2: Pricing Logic (The "Brain")
**Goal**: Convert resource definitions into monthly dollar amounts.
- **Story 2.1**: AWS Price List API Integration (Fetch SKU prices).
- **Story 2.2**: Local Pricing Cache (SQLite database to prevent API rate limits/latency).
- **Story 2.3**: Resource Matcher (Logic to map `t3.large` + `us-east-1` -> Price SKU).
- **Story 2.4**: Estimation Engine (Calculate `HourlyRate * 730`).
- **Story 2.5**: Bundled Pricing DB (Ship a lightweight SQLite DB with the CLI for 0-latency first runs).

### Epic 3: The CLI Experience (The "Interface")
**Goal**: A polished, "magic" terminal interface.
- **Story 3.1**: `relia estimate` Command (Run parser -> pricer -> output).
- **Story 3.2**: "Cost Diff" View (Show cost delta like a code diff).
- **Story 3.3**: Interactive Tables (Rich tables with color-coded alerts).
- **Story 3.4**: Error Handling (Friendly, actionable errors for parsing failures).
- **Story 3.5**: Visual Cost Topology (Render a tree view of costs: `Project -> VPC -> Cluster -> Instance`).

### Epic 4: CI/CD Integration
**Goal**: The "Gatekeeper" in the development workflow.
- **Story 4.1**: `relia check` Command (Exit code 1 if over budget).
- **Story 4.2**: GitHub Action (Docker container wrapping Relia).
- **Story 4.3**: PR Commenter (Post markdown summary to GitHub PR).

### Epic 5: Policy & Configuration
**Goal**: User-defined guardrails.
- **Story 5.1**: Configuration File (`.relia.yml` loader).
- **Story 5.2**: Budget Logic (Total cap, per-resource cap).
- **Story 5.3**: Policy Enforcement (Block vs Warn thresholds).

### Epic 6: Documentation & Launch
**Goal**: Make it adoptable in <5 minutes.
- **Story 6.1**: Quickstart Guide (README with GIF).
- **Story 6.2**: Architecture Docs (How it works).
- **Story 6.3**: PyPI Release.
- **Story 6.4**: Docker Hub Registry (Publish `relia-io/relia`).

---

## Detailed Story Breakdown

### Epic 0: Foundation & Tooling
*(Completed in Phase 1)*
- **Story 0.1**: Repository Scaffold `[DONE]`
- **Story 0.2**: CI Pipeline `[DONE]`
- **Story 0.4**: Mock TUI `[DONE]` (Prototype showing potential output).

### Epic 1: Core Parsing Engine

#### Story 1.1: Terraform Parser
**What**: Read `.tf` files from a directory.
**Acceptance**:
- `Parser.parse("./infra")` returns a list of resource dictionaries.
- Handles multiple `.tf` files in a folder.

#### Story 1.2: Resource Normalization
**What**: transform raw HCL into `ReliaResource` model.
**Acceptance**:
- `aws_instance` becomes a standard class with `instance_type`, `region`, etc.
- Unknown/unsupported resources are gracefully ignored/logged.

### Epic 2: Pricing Logic

#### Story 2.1: AWS Price List API
**What**: Fetch pricing data from AWS.
**Acceptance**:
- Python client to query AWS Pricing API.
- Filter by Region and Service.

#### Story 2.2: Local Pricing Cache (SQLite)
**What**: Store pricing data locally to speed up subsequent runs.
**Acceptance**:
- First run: Fetches ~10MB pricing data (slow-ish).
- Second run: Instant (sub-second).
- Schema: `sku`, `attributes`, `price_per_unit`.

### Epic 3: The CLI Experience

#### Story 3.2: "Cost Diff" View
**What**: Show cost changes clearly.
**Acceptance**:
```diff
+ aws_instance.web  (t3.large)   +$68/mo
- aws_instance.old  (t3.micro)   -$12/mo
```
- Green for additions/savings, Red for cost increases (or vice versa? Usually Red = cost up).

### Epic 4: CI/CD Integration

#### Story 4.2: GitHub Action
**What**: `uses: relia-io/action@v1`.
**Acceptance**:
- `action.yml` defined.
- Runs `relia check`.
- Pass/Fail based on configured budget.
