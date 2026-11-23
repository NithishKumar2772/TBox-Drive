# Roadmap

## M0 Foundation
- [x] Repository conventions and docs
- [x] Provider interface skeleton
- [x] Chunk planner and tests
- [x] CI for Go and frontend
- [x] Retry policy and unit tests
- [x] SQLite account/transfer/part/session repository

## M1 TeraBox account and transfer foundation
- [x] Official OAuth device QR request and polling
- [x] Authorization-code/device-code exchange and refresh
- [x] Token information and account-specific API domains
- [x] Account connection service and OS credential vault adapter
- [x] Paginated file listing and quota lookup
- [x] Streaming download and multipart upload foundation
- [x] Durable transfer records and resumable part reconciliation
- [x] Desktop account connection, file listing, single-file transfer UI
- [x] Multi-file selection and sequential upload UI
- [x] Current-folder search, pagination control and appearance selector
- [x] Profile dropdown and application settings entry point
- [x] Configurable automatic folder refresh
- [x] GitHub latest-release checker and automatic startup check preference
- [x] Upload throughput and estimated-time-remaining display
- [ ] Add and verify automatic >4 GiB multipart upload acceptance test using an approved TeraBox account; confirm final object is one file
- [ ] Real TeraBox account verification with an approved Open Platform app
- [ ] Empty-file upload
- [x] Opt-in bounded multipart concurrency (1–4 workers) with sequential default and local HTTP concurrency test
- [ ] Benchmark safe provider concurrency against an authorized TeraBox account; keep default conservative until verified

## M2 File management and desktop explorer
- [x] Rename, move and delete API adapter methods (verify against authorized account)
- [ ] Officially documented directory creation method
- [ ] Folder upload preserving nested structure
- [ ] Explorer sorting, richer file properties and previews
- [ ] Windows Explorer shell status overlays (cloud-only, syncing, synced, offline, encrypted/locked/unlocked, conflict/error)
- [ ] Windows Explorer context menu actions (open, download, pin offline, free up space, sync, transfer status, lock vault)
- [ ] Windows Explorer drag/drop upload/download and safe overwrite prompts
- [ ] Explorer properties, thumbnails/preview handlers and notifications
- [ ] Windows Cloud Files API placeholder integration or documented WinFsp fallback for on-demand files
- [ ] Pause/resume/cancel controls and transfer queue recovery UX
- [ ] Download progress events and unified real-time upload/download throughput + ETA
- [x] Windows WinFsp virtual-drive prototype: remote listing, on-demand whole-file cache hydration, upload-on-flush/close, rename/move/delete
- [ ] Production-ready Windows virtual drive: byte-range streaming, offline pinning, cache policy, reconnect and conflict-safe writes
- [ ] macOS Finder/File Provider integration
- [ ] Linux mount/file-manager integration

## M3 Hybrid drive and synchronization
- [ ] Streaming, mirroring and hybrid per-folder policies
- [ ] On-demand file hydration and smart local cache
- [ ] Offline pinning and cache eviction policy
- [ ] Two-way change detection and synchronization journal
- [ ] Conflict detection and resolution UI
- [ ] Reconnect and background sync service

## M4 Security and media
- [ ] Application vault password setup and one-time unlock per app session
- [ ] Global lock-now action, idle auto-lock and clear keys on app shutdown
- [ ] Closed-lock/open-lock indicators on encrypted file rows and Explorer overlays
- [ ] Per-file encryption opt-in prompt on every upload
- [ ] First-time vault password, confirmation and optional password hint
- [ ] Password prompt once on app launch/vault lock; subsequent encrypted files open without repeat prompts until lock/exit
- [ ] Password change by re-wrapping file data keys
- [ ] Recovery-key generation, export, restore and forgotten-password workflow
- [ ] Reviewed Argon2id + XChaCha20-Poly1305 chunked file container and key lifecycle
- [ ] Per-upload encryption choice and local decryption
- [ ] Encrypted cache, secure temporary-file cleanup and recovery-key workflow
- [ ] Image/PDF/text/code preview and media gallery
- [ ] Audio/video streaming, subtitles, playback position and MPV integration
- [ ] ZIP/TAR/TAR.GZ import and optional extraction

## M5 Web, mobile and TV clients
- [ ] Browser-based web application (React) using a secure shared Go HTTP API; OAuth device authorization, file browsing, upload/download, previews and transfer history
- [ ] Secure account/device session registry and device authorization/revocation flows
- [ ] Multi-device session management UI: platform, device label, last active, current-session marker, revoke one/all sessions
- [ ] Cross-device settings and sync coordination; encrypted-key sharing must be separately designed and reviewed
- [ ] Android 7+ client
- [ ] iOS/iPadOS client
- [ ] Android TV / Google TV / Fire TV client
- [ ] Samsung Tizen TV client
- [ ] LG webOS TV client
- [ ] Shared 24-locale language framework and system-language detection

## M6 Release engineering
- [ ] Windows x64/ARM64 packaging
- [ ] macOS Intel/Apple Silicon packaging and signing
- [ ] Linux AppImage/DEB/RPM/Arch packaging
- [ ] Android/iOS/TV store packaging
- [ ] Accessibility, security, performance and upgrade testing
- [ ] Cryptographic threat model, external review and recovery drills
- [ ] Live-provider acceptance test suite
