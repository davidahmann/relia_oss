# Relia

Relia is a small gateway + CLI for policy-gated automation. It evaluates a YAML policy for an action request, optionally requires approval, issues (dev) AWS credentials, and produces signed receipts you can verify and pack.

## Status (v0.1)

- Auth is currently `RELIA_DEV_TOKEN` only (GitHub OIDC verification is not implemented yet).
- AWS credential minting is currently a dev broker (placeholder credentials).
- Slack support currently handles interactive callbacks (`/v1/slack/interactions`) and updates approval state; outbound Slack posting is not implemented yet.

## Quickstart

### Run locally (Go)

```bash
export RELIA_DEV_TOKEN=dev
go run ./cmd/relia-gateway
```

Then call the API with `Authorization: Bearer dev`.

### Run locally (Docker Compose)

```bash
export RELIA_DEV_TOKEN=dev
docker compose -f deploy/docker-compose.yml up --build
```

### Verify and pack

```bash
export RELIA_DEV_TOKEN=dev
go run ./cmd/relia-cli verify <receipt_id>
go run ./cmd/relia-cli pack <receipt_id> --out relia-pack.zip
```

## Docs

- `docs/QUICKSTART.md`
- `docs/SECURITY.md`
- `docs/TESTING.md`
- `product/spec.md`
