from relia.core.advisor import ReliaAdvisor
from relia.models import ReliaResource


def test_advisor_gp2_suggestion():
    advisor = ReliaAdvisor()
    r = ReliaResource(
        resource_type="aws_ebs_volume",
        resource_name="old_disk",
        attributes={"type": "gp2", "size": 100},
    )

    suggestions = advisor.analyze([r])
    assert r.id in suggestions
    assert any("Upgrade to gp3" in tip for tip in suggestions[r.id])


def test_advisor_t2_suggestion():
    advisor = ReliaAdvisor()
    r = ReliaResource(
        resource_type="aws_instance",
        resource_name="web",
        attributes={"instance_type": "t2.micro"},
    )

    suggestions = advisor.analyze([r])
    assert r.id in suggestions
    assert any("Consider t3.micro" in tip for tip in suggestions[r.id])


def test_advisor_graviton_suggestion():
    advisor = ReliaAdvisor()
    r = ReliaResource(
        resource_type="aws_instance",
        resource_name="app",
        attributes={"instance_type": "m5.large"},
    )

    suggestions = advisor.analyze([r])
    assert r.id in suggestions
    assert any("Consider Graviton (m5g.large)" in tip for tip in suggestions[r.id])
