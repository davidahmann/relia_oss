---
title: "Slack Approvals That Don’t Flake: Signatures, Idempotency, and Retries"
description: "Approvals are useless if they’re unreliable. This post explains how Relia verifies Slack signatures, prevents duplicate approvals, and retries outbound messages safely."
keywords: Slack approvals, interactive messages, signing secret, idempotency, retries, outbox pattern, reliability
date: "2025-12-21"
---

# Slack Approvals That Don’t Flake: Signatures, Idempotency, and Retries

Human approval is a high-signal control — but only if it’s **reliable**.

Relia integrates Slack approvals with three constraints:

1) **Authenticity**: only Slack can trigger approve/deny.
2) **Idempotency**: retries cannot create duplicates.
3) **Delivery**: transient Slack API failures shouldn’t stall or corrupt the authorization state.

## 1) Authenticity: Slack signature verification

Slack signs interactive requests. Relia verifies:

- `X-Slack-Request-Timestamp` is recent
- `X-Slack-Signature` matches the body using your signing secret

If verification fails, the request is rejected and no state changes.

## 2) Idempotency: approvals are single-shot

Slack can retry interaction callbacks.

Relia ensures that:

- approve/deny for an `approval_id` is idempotent
- repeat clicks don’t append multiple approvals to the receipt chain

The receipt remains stable and verifiable.

## 3) Delivery: outbound retries (outbox pattern)

Outbound Slack posting has a different failure mode: the user might never see the approval card.

Relia uses an **outbox-like queue** to:

- persist an “approval message to send” record
- attempt delivery to Slack
- retry with backoff if Slack is down or rate-limiting
- dedupe on `(approval_id, channel)` so retries don’t spam

The result: approval is reliable without turning Slack into a single point of failure.

## How to test locally

You can test the approval flow without a Slack workspace:

- Run the gateway (`docs/DEMO.md`).
- Trigger an approval-required authorize call.
- Simulate a signed Slack interaction payload for approve/deny.

To test outbound posting and retries for real, configure `SLACK_BOT_TOKEN` and `RELIA_SLACK_APPROVAL_CHANNEL` and use a real Slack app (see `docs/SLACK.md`).
