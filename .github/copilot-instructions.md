# Copilot instructions for DevPod

## Build, test, and lint commands

- CLI build: `CGO_ENABLED=0 go build -ldflags "-s -w" -o devpod-cli`
- Repo build helper: `make build` or `./hack/rebuild.sh` (may ask for sudo); build variables include `SKIP_INSTALL`, `BUILD_PLATFORMS`, and `BUILD_ARCHS`.
- Full Go unit test workflow: `./hack/unit-tests.sh`. This runs `go generate ./...`, builds `main.go` with `GOFLAGS=-mod=vendor`, then runs package tests with race and coverage.
- Single Go test/package: `GOFLAGS=-mod=vendor go test ./pkg/git -run TestNormalizeRepository` or `GOFLAGS=-mod=vendor go test ./pkg/devcontainer/config -run TestName`.
- E2E setup: from `e2e/`, run `BUILDDIR=bin SRCDIR=".." ../hack/build-e2e.sh` before tests.
- E2E tests: from `e2e/`, run `go test -v -ginkgo.v -timeout 3600s --ginkgo.label-filter=up-docker` for a label, or `go run github.com/onsi/ginkgo/v2/ginkgo -r --label-filter=build` on Windows-style workflows.
- Go lint: CI runs `golangci-lint run --timeout 30m` with `GOFLAGS=-mod=vendor`.
- Desktop install/build/checks: from `desktop/`, use `yarn install --frozen-lockfile`, `yarn build`, `yarn lint:ci`, `yarn format:check`, and `yarn types:check`.
- Desktop app dev/build: from `desktop/`, use `yarn desktop:dev:debug`, `yarn desktop:build`, or `yarn tauri build --config src-tauri/tauri-dev.conf.json`.
- Rust-only desktop checks may need a CLI resource first: build the CLI, then temporarily copy it to `desktop/src-tauri/bin/devpod-cli-$(rustc -Vv | awk '/host:/ {print $2}')` before running `cargo check` from `desktop/src-tauri/`.
- Docs: from `docs/`, use `yarn start` and `yarn build`.

## High-level architecture

- `main.go` only enters `cmd.Execute()`. CLI command wiring lives in `cmd/root.go`; add Cobra commands through `BuildRoot()` and keep command-specific flags/logic in `cmd/<command>.go` or a `cmd/<domain>/` package.
- Workspace commands resolve a provider/workspace/machine in `pkg/workspace`, then select a concrete client through `pkg/client` interfaces. Implementations are split between direct workspace clients, daemon clients, proxy clients, and machine clients under `pkg/client/clientimplementation`.
- Providers are declarative YAML command adapters. Built-in provider definitions live in `providers/{docker,kubernetes,pro}/provider.yaml` and are embedded by `providers/providers.go`; provider schema and runtime config types live in `pkg/provider`.
- Devcontainer handling is centered in `pkg/devcontainer`: config parsing/substitution is in `pkg/devcontainer/config`, Docker/Compose/Kubernetes execution paths are in sibling files, and missing devcontainer files can be replaced by language-detected defaults.
- The agent path is separate from the local CLI. CLI commands such as `up` invoke agent workflows, configure SSH/IDE integration locally, and then open IDE-specific packages under `pkg/ide`.
- Desktop is a Tauri app: Rust backend modules are in `desktop/src-tauri/src`, React/TypeScript frontend code is in `desktop/src`, and frontend API access is centralized through `desktop/src/client`. Some large payloads are fetched from the local Tauri server at `http://localhost:25842` instead of Tauri `invoke`.
- Documentation is a Docusaurus site under `docs/`; provider marketplace metadata is also reflected in `community.yaml` and docs pages when adding community providers.

## Key conventions

- This repository vendors Go dependencies. Preserve `GOFLAGS=-mod=vendor` when matching CI, and avoid changing vendored code unless explicitly updating dependencies.
- The canonical Go module declares Go 1.26.3; keep GitHub workflow Go versions in sync when changing the module directive.
- Errors are generally returned with context using `fmt.Errorf("...: %w", err)` or `github.com/pkg/errors` wrappers; CLI execution centralizes final logging and process exit handling in `cmd.Execute()`.
- Global CLI flags are created once in `cmd/flags` and passed into commands. Commands should respect `--debug`, `--silent`, `--context`, `--provider`, and `--devpod-home` through the existing `GlobalFlags` plumbing.
- Provider options often use custom JSON-friendly helper types such as `types.StrBool` and `types.StrArray`; reuse those when adding provider YAML-backed fields.
- E2E tests use Ginkgo labels through helper wrappers like `DevPodDescribe("[label] ...")`; choose labels that match the CI matrix (`build`, `ide`, `integration`, `machine`, `machineprovider`, `provider`, `ssh`, `up`, `up-docker`, `up-podman`, `up-docker-compose`, `up-docker-build`, `up-docker-compose-build`, `context`).
- Desktop frontend state favors React Query and context providers under `desktop/src/contexts`; command execution and Tauri interactions should go through typed client classes under `desktop/src/client`.
- Keep Rust/TypeScript event payloads synchronized. `desktop/src/client/client.ts` explicitly notes that channel payload types must match the Rust types.
- Desktop package management uses Yarn 1 (`packageManager: yarn@1.22.19`); do not use npm or pnpm for desktop lockfile changes.
