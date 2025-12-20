# Relia Quickstart (v0.1)

## Local setup

- Install Go 1.22+.
- Copy `relia.yaml` and update paths or environment variables as needed.

## Run the gateway (stub)

```bash
go run ./cmd/relia-gateway
```

## Run the CLI (stub)

```bash
go run ./cmd/relia-cli
```

## Optional pre-commit hook

```bash
ln -s ../../scripts/hooks/pre-commit .git/hooks/pre-commit
```
