# TBox Drive

Open-source, TeraBox-first cloud file manager and transfer application. The long-term goal is a provider-independent multi-cloud client with a reliable resumable transfer engine.

**Status:** early foundation. TeraBox authentication and live cloud operations are not implemented yet. Do not interpret scaffolding as a working provider integration.

## Stack
- Go backend and transfer engine
- Wails desktop shell (planned)
- React + TypeScript frontend
- SQLite persistence (planned)
- MIT license

## Development
Requirements: Go 1.23+, Node.js 20+.

```sh
go test ./...
go vet ./...
cd frontend && npm install && npm run build
```

See `AGENTS.md`, `docs/ARCHITECTURE.md`, `docs/TERABOX_INTEGRATION.md` and `docs/ROADMAP.md`.
