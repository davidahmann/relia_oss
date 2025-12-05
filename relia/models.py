from pydantic import BaseModel, Field
from typing import Dict, Any


class ReliaResource(BaseModel):
    id: str = Field(default_factory=lambda: "unknown")
    resource_type: str = Field(..., description="Terraform resource type")
    resource_name: str = Field(..., description="Terraform resource name")
    attributes: Dict[str, Any] = Field(default_factory=dict)
    file_path: str = Field(..., description="Source file path")

    def __init__(self, **data):
        super().__init__(**data)
        self.id = f"{self.resource_type}.{self.resource_name}"


class ReliaConfig(BaseModel):
    budget: float = 0.0
    rules: Dict[str, float] = {}  # e.g. {"aws_instance": 100.0} (max price per unit)
