from typing import List, Dict, Optional
from relia.models import ReliaResource


class ResourceMatcher:
    """
    Maps ReliaResource to AWS Pricing API filters.
    """

    def get_pricing_filters(
        self, resource: ReliaResource
    ) -> Optional[tuple[str, List[Dict[str, str]]]]:
        if resource.resource_type == "aws_instance":
            return "AmazonEC2", self._match_ec2(resource)
        # Add other resources here...
        return None

    def _match_ec2(self, resource: ReliaResource) -> List[Dict[str, str]]:
        instance_type = resource.attributes.get("instance_type")
        if not instance_type:
            # Maybe it's a variable; skip for MVP
            return []

        return [
            {"Type": "TERM_MATCH", "Field": "serviceCode", "Value": "AmazonEC2"},
            {
                "Type": "TERM_MATCH",
                "Field": "location",
                "Value": "US East (N. Virginia)",
            },
            {"Type": "TERM_MATCH", "Field": "instanceType", "Value": instance_type},
            {"Type": "TERM_MATCH", "Field": "operatingSystem", "Value": "Linux"},
            {"Type": "TERM_MATCH", "Field": "tenancy", "Value": "Shared"},
            {"Type": "TERM_MATCH", "Field": "preInstalledSw", "Value": "NA"},
            {"Type": "TERM_MATCH", "Field": "capacitystatus", "Value": "Used"},
        ]
