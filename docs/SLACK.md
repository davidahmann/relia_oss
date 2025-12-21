---
title: Slack approvals
description: "Configure Slack approvals for Relia: signature verification, interactive callbacks, outbound approval messages, and retry-safe delivery via an outbox."
keywords: slack approvals, interactive messages, signing secret, outbox, retries, idempotency, relia
---

# Slack integration (approvals)

Relia supports Slack-based approvals:

- **Outbound**: when an `/v1/authorize` request requires approval, Relia posts an approval request message to Slack.
- **Inbound**: when an approver clicks approve/deny, Slack sends an interactive callback to Relia at `/v1/slack/interactions`.

Outbound posting is implemented with a **durable outbox** (SQLite: `slack_outbox`, Postgres: `relia_slack_outbox`) so transient Slack failures are retried with backoff.

## What you need

- A Slack app installed into your workspace, with:
  - Interactivity enabled (Request URL points to your gateway’s `/v1/slack/interactions`).
  - A bot token with `chat:write`.
- A reachable gateway URL (public HTTPS for real Slack callbacks).

Use `examples/slack/slack-app-manifest.yml` as a starting point.

## Gateway configuration

Slack is enabled either by config file:

```yaml
slack:
  enabled: true
  signing_secret: "${RELIA_SLACK_SIGNING_SECRET}"
  approval_channel: "C0123456789"
```

…or by env:

- `RELIA_SLACK_ENABLED=true`

In both cases you must set:

- `RELIA_SLACK_SIGNING_SECRET` (required when Slack enabled)
- `RELIA_SLACK_BOT_TOKEN`
- `RELIA_SLACK_APPROVAL_CHANNEL` (channel ID, not name)

Optional:

- `RELIA_SLACK_OUTBOX_WORKER=0` disables the background retry worker.

## Local “E2E” (simulated Slack click)

This validates: enqueue outbox → posting attempt/backoff → approval transition → receipt finalization.
It does **not** validate Slack UI delivery (you need a real token/channel for that).

```bash
export RELIA_DEV_TOKEN=dev
export RELIA_SLACK_ENABLED=true
export RELIA_SLACK_SIGNING_SECRET=test-signing-secret
export RELIA_SLACK_BOT_TOKEN=xoxb-invalid
export RELIA_SLACK_APPROVAL_CHANNEL=C123

export RELIA_DB_DRIVER=sqlite
export RELIA_DB_DSN="file:/tmp/relia-slack-e2e.db?_journal_mode=WAL"

go run ./cmd/relia-gateway -config relia.yaml
```

Then, in another terminal:

```bash
curl -sS -H "Authorization: Bearer dev" -H "Content-Type: application/json" \
  -d '{"action":"terraform.apply","resource":"stack/prod","env":"prod"}' \
  http://localhost:8080/v1/authorize

# Inspect outbox retries (expect last_error=invalid_auth with dummy token)
sqlite3 /tmp/relia-slack-e2e.db \
  "select notification_id,status,attempt_count,next_attempt_at,last_error from slack_outbox;"
```

To simulate a Slack approve click, send an interact request with a valid signature:

```bash
python3 - <<'PY'
import hmac,hashlib,time,urllib.parse,json,os,sys
secret=os.environ["RELIA_SLACK_SIGNING_SECRET"]
approval_id=sys.argv[1]
payload=json.dumps({"actions":[{"action_id":"approve","value":approval_id}]})
body=urllib.parse.urlencode({"payload":payload})
ts=str(int(time.time()))
base=f"v0:{ts}:{body}".encode()
sig="v0="+hmac.new(secret.encode(),base,hashlib.sha256).hexdigest()
print(ts); print(sig); print(body)
PY approval-... > /tmp/slack_req.txt

TS=$(sed -n '1p' /tmp/slack_req.txt)
SIG=$(sed -n '2p' /tmp/slack_req.txt)
BODY=$(sed -n '3p' /tmp/slack_req.txt)

curl -sS -X POST http://localhost:8080/v1/slack/interactions \
  -H "X-Slack-Request-Timestamp: $TS" \
  -H "X-Slack-Signature: $SIG" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  --data "$BODY"
```

After approval, re-run `/v1/authorize` with the same request body to receive the final `allow` response and `receipt_id`.

## Real Slack E2E (recommended)

1. Deploy the gateway with a public HTTPS URL.
2. Set Slack config/env vars above.
3. Update Slack app Interactivity Request URL:
   - `https://<your-gateway>/v1/slack/interactions`
4. Trigger an approval-required request (example above with `env=prod`).
5. Click approve/deny in Slack and observe:
   - `/v1/approvals/<id>` transitions
   - outbox row transitions to `sent` and has `sent_at`

## Demo links (optional)

If you enable public verify:

```bash
export RELIA_PUBLIC_VERIFY=1
```

Slack messages include the `approval_id`; after approval you can share:

- `https://<your-gateway>/verify/<receipt_id>`
- `https://<your-gateway>/pack/<receipt_id>`
