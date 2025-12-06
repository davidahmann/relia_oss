from typing import List, Dict, Tuple
from relia.models import ReliaResource
from relia.core.parser import TerraformParser
from relia.core.pricing import PricingClient
from relia.core.matcher import ResourceMatcher
from relia.core.config import ConfigLoader
from relia.core.usage import UsageLoader
from relia.utils.logger import logger
from relia.core.constants import (
    HOURS_PER_MONTH,
    DEFAULT_EBS_SIZE_GB,
    DEFAULT_S3_STORAGE_GB,
    DEFAULT_LAMBDA_REQUESTS,
    DEFAULT_LAMBDA_DURATION_MS,
    DEFAULT_LAMBDA_MEMORY_MB,
    LAMBDA_REQUEST_PRICE_PER_MILLION,
)


def _safe_int(value, default=0):
    try:
        return int(value)
    except (ValueError, TypeError):
        return default


def _safe_float(value, default=0.0):
    try:
        return float(value)
    except (ValueError, TypeError):
        return default


class ReliaEngine:
    def __init__(self, config_path: str = ".relia.yaml", region: str = "us-east-1"):
        self.parser = TerraformParser()
        self.pricing = PricingClient()
        self.matcher = ResourceMatcher(region=region)
        self.config_loader = ConfigLoader()
        self.config = self.config_loader.load(config_path)
        self.usage_loader = UsageLoader()
        self.usage_loader.load()

    def _price_ec2(self, unit_price: float, resource: ReliaResource) -> float:
        return unit_price * HOURS_PER_MONTH

    def _price_nat(self, unit_price: float, resource: ReliaResource) -> float:
        logger.info(
            f"⚠️  {resource.id}: Data transfer costs not estimated. See .relia.usage.yaml."
        )
        return unit_price * HOURS_PER_MONTH

    def _price_ebs(self, unit_price: float, resource: ReliaResource) -> float:
        size = _safe_int(resource.attributes.get("size"), DEFAULT_EBS_SIZE_GB)
        return unit_price * size

    def _price_s3(self, unit_price: float, resource: ReliaResource) -> float:
        storage_gb = _safe_int(
            resource.attributes.get("storage_gb"), DEFAULT_S3_STORAGE_GB
        )
        return unit_price * storage_gb

    def _price_lambda(self, unit_price: float, resource: ReliaResource) -> float:
        requests = _safe_int(
            resource.attributes.get("monthly_requests"), DEFAULT_LAMBDA_REQUESTS
        )
        duration_ms = _safe_float(
            resource.attributes.get("avg_duration_ms"), DEFAULT_LAMBDA_DURATION_MS
        )
        memory_mb = _safe_int(
            resource.attributes.get("memory_size"), DEFAULT_LAMBDA_MEMORY_MB
        )

        gb_seconds = requests * (duration_ms / 1000) * (memory_mb / 1024)
        compute_cost = unit_price * gb_seconds
        request_cost = (requests / 1_000_000) * LAMBDA_REQUEST_PRICE_PER_MILLION
        return compute_cost + request_cost

    def _get_pricing_strategy(self, resource_type: str):
        strategies = {
            "aws_instance": self._price_ec2,
            "aws_db_instance": self._price_ec2,
            "aws_lb": self._price_ec2,
            "aws_elb": self._price_ec2,
            "aws_nat_gateway": self._price_nat,
            "aws_ebs_volume": self._price_ebs,
            "aws_s3_bucket": self._price_s3,
            "aws_lambda_function": self._price_lambda,
        }
        return strategies.get(resource_type)

    def run(self, path: str) -> Tuple[List[ReliaResource], Dict[str, float]]:
        # ... (rest of method until pricing loop)
        if self.matcher.region_name == "us-east-1":
            detected_region = self.parser.extract_provider_region(path)
            if detected_region:
                self.matcher.region_name = detected_region

        if path.endswith(".json"):
            resources = self.parser.parse_plan_json(path)
        else:
            resources = self.parser.parse_directory(path)

        if not resources:
            return [], {}

        costs = {}
        for resource in resources:
            resource.attributes = self.usage_loader.apply_usage(
                resource.id, resource.attributes
            )

            match_result = self.matcher.get_pricing_filters(resource)
            if match_result:
                service_code, filters = match_result
                unit_price = self.pricing.get_product_price(service_code, filters)

                if unit_price is not None:
                    strategy = self._get_pricing_strategy(resource.resource_type)
                    if strategy:
                        monthly_cost = strategy(unit_price, resource)
                    else:
                        monthly_cost = unit_price

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
