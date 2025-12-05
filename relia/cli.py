import typer
from relia.utils.output import print_estimate, print_diff, generate_markdown_report
from relia.core.engine import ReliaEngine

app = typer.Typer(
    name="relia",
    help="Relia: Cloud cost prevention and optimization.",
    add_completion=False,
)


@app.callback()
def main(
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Enable debug logging"),
):
    from relia.utils.logger import setup_logger

    setup_logger(verbose)


@app.command()
def estimate(
    path: str = typer.Argument(".", help="Path to infrastructure code"),
    topology: bool = typer.Option(
        False, "--topology", "-t", help="Show visual cost topology"
    ),
    diff: bool = typer.Option(False, "--diff", "-d", help="Show cost difference"),
    format: str = typer.Option(
        "table", "--format", "-f", help="Output format: table, json"
    ),
    region: str = typer.Option(
        "us-east-1", "--region", "-r", help="AWS Region used for pricing"
    ),
):
    """
    Estimate monthly cost for the user's infrastructure.
    """
    engine = ReliaEngine(region=region)
    resources, costs = engine.run(path)

    if not resources:
        if format == "json":
            import json

            typer.echo(json.dumps({"resources": [], "total_cost": 0.0}))
        else:
            typer.echo("No resources found.")
        return

    # Output
    if format == "json":
        import json

        output_data = {
            "resources": [
                {
                    "name": r.resource_name,
                    "type": r.resource_type,
                    "cost": costs.get(r.id, 0.0),
                    "attributes": r.attributes,
                }
                for r in resources
            ],
            "total_cost": sum(costs.values()),
        }
        typer.echo(json.dumps(output_data, indent=2))
        return

    print_estimate(resources, costs)

    if topology:
        from relia.utils.output import print_topology

        print_topology(resources, costs)

    if diff:
        print_diff(resources, costs)


@app.command()
def check(
    path: str = typer.Argument(".", help="Path to infrastructure code"),
    budget: float = typer.Option(
        None, help="Monthly budget limit in USD (overrides config)"
    ),
    config: str = typer.Option(".relia.yaml", help="Path to configuration file"),
    markdown_report: str = typer.Option(
        None, "--markdown-file", help="Output report to markdown file"
    ),
    dry_run: bool = typer.Option(
        False, "--dry-run", help="Simulate check without non-zero exit code"
    ),
):
    """
    Check if infrastructure cost exceeds budget or violates policies.
    Exits with code 1 if budget is exceeded or policies failed.
    """
    engine = ReliaEngine(config_path=config)
    resources, costs = engine.run(path)

    if not resources:
        typer.echo("No resources found.")
        return

    total_cost = sum(costs.values())

    # Logic: Option > Config > Default(0)
    final_budget = budget if budget is not None else engine.config.budget
    if final_budget == 0.0 and budget is None:
        # If no budget set anywhere, we might warn or just skip total check
        pass

    policy_violations = engine.check_policies(resources, costs)

    # Generate Report
    if markdown_report:
        md = generate_markdown_report(resources, costs, final_budget)
        if policy_violations:
            md += "\n## 🚨 Policy Violations\n"
            for v in policy_violations:
                md += f"- {v}\n"
        with open(markdown_report, "w") as f:
            f.write(md)

    exit_code = 0

    # Check Policies
    if policy_violations:
        typer.echo("🚨 Policy Violations:")
        for v in policy_violations:
            typer.echo(f"  - {v}")
        exit_code = 1

    # Check Budget
    if final_budget > 0 and total_cost > final_budget:
        typer.echo(
            f"🚨 Budget exceeded! Total: ${total_cost:,.2f}, Limit: ${final_budget:,.2f}"
        )
        exit_code = 1
    elif final_budget > 0:
        typer.echo(
            f"✅ Within budget. Total: ${total_cost:,.2f}, Limit: ${final_budget:,.2f}"
        )

    if exit_code != 0:
        if dry_run:
            typer.echo("⚠️  Check failed, but --dry-run is enabled. Exiting 0.")
        else:
            raise typer.Exit(code=exit_code)


@app.command()
def init():
    """
    Initialize Relia configuration files (.relia.yaml, .relia.usage.yaml).
    """
    from pathlib import Path

    config_path = Path(".relia.yaml")
    usage_path = Path(".relia.usage.yaml")

    # Create Config
    if not config_path.exists():
        config_content = """# Relia Configuration
budget: 50.0 # Monthly budget in USD
policy:
  max_resource_cost: 20.0
"""
        with open(config_path, "w") as f:
            f.write(config_content)
        typer.echo(f"✅ Created {config_path}")
    else:
        typer.echo(f"⚠️  {config_path} already exists. Skipping.")

    # Create Usage
    if not usage_path.exists():
        usage_content = """# Relia Usage Overlay
# Define usage assumptions for resources here.
usage:
  aws_lambda_function.example:
    monthly_requests: 1000000
    avg_duration_ms: 200
  aws_s3_bucket.example:
    storage_gb: 50
    monthly_requests: 10000
"""
        with open(usage_path, "w") as f:
            f.write(usage_content)
        typer.echo(f"✅ Created {usage_path}")
    else:
        typer.echo(f"⚠️  {usage_path} already exists. Skipping.")


@app.command()
def version():
    """
    Show version.
    """
    from relia import __version__

    print(f"Relia v{__version__}")


if __name__ == "__main__":
    app()
