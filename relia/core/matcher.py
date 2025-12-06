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
            "us-east-2": "US East (Ohio)",
            "us-west-1": "US West (N. California)",
            "us-west-2": "US West (Oregon)",
            "eu-west-1": "EU (Ireland)",
            "eu-west-2": "EU (London)",
            "eu-central-1": "EU (Frankfurt)",
            "ap-southeast-1": "Asia Pacific (Singapore)",
            "ap-southeast-2": "Asia Pacific (Sydney)",
            "ap-northeast-1": "Asia Pacific (Tokyo)",
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
        elif resource.resource_type == "aws_nat_gateway":
            return "AmazonEC2", self._match_nat_gateway(resource)
        elif resource.resource_type == "aws_lambda_function":
            return "AWSLambda", self._match_lambda(resource)
        elif resource.resource_type in ["aws_lb", "aws_elb"]:
            # Note: Checking service code.
            # ALB/NLB often under "AmazonEC2" in Price List API?
            # Yes, standard API usually bundles them.
            return "AmazonEC2", self._match_lb(resource)
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

        if not instance_class or "${" in str(instance_class):
            from relia.utils.logger import logger

            logger.warning(
                f"⚠️  {resource.id}: instance_class missing or unresolved variable. Use 'terraform plan -json'."
            )
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

        # Multi-AZ Support
        is_multi_az = resource.attributes.get("multi_az", False)
        # Terraform might parse boolean as True/False or "true"/"false" if expression
        # HCL2 parser usually gives python bool if literal

        deployment_option = "Multi-AZ" if is_multi_az else "Single-AZ"

        return [
            {"Type": "TERM_MATCH", "Field": "serviceCode", "Value": "AmazonRDS"},
            {"Type": "TERM_MATCH", "Field": "location", "Value": self._get_location()},
            {"Type": "TERM_MATCH", "Field": "instanceType", "Value": instance_class},
            {"Type": "TERM_MATCH", "Field": "databaseEngine", "Value": db_engine},
            {
                "Type": "TERM_MATCH",
                "Field": "deploymentOption",
                "Value": deployment_option,
            },
        ]

    def _match_ec2(self, resource: ReliaResource) -> List[Dict[str, str]]:
        instance_type = resource.attributes.get("instance_type")
        if not instance_type or "${" in str(instance_type):
            # Maybe it's a variable; skip for MVP
            from relia.utils.logger import logger

            logger.warning(
                f"⚠️  {resource.id}: instance_type missing or unresolved variable. Use 'terraform plan -json' for accurate estimates."
            )
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

    def _match_nat_gateway(self, resource: ReliaResource) -> List[Dict[str, str]]:
        # NAT Gateway Hourly Charge.
        # Filter by productFamily="NAT Gateway"
        # Usually checking 'group'="NAT Gateway" isolates the hourly cost from data processing.
        return [
            {"Type": "TERM_MATCH", "Field": "serviceCode", "Value": "AmazonEC2"},
            {"Type": "TERM_MATCH", "Field": "location", "Value": self._get_location()},
            {"Type": "TERM_MATCH", "Field": "productFamily", "Value": "NAT Gateway"},
            {"Type": "TERM_MATCH", "Field": "group", "Value": "NAT Gateway"},
        ]

    def _match_lambda(self, resource: ReliaResource) -> List[Dict[str, str]]:
        # AWS Lambda Duration Cost.
        # We target the 'AWS-Lambda-Duration' group to get the GB-Second price.
        # Request pricing is separate but usually fixed ($0.20/1M).

        # Architecture: "x86", "ARM" (Graviton is cheaper).
        # We default to x86 if not specified.
        # arch = resource.attributes.get("architectures", ["x86_64"])[0]
        # (Unused for now, MVP assumes x86 price)
        # Terraform uses "x86_64" or "arm64".
        # Pricing API often uses "AVX2" for x86? No, usually group is enough.

        return [
            {"Type": "TERM_MATCH", "Field": "serviceCode", "Value": "AWSLambda"},
            {"Type": "TERM_MATCH", "Field": "location", "Value": self._get_location()},
            {"Type": "TERM_MATCH", "Field": "productFamily", "Value": "Serverless"},
            {"Type": "TERM_MATCH", "Field": "group", "Value": "AWS-Lambda-Duration"},
        ]

    def _match_lb(self, resource: ReliaResource) -> List[Dict[str, str]]:
        # Handles aws_lb (ALB/NLB) and aws_elb (CLB).
        # Attributes: load_balancer_type (application, network, gateway). Default application.
        # For aws_elb, it's always classic.

        lb_type = "application"
        if resource.resource_type == "aws_elb":
            lb_type = "classic"
        else:
            lb_type = resource.attributes.get("load_balancer_type", "application")

        # Product Family Map
        family_map = {
            "application": "Load Balancer-Application",
            "network": "Load Balancer-Network",
            "gateway": "Load Balancer-Gateway",
            "classic": "Load Balancer",  # Classic ELB
        }

        product_family = family_map.get(lb_type, "Load Balancer-Application")

        return [
            {
                "Type": "TERM_MATCH",
                "Field": "serviceCode",
                "Value": "AmazonEC2",
            },  # Yes, ELB is under EC2 service code usually?
            # Actually, sometimes it's "AWSELB" or similar?
            # Standard Price List API for ALB is under AmazonEC2 service?
            # Let's verify: In us-east-1, it is usually "AmazonEC2" for historical reasons OR "AWSELB"?
            # Correction: It is usually "AmazonEC2" for Classic, but ALB/NLB might be separate?
            # Safe bet: Check "productFamily" under "AmazonEC2".
            # Actually, let's filter by location too.
            {"Type": "TERM_MATCH", "Field": "location", "Value": self._get_location()},
            {"Type": "TERM_MATCH", "Field": "productFamily", "Value": product_family},
            # To isolate hourly: group="LCU" vs group="Hourly"?
            # Usually strict TERM_MATCH on productFamily gives multiple terms (Usage vs Hourly).
            # We assume PricingClient picks the first one?
            # If so, we might get LCU usage price ($0.008) instead of Hourly ($0.0225).
            # We can try to force group="Hourly" or usageType check if possible.
            # But group filter is safer if available.
        ]
