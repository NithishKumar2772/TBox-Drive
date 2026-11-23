# TBox Drive

TBox Drive is an open-source, TeraBox-first cloud file manager and transfer application. The long-term goal is a provider-independent multi-cloud client with a reliable resumable transfer engine.

## Current backend capabilities
- Official TeraBox OAuth device authorization, polling and token refresh
- Account-specific API domain discovery
- OS credential vault integration for user tokens
- Paginated file listing and storage quota
- Streaming downloads
- Multipart uploads with MD5 verification, server-side resume discovery and retry after ambiguous part failures
- SQLite persistence for accounts, transfer tasks, upload sessions and completed parts

Live account use requires TeraBox-issued application credentials and an approved Open Platform application. The desktop UI is still being integrated with the backend.

## Technology
- Go backend and transfer engine
- Wails desktop shell (planned)
- React + TypeScript frontend
- SQLite persistence
- Windows Credential Manager / macOS Keychain / Linux Secret Service
- MIT license

## Development
Requirements: Go 1.23+, Node.js 20+.

```sh
go test ./...
go vet ./...
go build ./...
cd frontend && npm install && npm run build
```

See `AGENTS.md`, `docs/ARCHITECTURE.md`, `docs/TERABOX_INTEGRATION.md`, `docs/TRANSFER_ENGINE.md`, and `docs/ROADMAP.md`.

## Security
Never commit TeraBox client credentials, user tokens, refresh tokens or cookies. User tokens are stored in the OS credential vault; SQLite stores only credential references. On Linux, a Secret Service provider such as GNOME Keyring or KWallet must be available.
