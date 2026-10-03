# Tasks

## 1. Lint Configuration

- [x] 1.1 Add a committed `golangci-lint` configuration (`.golangci.yml`, schema version 2) scoped to the module; verify the configuration loads without a configuration error when the linter runs.
- [x] 1.2 Pin `golangci-lint` to release `v2.14.0` for CI and record it where a developer can find it (a comment in the config or the README); verify the recorded version matches the version the workflow installs.

## 2. CI Workflow Scaffold

- [ ] 2.1 Create `.github/workflows/ci.yml` with `pull_request` and `push` (main branch) triggers and a single job named `ci`; verify the file is valid YAML and the job runs when a pull request is opened.
- [ ] 2.2 Add a checkout step and an `actions/setup-go` step that reads the Go version from `go.mod` (`go-version-file: go.mod`) with module caching; verify the run reports the same Go version that `go.mod` declares.

## 3. Gate Checks

- [x] 3.1 Add the `go build ./...` step; verify a branch containing a package that does not compile fails the check.
- [x] 3.2 Add the `go vet ./...` step; verify a branch that introduces a vet finding fails the check.
- [x] 3.3 Add the `gofmt` cleanliness step (`test -z "$(gofmt -l .)"`) and the `go mod tidy` drift step (`go mod tidy` followed by `git diff --exit-code go.mod go.sum`); verify an unformatted file and a stale `go.sum` each fail the check.
- [x] 3.4 Add the `go test -race ./...` step; verify a branch with a failing test fails the check and the current suite passes.
- [x] 3.5 Add the `golangci-lint` step using the version pinned in task 1.2; verify a branch that introduces a lint finding within the configured scope fails the check.

## 4. Enforcement

- [ ] 4.1 Confirm the workflow reports a single status check named `ci` and that it is green on the current module; verify by observing the check on a pull request against `main`.
- [ ] 4.2 Add the `ci` check as a required status check on the `main` branch; verify a failing check blocks merge and a passing check allows it.

## 5. Documentation

- [x] 5.1 Update `README.md` to describe the quality gate and how to reproduce it locally (including the pinned lint version from task 1.2); verify the documented commands run as written against the current module.

## 6. Integration Verification

- [ ] 6.1 From a clean checkout, run the full gate and confirm it passes end to end; then on a throwaway branch introduce one failing condition per check (compile error, vet finding, unformatted file, stale module file, failing test, lint finding) and confirm the required `ci` check blocks merge; verify removing them restores a green, mergeable state.
