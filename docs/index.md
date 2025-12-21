---
title: Relia Documentation
description: "Relia is a policy-gated authorization gateway for production automation with zero standing secrets, optional Slack approvals, signed receipts, and audit packs."
keywords: relia, policy gate, approvals, receipts, audit pack, github oidc, aws sts, ci/cd security
---

# Relia Documentation

Relia is a small gateway that sits in front of production-changing automation (Terraform applies, deploys, migrations) and enforces:

- policy checks (allow / deny / require approval)
- optional Slack approvals
- just-in-time AWS credentials via GitHub OIDC
- signed, tamper-evident receipts you can verify offline
- audit packs you can attach to incident or audit tickets

## Start here

- `docs/DEMO.md` — the fastest “wow” path (target: < 15 minutes)
- `docs/QUICKSTART.md` — local + Docker + CLI usage

## Integrations

- `docs/AWS_OIDC.md` — GitHub OIDC → AWS STS (real creds)
- `docs/SLACK.md` — Slack approvals (inbound + outbound + retries)

## Reference

- `docs/POLICIES.md` — policy format + templates + simulation
- `docs/TESTING.md` — test matrix and how to run
- `docs/SECURITY.md` — vulnerability reporting and security posture
- `docs/RELEASE.md` — how releases work
- `docs/ROADMAP.md` — what’s next

## Blog

The blog is written for implementers and security/platform teams:

- `docs/blog/receipts-beat-logs.md`
- `docs/blog/zero-secrets-github-oidc-aws.md`
- `docs/blog/slack-approvals-with-retries.md`
- `docs/blog/audit-packs-incident-response.md`
- `docs/blog/receipt-quality-grade.md`
- `docs/blog/policy-simulator-instant-clarity.md`

