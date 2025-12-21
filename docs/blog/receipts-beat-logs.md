---
title: Why Signed Receipts Beat Logs for Production Automation
description: "Logs answer “did it run?” Receipts answer “was it authorized?” Learn why tamper-evident, offline-verifiable receipts are the missing artifact for audits and incident response."
keywords: receipts, audit trail, signed logs, tamper evident, incident response, compliance, CI/CD security
date: "2025-12-21"
---

# Why Signed Receipts Beat Logs for Production Automation

Production automation (Terraform applies, deploys, database migrations) usually leaves behind **logs**. Logs are useful, but they’re not a durable proof of authorization.

Relia’s thesis is simple:

> For production-changing automation, you need a **receipt** — a cryptographically signed, immutable artifact that can be verified offline.

## Logs are not authorization evidence

Logs typically can’t answer, reliably and later:

- Who approved this change?
- What policy was in force at the time?
- What identity was used, with what scope, for how long?
- Can we prove this without trusting the logging platform?

Even if you have “approval logs” in Slack/GitHub/Jira, those systems don’t naturally produce a **single, portable artifact** with integrity guarantees.

## What a receipt is (in Relia)

A Relia receipt is a signed record that binds together:

- **The request intent**: `action`, `resource`, `env`
- **Evidence digests**: e.g. `plan_digest`, `diff_url`
- **The policy snapshot**: `policy_id`, `policy_version`, `policy_hash`
- **Approval data** (optional): who approved, when, and what they approved
- **The minted credentials metadata**: role, TTL, expiration (not secrets)

The receipt is signed (Ed25519) and chainable, so any tampering is detectable.

## The key property: offline verification

If you can’t verify evidence offline, audits become trust exercises.

With Relia you can do:

- `relia verify <receipt_id>`
- Validate signature + integrity + chain
- Recompute the policy hash and compare

No “trust our logs” story required.

## The practical difference in an incident

During an incident, the question is rarely “did Terraform run?”

It’s:

- Was this prod action *authorized*?
- Who clicked approve (or did it bypass approval)?
- Was this action in-policy at the time?

Receipts turn that into a single artifact you can attach to an incident ticket — and verify without privileged access.

## Next

- Follow the “15-minute wow demo”: `docs/DEMO.md`
- See pack artifacts: `docs/QUICKSTART.md`

