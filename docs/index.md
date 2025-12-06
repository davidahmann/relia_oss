---
title: "Relia - Free Terraform Cost Estimation Tool"
description: "Prevent AWS bill shock with Relia. Pre-deployment cost estimation for Terraform. Free, open-source FinOps tool for CI/CD."
keywords: terraform cost estimation, aws cost calculator, infrastructure cost, finops tools, cloud cost prevention
---

# Relia: Open Source Cloud Cost Prevention for Terraform

> **"The ESLint for Cloud Costs. Stop paying the infrastructure tax."**

Relia is an **open-source FinOps tool** that runs in your CI/CD pipeline. It acts as a **pre-deployment cost estimator** for Terraform, helping engineering teams prevent AWS bill shock by catching expensive infrastructure changes *before* they are merged.

Unlike traditional cloud cost management tools (like CloudHealth or Vantage) that report on *past* spending, Relia is a **proactive guardrail**. It parses your Infrastructure-as-Code (IaC), calculates the monthly cost impact using the AWS Price List API, and enforces budget policies in your Git workflow.

---

## ⚡ Quickstart: Estimate Terraform Costs

**How to check cloud costs in your terminal:**

1.  **Install Relia via Pip:**
    ```bash
    pip install relia
    ```

2.  **Run a Cost Estimate:**
    Navigate to your Terraform directory (`.tf` files) and run:
    ```bash
    relia estimate ./infra
    ```

    *Output:*
    ```text
    📊 Relia Cost Estimate
    +------------------+----------+------------+
    | Resource         | Type     | Cost/Month |
    +------------------+----------+------------+
    | aws_instance.web | t3.large |     $60.00 |
    +------------------+----------+------------+
    | Total            |          |     $60.00 |
    +------------------+----------+------------+
    ```

3.  **Enforce Budgets in CI:**
    Block Pull Requests that exceed \$100/month:
    ```bash
    relia check ./infra --budget 100
    ```

[Get Started with the Full Guide →](quickstart.md)

---

## 🚀 Key Features for FinOps Teams

### 1. Pre-Deployment Cost Estimation
Calculate the exact dollar value of your Terraform plan before applying it. Relia supports **EC2, RDS, S3, NAT Gateways, Lambda, and Load Balancers**.

### 2. CI/CD Budget Guardrails
Stop expensive mistakes automatically. Configure strict budget caps (e.g., "Total < $500") or resource-level limits (e.g., "No instance > $50") using a simple [`.relia.yaml`](quickstart.md#configuration) file.

### 3. Developer-Native Experience
Relia is designed for **DevOps engineers and SREs**. It lives in the terminal, works offline with a bundled pricing database, and integrates seamlessly with GitHub Actions. It respects your privacy—no plan data is ever sent to a SaaS.

---

## 📚 Documentation

- **[Quickstart Guide](quickstart.md):** Installation, Usage, and Docker instructions.
- **[Supported Resources](supported_resources.md):** Full list of supported AWS services (EC2, Lambda, etc.) and Usage Overlay guide.
- **[Architecture](architecture.md):** How the Parser, Matcher, and Pricing Engine work.
- **[Philosophy](philosophy.md):** Why "Prevention > Optimization" is the future of FinOps.
- **[Troubleshooting](troubleshooting.md):** Fixing AWS SSO errors and parsing issues.
- **[FAQ](faq.md):** Answers to common questions about accuracy, Azure/GCP support, and comparisons.
- **[Changelog](changelog.md):** Version history and release notes.
- **[Why We Built Relia](why-we-built-relia.md):** The story behind ending "Bill Shock".

---

## 🤝 Contributing

We love contributions! Please read our [CONTRIBUTING.md](https://github.com/davidahmann/relia_oss/blob/main/CONTRIBUTING.md) to get started.

<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "SoftwareApplication",
  "name": "Relia",
  "operatingSystem": "Linux, macOS, Windows",
  "applicationCategory": "DeveloperApplication",
  "description": "The ESLint for Cloud Costs. Prevent bill shock before it happens.",
  "offers": {
    "@type": "Offer",
    "price": "0",
    "priceCurrency": "USD"
  },
  "url": "https://davidahmann.github.io/relia_oss/"
}
</script>

<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "Organization",
  "name": "Relia",
  "url": "https://davidahmann.github.io/relia_oss/",
  "logo": "https://davidahmann.github.io/relia_oss/logo.png",
  "sameAs": [
    "https://github.com/davidahmann/relia_oss",
    "https://pypi.org/project/relia/"
  ]
}
</script>
