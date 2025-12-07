# Changelog

## v1.3.2 (2025-12-07) "Community & E2E"
*   **Testing**: Added robust E2E test suite (`tests/test_e2e_live.py`) using subprocesses to verify binary behavior in CI.
*   **Community**: Added GitHub Templates (`bug_report`, `feature_request`) and specific `CONTRIBUTING` guide.
*   **Meta**: Updated contact email to `dahmann@lumyn.cc`.

---

## v1.3.1 (2025-12-07) "Offline & Demo Fixes"
*   **Data Fix**: Corrected pricing data in `bundled_pricing.db` for `t3.large` and `m5.large` instances to reflect accurate On-Demand rates (fixing a data quality issue in offline mode).
*   **Demo**: Included E2E demo script `demo/run_demo.sh` to showcase 12-factor configuration logic.
*   **Config**: Validated and cleaned up configuration precedence logic.

---

## v1.3.0 (2025-12-07) "Golden OSS"
**Features (12-Factor)**
*   **Env Var Config**: Relia now supports configuration via Environment Variables (e.g., `RELIA_BUDGET=500`), enabling cleaner CI/CD integration without checking in config files.

**Developer Experience**
*   **Makefile**: Added a root `Makefile` to standardize setup, testing, and linting (`make setup`, `make test`).
*   **CLI UX**: Added `relia --version` flag support.
*   **Docs**: Added explicit Prerequisites and updated Contributing guide.

**Breaking Changes**
*   Internal configuration loading logic has been refactored to use `pydantic-settings`. Behavior remains backward compatible with `.relia.yaml` files.

**Bug Fixes**
*   **CI/CD**: Fixed Mypy type checking errors (`pydantic.fields.FieldInfo`) and corrected test assertion logic for 12-factor config precedence.

---

**Documentation**
*   **Contributing Guide**: Added `CONTRIBUTING.md` to the root for new developers.
*   **Extension Guide**: Added `docs/how_to_add_resources.md` explaining how to add new resources to Relia's logic.

---

**Bug Fixes**
*   **Packaging**: Fixed an issue where `bundled_pricing.db` was excluded from the PyPI package, causing Offline Mode to fail in some environments.

---

**Linking & Canonicalization**
*   **Canonicals**: Added `<link rel="canonical">` tags to key documentation pages to prevent duplicate content issues.
*   **Internal Linking**: Added "Related Documentation" footers to all pages to improve crawlability and user navigation flow.

---

**Content & AEO**
*   **Comparisons**: Added `docs/comparisons.md` detailing the differences between Relia, Infracost, and CloudHealth for evaluators.
*   **PAA/FAQ**: Expanded FAQ with "People Also Ask" sections to target natural language queries (e.g., "How do I calculate EC2 cost from Terraform?").
*   **TL;DRs**: Added concise summaries to Quickstart guides for faster information retrieval by AI agents.

---

**Structured Data (AEO)**
*   **HowTo Schema**: Added structured data to `docs/quickstart.md` to help AI assistants extract installation steps.
*   **TechArticle Schema**: Added schema to `docs/architecture.md`.
*   **Organization Schema**: Added Identity schema to `docs/index.md` for Knowledge Graph optimizaton.

---

**Technical SEO**
*   **Frontmatter**: Added optimized `title`, `description`, and `keywords` YAML frontmatter to all documentation pages for better SERP visibility.
*   **H1 Structure**: Fixed header hierarchy in troubleshooting and philosophy documentation.
*   **Crawler Guidance**: Added `sitemap.xml` and `robots.txt` to `docs/` for search engine indexing.

---

**New Features**
*   **Active Advisor (Lambda)**: Now proactively suggests switching `x86_64` functions to `arm64` (Graviton2) for ~20% cost savings.
*   **Active Advisor (RDS)**:
    *   Recommends upgrading `gp2` storage to `gp3` for consistent performance and lower cost.
    *   Suggests **Aurora Serverless v2** for provisioned Aurora clusters to handle variable workloads efficiently.

---

**Tests**
*   **HTML Reporting**: Added dedicated tests (`tests/test_output_html.py`) to verify HTML report structure, version injection, and budget alerts using `BeautifulSoup` assertions logic (simulated).
*   **Regression Tests**: Expanded regression suite coverage.

---

**Infrastructure**
*   **CI/CD**: Upgraded GitHub Actions to latest versions (checkout@v4, setup-python@v5).
*   **Security**: Added `bandit` (SAST) and `pip-audit` (Dependency Check) to the standard test workflow.
*   **Release Safety**: Release workflow now runs full test suite before publishing to PyPI.

**Documentation**
*   **Quickstart**: Updated documentation for Active Advisor, Load Balancer rules, and HTML report generation.

---

**Improvements**
*   **Report Accuracy**: HTML Reports now correctly display the current package version instead of a hardcoded value.
*   **Load Balancer Pricing**: Updated `_match_lb` to strictly filter for `group="Load Balancer"` to ensure hourly pricing is returned (avoiding LCU confusion).
*   **CLI Robustness**: Added graceful error handling for file write operations (`--out`, `--markdown-file`) to prevent stack traces on permission errors.

---

**Fixes**
*   **Test Suite**: Updated `tests/test_check.py` to use `isolated_filesystem` to comply with new path traversal security checks introduced in v1.1.4.

---

