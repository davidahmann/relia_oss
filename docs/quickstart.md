# Quickstart Guide: Setting up Relia

This guide covers how to install Relia, run your first cost estimate, and integrate it into your CI/CD pipeline.

## Prerequisites
*   **Python:** 3.10+ (Recommended) OR Docker
*   **Terraform:** files (`.tf`) or a plan (`.json`)
*   **AWS Credentials:** (Optional) for real-time pricing, though Relia works offline.

---

## 1. Installation

### Option A: Install via Pip (Recommended)
Relia is available as a standard Python package.
```bash
pip install relia
```
*Verify installation:*
```bash
relia --version
```

### Option B: Run via Docker
If you prefer not to manage Python dependencies, use the official Docker image.
```bash
docker run --rm -v $(pwd):/app relia-io/relia estimate .
```

---

## 2. Basic Usage: Estimating Costs

Navigate to the root of your Terraform project (where your `main.tf` lives).

### Run a Standard Estimate
```bash
relia estimate .
```
This parses your HCL files and outputs a table of monthly resource costs.

### View Infrastructure Topology
Visualize your cost breakdown as a tree structure:
```bash
relia estimate . --topology
```

### Compare Against Baseline (Diff)
See how your current code differs from a previous state (requires connectivity):
```bash
relia estimate . --diff
```

---

## 3. Advanced: Budget Enforcement

You can use Relia as a linter to fail if costs are too high.

### CLI Flags
```bash
# Fail if monthly cost exceeds $100
relia check . --budget 100
```

### Configuration File (`.relia.yaml`)
For persistent governance, create a configuration file:
```yaml
# .relia.yaml
budget: 500.0          # Total project cap
rules:
  aws_instance: 100.0  # Max for single EC2
  aws_rds_instance: 200.0
```

Then simply run:
```bash
relia check .
```

---

## 4. CI/CD Integration (GitHub Actions)

To prevent expensive merges, add Relia to your Pull Request workflow.

**File:** `.github/workflows/cost-check.yml`
```yaml
name: Relia Cost Guardrail
on: [pull_request]

jobs:
  estimate-cost:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v3

      - name: Run Relia Check
        uses: relia-io/action@v1
        with:
          path: './infra'
          budget: '1000' # Fail PR if > $1000/mo
          markdown_report: 'relia_report.md'

      - name: Comment on PR
        uses: peter-evans/create-or-update-comment@v3
        if: always()
        with:
          issue-number: ${{ github.event.pull_request.number }}
          body-file: 'relia_report.md'
```
