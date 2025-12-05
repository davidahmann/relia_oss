# Why We Built Relia

## The "Bill Shock" Loop

1.  **Engineer** adds a NAT Gateway in Terraform to fix a connectivity issue.
2.  **Code Review** focuses on logic and security. No one calculates the cost.
3.  **Deploy** happens automatically.
4.  **30 Days Later**: Finance asks why the AWS bill jumped by $2,000/mo.
5.  **Panic**: The team scrambles to investigate, identify the NAT Gateway, and tear it down.

This loop is painful, expensive, and wasteful.

## The Missing Link

We have linters for style (`black`).
We have linters for security (`tfsec`).
**We needed a linter for cost.**

That's Relia. A simple, fast, pre-deploy check that asks: *"Can we afford this change?"*
