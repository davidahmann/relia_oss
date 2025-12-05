import typer
from relia.utils.output import print_estimate, print_diff, generate_markdown_report
from relia.core.engine import ReliaEngine

app = typer.Typer(
    name="relia",
    help="Relia: Cloud cost prevention and optimization.",
    add_completion=False,
)


@app.command()
def estimate(
    path: str = typer.Argument(".", help="Path to infrastructure code"),
    topology: bool = typer.Option(
        False, "--topology", "-t", help="Show visual cost topology"
    ),
    diff: bool = typer.Option(False, "--diff", "-d", help="Show cost difference"),
):
    """
    Estimate monthly cost for the user's infrastructure.
    """
    engine = ReliaEngine()
    resources, costs = engine.run(path)

    if not resources:
        typer.echo("No resources found.")
        return

    # Output
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
        raise typer.Exit(code=exit_code)


@app.command()
def version():
    """
    Show version.
    """
    from relia import __version__

    print(f"Relia v{__version__}")


if __name__ == "__main__":
    app()
