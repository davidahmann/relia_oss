from typer.testing import CliRunner
from relia.cli import app
from relia import __version__

runner = CliRunner()


def test_estimate():
    result = runner.invoke(app, ["estimate"])
    assert result.exit_code == 0
    assert "Relia Cost Estimate" in result.stdout


def test_version():
    result = runner.invoke(app, ["--version"])
    assert result.exit_code == 0
    assert f"Relia v{__version__}" in result.stdout
