# Changelog

All notable changes to this project will be documented in this file.

## [v1.1.0] - 2025-12-05

### 🚀 Major Features

*   **Production Stack (`MERIDIAN_ENV=production`)**: full support for running in production with Postgres (Async) and Redis.
*   **Point-in-Time Correctness**:
    *   Development: Uses DuckDB `ASOF JOIN` logic for zero leakage.
    *   Production: Uses Postgres `LATERAL JOIN` logic for zero leakage.
*   **Async Offline Store**: `PostgresOfflineStore` now uses `asyncpg` for high-throughput I/O.
*   **Hybrid Feature Fixes**: Correctly merges Python (on-the-fly) and SQL (batch) features in retrieval.

### 🐛 Bug Fixes

*   Fixed `AttributeError` in `prod_app.py` regarding `timedelta`.
*   Fixed data loss issue in `PostgresOfflineStore` where Python features were dropped during hybrid retrieval.
*   Fixed type casting issues in `RedisOnlineStore`.

### 📚 Documentation

*   Added [Feast Comparison](feast-alternative.md) page.
*   Added [FAQ](faq.md) page.
*   Added Use Case guides:
    *   [Churn Prediction](use-cases/churn-prediction.md) (PIT Focus)
    *   [Real-Time Recommendations](use-cases/real-time-recommendations.md) (Hybrid Focus)
*   Added Architecture Diagram to README.

## [v1.0.2] - 2025-12-04

### Added
*   Initial support for Hybrid Features (Python + SQL).
