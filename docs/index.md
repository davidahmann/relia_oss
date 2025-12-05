# Relia: The ESLint for Cloud Costs

> **"Stop paying the infrastructure tax. Start shipping."**

Relia is an open-source tool that prevents cloud waste before it ships. It acts as a **pre-commit hook for infrastructure**, estimating costs and enforcing budgets directly in your CI/CD pipeline.

---

## ⚡ Quickstart

**1. Install**
```bash
pip install relia
```

**2. Check Costs**
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

**3. Enforce Budget in CI**
```bash
relia check ./infra --budget 100
```

[Get Started Now →](quickstart.md)

---

## 🚀 Why Relia?

### 1. Prevention > Optimization
Existing tools (CloudHealth, Vantage) tell you about waste *after* you get the bill. Relia catches it *before* you merge the Pull Request.

### 2. Developer Native
No dashboards to log into. No "approval workflows". Relia lives in your terminal and your Git history. It feels like `black` or `ruff`, but for money.

### 3. Privacy First
Relia runs entirely in your environment (Local or CI runner). It never sends your Terraform code or credentials to a SaaS cloud.

### 4. Offline Ready
Relia ships with a bundled pricing database. You can get accurate estimates on an airplane without needing active AWS credentials.

---

## 📚 Documentation

- **[Quickstart](quickstart.md):** Install and run in 2 minutes.
- **[Architecture](architecture.md):** How the Parser, Matcher, and Pricing Engine work.
- **[Philosophy](philosophy.md):** Why we prioritized prevention over reporting.
- **[Troubleshooting](troubleshooting.md):** Common issues (AWS creds, parsing errors).
- **[FAQ](faq.md):** Supported clouds, accuracy, and enterprise features.
- **[Why We Built Relia](why-we-built-relia.md):** The story behind the tool.

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
