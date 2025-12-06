from pydantic import BaseModel, Field, model_validator
from typing import Dict, Any


class ReliaResource(BaseModel):
    id: str = Field(default_factory=lambda: "unknown")
    resource_type: str = Field(..., description="Terraform resource type")
    resource_name: str = Field(..., description="Terraform resource name")
    attributes: Dict[str, Any] = Field(default_factory=dict)
    file_path: str = Field(..., description="Source file path")

    @model_validator(mode="after")
    def compute_id(self) -> "ReliaResource":
        self.id = f"{self.resource_type}.{self.resource_name}"
        return self


class ReliaConfig(BaseModel):
    budget: float = 0.0
    rules: Dict[str, float] = {}  # e.g. {"aws_instance": 100.0} (max price per unit)
