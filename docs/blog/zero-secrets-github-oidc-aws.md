---
title: Zero Standing Secrets in CI with GitHub OIDC → AWS STS
description: "How Relia uses GitHub’s OIDC identity to mint short-lived AWS credentials, eliminating long-lived AWS keys in CI/CD."
keywords: GitHub OIDC, AWS STS, AssumeRoleWithWebIdentity, zero secrets, CI/CD security, workload identity
date: "2025-12-21"
---

# Zero Standing Secrets in CI with GitHub OIDC → AWS STS

Static cloud keys in CI are a liability:

- they leak (logs, forks, misconfig, vendor breaches)
- they’re hard to rotate safely
- they often end up over-scoped

Relia’s v0.1 design uses **workload identity** instead:

> GitHub Actions obtains an OIDC token → Relia validates it → Relia calls AWS STS → CI receives short-lived credentials.

## The flow

1) Your workflow requests an OIDC token (`id-token: write`).
2) The Relia GitHub Action calls `POST /v1/authorize` and includes:
   - `action`, `env`, `resource`
   - evidence like `plan_digest` and the GitHub run URL
3) Relia evaluates policy:
   - allow / deny / require approval
4) If allowed, Relia calls AWS STS `AssumeRoleWithWebIdentity` using the GitHub OIDC JWT.
5) The action exports AWS creds for the remainder of the job (minutes TTL).
6) Relia emits a signed receipt linking policy + identity + TTL + evidence digests.

## What you configure (users)

In AWS:

- An IAM OIDC provider for `https://token.actions.githubusercontent.com`
- An IAM role with a trust policy that allows your GitHub org/repo/branch (or environment) claims

In Relia:

- A policy rule that chooses the role ARN + TTL for a given action/env/resource

In GitHub Actions:

- `permissions: id-token: write`
- `RELIA_URL` secret pointing to your gateway

## How to test it

The simplest end-to-end test is:

1) Deploy the gateway with real AWS mode enabled (see `docs/AWS_OIDC.md`).
2) Run `examples/github-actions/terraform-prod.yml`.
3) Confirm:
   - the returned AWS creds are not `DEV_*`
   - AWS API calls succeed for the scoped role
   - the receipt references the correct role + TTL + GitHub run context

## Why this is better than “just use OIDC directly”

GitHub OIDC is great — but it does not by itself give you:

- policy evaluation on **intent** (“terraform.apply in prod”)
- optional human approvals
- signed receipts and offline verification
- audit packs you can attach to tickets

Relia turns “OIDC auth” into “authorization with durable evidence”.

