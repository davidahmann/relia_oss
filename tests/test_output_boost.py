import pytest
from relia.utils.output import (
    print_estimate,
    print_topology,
    print_diff,
    generate_markdown_report,
)
from relia.models import ReliaResource


@pytest.fixture
def sample_data():
    r1 = ReliaResource(
        resource_type="aws_instance",
        resource_name="web",
        attributes={"instance_type": "t3.large"},
        file_path="main.tf",
    )
    r2 = ReliaResource(
        resource_type="aws_rds_instance",
        resource_name="db",
        attributes={"class": "db.t3.micro"},
        file_path="main.tf",
    )
    costs = {r1.id: 50.0}  # r2 has no cost (unknown)
    return [r1, r2], costs


def test_print_estimate(sample_data):
    resources, costs = sample_data
    # Just verify it runs without error (exercising the Rich table building lines)
    # We can patch Console to avoid spamming output, but run is enough for coverage
    print_estimate(resources, costs)


def test_print_topology(sample_data):
    resources, costs = sample_data
    print_topology(resources, costs)


def test_print_diff(sample_data):
    resources, costs = sample_data
    # output.py logic has logic for cost > 0
    print_diff(resources, costs)


def test_markdown_report_zero_budget(sample_data):
    resources, costs = sample_data
    md = generate_markdown_report(resources, costs, budget=0.0)
    assert (
        "**Budget Status**" not in md
    )  # Logic might skip budget line if None, let's verify logic
    assert "Total Estimated Cost" in md


def test_markdown_report_no_budget(sample_data):
    resources, costs = sample_data
    md = generate_markdown_report(resources, costs, budget=None)
    assert "Budget Status" not in md
    assert "$50.00" in md
