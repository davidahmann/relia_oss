import sqlite3
import json
from pathlib import Path
from typing import Optional, Dict, Any
import time


class PricingCache:
    """
    Local SQLite cache for AWS pricing data.
    Schema:
    - sku (TEXT PRIMARY KEY)
    - price_data (JSON)
    - timestamp (REAL)
    """

    DB_PATH = Path.home() / ".relia" / "pricing_cache.db"
    TTL_SECONDS = 86400 * 7  # 7 days

    def __init__(self):
        self._init_db()

    def _init_db(self):
        self.DB_PATH.parent.mkdir(parents=True, exist_ok=True)

        # Check if we have a bundled DB (e.g. adjacent to this file)
        # In a real build, we'd copy this to the package dir
        bundled_db = Path(__file__).parent / "bundled_pricing.db"

        if not self.DB_PATH.exists() and bundled_db.exists():
            import shutil

            shutil.copy2(bundled_db, self.DB_PATH)

        with sqlite3.connect(self.DB_PATH) as conn:
            conn.execute(
                """
                CREATE TABLE IF NOT EXISTS pricing (
                    cache_key TEXT PRIMARY KEY,
                    price_data TEXT,
                    timestamp REAL
                )
            """
            )
            conn.commit()

    def get(self, cache_key: str) -> Optional[Dict[str, Any]]:
        """Retrieve from cache if not expired."""
        with sqlite3.connect(self.DB_PATH) as conn:
            cursor = conn.execute(
                "SELECT price_data, timestamp FROM pricing WHERE cache_key = ?",
                (cache_key,),
            )
            row = cursor.fetchone()

            if row:
                data_str, timestamp = row
                if time.time() - timestamp < self.TTL_SECONDS:
                    return json.loads(data_str)
                else:
                    # Cleanup expired
                    conn.execute(
                        "DELETE FROM pricing WHERE cache_key = ?", (cache_key,)
                    )
                    conn.commit()
            return None

    def set(self, cache_key: str, data: Dict[str, Any]):
        """Save to cache."""
        with sqlite3.connect(self.DB_PATH) as conn:
            conn.execute(
                "INSERT OR REPLACE INTO pricing (cache_key, price_data, timestamp) VALUES (?, ?, ?)",
                (cache_key, json.dumps(data), time.time()),
            )
            conn.commit()
