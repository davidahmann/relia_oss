from typing import List, Dict, Tuple
from relia.models import ReliaResource
from relia.core.parser import TerraformParser
from relia.core.pricing import PricingClient
from relia.core.matcher import ResourceMatcher
from relia.core.config import ConfigLoader


class ReliaEngine:
    def __init__(self, config_path: str = ".relia.yaml", region: str = "us-east-1"):
        self.parser = TerraformParser()
        self.pricing = PricingClient()
        self.matcher = ResourceMatcher(region=region)
        self.config_loader = ConfigLoader()
        self.config = self.config_loader.load(config_path)

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
            match_result = self.matcher.get_pricing_filters(resource)
            if match_result:
                service_code, filters = match_result
                price = self.pricing.get_product_price(service_code, filters)
                if price:
                    costs[resource.id] = price

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
