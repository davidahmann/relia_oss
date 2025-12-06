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

                if instance_type and instance_type.startswith(("m5.", "c5.", "r5.")):
                    g_type = instance_type.replace(".", "g.")
                    tips.append(
                        f"💡 Consider Graviton ({g_type}): Up to 20% savings for compatible workloads."
                    )

            # 3. RDS Optimization
            if resource.resource_type == "aws_db_instance":
                # Check for gp2 storage
                storage_type = resource.attributes.get("storage_type", "gp2")
                if storage_type == "gp2":
                    tips.append(
                        "💡 Upgrade storage to gp3: Consistent IOPS and lower price."
                    )

                # Check for Aurora Serverless opportunity (Generic advice)
                engine = resource.attributes.get("engine", "")
                if "aurora" in engine and "serverless" not in engine:
                    tips.append(
                        "💡 For variable workloads, consider Aurora Serverless v2 to auto-scale capacity."
                    )

            # 4. Lambda Optimization
            if resource.resource_type == "aws_lambda_function":
                archs = resource.attributes.get("architectures", [])
                # Terraform 'architectures' is a list, e.g. ["x86_64"] or ["arm64"]
                # If missing or explicitly x86, suggest ARM
                if not archs or "arm64" not in archs:
                    tips.append(
                        "💡 Switch to ARM64 (Graviton2): 20% cheaper per ms and often faster."
                    )

            if tips:
                suggestions[resource.id] = tips

        return suggestions
