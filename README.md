# Relia

Policy-gated automation with zero standing secrets. Signed receipts you can hand to security/audit.

Relia is a small gateway that sits in front of **production-changing automation** (CI/CD jobs, scripts, bots) and enforces:

- **Policy checks** (allow / deny / require approval)
- **Optional human approvals** (Slack)
- **Just-in-time credentials** (AWS STS via GitHub OIDC)
- **Signed, tamper-evident action receipts** (verifiable)
- **Audit packs** (`relia pack`) you can attach to an incident or audit ticket

Relia is automation-first: you can use it today for Terraform/deploy/migrations without adopting any agent framework.

## Status (v0.1)

- Auth supports `RELIA_DEV_TOKEN` (local/dev) and GitHub Actions OIDC JWTs (recommended).
- AWS credential minting defaults to a dev broker (placeholder credentials) and can use real AWS STS with `RELIA_AWS_MODE=real`.
- Slack supports interactive approvals (`/v1/slack/interactions`) and outbound approval requests with durable retries (outbox).
- Storage supports SQLite (default) and Postgres, with embedded migrations on startup.

## Why Relia exists

The standard toolchain answers “did it run?”. It does not reliably answer:

- Who approved this production action?
- What policy was in force at the time?
- What identity executed it, with what scope, for how long?
- Can we prove this after the fact, offline, without trusting a log platform?

Relia makes “production action authorization” a first-class record.

## What you get

- Remove long-lived cloud keys from CI (OIDC + STS).
- Require approval for high-risk actions (Slack optional).
- Mint short-lived, scoped AWS credentials only after approval.
- Generate a signed receipt for every decision/action.
- Generate a pack ZIP with checksums + a one-page `summary.html`.
- Verify receipts with `relia verify` (and a hosted verify page, optional).

## The 15-minute wow demo

Terraform apply in prod with zero AWS secrets + 1-click approval + instant audit artifacts.

- Run the demo: `docs/DEMO.md`

## Architecture (simple on purpose)

- `relia-gateway` (Go): REST API + policy eval + approvals + receipt ledger
- `relia-cli`: verify receipts + generate packs
- GitHub composite action: requests GitHub OIDC token → calls `/authorize` → exports AWS creds

Relia does not require a proxy/sidecar in v0.1.

## Quickstart

### Prereqs

- For GitHub Actions: workflow with `permissions: id-token: write`
- For real AWS creds: AWS role configured for GitHub OIDC trust (`docs/AWS_OIDC.md`)
- For Slack approvals: Slack app + bot token + signing secret (`docs/SLACK.md`)
- For stable receipt verification across restarts: configure a signing key (below).

### Run locally (Go, dev token)

```bash
export RELIA_DEV_TOKEN=dev
go run ./cmd/relia-gateway
```

Check it’s running:

```bash
curl -sS http://localhost:8080/healthz
```

### Policy simulator

```bash
go run ./cmd/relia-cli policy test --policy policies/relia.yaml --action terraform.apply --resource stack/prod --env prod
```

### Verify and pack

```bash
export RELIA_DEV_TOKEN=dev
go run ./cmd/relia-cli verify <receipt_id> --token dev
go run ./cmd/relia-cli pack <receipt_id> --out relia-pack.zip --token dev
```

Packs include `summary.html` and `summary.json` for a one-page audit summary.

```bash
unzip -l relia-pack.zip
```

### Hosted verify page (optional)

```bash
export RELIA_PUBLIC_VERIFY=1
go run ./cmd/relia-gateway
```

Then open `http://localhost:8080/verify/<receipt_id>` (and download `http://localhost:8080/pack/<receipt_id>`).

### Run locally (Docker Compose)

```bash
export RELIA_DEV_TOKEN=dev
docker compose -f deploy/docker-compose.yml up --build
```

By default, `deploy/docker-compose.yml` mounts:

- `./keys` → `/app/keys` (read-only) for receipt signing keys
- `./data` → `/app/data` for the SQLite DB (`/app/data/relia.db`)

Generate signing keys (recommended):

```bash
mkdir -p keys data
go run ./cmd/relia-cli keys gen --private keys/ed25519.key --public keys/ed25519.pub
```

## Docs

- `docs/QUICKSTART.md`
- `docs/DEMO.md`
- `docs/PLG.md`
- `docs/SLACK.md`
- `docs/AWS_OIDC.md`
- `docs/POLICIES.md`
- `docs/SECURITY.md`
- `docs/TESTING.md`
- `docs/RELEASE.md`
- `docs/ROADMAP.md`
- `product/spec.md`

## Contributing

- `CONTRIBUTING.md`
- `CODE_OF_CONDUCT.md`
- `LICENSE`
