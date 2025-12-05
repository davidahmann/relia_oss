# Relia

<div align="center">

![Relia Banner](https://placeholder-banner-url.com) *<!-- TODO: Add banner later -->*

**The "ESLint" for Cloud Costs.**
Prevent cloud waste before it ships to production.

[![CI](https://github.com/relia-oss/relia/actions/workflows/test.yml/badge.svg)](https://github.com/relia-oss/relia/actions)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Python](https://img.shields.io/badge/python-3.10+-blue.svg)](https://www.python.org/downloads/)
[![Code Style: Black](https://img.shields.io/badge/code%20style-black-000000.svg)](https://github.com/psf/black)

[Features](#-features) • [How It Works](#-how-it-works) • [Quick Start](#-quick-start) • [Roadmap](#-roadmap)

</div>

---

## ⚡ Problem: The "Bill Shock" Loop

Every engineering team knows the ritual:
1.  Engineer deploys a change (e.g., upsizing an RDS instance).
2.  CI/CD passes (tests pass, build succeeds).
3.  **30 days later**: Finance asks why the AWS bill jumped by $5,000.

Existing tools (CloudHealth, Vantage) are **reactive**. They tell you *after* you've wasted the money.

## 🛡️ Solution: Shift Left

**Relia** sits in your Pull Request. It parses your Terraform/Pulumi changes, estimates the monthly cost impact, and blocks the merge if you blow your budget.

> **Tagline**: *"Stop paying the infrastructure tax. Start shipping."*

---

## 🧩 How It Works

Relia runs as a CLI tool in your local environment or CI pipeline to minimize the feedback loop. See [ARCHITECTURE.md](ARCHITECTURE.md) for details.

```mermaid
flowchart LR
    A[Engineer] -->|`git push`| B(Pull Request)
    B --> C{Relia CI Check}

    subgraph "Relia Engine"
    D[Parser] -->|1. Read .tf files| E[Estimator]
    E -->|2. Fetch Prices| F[AWS Pricing API]
    F -->|3. Calculate Delta| G[Budget Policy]
    end

    C -->|Run| D

    G -->|Over Budget ❌| H[Block Merge]
    G -->|Under Budget ✅| I[Allow Merge]

    style H fill:#ffcccb,stroke:#ff0000
    style I fill:#ccffcc,stroke:#00ff00
```

---

## ✨ Features

### 1. Pre-Deploy Cost Estimation
Know exactly what a PR will cost **before** you click merge. Supports both standard `.tf` files and `terraform plan -json` output for high accuracy.

```bash
# Standard Estimate
$ relia estimate
📊 Relia Cost Estimate ...

# With Visual Topology
$ relia estimate --topology
🌳 Infrastructure Topology
☁️  Project
└── aws_instance
    └── 💻 web  $60.00/mo

# With Cost Diff
$ relia estimate --diff
📉 Cost Diff
+ aws_instance.web  +$60.00/mo
```

### 2. Budget Guardrails
Set clear policies in `.relia.yaml` and enforce them with `relia check`.

```yaml
# .relia.yaml
budget: 500.0
rules:
  aws_instance: 50.0 # Max price per instance
```

```bash
$ relia check --budget 500
✅ Within budget. Total: $450.00, Limit: $500.00
```

### 3. CI/CD Integration
Relia ships with a GitHub Action to block expensive PRs automatically.

---

## 🚀 Quick Start

### 1. Installation

Relia is available via PyPI (and Poetry):
```bash
pip install relia
# or
poetry add relia
```

### 2. Run Locally

Navigate to your Terraform project and run:

```bash
# Estimate cost
relia estimate ./infrastructure

# Check against budget
relia check ./infrastructure --budget 1000
```

### 3. Run with Docker

You can also run Relia without installing Python:

```bash
docker run --rm -v $(pwd):/app relia-io/relia estimate .
```

### 4. Add to CI/CD

Add Relia to your GitHub Actions workflow to block expensive PRs.

```yaml
# .github/workflows/cost-check.yml
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

---

## 🗺️ Roadmap

- [x] **Phase 1 (MVP)**: CLI, Terraform Support, AWS Pricing.
- [x] **Phase 2**: GitHub Action, Budget Policies via `.relia.yml`.
- [ ] **Phase 3**: Utilization Scanning ("Fix" Mode).

---

## 🤝 Contributing

We love contributions! Please check out [CONTRIBUTING.md](CONTRIBUTING.md) to get started.

## 📄 License

This project is licensed under the [Apache 2.0 License](LICENSE).
