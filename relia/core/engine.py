from typing import List, Dict, Tuple
from relia.models import ReliaResource
from relia.core.parser import TerraformParser
from relia.core.pricing import PricingClient
from relia.core.matcher import ResourceMatcher
from relia.core.config import ConfigLoader
from relia.core.usage import UsageLoader


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
                    if resource.resource_type in ["aws_instance", "aws_db_instance"]:
                        # Hourly -> Monthly
                        monthly_cost = unit_price * 730
                    elif resource.resource_type == "aws_ebs_volume":
                        # GB-Mo -> Monthly (multiply by size)
                        size = int(resource.attributes.get("size", 8))  # Default
                        monthly_cost = unit_price * size
                    elif resource.resource_type == "aws_s3_bucket":
                        # GB-Mo -> Monthly
                        # Default to 0 if not in usage file, to avoid scary assumptions
                        storage_gb = int(resource.attributes.get("storage_gb", 0))
                        monthly_cost = unit_price * storage_gb
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
