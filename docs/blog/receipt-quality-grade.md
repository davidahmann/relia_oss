---
title: A Simple “Quality Grade” for Authorization Receipts (A–F)
description: "Not all receipts are equally useful in audits. Relia assigns a lightweight A–F grade that quickly signals missing evidence or weak authorization posture."
keywords: receipt quality, auditability, policy hash, approvals, plan digest, security posture, governance
date: "2025-12-21"
---

# A Simple “Quality Grade” for Authorization Receipts (A–F)

Receipts are only valuable if they capture the right evidence.

In practice, teams often start with “it’s approved” and only later realize they’re missing:

- a policy hash
- a plan digest
- approver identity
- a scoped role and short TTL

Relia’s v0.1 “quality grade” is a small UX feature that makes gaps obvious.

## The rubric (lightweight)

This is not a scoring system. It’s a heuristic:

- **A**: approval (if required) + policy hash + plan digest + scoped role + short TTL
- **B**: missing one “nice-to-have” but still strongly auditable
- **C**: approval present but missing key evidence (e.g., no plan digest)
- **D**: weak identity or overly broad scope signals
- **F**: missing policy hash / invalid signature / cannot verify

## Where you see it

Relia surfaces the grade in places that create “product feel”:

- the hosted verify page (`/verify/<receipt_id>`)
- the GitHub Actions step summary (if enabled)
- CLI output (`relia verify <receipt_id>`)

## Why it works

Most “security posture” issues are invisible until an audit.

A grade makes it easy to:

- tighten policy by requiring specific evidence fields
- spot outliers (e.g., a prod apply with a C grade)
- standardize on A/B receipts as the norm

## Next

- See the verify page in the demo: `docs/DEMO.md`
- Verify receipts locally: `docs/QUICKSTART.md`

