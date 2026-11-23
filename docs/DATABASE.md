# SQLite persistence

The storage package is `internal/storage/sqlite` and uses the pure-Go `modernc.org/sqlite` driver.

Current schema:
- `accounts`: provider identity and credential-vault reference only; never raw tokens/cookies
- `transfers`: durable task state and byte progress
- `transfer_parts`: ordered part offsets, sizes, checksums, remote part IDs and status
- `provider_sessions`: opaque provider upload-session reference and expiry
- `transfer_events`: redacted lifecycle event types

The first migration is applied idempotently at store startup. SQLite is configured with foreign keys, WAL and a busy timeout. The transfer store currently supports create/get/list, progress updates, part upserts and provider session persistence. Wiring it into the application startup and transfer scheduler is the next integration step.
