import pytest
from typer.testing import CliRunner
from relia.cli import app
from unittest.mock import patch
from relia.models import ReliaResource

runner = CliRunner()


@pytest.fixture
def mock_engine():
    with patch("relia.cli.ReliaEngine") as MockEngine:
        engine_instance = MockEngine.return_value

        # Mock resources
        r1 = ReliaResource(
            resource_type="aws_instance",
            resource_name="web",
            attributes={"instance_type": "t3.large"},
            file_path="main.tf",
        )

        # Mock run return
        # Case 1: $60 cost
        # Case 1: $60 cost
        engine_instance.run.return_value = ([r1], {"aws_instance.web": 60.0})
        engine_instance.check_policies.return_value = []

        yield engine_instance


def test_check_within_budget(mock_engine):
    result = runner.invoke(app, ["check", ".", "--budget", "100"])
    assert result.exit_code == 0
    assert "Within budget" in result.stdout
    assert "Total: $60.00" in result.stdout


def test_check_exceeds_budget(mock_engine):
    result = runner.invoke(app, ["check", ".", "--budget", "50"])
    assert result.exit_code == 1
    assert "Budget exceeded" in result.stdout
    assert "Limit: $50.00" in result.stdout


def test_check_markdown_report(mock_engine, tmp_path):
    report_file = tmp_path / "report.md"
    result = runner.invoke(
        app, ["check", ".", "--budget", "100", "--markdown-file", str(report_file)]
    )

    assert result.exit_code == 0
    assert report_file.exists()
    content = report_file.read_text()
    assert "# ✅ Relia Cost Report" in content
    assert "Total Estimated Cost**: `$60.00/mo`" in content
