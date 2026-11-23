# TeraBox integration

## Official Open Platform authentication

TeraBox publishes OAuth integration documentation at https://www.terabox.com/integrations/docs?lang=en. Application developers must obtain their own `client_id`, `client_secret`, and `private_secret` from TeraBox. These are application credentials, not user login credentials, and must never be committed.

The initial OAuth client is in `internal/providers/terabox/oauth.go`:
- GET `/oauth/devicecode?client_id=...` to begin QR/device authorization
- POST multipart `/oauth/gettoken` for `device_code` or `authorization_code`
- POST form `/oauth/tokeninfo` to discover the account-specific API/upload domains
- POST multipart `/oauth/refreshtoken` to rotate tokens

The documented signature is MD5 of `client_id_timestamp_client_secret_private_secret`. Refresh tokens are single-use and rotate; persist the newly returned refresh token atomically. The official documentation describes access tokens as valid for two days and refresh tokens for 30 days. API domains are account-specific; cache token information for no more than one hour.

## Scope and limitations

OAuth is only the authentication layer. It does not implement cloud file listing, upload, download, multipart transfer, folder operations or sharing. Those operations must be implemented from the official Open Platform API specification and tested with an approved application. Do not substitute reverse-engineered PCS/NDUS web endpoints and label them official.

## Required setup

Obtain application credentials from TeraBox before end-to-end OAuth testing. Store them in an OS credential vault or deployment secret store. Never put secrets in frontend code, SQLite plaintext, logs, examples or GitHub Actions output.
