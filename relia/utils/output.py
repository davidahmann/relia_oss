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

    # Active Advisor Section
    all_tips = []
    for r in resources:
        if r.suggestions:
            for tip in r.suggestions:
                all_tips.append(f"[yellow]• {r.id}: {tip}[/yellow]")

    if all_tips:
        console.print("\n💡 [bold yellow]Active Advisor Tips[/bold yellow]")
        for tip in all_tips:
            console.print(tip)
        console.print("")


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


def generate_html_report(
    resources: List[ReliaResource],
    costs: Dict[str, float],
    budget: Optional[float] = None,
) -> str:
    total_cost = sum(costs.values())
    status_color = "green"
    status_text = "Within Budget"
    if budget and total_cost > budget:
        status_color = "red"
        status_text = "Over Budget"

    # Minimal CSS for a clean look
    css = """
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; max-width: 900px; margin: 40px auto; padding: 20px; color: #333; }
    h1 { color: #2c3e50; border-bottom: 2px solid #eee; padding-bottom: 10px; }
    .card { background: #f8f9fa; padding: 20px; border-radius: 8px; margin-bottom: 20px; border: 1px solid #e9ecef; }
    table { width: 100%; border-collapse: collapse; margin-top: 20px; }
    th, td { text-align: left; padding: 12px; border-bottom: 1px solid #ddd; }
    th { background-color: #f8f9fa; font-weight: 600; }
    tr:hover { background-color: #f1f1f1; }
    .cost { font-family: monospace; font-weight: bold; }
    .tip { background-color: #fff3cd; color: #856404; padding: 10px; margin-top: 5px; border-radius: 4px; border-left: 4px solid #ffeeba; font-size: 0.9em; }
    .footer { margin-top: 40px; font-size: 0.8em; color: #777; text-align: center; }
    """

    rows = ""
    for r in resources:
        cost = costs.get(r.id, 0.0)
        cost_str = f"${cost:,.2f}" if cost > 0 else "-"
        r_type = r.attributes.get("instance_type") or r.resource_type

        tips_html = ""
        if r.suggestions:
            tips_html = '<div class="tips">'
            for tip in r.suggestions:
                tips_html += f'<div class="tip">💡 {tip}</div>'
            tips_html += "</div>"

        rows += f"""
        <tr>
            <td>
                <strong>{r.id}</strong><br>
                <span style="color:#666; font-size:0.85em">{r_type}</span>
            </td>
            <td class="cost">{cost_str}</td>
        </tr>
        """
        if tips_html:
            rows += f'<tr><td colspan="2" style="border-bottom:none; padding-top:0;">{tips_html}</td></tr>'

    budget_html = ""
    if budget:
        budget_html = f'<p><strong>Budget Limit:</strong> ${budget:,.2f} (<span style="color:{status_color}">{status_text}</span>)</p>'

    html = f"""
    <!DOCTYPE html>
    <html>
    <head>
        <title>Relia Cost Report</title>
        <style>{css}</style>
    </head>
    <body>
        <h1>📊 Relia Cost Report</h1>

        <div class="card">
            <h2>Total Estimated Cost: <span style="color: #28a745">${total_cost:,.2f}/mo</span></h2>
            {budget_html}
            <p><small>Generated either locally or via CI/CD</small></p>
        </div>

        <h3>Resource Breakdown</h3>
        <table>
            <thead>
                <tr>
                    <th>Resource</th>
                    <th>Cost/Month</th>
                </tr>
            </thead>
            <tbody>
                {rows}
            </tbody>
        </table>

        <div class="footer">
            Generated by <strong>Relia OSS</strong>
        </div>
    </body>
    </html>
    """
    return html
