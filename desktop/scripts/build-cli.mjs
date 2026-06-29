// Builds the DevPod CLI sidecar binary that the Tauri app bundles as an
// `externalBin`. Tauri expects it at src-tauri/bin/devpod-cli-<rust-host-triple>,
// so we resolve the triple from `rustc`, build the Go CLI for the matching
// GOOS/GOARCH, and copy it into place. Runs before `vite`/`tauri build` so the
// desktop app starts without any manual sidecar steps.
import { execFileSync } from "node:child_process"
import { mkdirSync } from "node:fs"
import { dirname, join, resolve } from "node:path"
import { fileURLToPath } from "node:url"

const desktopDir = resolve(dirname(fileURLToPath(import.meta.url)), "..")
const repoRoot = resolve(desktopDir, "..")
const binDir = join(desktopDir, "src-tauri", "bin")

// Map rust host triples to GOOS/GOARCH for the platforms we ship.
const TRIPLE_TO_GO = {
  "aarch64-apple-darwin": ["darwin", "arm64"],
  "x86_64-apple-darwin": ["darwin", "amd64"],
  "aarch64-unknown-linux-gnu": ["linux", "arm64"],
  "x86_64-unknown-linux-gnu": ["linux", "amd64"],
  "x86_64-pc-windows-msvc": ["windows", "amd64"],
}

function hostTriple() {
  const out = execFileSync("rustc", ["-Vv"], { encoding: "utf8" })
  const match = out.match(/^host:\s*(.+)$/m)
  if (!match) {
    throw new Error("could not determine rust host triple from `rustc -Vv`")
  }

  return match[1].trim()
}

const triple = hostTriple()
const target = TRIPLE_TO_GO[triple]
if (!target) {
  throw new Error(`unsupported rust host triple: ${triple}`)
}
const [goos, goarch] = target

mkdirSync(binDir, { recursive: true })
const isWindows = goos === "windows"
const dest = join(binDir, `devpod-cli-${triple}${isWindows ? ".exe" : ""}`)

console.log(`[cli] building devpod-cli for ${goos}/${goarch} -> ${dest}`)
execFileSync("go", ["build", "-ldflags", "-s -w", "-o", dest, "."], {
  cwd: repoRoot,
  env: { ...process.env, CGO_ENABLED: "0", GOOS: goos, GOARCH: goarch },
  stdio: "inherit",
})
console.log("[cli] sidecar ready")
