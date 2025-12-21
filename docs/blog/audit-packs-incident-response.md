---
title: "Audit Packs for Incident Response: One Artifact, Zero Guesswork"
description: "Audit packs turn a receipt into a portable ZIP artifact you can hand to security, auditors, or incident commanders — including policy snapshot, hashes, approvals, and a human summary."
keywords: audit pack, incident response, compliance, evidence, receipt, reproducibility, offline verification
date: "2025-12-21"
---

# Audit Packs for Incident Response: One Artifact, Zero Guesswork

When production changes cause incidents, teams scramble across tools:

- GitHub run logs
- Slack threads
- policy files
- cloud audit trails

Relia’s “pack” concept collapses this into one portable artifact:

> A ZIP you can attach to an incident or audit ticket that contains everything needed to verify authorization offline.

## What’s inside a Relia pack (v0.1)

Relia produces a pack per receipt (downloadable via API or CLI). It includes:

- `receipt.json`: signed, chainable receipt
- `policy.yaml`: the exact policy used (or a snapshot)
- `approvals.json`: approval details (if present)
- `sha256sums.txt`: content hashes
- `summary.html` and `summary.json`: a human-friendly summary you can paste into tickets

The key property is that the pack is **self-contained**.

## Why this matters

Audit and incident workflows break when evidence lives only in systems that:

- require special access
- are mutable
- are hard to export cleanly

Packs let you:

- hand evidence to security without granting broad access
- preserve decision context even if systems change later
- reproduce verification in air-gapped environments

## How to generate a pack

After you have a `receipt_id`:

- `relia verify <receipt_id>`
- `relia pack <receipt_id> --out relia-pack.zip`

Or use the hosted endpoints if enabled:

- `GET /verify/<receipt_id>`
- `GET /pack/<receipt_id>`

See `docs/QUICKSTART.md` and `docs/DEMO.md`.
