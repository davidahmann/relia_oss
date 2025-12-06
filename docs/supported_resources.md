# Supported Resources & Usage Guide

Relia supports estimating costs for the following AWS resources. Some resources (like EC2) are estimated automatically from Terraform attributes, while others (like S3 or Lambda) require **Usage Overlays** to be accurate.

## 🟢 Auto-Estimated Resources
These resources are estimated purely based on your `.tf` files. Relia extracts the instance type, region, and engine to fetch the exact hourly rate.

| Resource Type | Terraform Resource | Notes |
| :--- | :--- | :--- |
| **EC2 Instances** | `aws_instance` | Supports all families (`t3`, `m5`, `c5`, etc.) and Operating Systems. |
| **RDS Instances** | `aws_db_instance` | Supports Single-AZ and Multi-AZ. Matches engine (`postgres`, `mysql`). |
| **Load Balancers** | `aws_lb`, `aws_elb` | Estimates the fixed hourly cost per load balancer. |
| **NAT Gateways** | `aws_nat_gateway` | Estimates the fixed hourly cost (~$32/mo). **Note:** Data processing fees are not yet estimated. |
| **EBS Volumes** | `aws_ebs_volume` | Estimates storage cost based on `size` (GB) and type (`gp2`, `gp3`, `io1`). |

---

## 🟡 Usage-Based Resources (Requires Overlay)
Some resources have costs that depend heavily on *usage* (e.g., requests, storage used), which cannot be seen in Terraform code. Relia uses a `.relia.usage.yaml` file to apply assumptions.

### 1. AWS Lambda (`aws_lambda_function`)
Lambda costs depend on the number of requests and execution duration.

**Configuration:**
Create or edit `.relia.usage.yaml` in your project root:

```yaml
usage:
  aws_lambda_function.my_api_handler:
    monthly_requests: 1000000   # 1 Million requests
    avg_duration_ms: 200        # Average execution time
```

**Calculation:**
*   **Compute:** `requests * (duration/1000) * (memory/1024) * price_per_gb_second`
*   **Requests:** `(requests / 1M) * $0.20`

---

### 2. S3 Buckets (`aws_s3_bucket`)
S3 costs are purely driven by how much data you store and API requests.

**Configuration:**
```yaml
usage:
  aws_s3_bucket.data_lake:
    storage_gb: 500             # Amount of data stored
    monthly_requests: 10000     # Tier 1 requests (PUT/COPY/POST)
```

---

## ⚠️ Un-Estimated Costs
The following costs are **NOT** currently estimated by Relia (v1.1.1):

*   **Data Transfer (Bandwidth)**: Egress/Ingress fees.
*   **Spot Instances**: All instances are priced as On-Demand.
*   **Reserved Instances / Savings Plans**: Relia assumes On-Demand public pricing.