**Security Hardening**
*   **Docker Integration**: Pinned Python base image to specific SHA256 digest for immutable builds.
*   **CLI Security**: Added strict path traversal checks (`Path.resolve().is_relative_to()`) for report outputs.
*   **Dependencies**: Added `pip-audit` to dev dependencies.

**Developer Experience**
*   **Pre-Commit**: Added `detect-secrets`, `commitizen`, and `poetry-check` hooks.

**CLI Features**
*   **Cache Management**: Added new `relia cache` command group.
    *   `relia cache status`: View pricing cache size and location.
    *   `relia cache clear`: Safely delete the local pricing database.

---

**Fix: CI Test Alignment**
*   Updated `tests/test_lambda_pricing.py` to correctly partial-mock the new API-driven pricing logic introduced in v1.1.2.

---

**Fix: Lambda Pricing Accuracy**
*   **API-Driven Lambda Requests**: Removed hardcoded $0.20/million constant for Lambda requests. Now fetches the exact real-time price from AWS API.

---

**Enhancement: Premium Reports & Expanded Advisor**

### ✨ Features
- **Premium HTML Reports**: Completely redesigned HTML output with:
  - **Topology Graph**: Embedded Mermaid.js visualization.
  - **Interactive Tables**: Client-side sorting and filtering.
  - **Stats Cards**: Revenue/Cost summary cards.
  - **Modern UI**: Inter font, polished CSS.
- **Enhanced Advisor**: Added support for **RDS Storage** optimization (gp2 -> gp3).

---

**Major Features: Active Advisor & Reports**

### ✨ Features
- **Active Advisor**: Proactively suggests cost optimizations:
  - **EBS**: Recommends upgrading from `gp2` to `gp3` (up to 20% savings).
  - **EC2**: Suggests newer generation instances (`t2` -> `t3`) and Graviton tips.
- **Shareable Reports**: Generate beautiful, self-contained HTML reports with `--format html --out report.html`. Perfect for sharing with finance or managers.

### 📝 Improvements
- **Output**: Enhanced CLI tables to display "Tips" when opportunities are found.

---

## v1.0.1 (2025-12-06)

## v1.0.0 (2025-12-06)
**Major Release: Production Ready**

### ✨ Features
- **Region Expansion**: Added support for 10+ new AWS regions including `us-east-2`, `eu-central-1`, `ap-southeast-1`.
- **Relia Init**: New command to generate `.relia.yaml` and usage overlays instantly.
- **Path Security**: Enhanced security with path resolution and traversal checks.
- **Improved Parsing**: Robust handling of single files and Terraform variables.

### 🛡️ Security
- **Docker**: Non-root user `relia` and pinned SHA digest.
- **Dependencies**: Added `bandit` security scanning.
- **CI/CD**: Strict linting and type checking enforced.

### 🐛 Fixes
- Fixed configuration key mismatch (`policy` -> `rules`).
- Fixed crashes on malformed variables (`_safe_int`).
- Fixed silent exception swallowing in parser.

---

## v0.3.0 (2025-12-06) "Pragmatic Improvements"
*   **Feature**: **NAT Gateway**: Added support for hourly cost and data transfer warnings (Bill Shock prevention).
*   **Feature**: **Lambda**: Added support for usage-based pricing (Duration/Requests).
*   **Feature**: **ALB/ELB**: Added support for Load Balancer hourly costs.
*   **Enhancement**: **Accuracy**: Auto-detects region from `provider "aws"` block and supports Multi-AZ RDS.
*   **Enhancement**: **Developer Experience**: Added `relia init`, `relia check --dry-run`, and better variable warnings.
*   **Fix**: **Foundation**: Robust bundled pricing DB seeding, structured logging, and verified `--verbose`.

## v0.2.2 (2025-12-05) "Expansion Pack"
*   **Feature**: **Resource Expansion**: Added support for **EBS** (`aws_ebs_volume`).
*   **Feature**: **Usage Overlay**: Introduced `.relia.usage.yaml` to support usage-based resources like **S3** (`aws_s3_bucket`) and Lambda.
*   **Feature**: **Complex Parsing**: Added support for estimating costs from Terraform Plan JSON (`terraform show -json`) with recursive module support.
*   **Improvement**: **Maintenance**: Documented bundled database update process.
*   **Docs**: Comprehensive documentation updates for all new features.

## v0.2.0 (2025-12-05) "Polish"
*   **Feature**: **Resource Expansion**: Added support for **RDS** (`aws_db_instance`).
*   **Feature**: **Dynamic Regions**: Added `--region` flag and `AWS_REGION` env var support.
*   **Improvement**: **Operational Visibility**: Replaced print statements with structured logging and added `--verbose` flag.
*   **Fix**: Resolved CI/CD build issues and test coverage reporting.

## v0.1.2 (2025-12-05)
*   **Feature**: Added JSON output support (`relia estimate --format json`) for easier pipeline integration.
*   **Feature**: Added official Pre-Commit hooks support (`.pre-commit-hooks.yaml`).
*   **Improvement**: Enhanced PyPI metadata for better discoverability.
*   **Fix**: Renamed package to `relia_oss` to resolve PyPI naming conflict.

## v0.1.0 (2025-12-05)
*   **Initial Release**: MVP Launch.
*   **Feature**: `relia estimate` with AWS EC2/RDS support.
*   **Feature**: `relia check` for budget guardrails.
*   **Feature**: `.relia.yaml` configuration.
*   **Feature**: Docker support.
