from unittest.mock import MagicMock, patch, mock_open
from relia.core.engine import ReliaEngine
from relia.models import ReliaResource


def test_lambda_pricing():
    # Mock usage
    usage_yaml = """
usage:
  aws_lambda_function.my_func:
    monthly_requests: 1000000
    avg_duration_ms: 1000
"""
    # Duration 1000ms = 1s.

    # Mock resource (Terraform attribs)
    # Memory defaults to 128 if not set? No, engine logic sets default.
    # Let's set memory explicitly in TF mock
    lambda_res = ReliaResource(
        resource_type="aws_lambda_function",
        resource_name="my_func",
        attributes={"memory_size": 1024, "function_name": "foo"},
        file_path="main.tf",
    )

    # Mock Components
    mock_matcher = MagicMock()
    mock_matcher.get_pricing_filters.return_value = (
        "AWSLambda",
        [{"Type": "Serverless"}],
    )

    mock_matcher.get_lambda_request_filters.return_value = [{"Type": "Request"}]

    mock_pricing = MagicMock()

    def price_side_effect(service, filters):
        # If filters match duration (from get_pricing_filters mock)
        if filters == [{"Type": "Serverless"}]:
            return 0.0000166667
        # If filters match requests
        if filters == [{"Type": "Request"}]:
            return 0.0000002  # $0.20 per million
        return 0.0

    mock_pricing.get_product_price.side_effect = price_side_effect

    with patch("builtins.open", mock_open(read_data=usage_yaml)):
        with patch("pathlib.Path.exists", return_value=True):
            with patch("relia.core.engine.ResourceMatcher", return_value=mock_matcher):
                with patch(
                    "relia.core.engine.PricingClient", return_value=mock_pricing
                ):
                    with patch("relia.core.engine.ConfigLoader"):
                        engine = ReliaEngine()
                        engine.parser.parse_directory = MagicMock(
                            return_value=[lambda_res]
                        )

                        # 1M Requests, 1000ms, 1024MB Memory
                        # GB-Seconds:
                        # 1,000,000 * (1000/1000) * (1024/1024) = 1,000,000 GB-Seconds

                        # Compute Cost: 1,000,000 * 0.0000166667 = $16.6667
                        # Request Cost: (1M / 1M) * 0.20 = $0.20
                        # Total: $16.8667

                        resources, costs = engine.run(".")

                        assert round(costs["aws_lambda_function.my_func"], 4) == 16.8667
