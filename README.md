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

TBox Drive is an open-source, TeraBox-first cloud file manager and transfer application. Its goal is to make a personal TeraBox account usable through a reliable desktop workspace today, with a shared architecture for future web, mobile, TV, and multi-device clients.

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
| Browser-based web application | Planned |
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
- Browser client using shared Go application services.
- Authorized-device listing, session revocation and cross-device coordination.
- Native Android, iOS and TV clients.

Planned items are not claims of current functionality. Their implementation and verification are tracked in the roadmap.

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
