from unittest.mock import MagicMock, patch
from relia.core.engine import ReliaEngine
from relia.models import ReliaResource


def test_nat_gateway_pricing():
    # Mock resource
    nat = ReliaResource(
        resource_type="aws_nat_gateway",
        resource_name="main",
        attributes={"allocation_id": "eip-123", "subnet_id": "subnet-abc"},
        file_path="main.tf",
    )

    # Mock Components
    mock_matcher = MagicMock()
    mock_matcher.get_pricing_filters.return_value = (
        "AmazonEC2",
        [{"Type": "TERM_MATCH", "Field": "productFamily", "Value": "NAT Gateway"}],
    )

    mock_pricing = MagicMock()
    # NAT Gateway is approx $0.045/hr
    mock_pricing.get_product_price.return_value = 0.045

    with patch("relia.core.engine.ResourceMatcher", return_value=mock_matcher):
        with patch("relia.core.engine.PricingClient", return_value=mock_pricing):
            with patch("relia.core.engine.ConfigLoader"):
                engine = ReliaEngine()
                # Mock parser
                engine.parser.parse_directory = MagicMock(return_value=[nat])

                resources, costs = engine.run(".")

                # Verify Cost: 0.045 * 730 = 32.85
                assert costs["aws_nat_gateway.main"] == 32.85
