from typing import List, Dict
from relia.models import ReliaResource


class ReliaAdvisor:
    """
    Analyzes resources for cost optimization opportunities.
    """

    def analyze(self, resources: List[ReliaResource]) -> Dict[str, List[str]]:
        """
        Returns a map of resource_id -> list of suggestions.
        """
        suggestions: Dict[str, List[str]] = {}

        for resource in resources:
            tips = []

            # 1. EBS Optimization: gp2 -> gp3
            if resource.resource_type == "aws_ebs_volume":
                vol_type = resource.attributes.get("type", "gp2")
                if vol_type == "gp2":
                    tips.append(
                        "💡 Upgrade to gp3: 20% cheaper and better baseline performance."
                    )

            # 2. EC2 Optimization: t2 -> t3 (Example)
            if resource.resource_type == "aws_instance":
                instance_type = resource.attributes.get("instance_type", "")
                if instance_type and "t2." in instance_type:
                    tips.append(
                        f"💡 Consider {instance_type.replace('t2.', 't3.')}: Newer generation, often cheaper."
                    )

                # Check for Graviton opportunity? (Simple heuristic)
                if instance_type and instance_type.startswith(("m5.", "c5.", "r5.")):
                    g_type = instance_type.replace(".", "g.")
                    tips.append(
                        f"💡 Consider Graviton ({g_type}): Up to 20% savings for compatible workloads."
                    )

            if tips:
                suggestions[resource.id] = tips

        return suggestions
