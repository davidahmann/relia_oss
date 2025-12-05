from pathlib import Path
from relia.core.pricing import PricingClient


def seed_db():
    print("🚀 Starting Bundled DB Expansion...")

    # Target the bundled DB file explicitly
    db_path = Path("relia/core/bundled_pricing.db").resolve()
    print(f"📦 Target DB: {db_path}")

    # Initialize client with that DB
    # Note: We use us-east-1 for the API client itself, but can query for other locations
    client = PricingClient(region="us-east-1", cache_path=str(db_path))

    # Configuration
    regions = ["US East (N. Virginia)", "US West (Oregon)", "EU (Frankfurt)"]

    instance_types = [
        "t3.nano",
        "t3.micro",
        "t3.small",
        "t3.medium",
        "t3.large",
        "t3.xlarge",
        "t3.2xlarge",
        "m5.large",
        "m5.xlarge",
        "m5.2xlarge",
        "m5.4xlarge",
        "c5.large",
        "c5.xlarge",
        "c5.2xlarge",
        "r5.large",
        "r5.xlarge",
    ]

    rds_classes = [
        "db.t3.micro",
        "db.t3.small",
        "db.t3.medium",
        "db.m5.large",
        "db.m5.xlarge",
        "db.r5.large",
    ]

    count = 0
    errors = 0

    print("--- 💻 Seeding EC2 Instances ---")
    for region in regions:
        print(f"  Region: {region}")
        for instance in instance_types:
            filters = [
                {"Type": "TERM_MATCH", "Field": "serviceCode", "Value": "AmazonEC2"},
                {"Type": "TERM_MATCH", "Field": "location", "Value": region},
                {"Type": "TERM_MATCH", "Field": "instanceType", "Value": instance},
                {"Type": "TERM_MATCH", "Field": "operatingSystem", "Value": "Linux"},
                {"Type": "TERM_MATCH", "Field": "preInstalledSw", "Value": "NA"},
                {"Type": "TERM_MATCH", "Field": "tenancy", "Value": "Shared"},
                {"Type": "TERM_MATCH", "Field": "capacitystatus", "Value": "Used"},
            ]
            price = client.get_product_price("AmazonEC2", filters)
            if price:
                count += 1
                # print(f"    ✅ {instance}: ${price}")
            else:
                errors += 1
                print(f"    ❌ {instance}: Failed")

    print("--- 🛢️ Seeding RDS Instances ---")
    for region in regions:
        print(f"  Region: {region}")
        for rds in rds_classes:
            filters = [
                {"Type": "TERM_MATCH", "Field": "serviceCode", "Value": "AmazonRDS"},
                {"Type": "TERM_MATCH", "Field": "location", "Value": region},
                {"Type": "TERM_MATCH", "Field": "instanceType", "Value": rds},
                {
                    "Type": "TERM_MATCH",
                    "Field": "databaseEngine",
                    "Value": "PostgreSQL",
                },  # MVP: Assume Postgres for now
                {
                    "Type": "TERM_MATCH",
                    "Field": "deploymentOption",
                    "Value": "Single-AZ",
                },  # MVP
            ]
            price = client.get_product_price("AmazonRDS", filters)
            if price:
                count += 1
            else:
                errors += 1
                print(f"    ❌ {rds}: Failed")

    print("--- 💾 Seeding EBS Volumes ---")
    # EBS Volume logic (simplified for MVP seeding)
    volume_map = {
        "gp2": "General Purpose",
        "gp3": "General Purpose gp3",
        "io1": "Provisioned IOPS",
        "standard": "Magnetic",
    }

    for region in regions:
        print(f"  Region: {region}")
        for code, api_name in volume_map.items():
            filters = [
                {"Type": "TERM_MATCH", "Field": "serviceCode", "Value": "AmazonEC2"},
                {"Type": "TERM_MATCH", "Field": "location", "Value": region},
                {"Type": "TERM_MATCH", "Field": "productFamily", "Value": "Storage"},
                {"Type": "TERM_MATCH", "Field": "volumeType", "Value": api_name},
            ]
            price = client.get_product_price("AmazonEC2", filters)
            if price:
                count += 1
            else:
                errors += 1
                print(f"    ❌ {code}: Failed")

    print(f"\n✨ Done! Seeded {count} prices. Errors: {errors}")
    print(f"📂 Database Updated: {db_path}")


if __name__ == "__main__":
    seed_db()
