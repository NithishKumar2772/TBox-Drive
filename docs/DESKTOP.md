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

The Vite build writes embedded assets to `internal/app/web/dist`. The Windows CI job compiles the desktop executable and publishes it as a workflow artifact.

## TeraBox application configuration

Register an application with the TeraBox Open Platform and configure these environment variables before launching the desktop app:

- `TBOX_TERABOX_CLIENT_ID`
- `TBOX_TERABOX_CLIENT_SECRET`
- `TBOX_TERABOX_PRIVATE_SECRET`
- `TBOX_TERABOX_APP_FOLDER`

The application folder must be the exact cloud path assigned to the TeraBox application, for example `/From: Other Applications/YourApp-YourAppID`. Do not guess this path.

User access and refresh tokens are stored in the OS credential vault. SQLite stores only the credential reference and local transfer state. On Linux, a desktop Secret Service provider such as GNOME Keyring or KWallet must be running.

## Current desktop features

- TeraBox QR/device authorization
- Account selection and disconnection
- File/folder listing
- Storage quota display
- File upload with progress events
- File download with save dialog
- Transfer history

Folder creation/rename/delete, pause/resume controls, multi-provider account management and cross-cloud transfer remain future work.
