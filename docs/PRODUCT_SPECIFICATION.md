# TBox Drive product specification

## Product direction
TBox Drive is a TeraBox-only cloud file manager with a Google Drive for Desktop-like virtual-drive experience. The shared Go core should remain provider-adapter-ready, but TeraBox is the only enabled storage provider in this product scope.

## Confirmed product decisions
- Storage provider: TeraBox only.
- Default storage mode: Hybrid (streaming + mirroring + selective offline access).
- Synchronization: two-way.
- Upload encryption: ask the user for each upload.
- Encryption algorithm requested: XChaCha20-Poly1305, performed client-side.
- Development priority: all requested platforms, delivered in phases.
- Windows file opening: smart automatic mode (stream where possible, fetch on demand, cache intelligently, fully download when required).

## Account profile and application settings
- A profile menu in the top-right corner of the desktop application.
- Profile summary, connected TeraBox account, account switching, sign-out and access to settings.
- General preferences: theme, automatic folder refresh, automatic release checks and diagnostics.
- Security settings: per-file encryption, password changes, password hints, recovery-key management and encrypted-file recovery.
- Application update checker: query the official TBox Drive GitHub releases endpoint, show the latest release and open its release notes/download page. Silent background installation must not occur without explicit user consent.
- Automatic folder refresh should be configurable and avoid overlapping requests or disrupting active transfers.

## Desktop experience
- Windows 10/11, macOS Intel/Apple Silicon, and Linux.
- A Windows WinFsp-backed virtual-drive prototype is implemented: it lists remote content, hydrates whole files into a local cache on open, and uploads edits on flush/close. It is not yet byte-range streaming or a full offline-sync engine.
- macOS Finder/File Provider and Linux file-manager integrations remain target capabilities, not yet implemented.
- Windows target: File Explorer integration, online-only files, on-demand hydration, offline pinning, cache management, status icons, context actions, drag/drop, search, sort, properties and reconnect.
- Windows Explorer experience target:
  - Shell status overlays for cloud-only, downloading, synced, offline-pinned, encrypted/locked, encrypted/unlocked, conflict and error states.
  - Right-click actions: Open in TBox Drive, download now, make available offline, free up local space, upload/sync now, view transfer status, copy cloud path/link where supported, and lock/unlock the TBox encryption vault.
  - Drag files/folders into the T: drive to upload; drag/copy from the drive to local storage to download. Confirm overwrite and conflict cases.
  - File/folder properties page with remote size, cloud path, account, sync state, cache size, last sync and encryption state.
  - Explorer preview/thumbnail integration for supported media and documents, with local decryption only after vault unlock.
  - Windows notifications for upload/download completion, errors, authentication expiry and conflicts.
  - Offline pinning and "Free up space" actions must preserve the remote original and never remove a pinned local copy.
  - Context menu, shell extension and icon-overlay registration must be signed, versioned and uninstall cleanly. Prefer modern Windows Cloud Files API placeholders where practical; keep WinFsp fallback compatibility.
- macOS target: Finder/File Provider integration.
- Linux target: file-manager integration through an appropriate mount provider such as FUSE.
- Light, dark and system themes.

## File management
- Browse nested folders; create, rename, move, copy and delete where supported by the official TeraBox API.
- Upload multiple files and complete folder trees.
- Search, sorting, thumbnails, previews and file properties.
- Do not implement undocumented TeraBox endpoints. The current official Open Platform file-manager documentation lists copy, move, rename and delete; directory creation must remain gated until an approved, documented method is verified.
- File/folder drag-and-drop and native OS integration are planned.

## Storage modes and synchronization
- Streaming: list remote metadata and retrieve content on demand. The current Windows prototype hydrates whole files on demand; byte-range streaming is not yet implemented.
- Mirroring: maintain local copies and synchronize changes.
- Hybrid: per-folder/per-file streaming, mirroring and offline pinning.
- Two-way sync with change detection, conflict handling, retry queue, offline journal, version-aware reconciliation and safe deletion semantics.
- Cache size/location/exclusions must be configurable. Cache eviction must never delete remote originals or pinned offline files.

## Transfers
- Multipart upload, large-file support, retries, resume, integrity verification, transfer queue, progress/speed, cancellation and bounded concurrency.
- Preserve the original logical file in the UI; internal parts must not appear as separate user files.
- Archive import for ZIP, TAR and TAR.GZ with optional extraction.
- Do not claim provider-specific size limits until confirmed against official documentation and a real authorized account.

## Professional application enhancements
- Transfer queue controls: pause, resume, retry, cancel, retry failed items and bounded parallel transfers.
- Optional bandwidth limits for uploads and downloads, with separate upload/download caps and a clear unlimited mode.
- Transfer notifications for completion, failure, authentication expiry and low-storage conditions.
- Per-file and per-transfer details: source/destination, size, checksum status, elapsed time, average/current speed and ETA.
- Safe conflict handling: ask before overwrite, keep both, rename incoming copy or skip; never silently discard local changes.
- Automatic retry with backoff and resume after app restart or network interruption.
- Configurable cache location/size, cache usage meter, pinned-file protection and secure temporary-file cleanup.
- App lock on idle, optional Windows Hello/device unlock, and a clear lock-now action. Device unlock must protect access to local secrets; it must not replace file-password/recovery-key security.
- Version history/restore where the TeraBox API exposes enough verified metadata; never promise cloud versioning without provider support.
- Diagnostics page with redacted logs, connection test and exportable support bundle. Never include passwords, access tokens, recovery keys or decrypted file content.
- Accessibility, keyboard navigation, drag/drop, sortable columns, context menus and high-DPI/responsive layouts.

## Security
- Ask before each upload whether to encrypt.
- Encrypted content must be encrypted before leaving the device and decrypted locally for supported previews/streaming. A single vault password unlocks per-file wrapped keys for the current application session.
- XChaCha20-Poly1305 is the requested AEAD. Key derivation, key backup/recovery, nonce layout, authenticated metadata, encrypted cache, and multi-device key sharing must be designed and reviewed before enabling production encryption.
- Never store plaintext keys or provider tokens in SQLite or logs. Use OS credential/key storage where available.
- Use official TeraBox OAuth; never collect TeraBox/Google passwords in a cloned login form.

## Media
- Photo/video gallery, image/PDF/text/code preview, audio and video playback, thumbnails, subtitle/audio track controls, seeking and locally saved playback position.
- MPV integration on supported desktop systems for formats unsupported by the system webview.
- Mobile playback gestures and TV remote navigation are platform-specific follow-up milestones.

## Platforms
- Windows 10/11.
- macOS Intel and Apple Silicon.
- Linux (x86_64 first; ARM64 where feasible).
- Android 7+.
- iOS and iPadOS (supported current Apple device generations; minimum deployment target to be selected during native client planning).
- Android TV / Google TV / Fire TV.
- Samsung Tizen and LG webOS TV applications.

## Languages
- 24 selectable locales plus automatic system-language selection.

## Delivery principles
- Implement in verifiable milestones.
- Keep platform-specific integrations separate from the shared Go core.
- Test with local HTTP fixtures and synthetic files; separately record live TeraBox integration results.
- Never mark a feature complete until its tests and platform/provider verification pass.
