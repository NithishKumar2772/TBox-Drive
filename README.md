<div align="center">

# TBox Drive

### Open-source TeraBox file manager, transfer workspace and self-hosted cloud client

A Windows-first application for managing files stored in your own TeraBox account. Built with Go, Wails, React and SQLite, with a shared transfer engine and a self-hosted browser interface.

[![CI](https://github.com/NithishKumar2772/TBox-Drive/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/NithishKumar2772/TBox-Drive/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/NithishKumar2772/TBox-Drive?display_name=tag)](https://github.com/NithishKumar2772/TBox-Drive/releases)
[![Downloads](https://img.shields.io/github/downloads/NithishKumar2772/TBox-Drive/total)](https://github.com/NithishKumar2772/TBox-Drive/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26.11%2B-00ADD8.svg?logo=go&logoColor=white)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18-61DAFB.svg?logo=react&logoColor=black)](https://react.dev/)

[Download / Releases](https://github.com/NithishKumar2772/TBox-Drive/releases) · [Documentation](docs/) · [Roadmap](docs/ROADMAP.md) · [Issues](https://github.com/NithishKumar2772/TBox-Drive/issues) · [Security](SECURITY.md)

</div>

---

> [!IMPORTANT]
> **Development status:** TBox Drive is not yet a complete multi-platform product. The current repository has a Windows desktop foundation, a Windows self-hosted browser executable and a Windows virtual-drive prototype. macOS, Linux desktop packaging, Android, iOS and TV applications are not complete and must not be treated as downloadable products. No public release is available yet. Follow [GitHub Releases](https://github.com/NithishKumar2772/TBox-Drive/releases) for verified packages when published.

> [!IMPORTANT]
> TBox Drive is an independent project and is not affiliated with Flextech Inc. or TeraBox. It does not provide independent cloud storage or bypass TeraBox account, permission, quota, file-size or service limits. Live use requires an approved TeraBox Open Platform application and credentials issued for that application.

## Contents

- [Current status](#current-status)
- [Platform availability](#platform-availability)
- [What works today](#what-works-today)
- [Screenshots and product areas](#screenshots-and-product-areas)
- [Architecture](#architecture)
- [Quick start](#quick-start)
- [Build on Windows](#build-on-windows)
- [Self-hosted browser mode](#self-hosted-browser-mode)
- [Virtual drive](#virtual-drive)
- [Security and privacy](#security-and-privacy)
- [Known limitations](#known-limitations)
- [Quality checks](#quality-checks)
- [Roadmap and platform order](#roadmap-and-platform-order)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)

## Current status

TBox Drive is an early-stage, TeraBox-first application. The table separates current code from prototypes and planned work.

| Component | Status | What that means |
| --- | --- | --- |
| Go application and TeraBox adapter | Implemented foundation | OAuth device authorization, token refresh, file listing, quota and provider operations are in the repository |
| Credential vault | Implemented foundation | Provider credentials are stored through the OS credential vault; SQLite stores credential references |
| Transfer engine | Implemented foundation | Multipart upload, part verification, resume reconciliation, persisted transfer state and bounded workers |
| Windows desktop | Build verified in CI | Wails desktop executable compiles; live-account and installed-device acceptance is still required |
| Windows self-hosted Web | Build verified in CI | Standalone local browser executable compiles; limited browser file-management client |
| Windows virtual drive | Prototype | WinFsp whole-file cache hydration and upload-on-close; not production streaming/offline sync |
| macOS | Not implemented | No macOS application target or package is currently published |
| Linux desktop | Not implemented | No supported Linux application target or package is currently published |
| Android / iOS | Not implemented | No native mobile application or installable package is currently published |
| Android TV / Google TV / Fire TV | Not implemented | No TV app or remote-navigation acceptance tests are currently published |
| Samsung Tizen / LG webOS | Not implemented | No TV platform apps are currently published |
| Public releases | Not published | The Releases page currently has no production release assets |

A successful GitHub Actions run proves only the steps configured in that run. It does not prove a real TeraBox account works, a physical device has been accepted, or a product is ready for general release.

## Platform availability

**Download only packages that appear as attached assets in a published GitHub Release.** A CI artifact is a temporary build/test output, not a supported public release.

| Platform | Application | Current availability |
| --- | --- | --- |
| Windows x64 | Wails desktop | Build verified in CI; no public release package yet |
| Windows x64 | Self-hosted browser server | Build verified in CI; no public release package yet |
| Windows x64 | CLI | Internal Go command exists; not a separately packaged supported CLI release |
| macOS Intel / Apple Silicon | Desktop | Not implemented |
| Linux x64 / ARM64 | Desktop | Not implemented |
| Android phones / tablets | Native app | Not implemented |
| iOS / iPadOS | Native app | Not implemented |
| Android TV / Google TV / Fire TV | TV app | Not implemented |
| Samsung Tizen / LG webOS | TV app | Not implemented |

[Open TBox Drive Releases](https://github.com/NithishKumar2772/TBox-Drive/releases)

## What works today

### Account and cloud access

- Official TeraBox Open Platform device authorization flow.
- Token refresh and account-specific API domain handling.
- OS credential-vault integration for user access and refresh tokens.
- SQLite persistence for account references and transfer state.
- Paginated file listing and storage quota lookup.

### Transfers

- Multipart upload foundation for large files.
- Verification of uploaded parts and reconciliation of ambiguous failures.
- Persistent transfer tasks, upload sessions and completed-part records.
- Bounded parallel multipart requests with conservative desktop concurrency.
- Streaming downloads in the Go provider layer.
- Desktop upload progress and persisted transfer history.

### Windows desktop

- Wails v2 desktop application with React and TypeScript interface.
- TeraBox device authorization and account management.
- File browsing, search, pagination, quota, upload and download.
- Rename, move and delete operations where supported by the provider API.
- Transfer history and appearance selection.
- Safe cloud move and copy operations from the desktop file explorer.
- Active upload/download cancellation with persisted transfer state and progress reporting.
- Windows WinFsp virtual-drive prototype.

### Self-hosted browser client

- Runs on the user's own Windows computer.
- Serves the browser UI and Go API from a single loopback-only origin.
- Account connection, file/folder browsing, quota, uploads, downloads, rename, delete and transfer history.
- Browser responses omit provider credential references; tokens remain in the Go process and OS credential vault.
- Private startup launch URL exchanged for an HttpOnly, SameSite=Strict session cookie.
- Same-origin mutation checks, Content Security Policy, no-store API responses and cloud-path confinement.

The browser client is an initial local-only foundation. It is not a remote multi-user service or a completed replacement for Google Drive for Desktop.

## Screenshots and product areas

There are no verified product screenshots committed in the current repository yet. Screenshots will be added only from a tested application build, with account details, filenames, paths and all private data removed.

The current interface areas include:

- Windows desktop file manager
- TeraBox account connection and profile
- Upload/download and transfer history
- Self-hosted browser file manager
- Windows virtual-drive prototype

## Architecture

TBox Drive uses one Go application core and one TeraBox provider adapter. Desktop and browser clients use the shared service layer instead of implementing provider authentication and transfer logic independently.

```text
                    TBox Drive
                        |
              Shared Go application core
                        |
        +---------------+----------------+
        |               |                |
   TeraBox adapter  Transfer engine   SQLite storage
        |               |                |
        +---------------+----------------+
                        |
              +---------+---------+
              |                   |
       Windows desktop     Self-hosted Web
         Wails + React      Browser + React
              |                   |
              +--- shared Go -----+
                   services

    Planned: macOS / Linux / Android / iOS / TV
    Planned: byte-range drive / cache / sync / devices
```

**Technology stack**

| Layer | Technology |
| --- | --- |
| Core and provider services | Go |
| Desktop shell | Wails v2 |
| Browser and desktop UI | React, TypeScript, Vite |
| Local metadata | SQLite |
| User secrets | Windows Credential Manager / macOS Keychain / Linux Secret Service adapters |
| Windows virtual-drive prototype | WinFsp |

## Quick start

### Requirements

- Windows 10/11 x64 for the currently build-verified desktop/browser targets.
- Go 1.26.11 or newer.
- Node.js 24 or newer and npm.
- Microsoft Edge WebView2 Runtime for the Wails desktop application.
- WinFsp runtime only if testing the Windows virtual-drive prototype.
- An approved TeraBox Open Platform application with its own Client ID, Client Secret, Private Secret and exact assigned application folder.

Do not use another application's credentials. Never place real credentials in source code, a README, issue, PR or build log.

### Clone

```powershell
git clone https://github.com/NithishKumar2772/TBox-Drive.git
cd TBox-Drive
```

### Configure TeraBox application settings

Supply the values through environment variables in the current PowerShell session:

```powershell
$env:TBOX_TERABOX_CLIENT_ID = "your-client-id"
$env:TBOX_TERABOX_CLIENT_SECRET = "your-client-secret"
$env:TBOX_TERABOX_PRIVATE_SECRET = "your-private-secret"
$env:TBOX_TERABOX_APP_FOLDER = "the-exact-folder-assigned-by-terabox"
```

The application folder must be the exact path assigned by the TeraBox Open Platform. Do not guess it. Account tokens are stored in the OS credential vault after successful authorization.

## Build on Windows

Use the checked-in build script from the repository root:

```powershell
.\scripts\build-windows.ps1
```

The script installs frontend dependencies, builds the React frontend and compiles both Windows applications into `dist/`:

- `TBoxDrive.exe` — Wails desktop application.
- `TBoxDrive-Web.exe` — self-hosted browser server.

The build script stops when a command fails. The current repository does not yet track `frontend/package-lock.json`, so the script falls back to `npm install` with a warning; a lockfile should be committed before a reproducible stable release. WinFsp must be installed separately to test virtual-drive mounting.

### Run the desktop application

```powershell
$env:TBOX_TERABOX_CLIENT_ID = "your-client-id"
$env:TBOX_TERABOX_CLIENT_SECRET = "your-client-secret"
$env:TBOX_TERABOX_PRIVATE_SECRET = "your-private-secret"
$env:TBOX_TERABOX_APP_FOLDER = "the-exact-folder-assigned-by-terabox"
.\dist\TBoxDrive.exe
```

The desktop app reports setup errors in the UI when required configuration is missing. On successful setup, connect using the official device authorization flow.

## Self-hosted browser mode

Start the web executable from PowerShell after configuring the four TeraBox variables:

```powershell
.\dist\TBoxDrive-Web.exe
```

Open the **private launch URL printed in the console**. It contains a random startup token; do not share it. The default listener uses `127.0.0.1:8765`. The server rejects non-loopback bind addresses. Stop it with Ctrl+C.

### Browser feature limits

- Browser uploads are staged to a temporary file before provider upload; temporary free disk space must be sufficient for the file.
- Set `TBOX_UPLOAD_TEMP_DIR` to choose the upload staging directory.
- Browser page-by-page listing, cloud move/copy, upload request progress, persisted provider-side upload/download progress and cancellation of the active upload are implemented. Browser downloads stream directly and can be canceled from the browser's download manager.
- Nested folder upload, browser previews, pause/resume/retry queue controls, in-app download cancellation and full transfer queue management are not yet complete. Browser request progress measures bytes sent to the local service; provider-side transfer progress is shown separately in Transfers.
- Remote/LAN access is intentionally disabled. Do not expose this server using port forwarding, tunnels or reverse proxies.
- This local browser service is single-user and is not a production public Web service.

See [docs/WEB.md](docs/WEB.md) for API details, security behaviour and troubleshooting.

## Virtual drive

The current Windows WinFsp prototype can list TeraBox files, hydrate whole files into a local cache when opened, and upload changes on flush/close. It also supports basic rename, move and delete operations.

**Not yet production-ready:** byte-range streaming, reliable offline pinning, bounded cache eviction, background reconnect, two-way synchronization, conflict-safe writes and full Explorer shell integration remain incomplete. Folder creation is disabled until an officially documented and verified provider method is available.

Install WinFsp before testing. Use a non-critical TeraBox test folder and keep an independent backup of important data.

## Security and privacy

- User access and refresh tokens are stored through the operating-system credential vault. SQLite stores credential references, not provider tokens.
- Refresh tokens are single-use; refresh persistence is serialized within the service and coordinated across desktop/Web processes with a SQLite lease.
- The browser service binds to loopback and requires the private launch-session cookie for API access.
- Mutating browser requests require exact same-origin Origin headers. Cross-site Fetch Metadata is rejected.
- API responses are marked `no-store`; a restrictive Content Security Policy is applied.
- Never log or publish Client Secrets, access/refresh tokens, cookies, launch URLs, account details, private filenames or downloaded content.
- On Linux, a desktop Secret Service provider such as GNOME Keyring or KWallet is required for the credential-vault adapter.
- Per-file encryption, vault password, recovery keys and encrypted cache are not implemented. Do not assume uploaded files are encrypted by TBox Drive.
- The project has not received an independent security audit or live-account production acceptance.

Read [SECURITY.md](SECURITY.md) and [docs/SECURITY.md](docs/SECURITY.md) before using real account data.

## Known limitations

| Area | Limitation |
| --- | --- |
| Provider acceptance | No approved real-account end-to-end acceptance test has been completed |
| Large files | >4 GiB provider acceptance still needs testing against an approved account |
| Empty files | Empty-file upload is not yet supported |
| Folder operations | Folder-tree upload and documented directory creation are incomplete |
| Transfer controls | Unified pause/resume/cancel and queue-recovery UI are incomplete |
| Browser upload | Temporary disk staging is required |
| Virtual drive | Windows whole-file cache prototype; no production byte-range streaming |
| Sync | Mirroring, offline pinning, conflict resolution and two-way sync are incomplete |
| Encryption | Per-file encryption/vault/recovery are not implemented |
| Platforms | macOS, Linux, Android, iOS and TV native applications are not implemented |
| Security | Independent review and real-device acceptance remain outstanding |

See [docs/ROADMAP.md](docs/ROADMAP.md) for the full milestone list.

## Quality checks

The existing main-branch CI validates Go formatting/module files, Go tests, `govulncheck`, vet, build, frontend audit/build and Windows desktop/Web executables. The merged Windows/Web foundation passed all three jobs.\n\nThe release workflow in `.github/workflows/release-windows.yml` adds a separate tag/manual-triggered gate: repeat source validation, build both Windows executables, package a portable ZIP, calculate SHA-256 checksums, and only then publish a clearly labelled Windows early-preview release. It does not generate packages for platforms that have no native implementation. It has been added to this branch but has not yet been tested by a GitHub Actions run.\n\nA passing CI run proves only its configured checks. It does not replace a live TeraBox account acceptance test, WinFsp installed-machine test, manual upgrade test, platform-device testing or an independent security audit.

[View GitHub Actions](https://github.com/NithishKumar2772/TBox-Drive/actions)

## Roadmap and platform order

The planned delivery order is:

1. **Windows desktop + self-hosted Web** — current foundation; complete browser transfer controls, Windows drive reliability and live-account acceptance.
2. **macOS desktop** — native Wails packaging, Keychain validation, Finder/File Provider integration and Intel/Apple Silicon acceptance.
3. **TV clients** — Android TV / Google TV / Fire TV first, then Samsung Tizen and LG webOS with remote-control navigation.
4. **Mobile clients** — Android first, then iOS/iPadOS with platform-specific credential storage, file selection, background-transfer and signing tests.
5. **Device and sync services** — authenticated device registration/revocation, safe cross-device sync and key-sharing design only after threat-model review.

These are goals, not claims of current platform support. Platform completion requires a native implementation, reproducible CI build, installer/package, tests on the actual platform and release notes describing known limitations.

## Documentation

| Document | Purpose |
| --- | --- |
| [Product specification](docs/PRODUCT_SPECIFICATION.md) | Product requirements and design decisions |
| [Roadmap](docs/ROADMAP.md) | Milestones and implementation status |
| [Architecture](docs/ARCHITECTURE.md) | Shared Go application and client boundaries |
| [Transfer engine](docs/TRANSFER_ENGINE.md) | Multipart upload, retries and recovery |
| [TeraBox integration](docs/TERABOX_INTEGRATION.md) | Provider authentication and API details |
| [Provider interface](docs/PROVIDER_INTERFACE.md) | Provider abstraction |
| [Database design](docs/DATABASE.md) | SQLite persistence |
| [Desktop guide](docs/DESKTOP.md) | Wails desktop setup and limitations |
| [Web guide](docs/WEB.md) | Self-hosted browser client and security boundaries |
| [Security design](docs/SECURITY.md) | Credential handling and security assumptions |
| [Release notes](CHANGELOG.md) | Version history |

## Contributing

Before submitting a pull request:

1. Read the applicable product specification and roadmap items.
2. Inspect existing architecture and upstream provider documentation.
3. Add tests for new behaviour and failure cases.
4. Run Go tests, vet/build, frontend audit/build and relevant platform builds.
5. State what is implemented, how it was tested and what remains incomplete.

Never attach credentials, tokens, private file metadata or personal files to an issue or PR. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

TBox Drive is licensed under the [MIT License](LICENSE).

---

<div align="center">

**TBox Drive** — a dependable, multi-device TeraBox workspace, built one verified platform at a time.

</div>
