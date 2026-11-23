# AI Agent Instructions — TBox Drive

## Mission
Build TBox Drive as a production-quality, TeraBox-first cloud manager. Keep core logic provider-neutral; provider-specific behaviour belongs in `internal/providers/<provider>`.

## Before every task
1. Read this file and relevant docs.
2. Inspect existing code and git diff.
3. Convert the complete issue into acceptance criteria.
4. Verify API assumptions against current reliable sources.
5. Prepare a concise implementation plan.

## Engineering rules
- Run gofmt, `go test ./...`, and `go vet ./...` for Go changes.
- Use context.Context for network, I/O and long-running operations.
- Stream large files; never load an entire large file into RAM.
- Bound concurrency, memory, retries and queued work.
- Persist transfer state for crash recovery.
- Use typed errors and redact credentials from logs.
- Never invent undocumented API endpoints, headers, signatures, limits or checksum semantics.
- Never claim provider functionality works until tested against the real provider with an authorized test account.
- Never commit tokens, cookies, credentials or personal files.
- Do not force-push or commit directly to main. Use focused branches and PRs.

## TeraBox rules
Verify current authentication and upload protocol, respect service limits/terms/rate limits, and implement provider-specific session creation, part upload and finalization only when verified. Do not mark complete before remote finalization and verification.

## Completion report
List requirements met, files changed, commands actually run, results, limitations and follow-up work. Never fabricate tests or build results.
