#!/usr/bin/env markdown
# AWS via GitHub OIDC (real STS)

Relia can mint real AWS credentials by calling `AssumeRoleWithWebIdentity` using the GitHub Actions OIDC JWT presented to `/v1/authorize`.

## What you need

- A public, reachable Relia gateway URL (for GitHub Actions to call).
- An AWS IAM role that trusts GitHub Actions OIDC and allows `sts:AssumeRoleWithWebIdentity`.
- A policy rule that returns an `aws_role_arn` for the target action/env.

## Gateway configuration

Enable real STS on the gateway:

- `RELIA_AWS_MODE=real`
- Provide a region via config `aws.sts_region_default` or env `AWS_REGION`

Example `relia.yaml`:

```yaml
aws:
  sts_region_default: us-east-1
```

If `RELIA_AWS_MODE` is not `real`, Relia returns **placeholder** credentials (`DEV_*`) for local development.

## AWS IAM setup (high level)

1. Ensure your AWS account has an OIDC provider for GitHub Actions:
   - Provider URL: `https://token.actions.githubusercontent.com`
2. Create an IAM role with a trust policy that limits:
   - `aud` to `relia` (Relia’s default audience)
   - `sub` to your repo and branch/environment constraints
3. Attach a permissions policy to the role for what the workflow should do.

## GitHub Actions setup

1. Set a repo secret `RELIA_URL` to your deployed gateway base URL, e.g. `https://relia.example.com`.
2. Ensure your workflow has:

```yaml
permissions:
  id-token: write
  contents: read
```

3. Use the composite action in `.github/actions/relia-authorize`. Example:
   - `examples/github-actions/terraform-prod.yml`

The action:
- fetches a GitHub OIDC JWT with audience `relia`
- calls `POST $RELIA_URL/v1/authorize` with `Authorization: Bearer <jwt>`
- exports returned AWS creds as `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`

## How to test end-to-end

In GitHub Actions:

1. Trigger `examples/github-actions/terraform-prod.yml` (or copy it into `.github/workflows/`).
2. Confirm the “Sanity check AWS creds” step:
   - fails if credentials are `DEV_*`
   - succeeds and prints `aws sts get-caller-identity` output for real creds

Locally (partial):

- You can validate request/receipt flow with `RELIA_AWS_MODE=dev` (dev creds) using `go test ./tests/smoke -run TestSmoke`.
- Full real-ST S validation requires the GitHub Actions OIDC runtime + AWS trust configuration.
