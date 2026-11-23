# Security design

- TeraBox OAuth client credentials are application secrets and must be supplied through a secret manager/environment configuration, never committed.
- User access and refresh tokens are stored in the operating-system credential vault through `internal/security/keyring`: Windows Credential Manager, macOS Keychain, or Linux Secret Service.
- SQLite stores only a credential reference, never raw tokens or cookies.
- Refresh tokens are single-use; refresh and credential persistence are serialized per service instance.
- Never log Authorization, Cookie, Set-Cookie, access tokens, refresh tokens, signatures or token-bearing URLs.
- Validate remote paths and keep file operations inside the TeraBox application folder.
- Use TLS with certificate validation.
- Keep test fixtures synthetic and avoid logging personal filenames or file contents.
- On Linux, configure a desktop Secret Service (for example GNOME Keyring or KWallet) before connecting a cloud account.
