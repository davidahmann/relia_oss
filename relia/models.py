from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple, Type

import yaml  # type: ignore
from pydantic import BaseModel, Field, model_validator
from pydantic_settings import (
    BaseSettings,
    PydanticBaseSettingsSource,
    SettingsConfigDict,
)


class ReliaResource(BaseModel):
    id: str = Field(default_factory=lambda: "unknown")
    resource_type: str = Field(..., description="Terraform resource type")
    resource_name: str = Field(..., description="Terraform resource name")
    attributes: Dict[str, Any] = Field(default_factory=dict)
    file_path: Optional[str] = None
    suggestions: List[str] = []

    @model_validator(mode="after")
    def compute_id(self) -> "ReliaResource":
        self.id = f"{self.resource_type}.{self.resource_name}"
        return self


class YamlConfigSettingsSource(PydanticBaseSettingsSource):
    def get_field_value(self, field: Field, field_name: str) -> Tuple[Any, str, bool]:
        # Not used directly, we implement __call__
        return None, field_name, False

    def __call__(self) -> Dict[str, Any]:
        config_path = Path(".relia.yaml")
        if not config_path.exists():
            config_path = Path(".relia.yml")

        if not config_path.exists():
            return {}

        try:
            with open(config_path, "r") as f:
                return yaml.safe_load(f) or {}
        except Exception:
            return {}


class ReliaConfig(BaseSettings):
    budget: float = 0.0
    rules: Dict[str, float] = {}

    model_config = SettingsConfigDict(env_prefix="RELIA_")

    @classmethod
    def settings_customise_sources(
        cls,
        settings_cls: Type[BaseSettings],
        init_settings: PydanticBaseSettingsSource,
        env_settings: PydanticBaseSettingsSource,
        dotenv_settings: PydanticBaseSettingsSource,
        file_secret_settings: PydanticBaseSettingsSource,
    ) -> Tuple[PydanticBaseSettingsSource, ...]:
        return (
            init_settings,
            env_settings,
            YamlConfigSettingsSource(settings_cls),
            file_secret_settings,
        )
