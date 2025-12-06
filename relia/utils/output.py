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
    resource_count = len(resources)

    # Calculate potential savings if suggestions exist (rough estimate placeholder logic or just count tips)
    optimization_count = sum(1 for r in resources if r.suggestions)

    status_text_color = "#155724" if not budget or total_cost <= budget else "#dc2626"
    status_text = (
        "Within Budget" if not budget or total_cost <= budget else "Over Budget"
    )

    # Premium CSS
    css = """
    :root { --primary: #2563eb; --bg: #f8fafc; --card-bg: #ffffff; --text: #1e293b; --border: #e2e8f0; }
    body { font-family: 'Inter', system-ui, -apple-system, sans-serif; background-color: var(--bg); color: var(--text); padding: 40px 20px; line-height: 1.5; }
    .container { max-width: 1000px; margin: 0 auto; }
    .header { margin-bottom: 30px; border-bottom: 1px solid var(--border); padding-bottom: 20px; }
    h1 { margin: 0; font-size: 1.8rem; display: flex; align-items: center; gap: 10px; }

    .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 40px; }
    .stat-card { background: var(--card-bg); padding: 20px; border-radius: 12px; box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1); border: 1px solid var(--border); }
    .stat-label { font-size: 0.875rem; color: #64748b; font-weight: 500; }
    .stat-value { font-size: 1.5rem; font-weight: 700; margin-top: 5px; color: var(--text); }
    .stat-sub { font-size: 0.75rem; color: #94a3b8; margin-top: 5px; }

    .section-title { font-size: 1.2rem; font-weight: 600; margin-bottom: 15px; margin-top: 40px; }

    .controls { margin-bottom: 15px; display: flex; gap: 10px; }
    input.search { padding: 8px 12px; border: 1px solid var(--border); border-radius: 6px; width: 300px; }

    table { width: 100%; border-collapse: separate; border-spacing: 0; background: var(--card-bg); border-radius: 8px; overflow: hidden; box-shadow: 0 1px 3px rgb(0 0 0 / 0.1); }
    th { background: #f1f5f9; padding: 12px 16px; text-align: left; font-weight: 600; cursor: pointer; user-select: none; }
    th:hover { background: #e2e8f0; }
    td { padding: 12px 16px; border-bottom: 1px solid var(--border); vertical-align: top; }
    tr:last-child td { border-bottom: none; }

    .cost-val { font-family: 'JetBrains Mono', monospace; font-weight: 600; }
    .res-type { font-size: 0.85em; color: #64748b; background: #f1f5f9; padding: 2px 6px; border-radius: 4px; display: inline-block; margin-top: 4px; }

    .tips-box { margin-top: 8px; }
    .tip { display: flex; gap: 8px; font-size: 0.85rem; color: #854d0e; background: #fefce8; padding: 8px; border-radius: 6px; border: 1px solid #fef08a; margin-top: 4px; }

    .mermaid-box { background: var(--card-bg); padding: 20px; border-radius: 12px; border: 1px solid var(--border); overflow-x: auto; }
    """

    # JS for Filtering and Sorting
    script = """
    function filterTable() {
        const query = document.getElementById('search').value.toLowerCase();
        const rows = document.querySelectorAll('tbody tr');
        rows.forEach(row => {
            const text = row.innerText.toLowerCase();
            row.style.display = text.includes(query) ? '' : 'none';
        });
    }

    function sortTable(n) {
        var table, rows, switching, i, x, y, shouldSwitch, dir, switchcount = 0;
        table = document.getElementById("costTable");
        switching = true;
        dir = "asc";
        while (switching) {
            switching = false;
            rows = table.rows;
            for (i = 1; i < (rows.length - 1); i++) {
                shouldSwitch = false;
                x = rows[i].getElementsByTagName("TD")[n];
                y = rows[i + 1].getElementsByTagName("TD")[n];
                // Check if numeric (cost column index 1)
                let xVal = x.innerText.replace(/[$,]/g, "");
                let yVal = y.innerText.replace(/[$,]/g, "");
                let isNum = !isNaN(parseFloat(xVal)) && n === 1;

                if (isNum) {
                    if (dir == "asc") {
                        if (parseFloat(xVal) > parseFloat(yVal)) { shouldSwitch = true; break; }
                    } else {
                        if (parseFloat(xVal) < parseFloat(yVal)) { shouldSwitch = true; break; }
                    }
                } else {
                    if (dir == "asc") {
                        if (x.innerHTML.toLowerCase() > y.innerHTML.toLowerCase()) { shouldSwitch = true; break; }
                    } else {
                        if (x.innerHTML.toLowerCase() < y.innerHTML.toLowerCase()) { shouldSwitch = true; break; }
                    }
                }
            }
            if (shouldSwitch) {
                rows[i].parentNode.insertBefore(rows[i + 1], rows[i]);
                switching = true;
                switchcount ++;
            } else {
                if (switchcount == 0 && dir == "asc") {
                    dir = "desc";
                    switching = true;
                }
            }
        }
    }
    """

    # Generate Topology (Mermaid) - Simple Graph by Type
    # A real topology would inspect dependencies, but grouping by type is a good visual start
    mermaid_graph = "graph TD;\\n"
    mermaid_graph += "    ROOT[Infrastructure];\\n"
    type_groups: Dict[str, List[ReliaResource]] = {}
    for r in resources:
        if r.resource_type not in type_groups:
            type_groups[r.resource_type] = []
        type_groups[r.resource_type].append(r)

    for r_type, items in type_groups.items():
        clean_type = r_type.replace("_", "-")
        mermaid_graph += f"    ROOT --> {clean_type}[{r_type}];\\n"
        for item in items:
            # Escape names for mermaid
            safe_name = item.resource_name.replace('"', "").replace("'", "")
            safe_id = item.id.replace(".", "_").replace("-", "_")
            cost = costs.get(item.id, 0.0)
            label = f"{safe_name}<br/>${cost:,.2f}"
            mermaid_graph += f'    {clean_type} --> {safe_id}("{label}");\\n'
            # Style nodes based on cost?
            if cost > 50:
                mermaid_graph += f"    style {safe_id} fill:#fecaca,stroke:#dc2626;\\n"

    rows_html = ""
    for r in resources:
        cost = costs.get(r.id, 0.0)
        cost_str = f"${cost:,.2f}" if cost > 0 else "-"
        r_type = r.attributes.get("instance_type") or r.resource_type

        tips_section = ""
        if r.suggestions:
            for tip in r.suggestions:
                tips_section += f'<div class="tip">💡 {tip}</div>'
            if tips_section:
                tips_section = f'<div class="tips-box">{tips_section}</div>'

        rows_html += f"""
        <tr>
            <td>
                <div><strong>{r.id}</strong></div>
                <div class="res-type">{r_type}</div>
                {tips_section}
            </td>
            <td class="cost-val">{cost_str}</td>
        </tr>
        """

    return f"""
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Relia Cost Report</title>
        <link rel="preconnect" href="https://fonts.googleapis.com">
        <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@500&display=swap" rel="stylesheet">
        <script type="module">
            import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.esm.min.mjs';
            mermaid.initialize({{ startOnLoad: true, theme: 'neutral' }});
        </script>
        <style>{css}</style>
    </head>
    <body>
        <div class="container">
            <div class="header">
                <h1>📊 Relia Cost Report</h1>
                <div style="color: #64748b; font-size: 0.9rem;">Generated by Relia OSS v1.1.0</div>
            </div>

            <div class="stats-grid">
                <div class="stat-card">
                    <div class="stat-label">Total Monthly Cost</div>
                    <div class="stat-value" style="color: #2563eb">${total_cost:,.2f}</div>
                    <div class="stat-sub" style="color: {status_text_color}; font-weight: 500;">{status_text}</div>
                </div>
                <div class="stat-card">
                    <div class="stat-label">Resources Scanned</div>
                    <div class="stat-value">{resource_count}</div>
                    <div class="stat-sub">Infrastructure Units</div>
                </div>
                <div class="stat-card">
                    <div class="stat-label">Optimization Tips</div>
                    <div class="stat-value" style="color: { '#eab308' if optimization_count > 0 else '#64748b' }">{optimization_count}</div>
                    <div class="stat-sub">Actionable Improvements</div>
                </div>
            </div>

            <div class="section-title">Infrastructure Topology</div>
            <div class="mermaid-box">
                <div class="mermaid">
                {mermaid_graph}
                </div>
            </div>

            <div class="section-title">Cost Breakdown</div>
            <div class="controls">
                <input type="text" id="search" class="search" onkeyup="filterTable()" placeholder="🔍 Filter resources...">
            </div>

            <table id="costTable">
                <thead>
                    <tr>
                        <th onclick="sortTable(0)">Resource ↕</th>
                        <th onclick="sortTable(1)">Cost/Mo ↕</th>
                    </tr>
                </thead>
                <tbody>
                    {rows_html}
                </tbody>
            </table>

            <div style="margin-top: 50px; text-align: center; color: #94a3b8; font-size: 0.8rem;">
                Powered by <strong>Relia OSS</strong> &bull; Open Source Cloud Cost Estimation
            </div>
        </div>

        <script>{script}</script>
    </body>
    </html>
    """
