# Desktop application

TBox Drive uses Wails v2 with a React/Vite frontend and Go backend.

## Windows build

Requirements:
- Go 1.25+
- Node.js 20+
- Microsoft Edge WebView2 Runtime

From the repository root:

```powershell
cd frontend
npm install
npm run build
cd ..
go build -tags desktop -o dist/TBoxDrive.exe ./cmd/tboxdrive-desktop
```

The Vite build writes embedded assets to `internal/app/web/dist`. The Windows CI job compiles the desktop executable and publishes it as a workflow artifact. The executable uses the WinFsp Go binding; users must install the WinFsp runtime separately before mounting a drive.

## TeraBox application configuration

Register an application with the TeraBox Open Platform and configure these environment variables before launching the desktop app:

- `TBOX_TERABOX_CLIENT_ID`
- `TBOX_TERABOX_CLIENT_SECRET`
- `TBOX_TERABOX_PRIVATE_SECRET`
- `TBOX_TERABOX_APP_FOLDER`

The application folder must be the exact cloud path assigned to the TeraBox application, for example `/From: Other Applications/YourApp-YourAppID`. Do not guess this path.

User access and refresh tokens are stored in the OS credential vault. SQLite stores only the credential reference and local transfer state. On Linux, a desktop Secret Service provider such as GNOME Keyring or KWallet must be running.

## Current desktop features

- TeraBox QR/device authorization and account selection/disconnection
- File browsing, current-folder search, category filters and paginated listing
- Storage quota display
- Multi-file upload with progress events and per-file error reporting
- File download, rename and delete
- Transfer history
- System/light/dark appearance selection
- Profile dropdown and settings screen
- Configurable automatic current-folder and transfer-list refresh (30-second interval)
- Automatic GitHub latest-release check preference and manual check
- Upload throughput and estimated-time-remaining display\n- TeraBox multipart upload already divides large files into internal 8 MiB parts and finalizes them as one cloud file; files over 4 GiB still need real-account acceptance testing.
- Windows virtual-drive mount/unmount controls

## Profile, settings and transfer monitoring

The top-right profile menu provides access to account information, settings, sign-out and release checks. Settings include automatic folder refresh and GitHub release checks. Transfer progress events display uploaded bytes, current throughput and estimated remaining time. Download speed/ETA is not yet wired to the same progress-event contract.

The release checker reads the latest public GitHub release metadata. It does not silently download or install updates; users open the release page to review and install a release.

The application-vault session unlock model, per-file encryption, lock indicators, password hints, change-password and recovery-key flows are specified in `docs/PRODUCT_SPECIFICATION.md`, but are not enabled in this build. The intended behaviour is one vault unlock per app session, then no repeated password prompt for each encrypted file; closing/reopening the app requires unlocking again. Do not use the current app to assume files are encrypted.

## Windows virtual-drive prototype

The Windows build uses WinFsp to expose the configured TeraBox application folder as a drive letter. The first prototype:
- Lists remote folders and files in File Explorer.
- Downloads a file into the local cache when an application opens it.
- Reuses cached file content when its metadata-derived cache key is unchanged.
- Uploads edited/new files when the file handle is flushed or closed.
- Supports remote rename, move and delete operations through the documented Open Platform file-manager API.
- Prevents deleting or renaming the application root.

Limitations:
- This is on-demand whole-file hydration, not byte-range streaming.
- Folder creation is disabled because the documented Open Platform file-manager API does not currently expose a verified directory-create operation.
- The mount is tied to the selected account for the lifetime of the mount.
- Install WinFsp before mounting. The mountpoint must be an available drive letter such as `T:`.
- Test with non-critical files first; live TeraBox integration has not yet been validated with the user's account.

Folder-tree upload, robust offline sync/conflict resolution, pinning, cache eviction controls and background reconnect remain in development.

## Windows File Explorer roadmap

Planned Explorer improvements include status overlays, right-click actions, drag-and-drop upload/download, offline pinning, free-up-space, file properties, thumbnails/previews, notifications, conflict prompts and Cloud Files API placeholder support. These are roadmap items, not current capabilities. The existing WinFsp prototype provides the mounted drive and basic remote file operations only.
