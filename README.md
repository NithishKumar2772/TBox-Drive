# TBox Drive

**TBox Drive is a TeraBox-only desktop file manager and transfer client.** It is not a multi-cloud application.

## Product scope
- Connect to a TeraBox account using a verified, permitted authentication flow.
- Browse and manage TeraBox files and folders.
- Upload/download large files with streaming, bounded concurrency, retry and resumable state.
- Keep uploaded content as one logical file; splitting is an internal transfer mechanism only where the verified TeraBox protocol supports it.

## Technology
- Go transfer engine/backend
- React + TypeScript UI
- Wails desktop shell (planned)
- Local durable transfer journal; SQLite may be used for task/account indexing
- MIT license

## Current status
The repository contains the initial app foundation and transfer-engine foundation. TeraBox authentication, real API calls, upload session creation, multipart part upload, finalization and live downloads are not implemented yet. Local chunk/journal code is not a completed TeraBox integration.

## Development
Requirements: Go 1.23+, Node.js 20+.

```sh
gofmt -w ./cmd ./internal
go test ./...
go vet ./...
cd frontend && npm install && npm run build
```

Read AGENTS.md, docs/ARCHITECTURE.md, docs/TERABOX_INTEGRATION.md and docs/TERABOX_TRANSFER_ENGINE.md.
