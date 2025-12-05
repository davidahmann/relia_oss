import hcl2
from typing import List
from relia.models import ReliaResource
from pathlib import Path
from relia.utils.logger import logger


class TerraformParser:
    """Parses Terraform files into ReliaResource objects."""

    def parse_directory(self, directory: str) -> List[ReliaResource]:
        """Parses all .tf files in a directory recursively."""
        resources: List[ReliaResource] = []
        path = Path(directory)

        if not path.exists():
            return []

        for file_path in path.rglob("*.tf"):
            resources.extend(self.parse_file(str(file_path)))

        return resources

    def parse_file(self, file_path: str) -> List[ReliaResource]:
        """Parses a single .tf file."""
        resources: List[ReliaResource] = []

        try:
            with open(file_path, "r") as file:
                dict_structure = hcl2.load(file)

            # hcl2 returns {'resource': [{'aws_instance': {'name': {...}}}]}
            # We need to flatten this.

            if "resource" in dict_structure:
                for resource_block in dict_structure["resource"]:
                    # resource_block is like {'aws_instance': {'web': {...}}}
                    for r_type, r_instances in resource_block.items():
                        # r_instances is usually a dict of name -> config
                        # OR if multiple instances of same type, it might be different structure in some parsers
                        # In python-hcl2, it's typically {name: {config}}

                        for r_name, r_config in r_instances.items():
                            resources.append(
                                ReliaResource(
                                    resource_type=r_type,
                                    resource_name=r_name,
                                    attributes=r_config,
                                    file_path=file_path,
                                )
                            )

            return resources

        except Exception as e:
            # We don't want to crash the whole scan on one bad file
            logger.warning(f"Error parsing {file_path}: {e}")
            return []

    def parse_plan_json(self, file_path: str) -> List[ReliaResource]:
        """Parses a terraform show -json output file."""
        import json

        resources: List[ReliaResource] = []

        try:
            with open(file_path, "r") as f:
                data = json.load(f)

            # Navigate to resources: planned_values -> root_module
            root = data.get("planned_values", {}).get("root_module", {})

            def extract_resources(module: dict) -> List[ReliaResource]:
                found = []
                # 1. Local resources
                for r in module.get("resources", []):
                    r_type = r.get("type")
                    r_name = r.get("name")
                    r_values = r.get("values", {})

                    if r_type and r_name:
                        found.append(
                            ReliaResource(
                                resource_type=r_type,
                                # Use address if available for uniqueness, else name
                                resource_name=r.get("address", r_name),
                                attributes=r_values,
                                file_path=file_path,
                            )
                        )

                # 2. Child modules (Recursion)
                for child in module.get("child_modules", []):
                    found.extend(extract_resources(child))

                return found

            resources = extract_resources(root)

            return resources
        except Exception as e:
            logger.warning(f"Error parsing JSON plan {file_path}: {e}")
            return []

    def extract_provider_region(self, directory: str) -> str | None:
        """
        Scans .tf files in directory for a 'provider "aws"' block with a region.
        Returns the first valid region found.
        """
        path = Path(directory)
        if not path.exists():
            return None

        # Reuse parsing logic? Or simpler scan?
        # Parse logic uses hcl2.load which is robust.
        # But parsing *all* files can be slow.
        # For MVP, just iterate parsing until we find one.

        for file_path in path.rglob("*.tf"):
            try:
                with open(file_path, "r") as f:
                    data = hcl2.load(f)

                # Check providers
                # Structure: {'provider': [{'aws': {'region': '...'}}]}
                providers = data.get("provider", [])
                for provider in providers:
                    if "aws" in provider:
                        region = provider["aws"].get("region")
                        if (
                            region
                            and isinstance(region, str)
                            and not region.startswith("$")
                        ):
                            # Found explicit string region
                            return region
            except Exception:
                continue

        return None
