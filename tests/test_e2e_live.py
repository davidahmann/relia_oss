import subprocess
import sys
import os
import pytest
from pathlib import Path

# Paths
ROOT_DIR = Path(__file__).parent.parent
DEMO_TF = ROOT_DIR / "demo" / "main.tf"


@pytest.mark.e2e
def test_e2e_estimate_success():
    """Verify 'relia estimate' runs and produces output."""
    if not DEMO_TF.exists():
        pytest.skip("Demo TF file missing")

    cmd = [sys.executable, "-m", "relia.main", "estimate", str(DEMO_TF)]

    # Run process
    result = subprocess.run(cmd, cwd=ROOT_DIR, capture_output=True, text=True)

    assert result.returncode == 0
    assert "Relia Cost Estimate" in result.stdout
    assert "t3.large" in result.stdout


@pytest.mark.e2e
def test_e2e_check_budget_fail():
    """Verify 'relia check' enforces budget failures (Exit Code 1)."""
    if not DEMO_TF.exists():
        pytest.skip("Demo TF file missing")

    cmd = [sys.executable, "-m", "relia.main", "check", str(DEMO_TF)]

    env = os.environ.copy()
    env["RELIA_BUDGET"] = "50"  # Too low

    result = subprocess.run(cmd, cwd=ROOT_DIR, capture_output=True, text=True, env=env)

    assert result.returncode == 1
    assert "Budget exceeded" in result.stdout
    assert "Limit: $50.00" in result.stdout


@pytest.mark.e2e
def test_e2e_check_budget_pass():
    """Verify 'relia check' passes within budget (Exit Code 0)."""
    if not DEMO_TF.exists():
        pytest.skip("Demo TF file missing")

    cmd = [sys.executable, "-m", "relia.main", "check", str(DEMO_TF)]

    env = os.environ.copy()
    env["RELIA_BUDGET"] = "500"  # Sufficient

    result = subprocess.run(cmd, cwd=ROOT_DIR, capture_output=True, text=True, env=env)

    assert result.returncode == 0
    assert "Within budget" in result.stdout
