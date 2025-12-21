---
title: Policies and templates
description: "Relia policies match (action, env, resource) and return an effect: allow/deny/require approval plus role + TTL. Includes templates and a simulator."
keywords: relia policy, policy templates, approval policy, authorization rules, terraform apply
---

# Policy templates

Relia policies are simple YAML: they match `(action, env, resource)` and return an effect:

- `require_approval` (bool)
- `ttl_seconds` (int)
- `aws_role_arn` (string)
- `risk` / `reason` (strings)

## Templates

See `policies/templates/`:

- `policies/templates/terraform-prod-apply.yaml`
- `policies/templates/db-migration-prod.yaml`
- `policies/templates/deploy-prod-main.yaml`

## Simulate a policy

```bash
go run ./cmd/relia-cli policy test \
  --policy policies/templates/terraform-prod-apply.yaml \
  --action terraform.apply \
  --resource stack/prod \
  --env prod
```
