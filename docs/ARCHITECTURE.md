# Architecture

TBox Drive is a provider-neutral application with a TeraBox-first adapter.

## Layers
1. UI clients: Wails desktop initially; mobile/web later.
2. Application services: accounts, explorer, transfer lifecycle and settings.
3. Core engines: planning, bounded workers, retry, persistence and integrity.
4. Provider interface: stable common operations and explicit capability flags.
5. Provider adapters: authentication and protocol-specific behaviour.

Core packages must not import provider implementations. The application composition root wires adapters into services. Completion requires provider finalization and remote verification. Virtual drive and other providers are deferred until TeraBox basics and transfer reliability are proven.
