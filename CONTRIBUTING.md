# Contributing to Relia

Thank you for your interest in contributing to Relia! We are an open-source project dedicated to preventing cloud bill shock for engineering teams.

## 🛠️ Development Setup

Relia uses standard Python tooling managed by a **Makefile**.

1.  **Fork and Clone**
    ```bash
    git clone https://github.com/YOUR_USERNAME/relia_oss.git
    cd relia_oss
    ```

2.  **Setup Environment**
    Run the setup command to install `poetry` and all dependencies:
    ```bash
    make setup
    ```

3.  **Verify Install**
    ```bash
    make test
    ```

## 🏗️ Project Structure

*   `relia/`: Source code.
*   `relia/core/bundled_pricing.db`: SQLite database for offline caching.
*   `tests/`: Pytest suite.
*   `.github/`: Workflows and Issue Templates.

## 🧪 Testing & Configuration

### Running Tests
We require >85% code coverage.
```bash
make test
```

### Configuration for Testing
You can control Relia's behavior during tests or local runs using **Environment Variables**:

*   `RELIA_BUDGET`: Override budget limit (e.g. `500`).
*   `RELIA_CONFIG_PATH`: Point to a specific config file (default: `.relia.yaml`).

### Offline vs Online
Relia includes a `bundled_pricing.db` for offline estimates. If you change pricing logic, verify it works in offline mode (disconnect internet or unset AWS creds).

## 🛡️ Quality Standards

Before submitting a Pull Request, verify your code meets our standards:

```bash
make lint           # Check style (ruff) and types (mypy)
make check-security # Run security audit (bandit, pip-audit)
make format         # Auto-format code
```

## 📝 Pull Request Process

1.  Create a feature branch (`git checkout -b feat/my-feature`).
2.  Commit your changes following [Conventional Commits](https://www.conventionalcommits.org/).
3.  Push and open a Pull Request.
4.  Fill out the **Pull Request Template** checklist.

## 🤝 Code of Conduct

Please be respectful. We adhere to the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md).
