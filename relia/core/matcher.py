from typing import List, Dict, Optional
from relia.models import ReliaResource


class ResourceMatcher:
    """
    Maps ReliaResource to AWS Pricing API filters.
    """

    def __init__(self, region: str = "us-east-1"):
        self.region_name = region
        self.location_map = {
            "us-east-1": "US East (N. Virginia)",
            "us-west-2": "US West (Oregon)",
            "eu-west-1": "EU (Ireland)",
            # Add more mappings as needed or use a robust lookup lib
        }

    def _get_location(self) -> str:
        return self.location_map.get(self.region_name, "US East (N. Virginia)")

    def get_pricing_filters(
        self, resource: ReliaResource
    ) -> Optional[tuple[str, List[Dict[str, str]]]]:
        if resource.resource_type == "aws_instance":
            return "AmazonEC2", self._match_ec2(resource)
        elif resource.resource_type == "aws_db_instance":
            return "AmazonRDS", self._match_rds(resource)
        return None

    def _match_rds(self, resource: ReliaResource) -> List[Dict[str, str]]:
        # Example: db.t3.micro
        instance_class = resource.attributes.get(
            "instance_class"
        ) or resource.attributes.get("class")
        engine = resource.attributes.get("engine", "mysql")  # Default guess

        if not instance_class:
            return []

        # Map engine to "Database Engine"
        # simplistic mapping
        engine_map = {
            "mysql": "MySQL",
            "postgres": "PostgreSQL",
            "mariadb": "MariaDB",
            "oracle": "Oracle",
            "sqlserver": "SQL Server",
        }
        db_engine = engine_map.get(engine, "MySQL")

        return [
            {"Type": "TERM_MATCH", "Field": "serviceCode", "Value": "AmazonRDS"},
            {"Type": "TERM_MATCH", "Field": "location", "Value": self._get_location()},
            {"Type": "TERM_MATCH", "Field": "instanceType", "Value": instance_class},
            {"Type": "TERM_MATCH", "Field": "databaseEngine", "Value": db_engine},
            {
                "Type": "TERM_MATCH",
                "Field": "deploymentOption",
                "Value": "Single-AZ",
            },  # Default assumption
        ]

    def _match_ec2(self, resource: ReliaResource) -> List[Dict[str, str]]:
        instance_type = resource.attributes.get("instance_type")
        if not instance_type:
            # Maybe it's a variable; skip for MVP
            return []

        return [
            {"Type": "TERM_MATCH", "Field": "serviceCode", "Value": "AmazonEC2"},
            {
                "Type": "TERM_MATCH",
                "Field": "location",
                "Value": self._get_location(),
            },
            {"Type": "TERM_MATCH", "Field": "instanceType", "Value": instance_type},
            {"Type": "TERM_MATCH", "Field": "operatingSystem", "Value": "Linux"},
            {"Type": "TERM_MATCH", "Field": "tenancy", "Value": "Shared"},
            {"Type": "TERM_MATCH", "Field": "preInstalledSw", "Value": "NA"},
            {"Type": "TERM_MATCH", "Field": "capacitystatus", "Value": "Used"},
        ]
