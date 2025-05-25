# Contributing to go-term-latex

## Getting Started

### Prerequisites

- Go 1.26+
- One TeX backend for integration tests: `texlive-latex-base` + `poppler-utils`
  (Linux), MacTeX (macOS), or `tectonic`.
- Unit tests (no TeX) run without any backend.

### Setup

```bash
git clone https://github.com/floatpane/go-term-latex.git
cd go-term-latex
go mod tidy
```

### Build & Test

```bash
go build ./...
go test ./...                       # unit tests only, no TeX needed
go test -tags integration ./...     # requires a TeX backend in PATH
```

### Linting

```bash
gofmt -l .
go vet ./...
golangci-lint run
```

## Making Changes

### Branch Naming

- `feature/` — new functionality
- `fix/` — bug fixes
- `docs/` — documentation
- `refactor/` — no behavior change

### Commit Messages

[Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): short description
```

Examples:

```
feat(backend): add miktex backend
fix(render): handle pdftoppm -1.png suffix on older poppler
docs: add physics package example to README
```

### Before Submitting a PR

1. `gofmt -l .` is clean.
2. `go vet ./...` is clean.
3. `go test ./...` passes.
4. If you add or change a backend, test against a real TeX installation and note
   the backend version and OS in the PR description.
5. Keep PRs focused — one logical change per PR.

### Security-sensitive changes

If you modify the shell-out commands (`exec.Command`) or the document template
(`buildDoc`), re-read [SECURITY.md](SECURITY.md) and note in the PR that you've
considered injection and file-read risks.

## AI Policy

We welcome AI-assisted contributions. Contributors are fully responsible for any
code they submit.

**What we expect:** understand what you submit, review AI output carefully, no
AI-generated issues or reviews, no tests that don't actually test anything.

**What we won't accept:** bulk AI refactors, hallucinated APIs, or contributions
where the author clearly doesn't understand the changes.

## Code of Conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md).
