# AI Agent Instructions — TBox Drive

## Product definition
TBox Drive is a **TeraBox-only** application. Do not design, implement, or add abstractions for Google Drive, OneDrive, Dropbox, S3, WebDAV or any other cloud provider. The only remote cloud target is TeraBox.

## Mission
Build a reliable desktop TeraBox file manager with large-file upload/download, safe chunking, bounded concurrency, retry, pause/resume, crash recovery and integrity verification.

## Before every task
1. Read this file and relevant docs.
2. Inspect current code and git diff.
3. Convert the complete issue into acceptance criteria.
4. Verify TeraBox protocol assumptions from reliable, current sources.
5. Prepare a concise implementation plan.

## Engineering rules
- Run gofmt, go test ./..., and go vet ./... for Go changes.
- Use context.Context for network, I/O and long-running operations.
- Stream large files; never load an entire large file into RAM.
- Bound concurrency, memory, retries and queued work.
- Persist transfer state around irreversible transitions.
- Revalidate source file identity before resuming.
- Use typed errors and redact credentials from logs.
- Never invent undocumented endpoints, headers, signatures, limits or checksum semantics.
- Never claim TeraBox operations work until tested against the real service with an authorized account.
- Never commit tokens, cookies, credentials or personal files.
- Use focused branches and PRs; do not force-push or commit directly to main.

## TeraBox protocol rules
Verify permitted authentication, session lifecycle, listing, upload session creation, accepted part size/count, part request format, idempotency, retries, finalization, download ranges and rate limits. Do not bypass access controls, CAPTCHAs or anti-abuse measures. Do not mark a transfer complete before remote finalization and verification.

## Completion report
List requirements met, files changed, commands/tests actually run, results, limitations and follow-up work. Never fabricate tests or build results.
