import boto3
import json
from typing import Optional, Dict, List
from relia.core.cache import PricingCache


class PricingClient:
    """
    Client for AWS Price List Service API.
    Handles caching and filtering.
    """

    def __init__(self, region="us-east-1"):
        # Pricing API is only available in us-east-1 and ap-south-1
        self.client = boto3.client("pricing", region_name="us-east-1")
        self.cache = PricingCache()

    def get_product_price(
        self, service_code: str, filters: List[Dict[str, str]]
    ) -> Optional[float]:
        """
        Fetch monthly price for a product.
        Returns price in USD.
        """
        # Create a cache key from filters
        # Sort filters to ensure consistent key
        sorted_filters = sorted(
            [f"{f['Type']}:{f['Field']}:{f['Value']}" for f in filters]
        )
        cache_key = f"{service_code}|" + "|".join(sorted_filters)

        # Check cache
        cached_price = self.cache.get(cache_key)
        if cached_price is not None:
            return cached_price.get("price")

        try:
            response = self.client.get_products(
                ServiceCode=service_code, Filters=filters, MaxResults=1
            )

            price_list = response.get("PriceList", [])
            if not price_list:
                return None

            # PriceList returns string JSON, we need to parse it
            product_data = json.loads(price_list[0])
            price = self._extract_price(product_data)

            if price is not None:
                self.cache.set(cache_key, {"price": price})

            return price

        except Exception as e:
            # Fallback or log error
            print(f"Error fetching price for {cache_key}: {e}")
            return None

    def _extract_price(self, product_data: Dict) -> Optional[float]:
        """
        Extract OnDemand hourly price from complex Price List JSON.
        We assume OnDemand and USD for MVP.
        """
        try:
            terms = product_data.get("terms", {}).get("OnDemand", {})
            if not terms:
                return None

            # Grab the first term (usually there's only one for OnDemand w/o contract)
            offer_term = next(iter(terms.values()))

            price_dimensions = offer_term.get("priceDimensions", {})
            if not price_dimensions:
                return None

            # Grab the first price dimension
            price_dimension = next(iter(price_dimensions.values()))

            price_per_unit = price_dimension.get("pricePerUnit", {}).get("USD")
            if price_per_unit:
                return float(price_per_unit)

        except Exception as e:
            print(f"Error parsing price JSON: {e}")
            return None

        return None
