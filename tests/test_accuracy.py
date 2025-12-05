from unittest.mock import MagicMock, patch, mock_open
from relia.core.engine import ReliaEngine
from relia.models import ReliaResource
from relia.core.parser import TerraformParser


def test_extract_provider_region():
    parser = TerraformParser()
    mock_tf = """
    provider "aws" {
        region = "us-west-2"
    }
    """
    # hcl2 returns dict
    mock_dict = {"provider": [{"aws": {"region": "us-west-2"}}]}

    with patch("builtins.open", mock_open(read_data=mock_tf)):
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.rglob", return_value=[("main.tf")]):
                with patch("hcl2.load", return_value=mock_dict):
                    region = parser.extract_provider_region(".")
                    assert region == "us-west-2"


def test_multi_az_rds():
    # Mock resource
    rds = ReliaResource(
        resource_type="aws_db_instance",
        resource_name="db",
        attributes={"class": "db.m5.large", "engine": "postgres", "multi_az": True},
        file_path="main.tf",
    )

    # Mock Engine
    mock_matcher = MagicMock()
    mock_matcher.region_name = "us-east-1"
    # We want to check filters
    from relia.core.matcher import ResourceMatcher

    real_matcher = ResourceMatcher()

    filters = real_matcher._match_rds(rds)

    # Assert deploymentOption is Multi-AZ
    dep_opt = next(f for f in filters if f["Field"] == "deploymentOption")
    assert dep_opt["Value"] == "Multi-AZ"


def test_engine_updates_region():
    # Test Auto-Detection Logic
    with patch("relia.core.engine.TerraformParser"):
        with patch("relia.core.engine.PricingClient"):
            with patch("relia.core.engine.ConfigLoader"):
                engine = ReliaEngine()
                engine.matcher = MagicMock()
                engine.matcher.region_name = "us-east-1"

                # Mock Parser to return region
                engine.parser.extract_provider_region.return_value = "eu-central-1"
                engine.parser.parse_directory.return_value = []

                engine.run(".")

                assert engine.matcher.region_name == "eu-central-1"
