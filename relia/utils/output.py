from rich.console import Console
from rich.table import Table
from rich import box
from typing import List, Dict, Optional
from relia.models import ReliaResource

console = Console()


def print_estimate(resources: List[ReliaResource], costs: Dict[str, float]):
    console.print("\n📊 [bold blue]Relia Cost Estimate[/bold blue]\n")

    table = Table(box=box.ROUNDED, show_header=True, header_style="bold magenta")
    table.add_column("Resource", style="cyan")
    table.add_column("Type", style="dim")
    table.add_column("Cost/Month", justify="right", style="green")
    table.add_column("Status", justify="center")

    total_cost = 0.0

    for r in resources:
        cost = costs.get(r.id, 0.0)
        total_cost += cost

        # Determine "Type" to show
        r_type_val = r.attributes.get("instance_type") or r.resource_type

        cost_str = f"${cost:,.2f}" if cost > 0 else "-"
        status = "✅"  # Logic can be smarter later

        table.add_row(r.id, r_type_val, cost_str, status)

    table.add_section()
    table.add_row("[bold]Total[/bold]", "", f"[bold]${total_cost:,.2f}[/bold]", "")

    console.print(table)


def print_topology(resources: List[ReliaResource], costs: Dict[str, float]):
    from rich.tree import Tree

    console.print("\n🌳 [bold blue]Infrastructure Topology[/bold blue]\n")

    root = Tree("☁️  [bold]Project Infrastructure[/bold]")

    # Group by type for now (simple topology)
    # A real topology would need dependency graph analysis
    type_groups: Dict[str, List[ReliaResource]] = {}
    for r in resources:
        if r.resource_type not in type_groups:
            type_groups[r.resource_type] = []
        type_groups[r.resource_type].append(r)

    for r_type, items in type_groups.items():
        type_node = root.add(f"[cyan]{r_type}[/cyan]")
        for r in items:
            cost = costs.get(r.id, 0.0)
            cost_str = f"[green]${cost:,.2f}/mo[/green]" if cost > 0 else "[dim]-[/dim]"

            # Icon selection
            icon = "📦"
            if "instance" in r_type:
                icon = "💻"
            elif "db" in r_type or "rds" in r_type:
                icon = "🛢️ "
            elif "bucket" in r_type:
                icon = "💾"
            elif "vpc" in r_type or "subnet" in r_type:
                icon = "🕸️ "

            type_node.add(f"{icon} {r.resource_name}  {cost_str}")

    console.print(root)


def print_diff(resources: List[ReliaResource], costs: Dict[str, float]):
    console.print("\n📉 [bold blue]Cost Diff (Estimated)[/bold blue]\n")

    # Simple logic: assume everything is new (+) for now,
    # unless we eventually implement state tracking.

    for r in resources:
        cost = costs.get(r.id, 0.0)
        if cost > 0:
            console.print(f"[green]+ {r.id:<30} +${cost:,.2f}/mo[/green]")
        else:
            console.print(f"[dim]  {r.id:<30}   -[/dim]")

    total = sum(costs.values())
    console.print(f"\n[bold green]+ ${total:,.2f}/mo (Total New Cost)[/bold green]\n")


def generate_markdown_report(
    resources: List[ReliaResource],
    costs: Dict[str, float],
    budget: Optional[float] = None,
) -> str:
    total_cost = sum(costs.values())
    status_emoji = "✅"
    if budget and total_cost > budget:
        status_emoji = "🚨"

    md = f"# {status_emoji} Relia Cost Report\n\n"
    if budget:
        md += f"**Budget Status**: {'Over Budget' if total_cost > budget else 'Within Budget'}\n"
        md += f"**Limit**: `${budget:,.2f}`\n\n"

    md += f"**Total Estimated Cost**: `${total_cost:,.2f}/mo`\n\n"

    md += "## Resource Breakdown\n\n"
    md += "| Resource | Type | Cost/Mo |\n"
    md += "| :--- | :--- | :---: |\n"

    for r in resources:
        cost = costs.get(r.id, 0.0)
        cost_str = f"${cost:,.2f}" if cost > 0 else "-"
        r_type_val = r.attributes.get("instance_type") or r.resource_type
        md += f"| `{r.id}` | `{r_type_val}` | {cost_str} |\n"

    return md
