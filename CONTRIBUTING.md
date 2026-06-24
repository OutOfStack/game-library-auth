# Contributing

## Commit messages and releases

This repository uses Conventional Commits on `main` to decide whether GitHub Actions should create a SemVer Git tag. CI uses [`svu`](https://github.com/caarlos0/svu) to calculate the next tag from Git history.

- `fix: ...` creates a patch tag, for example `v1.0.1`.
- `feat: ...` creates a minor tag, for example `v1.1.0`.
- `feat!: ...` creates a major tag, for example `v2.0.0`.
- A commit body with `BREAKING CHANGE: ...` also creates a major tag.
- `chore:`, `docs:`, `test:`, `refactor:`, and `ci:` do not create a SemVer tag.

Examples:

```text
fix: handle empty health metadata
feat: expose app version in health response
feat!: change infoapi company exists response
chore: regenerate mocks
```

When a merge to `main` creates a SemVer tag, CI builds and deploys the Docker image with that tag. When a merge to `main` does not create a SemVer tag, CI still deploys the image tagged with the commit SHA.
