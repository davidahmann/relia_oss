# Contributing to Relia

Thank you for your interest in contributing to Relia! We are an open-source project dedicated to preventing cloud bill shock for engineering teams.

## 🛠️ Development Setup

Relia uses **Poetry** for dependency management and **Pre-commit** for code quality.

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/davidahmann/relia_oss.git
    cd relia_oss
    ```

2.  **Install dependencies:**
    ```bash
    pip install poetry
    poetry install
    ```

3.  **Install pre-commit hooks:**
    ```bash
    poetry run pre-commit install
    ```
    This ensures all code is linted (Ruff) and checked for secrets *before* you commit.

## 🧪 Running Tests

We aim for high test coverage (>85%). Please run the full suite before submitting a PR.

```bash
poetry run pytest
```

To run a specific test file:
```bash
poetry run pytest tests/test_matcher_unit.py
```

## 📐 adding New Resources

Want to add support for a new AWS resource (e.g., DynamoDB, Redshift)?
Please read our **[Guide to Adding Resources](docs/how_to_add_resources.md)**.

## 📝 Pull Request Process

1.  Create a feature branch (`git checkout -b feat/my-feature`).
2.  Commit your changes (`git commit -m "feat: Add DynamoDB support"`).
    *   We follow [Conventional Commits](https://www.conventionalcommits.org/).
3.  Push to the branch and open a Pull Request.
4.  Ensure all CI checks pass.

## 🤝 Code of Conduct

Please be respectful and kind. We welcome contributors of all skill levels.
