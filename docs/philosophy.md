# The Philosophy of Relia: Shift Left FinOps

Relia was built on three core beliefs about the future of cloud engineering.

## 1. Prevention > Optimization
The current state of "Cloud Cost Management" is reactive. Tools like AWS Cost Explorer, CloudHealth, and Vantage are essentially **autopsies**—they tell you exactly how you bled money *last month*.

We believe in **Shift Left FinOps**. Just as we shifted security (DevSecOps) and testing (CI/CD) to the left, we must shift cost awareness to the Pull Request.
*   **Old Way:** Deploy -> Bill Shock -> Fix.
*   **Relia Way:** Estimate -> Budget Check -> Deploy.

## 2. Developer Experience (DX) is King
Adoption is the biggest hurdle for governance tools. If a tool is slow, clunky, or requires a separate login, engineers will route around it.
*   **Speed:** Relia gives you an answer in milliseconds, not minutes.
*   **Aesthetics:** We use [Rich](https://github.com/Textualize/rich) to make cost data visually intuitive and readable in the terminal.
*   **Determinism:** Infrastructure-as-Code should have deterministic costs. same `.tf` = same `$`.

## 3. Sovereign Infrastructure
Your infrastructure logic is your intellectual property. You shouldn't have to upload your entire Terraform state to a third-party SaaS just to know what an EC2 instance costs.
*   **Local First:** Relia runs on *your* machine.
*   **No Secrets:** We analyze resource attributes (type, region), not sensitive data (passwords, keys).
*   **Open Source:** You can audit exactly how we calculate every cent.
