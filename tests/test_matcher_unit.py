import pytest
from relia.core.matcher import ResourceMatcher
from relia.models import ReliaResource


@pytest.fixture
def matcher():
    return ResourceMatcher(region="us-east-1")


def test_match_ec2_standard(matcher):
    r = ReliaResource(
        resource_type="aws_instance",
        resource_name="web",
        attributes={"instance_type": "t3.micro"},
        file_path=".",
    )
    filters = matcher._match_ec2(r)
    assert ("instanceType", "t3.micro") in [(f["Field"], f["Value"]) for f in filters]
    assert ("operatingSystem", "Linux") in [(f["Field"], f["Value"]) for f in filters]


def test_match_ec2_missing_type(matcher):
    r = ReliaResource(
        resource_type="aws_instance", resource_name="web", attributes={}, file_path="."
    )
    # Should warn and return empty list []
    assert matcher._match_ec2(r) == []


def test_match_rds_multi_az(matcher):
    r = ReliaResource(
        resource_type="aws_db_instance",
        resource_name="db",
        attributes={
            "instance_class": "db.t3.micro",
            "multi_az": True,
            "engine": "mysql",
        },
        file_path=".",
    )
    filters = matcher._match_rds(r)
    # Check explicitly for Multi-AZ
    kv_pairs = [(f["Field"], f["Value"]) for f in filters]
    assert ("deploymentOption", "Multi-AZ") in kv_pairs
    assert ("databaseEngine", "MySQL") in kv_pairs


def test_match_rds_single_az(matcher):
    r = ReliaResource(
        resource_type="aws_db_instance",
        resource_name="db",
        attributes={
            "instance_class": "db.t3.micro",
            "multi_az": False,
            "engine": "mysql",
        },
        file_path=".",
    )
    filters = matcher._match_rds(r)
    kv_pairs = [(f["Field"], f["Value"]) for f in filters]
    assert ("deploymentOption", "Single-AZ") in kv_pairs


def test_match_nat_gateway(matcher):
    r = ReliaResource(
        resource_type="aws_nat_gateway",
        resource_name="nat",
        attributes={},
        file_path=".",
    )
    filters = matcher._match_nat_gateway(r)
    kv_pairs = [(f["Field"], f["Value"]) for f in filters]
    assert ("group", "NAT Gateway") in kv_pairs


def test_match_s3(matcher):
    r = ReliaResource(
        resource_type="aws_s3_bucket",
        resource_name="bucket",
        attributes={},
        file_path=".",
    )
    filters = matcher._match_s3(r)
    kv_pairs = [(f["Field"], f["Value"]) for f in filters]
    assert ("volumeType", "Standard") in kv_pairs


def test_match_lambda(matcher):
    # Current implementation ignores architecture, defaults to standard duration
    r = ReliaResource(
        resource_type="aws_lambda_function",
        resource_name="func",
        attributes={"architectures": ["arm64"]},
        file_path=".",
    )
    filters = matcher._match_lambda(r)
    kv_pairs = [(f["Field"], f["Value"]) for f in filters]
    assert ("group", "AWS-Lambda-Duration") in kv_pairs


def test_match_lambda_x86(matcher):
    r = ReliaResource(
        resource_type="aws_lambda_function",
        resource_name="func",
        attributes={},
        file_path=".",
    )
    filters = matcher._match_lambda(r)
    kv_pairs = [(f["Field"], f["Value"]) for f in filters]
    assert ("group", "AWS-Lambda-Duration") in kv_pairs


def test_unsupported_resource(matcher):
    r = ReliaResource(
        resource_type="aws_unknown_thing",
        resource_name="unk",
        attributes={},
        file_path=".",
    )
    assert matcher.get_pricing_filters(r) is None
