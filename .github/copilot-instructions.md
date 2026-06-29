# Copilot instructions for DevPod

## Build, test, and lint commands

- CLI build: `CGO_ENABLED=0 go build -ldflags "-s -w" -o devpod-cli`
- Local release-like install: `go install -ldflags "-s -w -X github.com/loft-sh/devpod/pkg/version.version=v0.0.1" .`; plain `go install .` leaves `version.GetVersion()` at `v0.0.0` and uses the fork's latest release agent URL instead of a versioned release URL.
- Repo build helper: `make build` or `./hack/rebuild.sh` (may ask for sudo); build variables include `SKIP_INSTALL`, `BUILD_PLATFORMS`, and `BUILD_ARCHS`.
- Full Go unit test workflow: `./hack/unit-tests.sh`. This runs `go generate ./...`, builds `main.go` with `GOFLAGS=-mod=vendor`, then runs package tests with race and coverage.
- Single Go test/package: `GOFLAGS=-mod=vendor go test ./pkg/git -run TestNormalizeRepository` or `GOFLAGS=-mod=vendor go test ./pkg/devcontainer/config -run TestName`.
- E2E setup: from `e2e/`, run `BUILDDIR=bin SRCDIR=".." ../hack/build-e2e.sh` before tests.
- E2E tests: from `e2e/`, run `go test -v -ginkgo.v -timeout 3600s --ginkgo.label-filter=up-docker` for a label, or `go run github.com/onsi/ginkgo/v2/ginkgo -r --label-filter=build` on Windows-style workflows.
- CI runs a vendored Go CLI build plus focused Go tests; release tags build CLI assets for `scaryrawr/devpod` GitHub releases and publish them with `softprops/action-gh-release`.
- Release CLI builds use `CGO_ENABLED=0`; `zig cc`/`zig c++` are not involved. Windows named pipe code uses the already-vendored `github.com/Microsoft/go-winio`, so Windows amd64 and arm64 CLI assets can be cross-compiled without `gopkg.in/natefinch/npipe.v2`.
- Desktop install/build/checks: from `desktop/`, use `pnpm install --frozen-lockfile`, `pnpm build`, `pnpm lint:ci`, `pnpm format:check`, and `pnpm types:check`.
- Desktop app dev/build: from `desktop/`, use `pnpm desktop:dev:debug`, `pnpm desktop:build`, or `pnpm tauri build --config src-tauri/tauri-dev.conf.json`.
- The desktop app bundles the Go CLI as a Tauri `externalBin` sidecar at `desktop/src-tauri/bin/devpod-cli-<rust-host-triple>`. `pnpm desktop:dev`/`desktop:build` auto-build it via `pnpm cli:build` (`desktop/scripts/build-cli.mjs`); a Rust-only `cargo check` still needs it built manually (run `pnpm cli:build`). Without it the build fails with `resource path bin/devpod-cli-... doesn't exist`. `bin/` is gitignored except `.gitkeep`.
- Desktop frontend CLI flags must match the current CLI: command flags are sent literally and an unknown flag aborts the call. After Pro removal there is no `--skip-pro`; `client.workspaces.listAll()` must not pass it.
- Docs: from `docs/`, use `yarn start` and `yarn build`.

## High-level architecture

- `main.go` only enters `cmd.Execute()`. CLI command wiring lives in `cmd/root.go`; add Cobra commands through `BuildRoot()` and keep command-specific flags/logic in `cmd/<command>.go` or a `cmd/<domain>/` package.
- Workspace commands resolve a provider/workspace/machine in `pkg/workspace`, then select a concrete client through `pkg/client` interfaces. Implementations are split between direct workspace clients, daemon clients, proxy clients, and machine clients under `pkg/client/clientimplementation`.
- Providers are declarative YAML command adapters. Built-in provider definitions live in `providers/{docker,kubernetes}/provider.yaml` and are embedded by `providers/providers.go`; provider schema and runtime config types live in `pkg/provider`.
- Devcontainer handling is centered in `pkg/devcontainer`: config parsing/substitution is in `pkg/devcontainer/config`, Docker/Compose/Kubernetes execution paths are in sibling files, and missing devcontainer files can be replaced by language-detected defaults.
- The agent path is separate from the local CLI. CLI commands such as `up` invoke agent workflows, configure SSH/IDE integration locally, and then open IDE-specific packages under `pkg/ide`.
- Agent binary downloads default to `https://github.com/scaryrawr/devpod/releases/...`; `DEVPOD_AGENT_URL` can override this for local testing or alternate release locations.
- Desktop is a Tauri app: Rust backend modules are in `desktop/src-tauri/src`, React/TypeScript frontend code is in `desktop/src`, and frontend API access is centralized through `desktop/src/client`. Some large payloads are fetched from the local Tauri server at `http://localhost:25842` instead of Tauri `invoke`.
- Documentation is a Docusaurus site under `docs/`; provider marketplace metadata is also reflected in `community.yaml` and docs pages when adding community providers.

## Key conventions

- This repository vendors Go dependencies. Preserve `GOFLAGS=-mod=vendor` when matching CI, and avoid changing vendored code unless explicitly updating dependencies.
- The canonical Go module declares Go 1.26.3; keep GitHub workflow Go versions in sync when changing the module directive.
- Errors are generally returned with context using `fmt.Errorf("...: %w", err)` or `github.com/pkg/errors` wrappers; CLI execution centralizes final logging and process exit handling in `cmd.Execute()`.
- Logging goes through the internal `github.com/loft-sh/devpod/pkg/log` package and its subpackages; do not import `github.com/loft-sh/log`. Logger format methods are for printing only, so use `%v` for logged errors and reserve `%w` for returned/wrapped errors.
- Global CLI flags are created once in `cmd/flags` and passed into commands. Commands should respect `--debug`, `--silent`, `--context`, `--provider`, and `--devpod-home` through the existing `GlobalFlags` plumbing.
- Provider options often use custom JSON-friendly helper types such as `types.StrBool` and `types.StrArray`; reuse those when adding provider YAML-backed fields.
- Devcontainer spec changes should be checked against https://containers.dev/implementors/json_reference/ across both `devcontainer.json` and `devcontainer.metadata` image-label paths; add focused tests for string/array/object forms and merge/default behavior.
- E2E tests use Ginkgo labels through helper wrappers like `DevPodDescribe("[label] ...")`; choose labels that match the CI matrix (`build`, `ide`, `integration`, `machine`, `machineprovider`, `provider`, `ssh`, `up`, `up-docker`, `up-podman`, `up-docker-compose`, `up-docker-build`, `up-docker-compose-build`, `context`).
- Desktop frontend state favors React Query and context providers under `desktop/src/contexts`; command execution and Tauri interactions should go through typed client classes under `desktop/src/client`.
- Keep Rust/TypeScript event payloads synchronized. `desktop/src/client/client.ts` explicitly notes that channel payload types must match the Rust types.
- Desktop package management uses pnpm (`packageManager: pnpm@11.9.0`) on Node 24 (`desktop/.nvmrc`); do not use npm or yarn for desktop lockfile changes. pnpm 11 reads project settings only from `desktop/pnpm-workspace.yaml` (not the `package.json` `pnpm` field, and only auth/registry from `.npmrc`): `overrides` replaces yarn `resolutions`, dependency build scripts must be approved via `allowBuilds` (e.g. `esbuild: true`), and `savePrefix: ""` pins exact versions. pnpm 11 also defaults `minimumReleaseAge` to 1440 (1-day supply-chain delay) — keep it unless a freshly published, already-locked dep must be allowed.
