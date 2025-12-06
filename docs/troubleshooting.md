---
title: "Troubleshooting Relia: Common Issues & Solutions"
description: "Fix common Relia errors like AWS SSO token expiry, missing resource costs, and installation issues."
keywords: troubleshoot relia, aws sso error, resource not found pricing
---

# Troubleshooting Relia: Common Issues & Solutions

## AWS SSO Token Expiry
If you see errors like:
```text
Error when retrieving token from sso: Token has expired and refresh failed
```
It means your local AWS session is dead.

**Fix:**
Relia will automatically fallback to its **Bundled Pricing Database** or **Local Cache**, so you can typically ignore this warning if you are estimating common resources (e.g., standard EC2 families).
To fix it properly, refresh your credentials:
```bash
aws sso login
```

## "Resource Not Found" in Pricing
If a resource shows cost `-` or `$0.00`:
1.  **Unsupported Type**: We might not support that resource yet (currently focusing on EC2, RDS).
2.  **Complex Filters**: Some specific configurations (e.g. BYOL, Dedicated Hosts) might miss our matcher rules.

Please [open an issue](https://github.com/davidahmann/relia_oss/issues) with your `.tf` snippet!

---
## Related Documentation
- [Quickstart Guide](quickstart.md) - Basic configuration
- [FAQ](faq.md) - Why costs might show as $0.00
- [GitHub Issues](https://github.com/davidahmann/relia_oss/issues) - Report a bug
