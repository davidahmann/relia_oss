#!/usr/bin/env markdown
# One-path demo (v0.1)

Goal: **prod apply with zero secrets + approval + instant audit artifacts**.

## Local demo (single machine)

This runs the gateway locally and simulates a Slack approval click (no Slack workspace required).

### 1) Run the gateway

```bash
export RELIA_DEV_TOKEN=dev
export RELIA_POLICY_PATH=policies/relia.yaml
export RELIA_PUBLIC_VERIFY=1

# Optional: simulate Slack approval flow locally (outbound posts will fail with invalid token, but retries work)
export RELIA_SLACK_ENABLED=true
export RELIA_SLACK_SIGNING_SECRET=test-signing-secret
export RELIA_SLACK_BOT_TOKEN=xoxb-invalid
export RELIA_SLACK_APPROVAL_CHANNEL=C123

go run ./cmd/relia-gateway
```

Wait until `GET /healthz` returns ok:

```bash
curl -sS http://localhost:8080/healthz
```

### 2) Trigger an approval-required request

```bash
curl -sS -H "Authorization: Bearer dev" -H "Content-Type: application/json" \
  -d '{"action":"terraform.apply","resource":"stack/prod","env":"prod","evidence":{"plan_digest":"sha256:deadbeef","diff_url":"https://example.com/diff"}}' \
  http://localhost:8080/v1/authorize
```

Copy `approval.approval_id` from the response.

### 3) Simulate “Approve” (signed Slack interaction)

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

### 4) Re-run authorize to mint creds + final receipt

```bash
curl -sS -H "Authorization: Bearer dev" -H "Content-Type: application/json" \
  -d '{"action":"terraform.apply","resource":"stack/prod","env":"prod","evidence":{"plan_digest":"sha256:deadbeef","diff_url":"https://example.com/diff"}}' \
  http://localhost:8080/v1/authorize
```

Copy `receipt_id` from the response.

### 5) “Exec wow” artifacts

- Verify page: `http://localhost:8080/verify/<receipt_id>`
- Pack download: `http://localhost:8080/pack/<receipt_id>`
- CLI verify:

```bash
go run ./cmd/relia-cli verify <receipt_id> --token dev
```

Unzip the pack and open `summary.html` (one-page audit summary).

## GitHub Actions demo (real OIDC)

1. Deploy the gateway with `RELIA_AWS_MODE=real` (see `docs/AWS_OIDC.md`).
2. Set `RELIA_URL` secret in your repo.
3. Run `examples/github-actions/terraform-prod.yml`.

The action writes a human-readable summary (receipt + grade + links) to the GitHub Actions step summary.

