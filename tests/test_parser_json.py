from unittest.mock import mock_open, patch
import json
from relia.core.parser import TerraformParser


def test_parse_plan_json_recursive():
    # Mock efficient plan JSON structure
    mock_plan = {
        "planned_values": {
            "root_module": {
                "resources": [
                    {
                        "type": "aws_instance",
                        "name": "root_web",
                        "address": "aws_instance.root_web",
                        "values": {"instance_type": "t3.micro"},
                    }
                ],
                "child_modules": [
                    {
                        "address": "module.db",
                        "resources": [
                            {
                                "type": "aws_db_instance",
                                "name": "main",
                                "address": "module.db.aws_db_instance.main",
                                "values": {
                                    "instance_class": "db.t3.micro",
                                    "engine": "postgres",
                                },
                            }
                        ],
                        "child_modules": [],
                    }
                ],
            }
        }
    }

    parser = TerraformParser()

    with patch("builtins.open", mock_open(read_data=json.dumps(mock_plan))):
        resources = parser.parse_plan_json("plan.json")

    assert len(resources) == 2

    # Check Root Resource
    root_res = next(r for r in resources if r.resource_type == "aws_instance")
    assert root_res.resource_name == "aws_instance.root_web"
    assert root_res.attributes["instance_type"] == "t3.micro"

    # Check Module Resource
    mod_res = next(r for r in resources if r.resource_type == "aws_db_instance")
    assert mod_res.resource_name == "module.db.aws_db_instance.main"
    assert mod_res.attributes["instance_class"] == "db.t3.micro"
