import pytest
from unittest.mock import MagicMock, patch
from relia.core.pricing import PricingClient


@pytest.fixture
def mock_pricing_client():
    with patch("boto3.client"):
        client = PricingClient()
        client.client = MagicMock()
        yield client


def test_extract_price(mock_pricing_client):
    # Mock data structure from AWS Price List API
    mock_data = {
        "terms": {
            "OnDemand": {
                "term_id": {
                    "priceDimensions": {"dim_id": {"pricePerUnit": {"USD": "0.0416"}}}
                }
            }
        }
    }

    price = mock_pricing_client._extract_price(mock_data)
    # Unit price is 0.0416 (No longer multiplied by 730)
    assert price == 0.0416


def test_get_product_price_cache_hit(mock_pricing_client):
    # Manually populate cache
    mock_pricing_client.cache.set(
        "AmazonEC2|TERM_MATCH:instanceType:t3.micro", {"price": 10.0}
    )

    price = mock_pricing_client.get_product_price(
        "AmazonEC2",
        [{"Type": "TERM_MATCH", "Field": "instanceType", "Value": "t3.micro"}],
    )

    assert price == 10.0
    # Ensure we didn't call API
    mock_pricing_client.client.get_products.assert_not_called()
