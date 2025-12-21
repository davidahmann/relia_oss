---
title: Quickstart (v0.1)
description: "Run Relia locally, simulate policies and approvals, and verify/pack receipts with the CLI."
keywords: relia quickstart, policy simulator, slack approvals, verify receipt, audit pack
---

# Relia Quickstart (v0.1)

## Local setup

- Install Go 1.24+.
- Copy `relia.yaml` and update paths or environment variables as needed.

## One-path demo

- `docs/DEMO.md`

## Run the gateway

```bash
go run ./cmd/relia-gateway
```

## Signing keys (recommended)

For stable, offline-verifiable receipts across gateway restarts, configure a persistent Ed25519 signing key:

```bash
mkdir -p keys
go run ./cmd/relia-cli keys gen --private keys/ed25519.key --public keys/ed25519.pub
```

Then set in `relia.yaml`:

- `signing_key.private_key_path: "./keys/ed25519.key"`

## Policy simulator

```bash
go run ./cmd/relia-cli policy test --policy policies/relia.yaml --action terraform.apply --resource stack/prod --env prod
```

## Slack approvals (optional)

- `docs/SLACK.md`

## AWS via GitHub OIDC (optional)

- `docs/AWS_OIDC.md`

## Policy templates

- `docs/POLICIES.md`

## Hosted verify page (optional)

Enable public verify/pack endpoints:

```bash
RELIA_PUBLIC_VERIFY=1 go run ./cmd/relia-gateway
```

Then:

- `GET /verify/<receipt_id>` renders a human-friendly receipt view (includes a quality grade).
- `GET /pack/<receipt_id>` downloads a pack zip (includes `summary.html` + `summary.json`).

## Run the gateway (Docker)

```bash
RELIA_DEV_TOKEN=dev docker compose -f deploy/docker-compose.yml up --build
```

## Run the CLI

```bash
go run ./cmd/relia-cli --help
```

## CLI verify and pack

```bash
RELIA_DEV_TOKEN=dev go run ./cmd/relia-cli verify <receipt_id>
RELIA_DEV_TOKEN=dev go run ./cmd/relia-cli pack <receipt_id> --out relia-pack.zip
RELIA_DEV_TOKEN=dev go run ./cmd/relia-cli policy lint policies/relia.yaml
```

## GitHub Action example

Use the composite action in `.github/actions/relia-authorize` and the example
workflow in `examples/github-actions/terraform-prod.yml`. Make sure the workflow
has `id-token: write` permissions and set `RELIA_URL` in repo secrets.

## Optional pre-commit hook

```bash
ln -s ../../scripts/hooks/pre-commit .git/hooks/pre-commit
```
