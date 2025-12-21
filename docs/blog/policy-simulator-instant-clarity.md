---
title: "Policy Simulator: Instant Clarity Before You Ship a Rule"
description: "A fast feedback loop for security and platform teams: simulate a production action locally and see the matched rule, verdict, approval requirement, role, and TTL."
keywords: policy simulator, policy testing, shift left, authorization, CI/CD, governance
date: "2025-12-21"
---

# Policy Simulator: Instant Clarity Before You Ship a Rule

Policy mistakes are expensive:

- rules that are too permissive
- rules that are too strict and block releases
- rules that are ambiguous (“it depends”)

Relia includes a policy simulator command so you can test policy **before** it blocks prod.

## What it does

Given an action intent (`action`, `env`, `resource`, plus optional evidence), the simulator prints:

- matched rule ID
- verdict (allow/deny/require approval)
- required approval or not
- role ARN and TTL (if applicable)
- any missing required evidence

## Example

```bash
go run ./cmd/relia-cli policy test \
  --policy policies/relia.yaml \
  --action terraform.apply \
  --env prod \
  --resource stack/prod \
  --intent '{"change_id":"CHG-1234","plan_digest":"sha256:..."}'
```

## Why it matters (PLG)

This is the “time-to-value” lever:

- engineers see the policy decision instantly
- security gets deterministic evidence without meetings
- policy changes become testable artifacts, not tribal knowledge

If you want adoption to be bottom-up, you need tools that make the product legible in minutes.
