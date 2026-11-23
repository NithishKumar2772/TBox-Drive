# TeraBox integration

## Official Open Platform authentication

TeraBox publishes OAuth integration documentation at https://www.terabox.com/integrations/docs?lang=en. Application developers must obtain their own `client_id`, `client_secret`, and `private_secret` from TeraBox. These are application credentials, not user login credentials, and must never be committed.

Implemented:
- OAuth device QR authorization request, code exchange, token info and refresh in `oauth.go`
- API client with account-specific API/upload domains in `api.go`
- Paginated file listing and storage quota lookup
- Provider adapter directory listing after an authenticated API client is supplied

The documented signature is MD5 of `client_id_timestamp_client_secret_private_secret`. Refresh tokens are single-use and rotate; persist the newly returned refresh token atomically. The official docs describe access tokens as valid for two days and refresh tokens for 30 days. API domains are account-specific; cache token information for no more than one hour.

## Upload protocol researched

The official docs describe three stages: precreate with an ordered MD5 block list, upload each shard through the account's `upload_domain` endpoint, then call create with the same ordered MD5 list. For multi-shard uploads, each shard must be larger than 4 MiB. The final implementation must respect the precreate response's requested shard indexes and must not report completion before create succeeds. This protocol is not implemented yet.

## Scope and limitations

OAuth and file listing are not a complete cloud drive. Upload, download, folder creation/management, durable transfer sessions and desktop integration remain future milestones. Do not substitute reverse-engineered PCS/NDUS web endpoints and label them official.

## Required setup

Obtain application credentials from TeraBox before end-to-end testing. Store them in an OS credential vault or deployment secret store. Never put secrets in frontend code, SQLite plaintext, logs, examples or GitHub Actions output.
