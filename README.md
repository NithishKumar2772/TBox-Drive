# TBox Drive

TBox Drive is an open-source, TeraBox-first cloud file manager and transfer application. The long-term goal is a provider-independent multi-cloud client with a reliable resumable transfer engine.

## Current backend and desktop capabilities
- Official TeraBox OAuth device authorization, polling and token refresh
- OS credential vault integration for user tokens
- Paginated file listing, storage quota and streaming downloads
- Multipart uploads with MD5 verification, server-side resume discovery and recovery after ambiguous part failures
- SQLite persistence for accounts, transfer tasks, upload sessions and completed parts
- Wails desktop shell with React UI for account connection, file browsing, upload/download and transfer history

Live account use requires TeraBox-issued application credentials and an approved Open Platform application. The current desktop build is an early foundation, not yet a virtual drive or synchronization client.

Windows virtual-drive mounting requires the separately installed [WinFsp runtime](https://winfsp.dev/rel/). The first prototype hydrates complete files on open; byte-range streaming and production-grade offline synchronization are still in development.

See [docs/PRODUCT_SPECIFICATION.md](docs/PRODUCT_SPECIFICATION.md) for the confirmed product scope and [docs/ROADMAP.md](docs/ROADMAP.md) for implementation status.

## Technology
- Go backend and transfer engine
- Wails v2 desktop shell
- React + TypeScript + Vite
- SQLite persistence
- Windows Credential Manager / macOS Keychain / Linux Secret Service
- MIT license

## Development
Requirements: Go 1.26.11+, Node.js 24+.

```sh
go test ./...
go vet ./...
go build ./...
cd frontend && npm install && npm run build
```

For the Windows desktop executable, see [docs/DESKTOP.md](docs/DESKTOP.md).

## Security
Never commit TeraBox client credentials, user tokens, refresh tokens or cookies. User tokens are stored in the OS credential vault; SQLite stores only credential references. On Linux, a Secret Service provider such as GNOME Keyring or KWallet must be available.
