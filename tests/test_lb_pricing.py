from unittest.mock import MagicMock, patch
from relia.core.engine import ReliaEngine
from relia.models import ReliaResource


def test_lb_pricing():
    # Mock resource (ALB)
    alb = ReliaResource(
        resource_type="aws_lb",
        resource_name="app_lb",
        attributes={"load_balancer_type": "application"},
        file_path="main.tf",
    )

    # Mock Components
    mock_matcher = MagicMock()
    mock_matcher.get_pricing_filters.return_value = (
        "AmazonEC2",
        [{"Type": "Load Balancer-Application"}],
    )

    mock_pricing = MagicMock()
    # Approx ALB hourly price
    mock_pricing.get_product_price.return_value = 0.0225

    with patch("relia.core.engine.ResourceMatcher", return_value=mock_matcher):
        with patch("relia.core.engine.PricingClient", return_value=mock_pricing):
            with patch("relia.core.engine.ConfigLoader"):
                engine = ReliaEngine()
                engine.parser.parse_directory = MagicMock(return_value=[alb])

                resources, costs = engine.run(".")

                # Verify Cost: 0.0225 * 730 = 16.425
                assert costs["aws_lb.app_lb"] == 16.425
