# Product Requirements Document (PRD)
**Project Name:** Relia
**Version:** 1.0 (MVP)
**Status:** Approved
**Last Updated:** 2025-12-05

---

## 1. Executive Summary

**Relia** is the "ESLint for cloud costs" — an open-source, CI/CD-native tool that prevents cloud waste before it ships.

Unlike CloudHealth or Vantage which report on *past* spend, Relia acts as a strict **pre-commit hook for infrastructure**. It parses Terraform/Pulumi changes, estimates the cost impact, and blocks deploys that violate budget policies.

**The Vision:** Catch a $10k/month mistake in the Pull Request, not the monthly bill.
**Tagline:** *"Stop paying the infrastructure tax. Start shipping."*

---

## 2. Problem Statement

### The "Bill Shock" Loop
1.  **Engineers** change Terraform code (adding a NAT Gateway or upsizing an RDS).
2.  **Deploy** happens automatically via CI/CD.
3.  **Finance** gets the bill 30 days later: *“Why did our AWS bill jump $5k?”*
4.  **CTO** spends 3 days investigating and blames "lack of governance."

### The Gap
Tools exist for *Optimization* (finding unused resources later) and *Reporting* (pretty dashboards for Finance).
**No tool effectively exists for Prevention** in the engineer's workflow.

---

## 3. User Persona

### "Jordan" — The Product-Minded Tech Lead
- **Role:** Senior Engineer / Tech Lead at a Series B startup ($5k-$50k/mo cloud spend).
- **Motivation:** Wants to ship features, not manage budgets. Hates getting yelled at by Finance.
- **Behavior:** Lives in the terminal and GitHub. Rejects complex "enterprise" platforms.
- **Needs:** A tool that feels like `black` or `eslint` — simple, fast, deterministic.

---

## 4. Functional Requirements

### FR1: Pre-Deploy Cost Estimation
**Goal:** Predict the dollar impact of a PR before it merges.
*   **Input:** Terraform (`.tf`), Pulumi (`.yaml`/`.py`) directories.
*   **Process:**
    1.  Parse Infrastructure-as-Code (IaC) to identify resource changes (Add/Modify/Delete).
    2.  Map resources (e.g., `aws_instance.t3_large`) to Cloud Pricing APIs (AWS Price List API).
    3.  Calculate monthly cost delta.
*   **Output:** CLI table and PR comment showing `+$120.50/mo`.

### FR2: Budget Guardrails
**Goal:** Enforce spending limits automatically.
*   **Configuration:** `.relia.yml` defines policies.
*   **Logic:**
    *   *Warning:* "This change adds >$500/mo."
    *   *Blocker:* "Total Budget of $5,000/mo exceeded."
*   **Override:** Allow manual approval (e.g., via commit message flag or PR label).

### FR3: Utilization Analysis & "Fix" Mode
**Goal:** Identify and fix waste in *existing* infrastructure.
*   **Input:** Read-only access to AWS CloudWatch / GCP Monitoring.
*   **Process:** Find resources with low utilization (e.g., <5% CPU avg for 7 days).
*   **Interactive Fix:** `relia fix` presents a TUI to apply rightsizing recommendations (e.g., downsize RDS) by automatically patching the `.tf` file.

---

## 5. Non-Functional Requirements

### NFR1: The "Wow" Experience (TUI)
*   **Terminal-First:** No web UI. All interactions happen in the CLI.
*   **Visuals:** Use `Rich` libraries for beautiful tables, diffs, and progress bars.
*   **Speed:** Simple estimates must complete in <5 seconds.

### NFR2: Accuracy & Trust
*   **Explicit Confidence:** If exact pricing isn't known (e.g. data transfer costs), output a range or flag it as "Variable Cost."
*   **Transparency:** Show *how* the cost was calculated (e.g. "730 hours * $0.10/hr").

### NFR3: Safety
*   **Read-Only:** Relia never modifies cloud resources directly. It only modifies *code* (locally) or blocks *processes* (CI).
*   **No Secrets:** Never log credential values or sensitive resource data.

### NFR4: Distribution
*   **Containerized:** Must be available as a lightweight Docker container for effortless CI/CD integration and reproducible execution.

---

## 6. Success Metrics (MVP)

*   **Adoption:** 500+ GitHub Stars in first 3 months.
*   **Usage:** Used in 50+ active repositories (checked via telemetry or GH search).
*   **Value:** "Saved me money" testimonials from early adopters.

---

## 7. Roadmap & Scope

### Phase 1: The Linter (MVP)
*   CLI skeleton (Typer/Rich).
*   Basic Terraform parsing (HCL2).
*   AWS Pricing integration (EC2, RDS, Cache).
*   `relia estimate` command.

### Phase 2: The Guardrail (CI/CD)
*   GitHub Action wrapping the CLI.
*   PR Commenting bot.
*   Budget policy logic (`.relia.yml`).
*   `relia check` command.

### Phase 3: The Optimizer (Fixer)
*   CloudWatch integration.
*   Utilization scanning.
*   `relia fix` interactive TUI for rightsizing.
