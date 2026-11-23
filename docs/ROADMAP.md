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
- [ ] Real TeraBox account verification with an approved Open Platform app
- [ ] Empty-file upload and bounded parallel uploads

## M2 File management and desktop explorer
- [x] Rename, move and delete API adapter methods (verify against authorized account)
- [ ] Officially documented directory creation method
- [ ] Folder upload preserving nested structure
- [ ] Explorer sorting, richer file properties and previews
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
- [ ] Per-file encryption opt-in prompt on every upload
- [ ] First-time file password, confirmation and optional password hint
- [ ] Password prompt before encrypted file preview/open/download
- [ ] Password change by re-wrapping file data keys
- [ ] Recovery-key generation, export, restore and forgotten-password workflow
- [ ] Reviewed Argon2id + XChaCha20-Poly1305 chunked file container and key lifecycle
- [ ] Per-upload encryption choice and local decryption
- [ ] Encrypted cache, secure temporary-file cleanup and recovery-key workflow
- [ ] Image/PDF/text/code preview and media gallery
- [ ] Audio/video streaming, subtitles, playback position and MPV integration
- [ ] ZIP/TAR/TAR.GZ import and optional extraction

## M5 Mobile and TV clients
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
