import pytest
from unittest.mock import patch
from relia.core.parser import TerraformParser
from relia.core.pricing import PricingClient

# --- Parsing Tests ---


def test_parse_directory_not_exists():
    parser = TerraformParser()
    resources = parser.parse_directory("/non/existent/path")
    assert resources == []


def test_parse_file_bad_hcl(tmp_path):
    parser = TerraformParser()
    bad_file = tmp_path / "bad.tf"
    bad_file.write_text("resource 'aws_instance' { invalid_syntax }")
    # Should catch exception and return []
    resources = parser.parse_file(str(bad_file))
    assert resources == []


def test_parse_plan_json_bad_json(tmp_path):
    parser = TerraformParser()
    bad_file = tmp_path / "plan.json"
    bad_file.write_text("{ invalid json")
    resources = parser.parse_plan_json(str(bad_file))
    assert resources == []


def test_parse_plan_json_valid_empty(tmp_path):
    # Valid JSON but no resources
    parser = TerraformParser()
    good_file = tmp_path / "plan.json"
    good_file.write_text('{"planned_values": {"root_module": {}}}')
    resources = parser.parse_plan_json(str(good_file))
    assert resources == []


# --- Pricing Tests ---


@pytest.fixture
def mock_pricing_client():
    with patch("boto3.client") as mock_boto:
        client = PricingClient()
        client.client = mock_boto  # Override the mock created in __init__
        yield client, mock_boto


def test_pricing_api_error(mock_pricing_client):
    client, mock_boto = mock_pricing_client
    # Simulate API error
    mock_boto.get_products.side_effect = Exception("AWS Down")

    price = client.get_product_price("AmazonEC2", [])
    assert price is None


def test_pricing_empty_response(mock_pricing_client):
    client, mock_boto = mock_pricing_client
    # Simulate empty PriceList
    mock_boto.get_products.return_value = {"PriceList": []}

    price = client.get_product_price("AmazonEC2", [])
    assert price is None


def test_pricing_malformed_json_in_response(mock_pricing_client):
    client, mock_boto = mock_pricing_client
    # PriceList has valid JSON structure but unexpected schema
    mock_boto.get_products.return_value = {"PriceList": ['{"terms": {}}']}

    price = client.get_product_price("AmazonEC2", [])
    assert price is None  # _extract_price returns None


def test_pricing_extract_price_exceptions():
    # Test _extract_price directly to hit exception block
    client = PricingClient()
    # Pass an object that causes TypeError accessing 'terms'
    price = client._extract_price(None)
    assert price is None
