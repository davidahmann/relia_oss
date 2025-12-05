# FAQ

## Does Relia read my state file?
No. Relia parses your **code** (`.tf` files) or your **plan** (`terraform show -json`). It does not connect to your remote state bucket.

## Does it support Azure/GCP?
Not in v0.1.0. We are laser-focused on AWS for the MVP. GCP support is planned for v0.2.

## How accurate is it?
*   **EC2/RDS**: >99% accuracy for compute costs.
*   **Data Transfer/Storage**: Variable inputs (like "GB stored") are hard to guess from code alone. We default to base costs and allow overrides.

## Is there a SaaS version?
No. Relia is 100% open-source and self-hosted (local/CI).
