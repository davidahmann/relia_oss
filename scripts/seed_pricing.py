import sqlite3
import json
import time


def seed_db():
    conn = sqlite3.connect("relia/core/bundled_pricing.db")
    conn.execute(
        """
        CREATE TABLE IF NOT EXISTS pricing (
            cache_key TEXT PRIMARY KEY,
            price_data TEXT,
            timestamp REAL
        )
    """
    )

    # Add a mock price for t3.large used in demo so it works offline
    mock_price = {"price": 60.0}  # Approx $0.0822 * 730
    key = "AmazonEC2|TERM_MATCH:capacitystatus:Used|TERM_MATCH:instanceType:t3.large|TERM_MATCH:location:US East (N. Virginia)|TERM_MATCH:operatingSystem:Linux|TERM_MATCH:preInstalledSw:NA|TERM_MATCH:serviceCode:AmazonEC2|TERM_MATCH:tenancy:Shared"

    conn.execute(
        "INSERT OR REPLACE INTO pricing (cache_key, price_data, timestamp) VALUES (?, ?, ?)",
        (key, json.dumps(mock_price), time.time()),
    )

    mock_price_db = {"price": 70.0}
    key_db = "AmazonEC2|TERM_MATCH:capacitystatus:Used|TERM_MATCH:instanceType:m5.large|TERM_MATCH:location:US East (N. Virginia)|TERM_MATCH:operatingSystem:Linux|TERM_MATCH:preInstalledSw:NA|TERM_MATCH:serviceCode:AmazonEC2|TERM_MATCH:tenancy:Shared"

    conn.execute(
        "INSERT OR REPLACE INTO pricing (cache_key, price_data, timestamp) VALUES (?, ?, ?)",
        (key_db, json.dumps(mock_price_db), time.time()),
    )

    conn.commit()
    conn.close()
    print("Seeded relia/core/bundled_pricing.db")


if __name__ == "__main__":
    seed_db()
