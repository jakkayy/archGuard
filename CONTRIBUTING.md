# Contributing to ArchGuard

Thanks for your interest in improving ArchGuard!

## Development setup

```bash
git clone https://github.com/jakkayy/archGuard.git
cd archGuard
go build ./...
go test ./...
```

Requires Go 1.25+ (see `go.mod`).

## Before opening a pull request

```bash
gofmt -s -l .        # must print nothing
go vet ./...
go test -race ./...
go run ./cmd/archguard scan   # ArchGuard must pass on its own repository
```

## Branches & commits

- Open pull requests against `develop`; `main` holds released code.
- Use [Conventional Commits](https://www.conventionalcommits.org/), e.g.
  `feat(rule): add layer-boundary rule` or `fix(reporter): include line region in SARIF`.

## Adding a new rule

1. Implement the `policy.Rule` interface in `pkg/rule/`.
2. Register a factory for it in the rule registry.
3. Add unit tests covering both passing and violating fixtures.
4. Document the rule and its parameters in `README.md`.
