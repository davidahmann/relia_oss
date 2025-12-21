---
title: PLG checklist (Relia v0.1)
description: "A product-led growth checklist for Relia’s v0.1 wedge: generate shareable audit artifacts (receipts, verify page, packs) with fast time-to-value."
keywords: plg, product led growth, devtool adoption, audit artifacts, receipts
---

# PLG checklist (Relia v0.1)

Relia’s PLG wedge is **“produce an audit artifact people want to share”**:

- A receipt you can verify
- A one-page summary (`summary.html`)
- A shareable verify page (`/verify/<receipt_id>`, optional)
- A GitHub Actions step summary with links

## Brutal MVP test (Relia-specific)

A first-time user should be able to:

1. Run `relia-gateway` locally (no account, no calls).
2. Run one sample request that requires approval.
3. Complete approval (real Slack or simulated).
4. See a final `receipt_id`, open `/verify/<receipt_id>`, and download `/pack/<receipt_id>`.

Target: **time-to-first-receipt < 15 minutes**.

## What “self-serve” means here

- No demos required to understand: `docs/DEMO.md` is a single-path walkthrough.
- Outputs explain themselves: grade + links + pack summary.
- The GitHub Action is the “installer”: drop-in, no secrets, uses OIDC.

## Growth surfaces (v0.1)

- GitHub Actions step summary (receipt + verify/pack links)
- Pack zip attached as an artifact by the user’s workflow (optional)
- Hosted verify page link shared in PRs/incidents (optional)

## Instrumentation (later)

For OSS, start with lightweight, opt-in telemetry or log-based metrics:

- count authorize calls / verdicts
- approval latency distribution
- pack downloads
