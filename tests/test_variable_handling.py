from unittest.mock import MagicMock, patch
from relia.core.engine import ReliaEngine
from relia.models import ReliaResource


def test_engine_handles_variables_without_crashing():
    # Helper to create resource with variable attributes
    def create_res(rtype, attrs):
        return ReliaResource(
            resource_type=rtype,
            resource_name="test",
            attributes=attrs,
            file_path="test.tf",
        )

    resources = [
        create_res("aws_ebs_volume", {"size": "${var.size}"}),  # Should default to 8
        create_res(
            "aws_s3_bucket", {"storage_gb": "${var.storage}"}
        ),  # Should default to 0
        create_res(
            "aws_lambda_function",
            {
                "monthly_requests": "${var.reqs}",
                "avg_duration_ms": "${var.dur}",
                "memory_size": "${var.mem}",
            },
        ),
    ]

    # Mock Pricing
    mock_pricing = MagicMock()
    mock_pricing.get_product_price.return_value = 1.0  # Easy math

    mock_matcher = MagicMock()
    mock_matcher.get_pricing_filters.return_value = ("Service", [])
    mock_matcher.region_name = "us-east-1"

    with patch("relia.core.engine.ResourceMatcher", return_value=mock_matcher):
        with patch("relia.core.engine.PricingClient", return_value=mock_pricing):
            with patch("relia.core.engine.ConfigLoader"):
                engine = ReliaEngine()
                engine.parser.parse_directory = MagicMock(return_value=resources)

                # RUN - Should not crash
                _, costs = engine.run(".")

                # Verify Defaults Used

                # EBS: size default 8 * 1.0 = 8.0
                assert costs["aws_ebs_volume.test"] == 8.0

                # S3: storage default 0 * 1.0 = 0.0
                assert costs["aws_s3_bucket.test"] == 0.0

                # Lambda:
                # requests=0, duration=100.0, memory=128
                # gb_sec = 0
                # req_cost = 0
                # total = 0
                assert costs["aws_lambda_function.test"] == 0.0


def test_engine_handles_malformed_numbers():
    # Test "foo", empty string, etc
    resources = [
        ReliaResource(
            resource_type="aws_ebs_volume",
            resource_name="bad",
            attributes={"size": "large"},
            file_path=".",
        )
    ]

    mock_pricing = MagicMock()
    mock_pricing.get_product_price.return_value = 1.0
    mock_matcher = MagicMock()
    mock_matcher.get_pricing_filters.return_value = ("Service", [])
    mock_matcher.region_name = "us-east-1"

    with patch("relia.core.engine.ResourceMatcher", return_value=mock_matcher):
        with patch("relia.core.engine.PricingClient", return_value=mock_pricing):
            with patch("relia.core.engine.ConfigLoader"):
                engine = ReliaEngine()
                engine.parser.parse_directory = MagicMock(return_value=resources)

                _, costs = engine.run(".")

                # Should use default 8
                assert costs["aws_ebs_volume.bad"] == 8.0
