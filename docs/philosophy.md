# Philosophy

## 1. Prevention > Optimization

Most cloud cost tools (Vantage, CloudHealth, AWS Cost Explorer) operate on **lagging indicators**. They tell you that you spent too much money *last month*.

Relia operates on **leading indicators**. We integrate into the developer's workflow (CLI/CI) to catch expensive changes *before* they are deployed.

## 2. Developer Experience First

If a tool is annoying, developers won't use it.
*   **Fast**: Relia runs in milliseconds.
*   **Beautiful**: Output should be readable and visually appealing (`rich`).
*   **Deterministic**: Same input code = same cost estimate.

## 3. Privacy & Security

We believe your infrastructure code describes your competitive advantage.
*   **No SaaS**: Relia runs entirely on your machine or CI runner.
*   **No Secrets**: We never read or transmit your AWS credentials or `.tfvars` secrets.
*   **Offline First**: We bundle pricing data so you can work on an airplane.
