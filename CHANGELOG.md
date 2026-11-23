# Changelog

All notable changes to TBox Drive are recorded here. Unreleased work is not a published release.

## Unreleased

### Added
- TeraBox Open Platform OAuth device authorization, token refresh and OS credential-vault integration.
- SQLite account references, durable transfer records, multipart part/session persistence and recovery.
- TeraBox file listing, quota lookup, streaming downloads and multipart upload foundation.
- Bounded multipart concurrency with a conservative desktop default.
- Wails v2 + React Windows desktop foundation.
- Windows WinFsp virtual-drive prototype with whole-file cache hydration and upload-on-close.
- Loopback-only self-hosted Web server and React browser client for account connection, paginated browsing, quota, upload/download, rename/move/delete and transfer history.
- Browser upload request progress, provider-side upload/download transfer progress, active upload cancellation and native-browser download cancellation tracking.
- Cloud copy actions in both desktop and browser file explorers, with root/descendant path protections.
- Desktop file move, cancellable upload/download contexts, persisted download transfer records, byte progress, size verification and atomic temporary-file commit.
- Virtual-drive cache hydration serialization and fail-closed handling for cache files whose size differs from remote metadata, preserving possible unsynced local edits.
- Private browser launch-token exchange, HttpOnly session cookie, same-origin mutation checks, CSP/no-store response headers and cross-process SQLite credential leases.
- Windows dual-application build script and GitHub Actions release packaging workflow.
- Documentation for product scope, architecture, security, transfers, desktop, browser mode and roadmap.

### Verified by current main CI
- Go formatting, module verification, tests, vulnerability scan, vet and build.
- Frontend dependency audit and production build.
- Windows desktop and self-hosted Web executable build.

### Still required before an initial public release
- Live TeraBox account acceptance tests, including upload/download/recovery and provider size limits.
- Windows installed-machine acceptance for desktop, Web session handling, WinFsp mount/unmount, upgrade and uninstall.
- Browser folder uploads, pause/resume/retry, cancellable downloads and full queue management.
- Production virtual-drive byte-range streaming, cache eviction/offline pinning, sync and conflict recovery.
- Per-file encryption/vault/recovery design and implementation.
- Native macOS, Linux, Android, iOS and TV applications with platform builds and device acceptance.
- Independent security review and release signing/verification.

No public release has been published yet.
