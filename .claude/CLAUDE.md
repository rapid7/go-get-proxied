# go-get-proxied

A Go package (plus a small CLI in `main.go`) that retrieves the system-configured proxy for a given protocol and target URL on Windows, Linux and macOS.

This file is a navigation guide for agents working on the repo. For usage, CLI flags and the per-OS lookup order, read `README.md` and `proxy/doc.go`.

## Repository purpose and boundaries

- This repo is a public, BSD-3-licensed library. Downstream code depends on the exported `proxy` API (`Provider`, `Proxy`, `NewProvider`, `NewProxy`, the URL helpers) and on the CLI output format, so treat both as a contract.
- Module path is `github.com/rapid7/go-get-proxied`. Dependencies are vendored under `vendor/` for hermetic builds.

## Where to look

| Area | Path |
|------|------|
| CLI entry point (`-p`, `-c`, `-t`, `-j`, `-v`, `-l`) | `main.go` |
| Public interfaces, shared lookup logic | `proxy/provider.go` — `Provider` interface, config file (`readConfigFileProxy`, `unmarshalProxyConfigFile`), env vars (`readSystemEnvProxy`, `isProxyBypass`) |
| Per-OS providers and `NewProvider` | `proxy/provider_windows.go`, `proxy/provider_linux.go`, `proxy/provider_darwin.go` (selected by file-name suffix) |
| Proxy value type, JSON output | `proxy/proxy.go` |
| URL parsing helpers | `proxy/url.go` — `ParseURL`, `ParseTargetURL`, `SplitHostPort`, `IsLoopbackHost` |
| WinHTTP syscall wrappers (WPAD, PAC, IE settings, `netsh winhttp`) | `winhttp/winhttp.go` (Windows only) |
| Package docs / lookup order | `proxy/doc.go`, `README.md` |
| Service descriptor / owners | `cortex.yaml`, `.github/CODEOWNERS` |

## How lookup works

- Each OS provider checks sources in a fixed order (configuration file, then environment variables, then OS settings) and returns the first match. The order is documented in `README.md` and `proxy/doc.go`; keep both in sync with the code.
- The configuration file is JSON mapping protocol to proxy URL (e.g. `{"https":"http://host:8999"}`); keys are lower-cased on read and the file is capped at 1 MiB.
- Environment variables: `<PROTOCOL>_PROXY` (SOCKS maps to `ALL_PROXY`), with `NO_PROXY` / `no_proxy` bypass handling.
- macOS reads network settings by running `scutil`; Windows uses WinHTTP via `winhttp/`, with timeouts set by `SetTimeouts`.
- Every `Proxy` records its `Src()` (e.g. `ConfigurationFile`, `Environment[...]`, `WinHTTP:...`), which the CLI includes in JSON output.

## How to approach changes

- Failures return `nil` (no proxy) and are logged via the standard `log` package, not returned as errors. Keep that behaviour for callers.
- Keep code testable: environment and process access go through the `getEnv` (`getEnvAdapter`) and `proc` (`commandAdapter`) fields on `provider`, which tests replace with fakes. New OS calls should follow the same pattern.
- WinHTTP resources must be released (`closeHandle`, `freeWinHttpResource`, `Free()`); don't leak handles or buffers.
- When adding a dependency, update `go.mod` / `go.sum` and re-vendor (`go mod vendor`).

## Build and test entry points

- Build: `go build ./...` (CLI: `go build -o go-get-proxied .`).
- Unit tests: `go test ./...`. Tests use `github.com/stretchr/testify`.
- Tests are OS-specific: `provider_windows_test.go`, `provider_darwin_test.go` and `winhttp/winhttp_test.go` only run on their own OS. Running on one OS does not cover the others; don't claim cross-platform tests passed unless run on each platform.
- Static checks: `go vet ./...`.

## Sources of truth

- `README.md` — installation, library and CLI usage, lookup order.
- `proxy/*_test.go` — the most reliable reference for expected behaviour.
