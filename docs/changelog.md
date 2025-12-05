# Changelog

## v0.2.1 (2025-12-05) "Expansion Pack"
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
