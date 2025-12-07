import sqlite3
import os

db_path = "relia/core/bundled_pricing.db"

if not os.path.exists(db_path):
    print(f"DB not found at {db_path}")
    exit(1)

conn = sqlite3.connect(db_path)
# Update t3.large (approx $0.0832/hr -> ~$60/mo)
# The pricing logic in pricing.py returns hourly price.
# Usage.py multiplies by 730 hours.
# 0.0832 * 730 = 60.73
conn.execute(
    "UPDATE pricing SET price_data = '{\"price\": 0.0832}' WHERE cache_key LIKE '%t3.large%'"
)

# Update m5.large (approx $0.096/hr -> ~$70/mo)
# 0.096 * 730 = 70.08
conn.execute(
    "UPDATE pricing SET price_data = '{\"price\": 0.096}' WHERE cache_key LIKE '%m5.large%'"
)

conn.commit()
conn.close()
print("Bundled DB Updated Successfully")
