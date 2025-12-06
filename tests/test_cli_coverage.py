from typer.testing import CliRunner
from relia.cli import app
from relia import __version__
import json
from unittest.mock import MagicMock, patch

runner = CliRunner()


def test_version_command():
    result = runner.invoke(app, ["version"])
    assert result.exit_code == 0
    assert __version__ in result.output


def test_estimate_json_format():
    # Mock engine to avoid full scan
    with patch("relia.cli.ReliaEngine") as MockEngine:
        mock_instance = MockEngine.return_value
        # Use ReliaResource mock structure
        mock_res = MagicMock()
        mock_res.id = "aws_instance.test"
        mock_res.resource_type = "aws_instance"
        mock_res.resource_name = "test"  # Needed for JSON output
        mock_res.attributes = {}  # Needed to be serializable dict

        mock_instance.run.return_value = ([mock_res], {"aws_instance.test": 10.0})

        result = runner.invoke(app, ["estimate", "dummy_path", "--format", "json"])

        assert result.exit_code == 0
        data = json.loads(result.output)
        assert data["total_cost"] == 10.0
        assert data["resources"][0]["cost"] == 10.0


def test_check_dry_run():
    with patch("relia.cli.ReliaEngine") as MockEngine:
        mock_instance = MockEngine.return_value
        # MUST return resources, otherwise CLI exits early
        mock_instance.run.return_value = ([MagicMock()], {})
        mock_instance.check_policies.return_value = ["Violation 1"]
        # Ensure budget comparison works (Mock > 0 is weird)
        mock_instance.config.budget = 0.0

        # Without dry-run -> Exit 1
        result = runner.invoke(app, ["check", "dummy_path"])
        assert result.exit_code == 1
        assert "Violation 1" in result.output

        # With dry-run -> Exit 0
        result_dry = runner.invoke(app, ["check", "dummy_path", "--dry-run"])
        assert result_dry.exit_code == 0
        assert "Violation 1" in result_dry.output
        assert "dry-run is enabled" in result_dry.output


def test_invalid_path():
    # Currently Relia just returns "No resources found" for bad paths, which is exit 0.
    # This test verifies that behavior.
    result = runner.invoke(app, ["estimate", "/non/existent/path"])
    assert result.exit_code == 0
    assert "No resources found" in result.output
