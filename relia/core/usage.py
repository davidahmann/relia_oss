import yaml  # type: ignore
from pathlib import Path
from typing import Dict, Any
from relia.utils.logger import logger


class UsageLoader:
    """
    Loads usage assumptions from .relia.usage.yaml
    """

    def __init__(self, path: str = ".relia.usage.yaml"):
        self.path = Path(path)
        self.usage_data: Dict[str, Any] = {}

    def load(self) -> Dict[str, Any]:
        if not self.path.exists():
            return {}

        try:
            with open(self.path, "r") as f:
                data = yaml.safe_load(f) or {}
                self.usage_data = data.get("usage", {})
                return self.usage_data
        except Exception as e:
            logger.warning(f"Warning: Failed to load usage file {self.path}: {e}")
            return {}

    def apply_usage(
        self, resource_name: str, attributes: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Merges usage data into resource attributes.
        Usage data overrides existing attributes if keys collide.
        """
        # Match by name or address
        overrides = self.usage_data.get(resource_name, {})
        if overrides:
            # Return a new dict with usage merged in
            return {**attributes, **overrides}
        return attributes
