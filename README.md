# Relia

Relia is a small gateway + CLI for policy-gated automation. It evaluates a YAML policy for an action request, optionally requires approval, issues (dev) AWS credentials, and produces signed receipts you can verify and pack.

## Status (v0.1)

- Auth supports `RELIA_DEV_TOKEN` and GitHub Actions OIDC JWTs.
- AWS credential minting defaults to a dev broker (placeholder credentials) and can use real AWS STS with `RELIA_AWS_MODE=real`.
- Slack supports interactive approvals (`/v1/slack/interactions`) and outbound approval requests with durable retries (outbox).
- Storage supports SQLite (default) and Postgres, with embedded migrations on startup.

## Quickstart

### Run locally (Go)

```bash
export RELIA_DEV_TOKEN=dev
go run ./cmd/relia-gateway
```

Then call the API with `Authorization: Bearer dev`.

### Policy simulator

```bash
go run ./cmd/relia-cli policy test --policy policies/relia.yaml --action terraform.apply --resource stack/prod --env prod
```

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

Packs include `summary.html` and `summary.json` for a one-page audit summary.

### Hosted verify page (optional)

```bash
export RELIA_PUBLIC_VERIFY=1
go run ./cmd/relia-gateway
```

Then open `http://localhost:8080/verify/<receipt_id>` (and download `http://localhost:8080/pack/<receipt_id>`).

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
