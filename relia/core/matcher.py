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
        elif resource.resource_type == "aws_ebs_volume":
            return "AmazonEC2", self._match_ebs(resource)
        elif resource.resource_type == "aws_s3_bucket":
            return "AmazonS3", self._match_s3(resource)
        return None

    def _match_s3(self, resource: ReliaResource) -> List[Dict[str, str]]:
        # S3 Pricing is complex (Storage, Requests, Transfer).
        # We rely on Usage Overlay for: storage_gb, monthly_requests.
        # Fallback defaults: 1GB, 1000 requests.

        # We only price STORAGE for MVP to keep it simple.
        # "Timestorage-ByteHrs"

        return [
            {"Type": "TERM_MATCH", "Field": "serviceCode", "Value": "AmazonS3"},
            {"Type": "TERM_MATCH", "Field": "location", "Value": self._get_location()},
            {"Type": "TERM_MATCH", "Field": "productFamily", "Value": "Storage"},
            {"Type": "TERM_MATCH", "Field": "storageClass", "Value": "General Purpose"},
            {"Type": "TERM_MATCH", "Field": "volumeType", "Value": "Standard"},
        ]

    def _match_ebs(self, resource: ReliaResource) -> List[Dict[str, str]]:
        # Attributes: type (gp2, gp3, io1, io2, sc1, st1, standard), size (GB), iops (for io*)
        volume_type = resource.attributes.get("type", "gp2")
        # For MVP we default to bundling size in the usage or similar,
        # but the price is usually per GB-Month.
        # Pricing API VolumeType Map:
        # gp2 -> General Purpose
        # gp3 -> General Purpose gp3
        # io1 -> Provisioned IOPS
        # standard -> Magnetic

        type_map = {
            "gp2": "General Purpose",
            "gp3": "General Purpose gp3",
            "io1": "Provisioned IOPS",
            "io2": "System Operation",  # Approximate mapping, io2 is tricky in api
            "st1": "Throughput Optimized HDD",
            "sc1": "Cold HDD",
            "standard": "Magnetic",
        }

        api_type = type_map.get(volume_type, "General Purpose")

        # Note: EBS pricing is "EBS:VolumeUsage..." usually.
        # This is strictly for the STORAGE cost per GB-Month.

        return [
            {"Type": "TERM_MATCH", "Field": "serviceCode", "Value": "AmazonEC2"},
            {"Type": "TERM_MATCH", "Field": "location", "Value": self._get_location()},
            {"Type": "TERM_MATCH", "Field": "productFamily", "Value": "Storage"},
            {"Type": "TERM_MATCH", "Field": "volumeType", "Value": api_type},
        ]

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
