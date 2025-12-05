from typing import List, Dict, Tuple
from relia.models import ReliaResource
from relia.core.parser import TerraformParser
from relia.core.pricing import PricingClient
from relia.core.matcher import ResourceMatcher
from relia.core.config import ConfigLoader
from relia.core.usage import UsageLoader
from relia.utils.logger import logger


class ReliaEngine:
    def __init__(self, config_path: str = ".relia.yaml", region: str = "us-east-1"):
        self.parser = TerraformParser()
        self.pricing = PricingClient()
        self.matcher = ResourceMatcher(region=region)
        self.config_loader = ConfigLoader()
        self.config = self.config_loader.load(config_path)
        self.usage_loader = UsageLoader()
        self.usage_loader.load()

    def run(self, path: str) -> Tuple[List[ReliaResource], Dict[str, float]]:
        # 0. Auto-detect region if not locked
        # If the user passed default "us-east-1", we try to detect better
        # If user passed "us-west-2", we stick to it (matcher has it)

        # We assume if matcher.region_name is "us-east-1", it might be a default
        # Ideally we'd have an explicit flag. For MVP, we trust the Parser scan if current is us-east-1.
        if self.matcher.region_name == "us-east-1":
            detected_region = self.parser.extract_provider_region(path)
            if detected_region:
                self.matcher.region_name = detected_region
                # Also log this change if verbose?
                # logger.info(f"Auto-detected region: {detected_region}")

        # 1. Parse
        if path.endswith(".json"):
            resources = self.parser.parse_plan_json(path)
        else:
            resources = self.parser.parse_directory(path)

        if not resources:
            return [], {}

        # 2. Price
        costs = {}
        for resource in resources:
            # Apply Usage Overlay
            resource.attributes = self.usage_loader.apply_usage(
                resource.id, resource.attributes
            )

            match_result = self.matcher.get_pricing_filters(resource)
            if match_result:
                service_code, filters = match_result
                unit_price = self.pricing.get_product_price(service_code, filters)

                if unit_price is not None:
                    # Calculate quantity based on resource type
                    if resource.resource_type in [
                        "aws_instance",
                        "aws_db_instance",
                        "aws_lb",
                        "aws_elb",
                    ]:
                        # Hourly -> Monthly
                        monthly_cost = unit_price * 730
                    elif resource.resource_type == "aws_nat_gateway":
                        # Hourly -> Monthly
                        # Note: This is fixed cost only. Data processing is extra.
                        monthly_cost = unit_price * 730
                        # Warn about data transfer
                        # Only warn if not silent? logger.info is fine.
                        # The prompt asked for specific warning.
                        # We only warn if the user hasn't provided usage maybe?
                        # For now, just a helpful info message.
                        logger.info(
                            f"⚠️  {resource.id}: Data transfer costs not estimated. See .relia.usage.yaml."
                        )
                    elif resource.resource_type == "aws_ebs_volume":
                        # GB-Mo -> Monthly (multiply by size)
                        size = int(resource.attributes.get("size", 8))  # Default
                        monthly_cost = unit_price * size
                    elif resource.resource_type == "aws_s3_bucket":
                        # GB-Mo -> Monthly
                        # Default to 0 if not in usage file, to avoid scary assumptions
                        storage_gb = int(resource.attributes.get("storage_gb", 0))
                        monthly_cost = unit_price * storage_gb
                    elif resource.resource_type == "aws_lambda_function":
                        # Lambda Cost = Duration Cost + Request Cost
                        # Unit Price here is per GB-Second (approx $0.0000166667)

                        requests = int(resource.attributes.get("monthly_requests", 0))
                        duration_ms = float(
                            resource.attributes.get("avg_duration_ms", 100)
                        )
                        memory_mb = int(resource.attributes.get("memory_size", 128))

                        # 1. Duration Cost
                        # GB-Seconds = requests * (duration/1000) * (memory/1024)
                        gb_seconds = (
                            requests * (duration_ms / 1000) * (memory_mb / 1024)
                        )
                        compute_cost = unit_price * gb_seconds

                        # 2. Request Cost (Hardcoded $0.20 per 1M for standard x86/ARM)
                        # Fetching this dynamic secondary price is complex for MVP
                        request_cost = (requests / 1_000_000) * 0.20

                        monthly_cost = compute_cost + request_cost
                    else:
                        monthly_cost = unit_price  # Default / Fallback

                    costs[resource.id] = monthly_cost

        return resources, costs

    def check_policies(
        self, resources: List[ReliaResource], costs: Dict[str, float]
    ) -> List[str]:
        violations = []

        # 1. Per-Resource Rules
        for r in resources:
            limit = self.config.rules.get(r.resource_type)
            if limit:
                cost = costs.get(r.id, 0.0)
                if cost > limit:
                    violations.append(
                        f"Resource '{r.id}' (${cost:,.2f}) exceeds limit of ${limit:,.2f}"
                    )

        return violations
