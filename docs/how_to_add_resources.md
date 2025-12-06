---
title: "How to Add New AWS Resources to Relia | Developer Guide"
description: "Step-by-step developer guide on adding support for new AWS resources (DynamoDB, Redshift, etc.) to the Relia cost estimation engine."
keywords: contribute to relia, add aws resource, python terraform parser, aws pricing api development
---
<link rel="canonical" href="https://davidahmann.github.io/relia_oss/how_to_add_resources/" />

# How to Add New AWS Resources

Relia is designed to be extensible. Adding support for a new resource (e.g., `aws_dynamodb_table`) takes just 3 steps.

## Architecture Context
Costs are calculated in a linear pipeline:
1.  **Parser**: Reads `.tf` -> `ReliaResource("aws_dynamodb_table", ...)`
2.  **Matcher**: Maps `ReliaResource` -> AWS Pricing Filters (e.g. `AmazonDynamoDB`).
3.  **Pricing**: Fetches cost from cache or API.

You typically only need to touch the **Matcher**.

---

## Step 1: Identify Pricing Filters
First, find out how AWS prices the resource using the CLI.
Example for DynamoDB:
```bash
aws pricing get-products \
    --service-code AmazonDynamoDB \
    --filters Type=TERM_MATCH,Field=productFamily,Value="Database Storage" \
    --region us-east-1
```
*Goal: Find the unique combination of filters (Location, ProductFamily, etc.) that yields the price you want.*

## Step 2: Implement Matcher Logic
Open `relia/core/matcher.py` and add a new method.

```python
# relia/core/matcher.py

    def get_pricing_filters(self, resource: ReliaResource):
        # ... existing dispatch logic ...
        if resource.resource_type == "aws_dynamodb_table":
             return "AmazonDynamoDB", self._match_dynamodb(resource)

        return None

    def _match_dynamodb(self, resource: ReliaResource) -> List[Dict[str, str]]:
        # Extract attributes from Terraform
        billing_mode = resource.attributes.get("billing_mode", "PROVISIONED")

        # Define filters based on Step 1
        return [
            {"Type": "TERM_MATCH", "Field": "serviceCode", "Value": "AmazonDynamoDB"},
            {"Type": "TERM_MATCH", "Field": "location", "Value": self._get_location()},
            {"Type": "TERM_MATCH", "Field": "productFamily", "Value": "Database Storage"},
            {"Type": "TERM_MATCH", "Field": "group", "Value": "DDB-WriteUnits"}, # Example
        ]
```

## Step 3: Add Unit Tests
Ensure your logic holds up! Add a test case in `tests/test_matcher_unit.py`.

```python
def test_dynamodb_matching():
    # Mock a parsed resource
    res = ReliaResource(id="aws_dynamodb_table.users", attributes={"billing_mode": "PROVISIONED"})

    matcher = ResourceMatcher(region="us-east-1")
    service, filters = matcher.get_pricing_filters(res)

    assert service == "AmazonDynamoDB"
    assert {"Field": "productFamily", "Value": "Database Storage"} in filters
```

---

<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "HowTo",
  "name": "How to Add Support for a New AWS Resource in Relia",
  "description": "Guide for developers to extend Relia by adding parser and matcher logic for new Terraform resources.",
  "step": [{
    "@type": "HowToStep",
    "name": "Identify Pricing Filters",
    "text": "Use the AWS CLI to find the correct ServiceCode and Filters for the resource."
  }, {
    "@type": "HowToStep",
    "name": "Implement Matcher Logic",
    "text": "Add a handler method in relia/core/matcher.py mapping Terraform attributes to those filters."
  }, {
    "@type": "HowToStep",
    "name": "Add Unit Tests",
    "text": "Verify the logic with a new test case in tests/test_matcher_unit.py."
  }]
}
</script>

---
## Related Documentation
- [Architecture](architecture.md) - Understand the Matcher's role
- [GitHub Repository](https://github.com/davidahmann/relia_oss) - Submit your PR here
