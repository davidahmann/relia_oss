---
title: Security
description: "Security reporting and guidelines for Relia, including vulnerability disclosure and recommended hardening."
keywords: security policy, vulnerability disclosure, oss security, relia
---

# Security Policy

## Reporting a Vulnerability

If you believe you have found a security vulnerability, please report it
privately by emailing:

security@relia.dev

Include as much detail as possible to help us understand and reproduce the
issue. We will acknowledge receipt and work with you to resolve it.

## Supported Versions

Only the latest release line is supported with security updates.

## Secrets and configuration

Relia is designed to run without committing secrets to the repo. Prefer
environment variables and secret managers for:

- `RELIA_DEV_TOKEN` (dev-only; do not use in production)
- `RELIA_SLACK_SIGNING_SECRET`
- `RELIA_SLACK_BOT_TOKEN`

The gateway can also load `relia.yaml` (supports `${ENV_VAR}` expansion) via:

- `RELIA_CONFIG_PATH` (or `relia-gateway --config /path/to/relia.yaml`)
