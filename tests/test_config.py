import pytest
from typer.testing import CliRunner
from relia.cli import app
from relia.core.config import ConfigLoader
from relia.models import ReliaResource, ReliaConfig
from unittest.mock import patch

runner = CliRunner()


def test_config_loader(tmp_path):
    config_file = tmp_path / ".relia.yaml"
    config_file.write_text("budget: 500.0\nrules:\n  aws_instance: 100.0")

    loader = ConfigLoader()
    config = loader.load(str(config_file))

    assert config.budget == 500.0
    assert config.rules["aws_instance"] == 100.0


@pytest.fixture
def mock_engine_policy():
    with patch("relia.cli.ReliaEngine") as MockEngine:
        engine = MockEngine.return_value

        # Mock Config
        engine.config = ReliaConfig(budget=100.0, rules={"aws_instance": 10.0})

        # Mock Engine Run
        r1 = ReliaResource(
            resource_type="aws_instance",
            resource_name="web",
            attributes={},
            file_path="main.tf",
        )
        engine.run.return_value = (
            [r1],
            {"aws_instance.web": 20.0},
        )  # Costs 20, limit is 10
        engine.check_policies.return_value = [
            "Resource 'aws_instance.web' ($20.00) exceeds limit of $10.00"
        ]

        yield engine


def test_check_policy_violation(mock_engine_policy, tmp_path):
    # Should fail because of policy violation (20 > 10)
    result = runner.invoke(app, ["check", ".", "--config", "dummy.yaml"])

    assert result.exit_code == 1
    assert "Policy Violations" in result.stdout
    assert "exceeds limit of $10.00" in result.stdout


def test_check_budget_override(mock_engine_policy):
    # Config budget is 100, Usage is 20. Should pass budget check.
    # But policy check will fail it.
    # Let's mock check_policies to return empty to test budget logic specifically
    mock_engine_policy.check_policies.return_value = []

    # 1. Use config budget (100) -> Pass (20 < 100)
    result = runner.invoke(app, ["check", "."])
    assert result.exit_code == 0
    assert "Limit: $100.00" in result.stdout

    # 2. Override budget (10) -> Fail (20 > 10)
    result = runner.invoke(app, ["check", ".", "--budget", "10"])
    assert result.exit_code == 1
    assert "Limit: $10.00" in result.stdout
