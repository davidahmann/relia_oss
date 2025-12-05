from relia.core.parser import TerraformParser


def test_parser_basic(tmp_path):
    # Create a mock .tf file
    d = tmp_path / "infra"
    d.mkdir()
    p = d / "main.tf"
    p.write_text(
        """
    resource "aws_instance" "web" {
        instance_type = "t3.large"
        ami = "ami-12345"
    }

    resource "aws_s3_bucket" "data" {
        bucket = "my-bucket"
    }
    """
    )

    parser = TerraformParser()
    resources = parser.parse_directory(str(d))

    assert len(resources) == 2

    # Check Instance
    instance = next(r for r in resources if r.resource_type == "aws_instance")
    assert instance.resource_name == "web"
    assert instance.attributes["instance_type"] == "t3.large"
    assert instance.id == "aws_instance.web"

    # Check Bucket
    bucket = next(r for r in resources if r.resource_type == "aws_s3_bucket")
    assert bucket.resource_name == "data"
