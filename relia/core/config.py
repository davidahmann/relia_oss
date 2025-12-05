import yaml  # type: ignore
from pathlib import Path
from relia.models import ReliaConfig


class ConfigLoader:
    def load(self, path: str = ".relia.yaml") -> ReliaConfig:
        config_path = Path(path)
        if not config_path.exists():
            # Try .yml ending
            config_path = Path(path.replace(".yaml", ".yml"))

        if not config_path.exists():
            return ReliaConfig()

        try:
            with open(config_path, "r") as f:
                data = yaml.safe_load(f) or {}

            return ReliaConfig(
                budget=data.get("budget", 0.0), rules=data.get("rules", {})
            )
        except Exception as e:
            print(f"⚠️  Error loading config: {e}")
            return ReliaConfig()
