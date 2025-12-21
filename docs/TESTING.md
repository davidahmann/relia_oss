---
title: Testing
description: "How to run unit, integration, smoke, E2E, and coverage checks for Relia (target: 85%+)."
keywords: testing, coverage, unit tests, integration tests, e2e, smoke tests, go test
---

# Testing

## Unit + integration tests

Run the full suite (includes coverage gate in CI):

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | tail -n 1
```

## Smoke tests

Smoke tests validate the core HTTP flow (authorize → verify → pack) using an in-process server:

```bash
go test ./tests/smoke -run TestSmoke -count=1
```

## E2E tests

E2E tests are tagged and not run by default:

```bash
./scripts/e2e.sh
```

## Integration E2E (manual)

- Slack approvals: `docs/SLACK.md`
- AWS via GitHub OIDC: `docs/AWS_OIDC.md`

## Benchmarks

```bash
go test ./internal/crypto -run=^$ -bench=BenchmarkCanonicalize -benchmem
go test ./internal/api -run=^$ -bench=BenchmarkAuthorize -benchmem
```
