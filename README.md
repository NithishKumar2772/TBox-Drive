# TBox Drive

TBox Drive is an open-source, TeraBox-first cloud file manager and transfer application. The long-term goal is a provider-independent multi-cloud client with a reliable resumable transfer engine.

## Current status
The repository now includes official TeraBox OAuth, account-specific API domain discovery, file listing, quota lookup, streaming downloads, multipart uploads with part MD5 verification and precreate-based resume discovery, plus a SQLite persistence foundation. Live account verification still requires TeraBox-issued application credentials.

## Technology
- Go backend and transfer engine
- Wails desktop shell (planned)
- React + TypeScript frontend
- SQLite persistence
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

## Important
TeraBox integration requires an approved Open Platform application with `client_id`, `client_secret`, and `private_secret`. Never commit these credentials. File operations must stay within the app-specific TeraBox directory.
