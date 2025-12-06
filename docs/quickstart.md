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

### Option B: Run via Docker
```bash
docker run --rm -v $(pwd):/app relia-io/relia estimate .
```

---

## 2. Quick Setup (New!)
To get started quickly with budget policies and usage overlays, run:
```bash
relia init
```
This creates:
*   `.relia.yaml`: For budgets and per-resource cost rules.
*   `.relia.usage.yaml`: For usage assumptions (e.g. Lambda requests, S3 size).

---

## 3. Basic Usage: Estimating Costs

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

### Global Options
You can configure behavior with global flags:
```bash
# Enable verbose logging for debugging
relia estimate . --verbose

# Specify a target AWS region (Default: us-east-1)
relia estimate . --region eu-central-1
```

### Output Formats
Relia produces beautiful tables for humans, but you can also get machine-readable JSON for your pipelines:
```bash
relia estimate . --format json
```
*Example Output:*
```json
{
  "resources": [
    {
      "name": "aws_instance.web",
      "type": "aws_instance",
      "cost": 60.0,
      "attributes": { ... }
    }
  ],
  "total_cost": 60.0
}
```

### Compare Against Baseline (Diff)
See how your current code differs from a previous state (requires connectivity):
```bash
relia estimate . --diff
```

---

## 4. Advanced Usage

### Handling Complex Variables & Modules
For complex projects using variables, locals, or modules, Relia supports Terraform Plan JSON output.
1. Generate the plan JSON:
   ```bash
   terraform plan -out=tfplan
   terraform show -json tfplan > plan.json
   ```
2. Estimate costs using the JSON plan:
   ```bash
   relia estimate plan.json
   ```

### Usage Assumptions (S3, Lambda, etc.)
Some resources (like S3 or Lambda) depend on usage metrics not present in Terraform. You can define these in `.relia.usage.yaml`:

**Example `.relia.usage.yaml`:**
```yaml
usage:
  aws_s3_bucket.my_bucket:
    storage_gb: 500
    monthly_requests: 10000
```

Relia will automatically load this file and apply the usage data (e.g., 500GB of storage) when calculating costs.

---

## 5. Governance

### Pre-Commit Hook (Recommended)
The best way to save money is to prevent expensive code from ever being committed. Add this to your `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/davidahmann/relia_oss
    rev: v1.1.1
    hooks:
      - id: relia-estimate  # Shows you the cost
      - id: relia-check     # Blocks you if policies fail
```

### Budget Enforcement (CLI)
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

## 6. CI/CD Integration (GitHub Actions)

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
