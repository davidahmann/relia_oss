# Relia Quickstart (v0.1)

## Local setup

- Install Go 1.24+.
- Copy `relia.yaml` and update paths or environment variables as needed.

## Run the gateway

```bash
go run ./cmd/relia-gateway
```

## Slack approvals (optional)

- `docs/SLACK.md`

## AWS via GitHub OIDC (optional)

- `docs/AWS_OIDC.md`

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
