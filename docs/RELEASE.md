---
title: Release process
description: "How Relia releases are created (tags, changelog, and GitHub releases)."
keywords: release, versioning, changelog, github releases
---

# Release Process

This repository uses lightweight releases driven by Git tags.

## Versioning

- `v0.x.y` (pre-1.0): may include breaking changes in `0.x` minors.
- `v1.0.0+`: semantic versioning applies (breaking changes only on majors).

## Release checklist

1) Ensure `main` is green
- CI passes on GitHub Actions (tests, lint, security checks, coverage >= 85%).
- Smoke workflow passes.

2) Update docs
- Ensure `README.md`, `docs/QUICKSTART.md`, and `docs/SECURITY.md` reflect the release.
- Add an entry to `CHANGELOG.md` under a new version heading.

3) Tag the release

```bash
git pull
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

4) Create GitHub Release
- Use the `CHANGELOG.md` entry as the release notes.
- Attach any build artifacts if you start publishing binaries/images.

## Optional future automation

- Add a GitHub workflow that builds and publishes container images on tag push.
- Add release notes automation (e.g., `release-please` or changelog tooling).
