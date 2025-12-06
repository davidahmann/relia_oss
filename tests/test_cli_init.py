import yaml  # type: ignore
from typer.testing import CliRunner
from relia.cli import app
from pathlib import Path

runner = CliRunner()


def test_init_command_creates_valid_config():
    with runner.isolated_filesystem():
        result = runner.invoke(app, ["init"])
        assert result.exit_code == 0
        assert "Created .relia.yaml" in result.output

        config_path = Path(".relia.yaml")
        assert config_path.exists()

        with open(config_path) as f:
            data = yaml.safe_load(f)

        assert "budget" in data
        assert "rules" in data  # This was the bug (policy vs rules)
        assert data["rules"].get("aws_instance") == 20.0


def test_init_skips_existing():
    with runner.isolated_filesystem():
        with open(".relia.yaml", "w") as f:
            f.write("existing: true")

        result = runner.invoke(app, ["init"])
        assert "already exists" in result.output

        with open(".relia.yaml") as f:
            content = f.read()
        assert "existing: true" in content
