from unittest.mock import MagicMock, patch, mock_open
from relia.core.engine import ReliaEngine
from relia.models import ReliaResource


def test_usage_overlay_s3():
    # Mock usage file content
    usage_yaml = """
usage:
  aws_s3_bucket.my_data:
    storage_gb: 500
"""

    # Mock resource
    s3_res = ReliaResource(
        resource_type="aws_s3_bucket",
        resource_name="my_data",
        attributes={"bucket": "foo"},  # No size here
        file_path="main.tf",
    )

    # Mock Components
    mock_matcher = MagicMock()
    mock_matcher.get_pricing_filters.return_value = ("AmazonS3", [{"Type": "S3"}])

    mock_pricing = MagicMock()
    mock_pricing.get_product_price.return_value = 0.023  # $0.023 per GB

    # Patch everything including open() for the usage file
    with patch("builtins.open", mock_open(read_data=usage_yaml)):
        with patch("pathlib.Path.exists", return_value=True):
            with patch("relia.core.engine.ResourceMatcher", return_value=mock_matcher):
                with patch(
                    "relia.core.engine.PricingClient", return_value=mock_pricing
                ):
                    with patch("relia.core.engine.ConfigLoader"):
                        # We also need to patch the UsageLoader's open calls if they are separate
                        # But since engine inits usage loader, the outer patch works

                        engine = ReliaEngine()
                        # Mock parser
                        engine.parser.parse_directory = MagicMock(return_value=[s3_res])

                        resources, costs = engine.run(".")

                        # Verify Cost: 0.023 * 500 = 11.5
                        assert costs["aws_s3_bucket.my_data"] == 11.5

                        # Verify attribute was updated
                        assert s3_res.attributes["storage_gb"] == 500
