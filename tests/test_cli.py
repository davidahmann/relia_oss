from typer.testing import CliRunner
from relia.cli import app

runner = CliRunner()


def test_estimate():
    result = runner.invoke(app, ["estimate"])
    assert result.exit_code == 0
    assert "Relia Cost Estimate" in result.stdout
