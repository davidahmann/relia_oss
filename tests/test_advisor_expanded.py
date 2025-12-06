from relia.core.advisor import ReliaAdvisor
from relia.models import ReliaResource


def test_advisor_lambda_arm():
    advisor = ReliaAdvisor()
    # Case 1: Default (x86 implied)
    r1 = ReliaResource(
        resource_type="aws_lambda_function",
        resource_name="fn1",
        attributes={},
        file_path="main.tf",
    )
    tips1 = advisor.analyze([r1])
    # ID is aws_lambda_function.fn1
    assert "Switch to ARM64" in tips1["aws_lambda_function.fn1"][0]

    # Case 2: Explicit x86
    r2 = ReliaResource(
        resource_type="aws_lambda_function",
        resource_name="fn2",
        attributes={"architectures": ["x86_64"]},
        file_path="main.tf",
    )
    tips2 = advisor.analyze([r2])
    assert "Switch to ARM64" in tips2["aws_lambda_function.fn2"][0]

    # Case 3: Already ARM
    r3 = ReliaResource(
        resource_type="aws_lambda_function",
        resource_name="fn3",
        attributes={"architectures": ["arm64"]},
        file_path="main.tf",
    )
    tips3 = advisor.analyze([r3])
    assert "aws_lambda_function.fn3" not in tips3


def test_advisor_rds_optimization():
    advisor = ReliaAdvisor()
    # Case 1: gp2 storage
    r1 = ReliaResource(
        resource_type="aws_db_instance",
        resource_name="db1",
        attributes={"storage_type": "gp2"},
        file_path="main.tf",
    )
    tips1 = advisor.analyze([r1])
    assert any("Upgrade storage to gp3" in t for t in tips1["aws_db_instance.db1"])

    # Case 2: Aurora Provisioned
    r2 = ReliaResource(
        resource_type="aws_db_instance",
        resource_name="db2",
        attributes={"engine": "aurora-mysql", "storage_type": "io1"},
        file_path="main.tf",
    )
    tips2 = advisor.analyze([r2])
    assert any("Aurora Serverless v2" in t for t in tips2["aws_db_instance.db2"])
