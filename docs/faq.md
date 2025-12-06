---
title: "Relia FAQ: Terraform Cost Accuracy, Offline Mode & Comparisons"
description: "Answers to common questions about Relia. Learn about pricing accuracy, offline mode, and how it compares to other FinOps tools."
keywords: relia faq, terraform cost accuracy, offline cloud cost
---

# Frequently Asked Questions (FAQ)

<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "FAQPage",
  "mainEntity": [{
    "@type": "Question",
    "name": "Does Relia read my Terraform state file?",
    "acceptedAnswer": {
      "@type": "Answer",
      "text": "No. Relia operates purely on your **Infrastructure-as-Code (IaC)**. It parses your local `.tf` files or the JSON output of `terraform show -json`. It does not require access to your remote state bucket (S3/GCS) or state locking, making it safer and faster to run in CI."
    }
  }, {
    "@type": "Question",
    "name": "Does Relia support Azure or Google Cloud (GCP)?",
    "acceptedAnswer": {
      "@type": "Answer",
      "text": "Not in v0.1.0 (MVP). Relia is currently laser-focused on **AWS** support, specifically for EC2 and RDS resources. Support for Google Cloud Platform (GCP) and Azure is on the [roadmap](https://github.com/davidahmann/relia_oss) for v0.2."
    }
  }, {
    "@type": "Question",
    "name": "How accurate are Relia's cost estimates?",
    "acceptedAnswer": {
      "@type": "Answer",
      "text": "Relia provides **high accuracy (>99%)** for fixed-rate compute resources like **EC2 Instances** and **RDS Databases** by mapping them directly to the AWS Price List API. For usage-based costs like Data Transfer or S3 Storage (GB/month), Relia provides baseline estimates or defaults, as these cannot be determined purely from static code analysis."
    }
  }, {
    "@type": "Question",
    "name": "Is there a SaaS version of Relia?",
    "acceptedAnswer": {
      "@type": "Answer",
      "text": "No. Relia is **100% open-source software** that you host yourself. It runs locally on your machine or within your CI/CD runners (GitHub Actions, GitLab CI). This ensures your infrastructure code and cost data never leave your environment."
    }
  }]
}
</script>

## Core Functionality

### Does Relia read my Terraform state file?
**No.** Relia works by static analysis of your `.tf` code or by parsing the plan output. This approach ("Shift Left") allows us to provide feedback *before* state is even modified.

### Does it support Azure/GCP?
Currently, **AWS only**. We prioritize depth over breadth. We want to solve AWS billing perfectly before moving to other clouds.

5.  **Multi-Region**: We support estimates across multiple regions in the same project.

## Usage & Configuration

### How do I estimate Lambda or S3 costs?
Since Terraform code doesn't tell us how many requests your Lambda receives or how many GBs are in your S3 bucket, you need to provide a **Usage Overlay**.
1. Run `relia init` to generate a `.relia.usage.yaml` file.
2. Define your assumptions (e.g., `storage_gb: 50`).
See [Supported Resources](supported_resources.md) for examples.

### Does Relia work offline?
**Yes.** Relia ships with a bundled SQLite database containing prices for the most common instance types (t3, m5, etc.) in US/EU regions.
*   If you are offline, Relia uses the bundled DB (or your local cache).
*   If you are online, Relia fetches the latest prices from the AWS API and caches them.

### Can I run checks without breaking my build?
**Yes.** Use the `--dry-run` flag:
```bash
relia check --dry-run
```
This will print policy violations (like "Budget Exceeded") but exit with code `0`, allowing your pipeline to continue. Use this for "Warning Mode".

## Accuracy & Pricing

### How accurate is the cost estimation?
*   **Compute (EC2/RDS):** Extremely accurate. We match `instance_type`, `region`, and `operating_system` against the official AWS API.
*   **Storage (EBS/S3):** Accurate for provisioned capacity (e.g., `gp3` size).
*   **Usage (Data xfer):** Estimates based on simple defaults, as code doesn't predict traffic.

### How does Relia handle AWS Pricing/SSO?
Relia uses a **3-Layer Pricing Strategy**:
1.  **Bundled DB:** Common prices (t3, m5) are shipped with the tool (works offline).
2.  **Local Cache:** We cache API responses for 7 days in `~/.relia/`.
3.  **Real-time API:** We fetch fresh pricing from AWS if your credentials allow it.

## Comparison

### How is this different from Infracost?
Relia fills a similar niche but focuses on **simplicity and privacy**.
*   **Relia:** minimal setup, privacy-first (no SaaS), focused on "Linting" experience.
*   **Infracost:** Enterprise-heavy, often pushes towards their SaaS platform for policies.

### How is this different from CloudHealth/Vantage?
Those are **Reporting Tools** (Post-Deployment). Relia is a **Prevention Tool** (Pre-Deployment). You use Relia to *prevent* the spike that Vantage reports on 30 days later.

## People Also Ask

### What is the best open source Terraform cost tool?
Relia is a free, privacy-first option. It estimates AWS costs from `.tf` files without sending data to any SaaS. It is ideal for teams who want control and simplicity.

### How do I calculate EC2 cost from Terraform?
Run `relia estimate .` in your Terraform directory. Relia parses `instance_type` and fetches the exact hourly rate from AWS to give you a monthly projection.
