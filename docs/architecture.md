---
title: "Relia Architecture: How Terraform Cost Estimation Works"
description: "Deep dive into Relia's architecture: Parser, Matcher, and Pricing Engine. Understand how costs are calculated securely."
keywords: relia architecture, cost estimation logic, pricing engine, terraform parser
---

# Architecture Overview
Relia follows a linear pipeline architecture designed for speed and determinism.

```mermaid
flowchart LR
    A[Input Path] --> B[TerraformParser]
    B --> C[ReliaResources]
    C --> D[ReliaEngine]
    D --> E[PricingClient]
    E --> F[Cost Data]
    F --> G[CLI Output]
```

## 1. The Parser (`relia.core.parser`)
*   **Responsibility**: Converts Infrastructure-as-Code into standardized `ReliaResource` objects.
*   **HCL Support**: Uses `python-hcl2` to parse static `.tf` files.
*   **JSON Support**: Reads `terraform show -json` output for 100% accurate plan data.
*   **Output**: List of `ReliaResource(type, name, attributes)`.

## 2. The Matcher (`relia.core.matcher`)
*   **Responsibility**: Maps Terraform attributes to AWS Pricing API filters.
*   **Logic**:
    *   `aws_instance` -> `ServiceCode: AmazonEC2`
    *   `instance_type: t3.large` -> `instanceType: t3.large`
    *   `region: us-east-1` -> `location: US East (N. Virginia)`

## 3. The Pricing Client (`relia.core.pricing`)
*   **Responsibility**: Fetches monthly costs in USD.
*   **Layer 1 (Bundled DB)**: `bundled_pricing.db` (SQLite) ships with the package for instant lookup of common resources (e.g. t3/m5 instances).
*   **Layer 2 (Local Cache)**: `pricing_cache.db` stores API results for 7 days to minimize latency and API calls.
*   **Layer 3 (AWS API)**: Hits the AWS Price List API via `boto3` as a fallback.

## 4. The Engine (`relia.core.engine`)
*   **Responsibility**: Orchestrates the flow.
*   **Policy Check**: Loads `.relia.yaml` and verifies:
    1.  Total Budget Cap.
    2.  Per-resource cost limits.

## 5. The CLI (`relia.cli`)
*   **Built with**: `Typer` (commands) and `Rich` (tables, trees).
*   **Commands**:
    *   `estimate`: Shows cost table, topology tree (`--topology`), and diffs (`--diff`).
    *   `check`: Enforces budget for CI/CD integration.

<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "TechArticle",
  "headline": "Relia Architecture: How Terraform Cost Estimation Works",
  "description": "Technical deep-dive into Relia's parser, matcher, and pricing engine for estimating AWS infrastructure costs.",
  "author": {"@type": "Organization", "name": "Relia Team"},
  "keywords": "terraform parser, aws pricing api, cost estimation architecture"
}
</script>

---
## Related Documentation
- [Quickstart Guide](quickstart.md) - Install and run your first estimate
- [Philosophy](philosophy.md) - Our "Prevention > Optimization" approach
- [FAQ](faq.md) - Common questions and troubleshooting
