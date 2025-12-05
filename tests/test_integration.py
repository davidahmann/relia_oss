import pytest
from unittest.mock import MagicMock, patch, mock_open
from relia.core.engine import ReliaEngine


def test_full_pipeline_with_sample_tf():
    fixture_path = "tests/fixtures/integration"

    # Mock Pricing
    mock_pricing = MagicMock()

    # Side effects for different calls
    def get_price(service, filters):
        # NAT Gateway
        if service == "AmazonEC2" and any(f["Value"] == "NAT Gateway" for f in filters):
            return 0.045
        # EC2 t3.micro
        if service == "AmazonEC2" and any(f["Value"] == "t3.micro" for f in filters):
            return 0.0104
        # Lambda (Duration)
        if service == "AWSLambda":
            return 0.0000166667
        return 0.0

    mock_pricing.get_product_price.side_effect = get_price

    # Mock Usage for Lambda
    usage_overlay = """
usage:
  aws_lambda_function.api:
    monthly_requests: 1000000
    avg_duration_ms: 500
"""

    with patch("builtins.open", mock_open(read_data=usage_overlay)):
        # We need builtins.open to handle both config (write) and fixture (read)?
        # Actually verify engine reads fixture using hcl2.load which usually does explicit open.
        # Patching builtins.open might interfere with hcl2 reading the real file.
        # Better: Don't mock open for the TF file reading, only for the Usage file loading.
        # But usage file loading happens in engine initialization or run.
        # Let's mock UsageLoader specifically.
        pass

    with patch("relia.core.usage.UsageLoader.load") as mock_load:
        # Manually set usage
        mock_load.return_value = None  # It sets self.usage internally? No it sets self.usage on loader instance

        with patch("relia.core.engine.PricingClient", return_value=mock_pricing):
            engine = ReliaEngine()

            # Inject Usage Mock
            engine.usage_loader.usage = {
                "aws_lambda_function.api": {
                    "monthly_requests": 1000000,
                    "avg_duration_ms": 500,
                }
            }

            # Run on fixture
            resources, costs = engine.run(fixture_path)

            # 1. EC2: 0.0104 * 730 = 7.592
            # 2. NAT: 0.045 * 730 = 32.85
            # 3. Lambda:
            #    Compute: 1M * 0.5s * (128/1024) * 0.0000166667 = 500k * 0.125 * ... = 62500 * 0.0000166667 = 1.0416
            #    Request: 0.20
            #    Total: 1.2416

            assert "aws_instance.web" in costs
            assert "aws_nat_gateway.main" in costs

            assert (
                engine.matcher.region_name == "us-west-2"
            )  # Auto-detected from fixture

            total = sum(costs.values())
            expected = 7.592 + 32.85 + 1.2416
            assert pytest.approx(total, 0.1) == expected
