import pytest
from unittest.mock import MagicMock, patch
from relia.core.engine import ReliaEngine
from relia.models import ReliaResource


def test_ebs_pricing_logic():
    # Mock resources
    ebs = ReliaResource(
        resource_type="aws_ebs_volume",
        resource_name="data",
        attributes={"type": "gp3", "size": 100},  # 100 GB
        file_path="main.tf",
    )

    ec2 = ReliaResource(
        resource_type="aws_instance",
        resource_name="web",
        attributes={"instance_type": "t3.micro"},
        file_path="main.tf",
    )

    # Mock Matcher to return valid filters
    mock_matcher = MagicMock()
    mock_matcher.get_pricing_filters.side_effect = [
        ("AmazonEC2", [{"Type": "EBS"}]),  # EBS
        ("AmazonEC2", [{"Type": "EC2"}]),  # EC2
    ]

    # Mock Pricing to return raw unit prices
    # Standard gp3 price is ~$0.08/GB-Mo
    # t3.micro price is ~$0.0104/hr
    mock_pricing = MagicMock()
    mock_pricing.get_product_price.side_effect = [
        0.08,  # EBS Unit Price
        0.0104,  # EC2 Unit Price
    ]

    with patch("relia.core.engine.ResourceMatcher", return_value=mock_matcher):
        with patch("relia.core.engine.PricingClient", return_value=mock_pricing):
            # We also need to patch ConfigLoader to avoid FS calls or init issues
            with patch("relia.core.engine.ConfigLoader"):
                engine = ReliaEngine()
                # Bypass run() and verify calculation logic via injected loop or helper
                # but since logic is in run(), we mock parser
                engine.parser.parse_directory = MagicMock(return_value=[ebs, ec2])

                resources, costs = engine.run(".")

                # Verify EBS: 0.08 * 100 = 8.0
                assert costs["aws_ebs_volume.data"] == 8.0

                # Verify EC2: 0.0104 * 730 = 7.592
                assert costs["aws_instance.web"] == pytest.approx(7.592)
