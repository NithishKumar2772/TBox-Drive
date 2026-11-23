<div align="center">

# TBox Drive

### An open-source TeraBox cloud file manager and transfer workspace

Manage your TeraBox files with a desktop application, resumable transfers, integrity checks, and a shared Go transfer engine.

[![CI](https://github.com/NithishKumar2772/TBox-Drive/actions/workflows/ci.yml/badge.svg)](https://github.com/NithishKumar2772/TBox-Drive/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26%2B-00ADD8.svg?logo=go&logoColor=white)](https://go.dev/)
[![Frontend](https://img.shields.io/badge/Frontend-React%20%2B%20TypeScript-3178C6.svg?logo=typescript&logoColor=white)](frontend/)

[Repository](https://github.com/NithishKumar2772/TBox-Drive) · [Issues](https://github.com/NithishKumar2772/TBox-Drive/issues) · [Roadmap](docs/ROADMAP.md) · [Product specification](docs/PRODUCT_SPECIFICATION.md)

</div>

---

## Overview

TBox Drive is an open-source, TeraBox-first cloud file manager and transfer application. Its goal is to make a personal TeraBox account usable through a reliable Windows desktop application and a self-hosted browser workspace, with a shared architecture for future macOS, mobile, TV, and multi-device clients.

The project focuses on dependable large-file transfers, recovery after interruptions, file management, and eventually a virtual drive with streaming, mirroring, and offline access.

> [!IMPORTANT]
> TBox Drive is an independent project and is not affiliated with Flextech Inc. or TeraBox. Live account use requires valid TeraBox-issued application credentials and an approved Open Platform application. The application does not provide cloud storage independently of TeraBox or bypass TeraBox account and service limits.

## Project status

TBox Drive is under active development. The current repository contains working backend, transfer, and desktop foundations. Some capabilities below are prototypes or planned work, not production-ready features.

| Area | Current status |
| --- | --- |
| Go backend and provider foundation | Implemented |
| TeraBox OAuth device authorization and token refresh | Implemented |
| OS credential vault integration | Implemented |
| File listing, quota and streaming downloads | Implemented |
| Multipart upload, checksum verification and resume recovery | Implemented |
| Bounded multipart concurrency | Implemented; conservative two-worker desktop setting |
| SQLite transfer and upload-session persistence | Implemented |
| Wails + React desktop foundation | Implemented |
| Windows virtual-drive integration | Prototype; requires WinFsp |
| Production byte-range virtual-drive streaming | Planned |
| Mirroring, two-way sync and offline pinning | Planned |
| Self-hosted loopback browser client and Go API | Initial implementation; local-only, under development |
| Device registration and session management | Planned |
| Android, iOS and TV applications | Planned |

See [the detailed roadmap](docs/ROADMAP.md) for milestone-level progress and acceptance criteria.

## Features

### Account and cloud access
- Official TeraBox OAuth device authorization, polling and token refresh.
- Secure storage of user credentials through the operating system credential vault.
- Paginated remote file listing and storage quota information.
- Provider adapter architecture designed to keep cloud-provider code separate from the shared application core.

### Transfers
- Multipart upload foundation for large files.
- MD5 verification for uploaded parts.
- Resume discovery and recovery after ambiguous part-upload failures.
- Persistent transfer tasks, upload sessions and completed-part records in SQLite.
- Bounded parallel multipart uploads, with a conservative default of two workers in the desktop adapter.
- Streaming downloads and transfer progress reporting.

### Browser application (self-hosted foundation)
- Run the Go web service on the user's own Windows PC and open it at http://127.0.0.1:8765.
- Loopback-only binding; LAN and public network access are rejected.
- Same-origin checks and browser security headers on the HTTP API.
- Browser file listing, quota, uploads, downloads, rename, delete and transfer history.
- OAuth device authorization; provider tokens remain in the OS credential vault and are not returned to browser JavaScript.
- This is an early implementation, not a production security certification; remote device access is not enabled.

### Desktop application
- Wails desktop shell with React and TypeScript frontend.
- Account connection and file browsing.
- Upload and download workflows.
- Transfer history and persistent transfer state.
- Windows virtual-drive prototype using WinFsp.

### Planned workspace capabilities
- Production virtual drive with on-demand byte-range streaming.
- Dynamic cache management, offline pinning and mirroring.
- Safe two-way synchronization and conflict resolution.
- Expanded self-hosted browser client using shared Go application services.
- Authorized-device listing, session revocation and cross-device coordination.
- Native Android, iOS and TV clients.

**Platform delivery order:** Windows desktop + self-hosted Web → macOS → TV (Android TV / Google TV / Fire TV, Samsung Tizen, LG webOS) → Android and iOS mobile.\n\nPlanned items are not claims of current functionality. Their implementation and verification are tracked in the roadmap.

## Architecture

The project is organized around a shared Go core and provider-specific adapters. Desktop and future clients should use the same transfer and application services instead of maintaining separate upload implementations.

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
              Desktop application
                 Wails + React

       Planned clients: Web / Android / iOS / TV
       Planned services: Device sessions / Sync
```

For implementation details, see:
- [Architecture](docs/ARCHITECTURE.md)
- [Provider interface](docs/PROVIDER_INTERFACE.md)
- [Transfer engine](docs/TRANSFER_ENGINE.md)
- [Database design](docs/DATABASE.md)
- [TeraBox integration](docs/TERABOX_INTEGRATION.md)
- [Desktop setup](docs/DESKTOP.md)

## Download and installation

There is currently no production release advertised in this README. Use the repository's [GitHub Actions](https://github.com/NithishKumar2772/TBox-Drive/actions) and [Releases](https://github.com/NithishKumar2772/TBox-Drive/releases) pages for verified builds when they are published.

### TeraBox account requirements

Live account access requires:
1. A TeraBox Open Platform application.
2. Valid application credentials issued for that application.
3. Approval and permissions required by the TeraBox Open Platform.
4. A TeraBox account authorized through the official device authorization flow.

Do not use credentials copied from another application or commit credentials to the repository.

## Build from source

### Prerequisites

- **Go:** 1.26.11 or newer.
- **Node.js:** 24 or newer.
- **npm:** compatible with the installed Node.js version.
- **Windows virtual-drive prototype:** install the [WinFsp runtime](https://winfsp.dev/rel/) if you want to test drive mounting.

### Clone the repository

```bash
git clone https://github.com/NithishKumar2772/TBox-Drive.git
cd TBox-Drive
```

### Run backend checks

```bash
go mod download
go test ./...
go vet ./...
go build ./...
```

### Build the frontend

```bash
cd frontend
npm install
npm run build
```

### Build both Windows applications\n\nFrom PowerShell, run:\n\n```powershell\n.\\scripts\\build-windows.ps1\n```\n\nThe script builds the Wails desktop app and the self-hosted browser executable into the repository's `dist` directory.\n\n### Run the self-hosted browser client

Set the TeraBox Open Platform values in the environment. Do not put real credentials in source files or shell history.

**PowerShell:**

```powershell
$env:TBOX_TERABOX_CLIENT_ID = "your-client-id"
$env:TBOX_TERABOX_CLIENT_SECRET = "your-client-secret"
$env:TBOX_TERABOX_PRIVATE_SECRET = "your-private-secret"
$env:TBOX_TERABOX_APP_FOLDER = "/From: Other Applications/TBox-Drive"
go run ./cmd/tboxdrive
```

Open the private launch URL printed in the console. It contains a random startup token; do not share it. The server refuses non-loopback bind addresses. Use Ctrl+C to stop it. The same environment variables are used by the desktop app.

See [docs/WEB.md](docs/WEB.md) for the browser API, security boundary and troubleshooting.

### Build the desktop application

See [docs/DESKTOP.md](docs/DESKTOP.md) for platform-specific Wails setup and build instructions.

## Continuous integration and quality checks

The GitHub Actions workflow validates the project with:
- Go formatting and module verification.
- Go unit and integration tests.
- Go vulnerability scanning.
- `go vet` and `go build`.
- Frontend dependency audit and production build.
- Windows desktop application build.

A passing CI run verifies the checks in that workflow. It does not replace live TeraBox account acceptance, manual platform testing, or an independent security audit.

## Security and privacy

- Never commit TeraBox application secrets, access tokens, refresh tokens, cookies, or session data.
- User credentials are stored through the operating system credential vault; SQLite stores credential references rather than provider secrets.
- On Linux, a Secret Service provider such as GNOME Keyring or KWallet must be available.
- Use official TeraBox authorization flows. Do not collect a user's TeraBox password through an imitation login form.
- Redact tokens, account identifiers and private file information from logs and bug reports.
- Report security vulnerabilities privately and follow [SECURITY.md](SECURITY.md).

Read the [security documentation](docs/SECURITY.md) before testing with real account data.

## Documentation

| Document | Description |
| --- | --- |
| [Product specification](docs/PRODUCT_SPECIFICATION.md) | Product scope, requirements and platform goals |
| [Roadmap](docs/ROADMAP.md) | Milestones, implementation status and pending work |
| [Architecture](docs/ARCHITECTURE.md) | Application structure and component boundaries |
| [Transfer engine](docs/TRANSFER_ENGINE.md) | Upload, download, retry and recovery design |
| [TeraBox integration](docs/TERABOX_INTEGRATION.md) | Authentication and provider API integration |
| [Provider interface](docs/PROVIDER_INTERFACE.md) | Provider abstraction and extension points |
| [Database design](docs/DATABASE.md) | Local persistence and data model |
| [Desktop guide](docs/DESKTOP.md) | Desktop development and build instructions |
| [Web guide](docs/WEB.md) | Self-hosted browser client setup, API and security boundary |
| [Security](docs/SECURITY.md) | Credential handling and security requirements |

## Contributing

Contributions, bug reports and feature proposals are welcome.

Before opening a pull request:
1. Read the relevant product specification and roadmap items.
2. Inspect existing code and follow the current architecture.
3. Add or update tests for behavior changes.
4. Run the Go and frontend checks relevant to your changes.
5. Describe what was implemented, how it was tested, and any limitations.

For bugs, include reproducible steps and sanitized logs. Never attach credentials, tokens, cookies, private account data, or personal files.

## License

TBox Drive is licensed under the [MIT License](LICENSE).

---

<div align="center">

**TBox Drive** — building a dependable, multi-device cloud file workspace.

</div>
