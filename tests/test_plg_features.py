from pathlib import Path
from typer.testing import CliRunner
from relia.cli import app
from relia.core.advisor import ReliaAdvisor
from relia.models import ReliaResource

runner = CliRunner()


def test_advisor_logic():
    """Unit test for advisor rules"""
    advisor = ReliaAdvisor()

    # GP2 -> GP3
    r1 = ReliaResource(
        resource_type="aws_ebs_volume", resource_name="vol", attributes={"type": "gp2"}
    )
    assert "Upgrade to gp3" in advisor.analyze([r1])[r1.id][0]

    # T2 -> T3
    r2 = ReliaResource(
        resource_type="aws_instance",
        resource_name="ec2",
        attributes={"instance_type": "t2.micro"},
    )
    assert "Consider t3.micro" in advisor.analyze([r2])[r2.id][0]


def test_html_report_generation():
    """Test HTML report structure and content"""
    # Mock data
    r1 = ReliaResource(
        resource_type="aws_ebs_volume",
        resource_name="legacy_vol",
        attributes={"type": "gp2", "size": 100},
    )
    r1.suggestions = ["💡 Upgrade to gp3"]

    resources = [r1]
    costs = {r1.id: 10.0}

    from relia.utils.output import generate_html_report

    html = generate_html_report(resources, costs)

    assert "<!DOCTYPE html>" in html
    assert "Relia Cost Report" in html
    assert "$10.00" in html
    assert "Upgrade to gp3" in html  # Tip is present
    assert 'class="tip"' in html
    assert "mermaid" in html  # Check mermaid is present


def test_cli_html_output():
    """Test CLI command produces file"""
    with runner.isolated_filesystem():
        # Create a dummy TF file
        with open("main.tf", "w") as f:
            f.write('resource "aws_ebs_volume" "test" { type = "gp2" size = 10 }')

        result = runner.invoke(
            app, ["estimate", ".", "--format", "html", "--out", "report.html"]
        )
        assert result.exit_code == 0
        assert "Report saved to report.html" in result.output

        report_path = Path("report.html")
        assert report_path.exists()
        content = report_path.read_text()
        assert "Relia Cost Report" in content
