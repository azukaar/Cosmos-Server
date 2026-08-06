# Changelog

All notable changes to cosmos CLI will be documented here.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)  
Versioning: [Semantic Versioning](https://semver.org/spec/v2.0.0.html)  
Tags use the prefix `cli/` (e.g. `cli/v0.1.0`).

---

## [Unreleased]

---

## [0.1.0] - 2026-05-22

### Added
- Full coverage of Cosmos Server REST API via the auto-generated go-sdk
- Command groups: `configure`, `system`, `containers`, `routes`, `users`,
  `api-tokens`, `openid`, `constellation`, `backups`, `cron`, `metrics`, `storage`
- Profile-based authentication: URL in `~/.cosmos/config.yaml`,
  token in OS keychain (macOS Keychain, Linux libsecret, Windows Credential Manager)
- `cosmos configure` interactive wizard with `list`, `use`, `delete` subcommands
- `COSMOS_URL` / `COSMOS_TOKEN` / `COSMOS_PROFILE` env var overrides for CI/headless
- `--profile`, `--url`, `--token` persistent flags for one-off overrides
- `scripts/install.sh` one-liner installer
- CI workflow triggered on changes to `cli/` or `go-sdk/`
- Release workflow: cross-platform binaries (linux/darwin/windows, amd64/arm64)
  triggered by `cli/v*.*.*` tags

[Unreleased]: https://github.com/VHStark/Cosmos-Server/compare/cli/v0.1.0...HEAD
[0.1.0]: https://github.com/VHStark/Cosmos-Server/releases/tag/cli/v0.1.0
