# Roadmap

Relia is the **runtime enforcement** component in the broader Axiom “Inference Ledger” sequence:

- **Fabra**: Context Record (what the system knew)
- **Lumyn**: Decision Record (what it decided and why)
- **Relia**: Enforcement + signed receipts (ensure high-stakes actions are policy-bound at runtime)
- **Clyra** (future): Proof bundle verification and settlement workflows

## Relia v0.1 (current)

Wedge form factor: **Authorize Gateway + GitHub Action**.

Shipped/targeted capabilities:
- `/v1/authorize` policy-check + idempotency
- Optional Slack interactive approvals
- Receipt issuance and chaining
- `/v1/verify/{receipt_id}` and `/v1/pack/{receipt_id}`
- CLI: `relia verify`, `relia pack`, `relia policy lint`

Notes:
- GitHub Actions OIDC JWT validation is implemented (dev token still supported).
- AWS credential minting defaults to a dev broker placeholder; real AWS STS is supported with `RELIA_AWS_MODE=real`.

## v0.2 (proxy / sidecar)

Add an explicit tool proxy / sidecar for non-CI use cases:
- HTTP egress proxy for agent tools and automation (explicit routing)
- Authorize per outbound action and stamp receipts for SaaS/internal API calls

## v0.3 (SDKs)

Add thin SDKs/wrappers:
- TypeScript + Python helpers to call `/authorize`, attach receipts, and integrate approvals
- Integrations for common agent runtimes (e.g. LangGraph, CrewAI) as optional modules

## Alignment references

- Axiom master plan: `product/axiom.md`
- Wedge and next steps: `product/tech4.md`
- Relia v0.1 spec: `product/spec.md`
