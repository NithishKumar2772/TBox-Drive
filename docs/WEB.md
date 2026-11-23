# Self-hosted web client

## Deployment model

The first browser client is self-hosted. The Go process runs on the user's own computer and serves the React application and API from one origin:

- Default address: `http://127.0.0.1:8765`
- Default bind: IPv4 loopback only
- Database: the same per-user SQLite database used by the desktop application
- Provider credentials: the operating-system credential vault
- Cloud provider: TeraBox only

The service rejects non-loopback bind addresses. LAN access, public hosting, reverse proxies, remote device access and cloud-hosted operation are not supported by this first implementation.

## Start on Windows

Configure the TeraBox Open Platform application values in the current PowerShell session:

```powershell
$env:TBOX_TERABOX_CLIENT_ID = "your-client-id"
$env:TBOX_TERABOX_CLIENT_SECRET = "your-client-secret"
$env:TBOX_TERABOX_PRIVATE_SECRET = "your-private-secret"
$env:TBOX_TERABOX_APP_FOLDER = "/From: Other Applications/TBox-Drive"
go run ./cmd/tboxdrive
```

Open the **private launch URL printed in the console**. It contains a random startup token in the URL fragment; the browser exchanges it for an HttpOnly, SameSite=Strict session cookie and removes the fragment from the address bar. Do not share this launch URL. Stop the service with Ctrl+C. The application folder must be the authorized TeraBox application directory configured for the Open Platform app.

To use another local port, set `TBOX_WEB_ADDR` to a loopback address such as `127.0.0.1:9000`. Non-loopback IP addresses and hostnames are rejected.

## Browser capabilities in this initial implementation

- Connect a TeraBox account with the official device authorization flow.
- List files and folders and navigate directories.
- View storage quota.
- Upload one or more selected files.
- Download files through the browser.
- Rename, move (API endpoint), and delete cloud items.
- View persisted transfer records.
- Disconnect a TeraBox account.

Folder upload with nested directory preservation, pause/resume/cancel, live browser upload progress, full transfer queue control and virtual-drive mounting are not yet available in the browser client.

## API surface

| Method | Path | Purpose |
| --- | --- | --- |
| GET | /api/status | Local service status |
| GET | /api/accounts | Public account metadata (no credential references) |
| POST | /api/connect/start | Start device authorization |
| POST | /api/connect/complete | Wait for and complete authorization |
| POST | /api/accounts/disconnect | Disconnect an account |
| GET | /api/files | Paginated folder listing |
| GET | /api/quota | Storage quota |
| POST | /api/upload | Stream a browser upload to a temporary local file, then upload to TeraBox |
| GET | /api/download | Stream a TeraBox file to the browser |
| GET | /api/transfers | Recent persisted transfer records |
| POST | /api/files/rename | Rename a cloud item |
| POST | /api/files/delete | Delete a cloud item |
| POST | /api/files/move | Move a cloud item |

All cloud paths supplied by the browser are constrained to the configured TBox Drive application folder. Uploads are staged to a temporary local file and removed when the request finishes. A browser upload therefore needs temporary-disk space roughly equal to the file size before the provider upload can begin; select a temp drive with sufficient free space for very large files.

## Security boundary and limitations

- The listener binds to a loopback IP address only. Do not expose this process through a reverse proxy, tunnel, port forward or LAN interface.
- Requests with non-loopback Host values are rejected.
- Mutating browser requests require an exact same-origin Origin header. Cross-site Fetch Metadata values are rejected.
- The console prints a random launch token in a URL fragment. The browser exchanges it for an eight-hour HttpOnly, SameSite=Strict cookie; the fragment is removed from browser history. API requests require this local session.
- API responses are marked no-store and the app applies a restrictive Content Security Policy.
- The server does not enable CORS and applies security response headers.
- Browser responses omit credential references. Access and refresh tokens remain in the OS credential vault and are used only by the Go service.
- TeraBox OAuth application secrets are supplied to the process through environment variables. Do not commit them or paste them into issue reports.
- This first implementation is not a complete defense against malicious software already running as the same OS user. It is not designed as a multi-user service.
- The browser client does not provide remote device authentication, user accounts, TLS termination, rate limiting for hostile network clients, or an independent web security audit.

Do not change the bind address to a public or LAN interface. Remote access requires a separately reviewed device-session design with explicit authentication, TLS, revocation, CSRF protection and threat modelling.

## Validation

CI runs Go tests, vet, build and vulnerability scanning, plus the frontend production build and Windows desktop build. Unit tests cover loopback Host enforcement, same-origin mutation checks, cloud path confinement and rejection of public bind addresses. These checks do not replace testing with an approved TeraBox account.
