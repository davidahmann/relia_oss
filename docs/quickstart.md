# Quickstart

## 1. Installation

### Via Pip (Recommended)
```bash
pip install relia
```

### Via Docker
You can run Relia without installing anything locally:
```bash
docker run --rm -v $(pwd):/app relia-io/relia estimate .
```

---

## 2. Local Usage

### Estimate Costs
Navigate to your Terraform directory and run:
```bash
relia estimate .
```

**Options:**
*   `--topology`: View a tree visualization of your infrastructure cost.
*   `--diff`: Compare costs against a baseline (if available).

### Check Budget
To enforce a budget (e.g. preventing spend over $100):
```bash
relia check . --budget 100
```
This will exit with code `1` if the estimate exceeds $100.

---

## 3. Configuration

Create a `.relia.yaml` file in your root to define policies:

```yaml
budget: 500.0  # Total monthly budget cap
rules:
  aws_instance: 100.0  # Max cost for any single EC2 instance
  aws_rds_instance: 200.0
```

---

## 4. CI/CD Integration (GitHub Actions)

Add this to `.github/workflows/cost-check.yml`:

```yaml
name: Cost Check
on: [pull_request]
jobs:
  relia:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: relia-io/action@v1
        with:
          path: './infra'
          budget: '1000'
          markdown_report: 'relia_report.md'
```
