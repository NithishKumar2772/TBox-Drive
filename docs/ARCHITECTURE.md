# Architecture

TBox Drive uses one shared Go application and transfer core with a TeraBox-first provider adapter. Windows desktop and the self-hosted browser client are the first two clients.

## Layers

1. UI clients: Wails desktop and same-origin React browser client.
2. Application services: account lifecycle, file operations, transfer lifecycle and settings.
3. Core engines: chunk planning, bounded workers, retry, persistence and integrity.
4. Provider interface: stable common operations and explicit capability flags.
5. Provider adapters: TeraBox authentication and protocol-specific behaviour.
6. Persistence and secrets: SQLite metadata plus operating-system credential vault.

## Client topology

- Windows desktop: Wails hosts the React interface and calls bound Go application methods.
- Self-hosted Web: a local Go HTTP server serves the React bundle and exposes a same-origin API on loopback only.
- Both clients use the same account service, TeraBox API adapter, SQLite database and OS credential vault.
- Browser JavaScript receives public account metadata only. Provider access and refresh tokens remain in the Go process and credential vault.
- Remote/LAN access is intentionally disabled. A future remote mode requires a separate authenticated device-session architecture and security review.

```text
             Windows desktop (Wails)      Browser (loopback)
                       |                         |
                       +-----------+-------------+
                                   |
                         Shared Go application
                                   |
                    +--------------+--------------+
                    |              |              |
               TeraBox API    Transfer core    SQLite
                    |              |              |
                    +--------------+--------------+
                                   |
                         OS credential vault
```

Core packages must not import provider implementations. The application composition root wires adapters into services. Completion requires provider finalization and remote verification. macOS, TV and mobile clients are planned in phases after the Windows and self-hosted Web foundations. Production virtual-drive streaming and synchronization remain separate milestones.
