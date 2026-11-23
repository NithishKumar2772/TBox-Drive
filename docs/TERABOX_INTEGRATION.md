# TeraBox integration

## Official Open Platform and application setup

TeraBox publishes OAuth and file-operation documentation at https://www.terabox.com/integrations/docs?lang=en. Developers must obtain their own `client_id`, `client_secret`, and `private_secret` from TeraBox. These are application credentials, not user login credentials, and must never be committed to the repository.

The application must use the documented Open Platform API. Do not substitute reverse-engineered PCS/NDUS web endpoints and label them official. Live acceptance testing requires an approved TeraBox Open Platform application and its credentials.

## Implemented OAuth authentication

The OAuth client in `internal/providers/terabox/oauth.go` implements:

- GET `/oauth/devicecode?client_id=...` to begin device/QR authorization
- POST multipart `/oauth/gettoken` for device-code or authorization-code exchange
- POST form `/oauth/tokeninfo` to discover account-specific API and upload domains
- POST multipart `/oauth/refreshtoken` to rotate refresh tokens
- Device authorization polling that respects the server-provided interval, expiry, cancellation and pending-approval response

The documented signature is MD5 of `client_id_timestamp_client_secret_private_secret`. Refresh tokens are single-use and rotate; persist the newly returned refresh token atomically. The published documentation describes access tokens as valid for two days and refresh tokens for 30 days. Account-specific API domains should be cached for no more than one hour.

OAuth requests carry application credentials and rotating refresh tokens. The client rejects redirects so credentials cannot be forwarded to a different host.

## Implemented provider operations

The implementation in `internal/providers/terabox` includes:

- Account-specific API and upload domains
- Paginated file listing and storage quota lookup
- Streaming download URL lookup and file streaming
- Multipart upload precreate, balanced shard planning, sequential shard transfer, MD5 verification and final create
- Resume discovery and reconciliation after ambiguous shard failures
- Verification of upload ID, part sequence, part MD5 and source-file immutability
- Transfer task, part and session persistence through the SQLite repository
- File-manager rename, move, copy and delete operations using the documented operation set

The desktop service starts and completes device authorization, stores account metadata in SQLite, stores user OAuth tokens in the OS credential vault, and refreshes rotating tokens under a mutex.

## Multipart upload behaviour

The documented multipart flow has three stages:

1. Precreate the remote file with an ordered MD5 block list.
2. Upload the requested shards through the account's `upload_domain`.
3. Finalize once with the same ordered MD5 list.

For multi-shard uploads, each shard must be larger than 4 MiB. The planner balances shard sizes to avoid a too-small final part. Uploads are sequential in this version. If a part response is ambiguous, the client asks precreate for the authoritative missing-part list before retrying. Completion requires the final create call.

The current planner has tests for files larger than 4 GiB and contiguous part coverage. This is not evidence of a successful live-account upload larger than 4 GiB; the provider's accepted maximum size and end-to-end behaviour still require account-level acceptance testing.

## Security requirements

- Store application credentials in a secret store; never commit them.
- Store user OAuth tokens in the operating-system credential vault, not plaintext SQLite.
- Never expose tokens or application secrets in frontend code, logs, examples or CI output.
- Keep TLS verification enabled and reject credential-bearing redirects.
- Persist replacement refresh tokens atomically because refresh tokens are single-use.

## Remaining limitations

- Empty-file upload is not supported.
- Upload concurrency is sequential.
- Folder creation and folder-tree upload remain disabled because the documented Open Platform API used here does not provide a verified directory-create operation.
- Real account integration tests require an approved TeraBox Open Platform application.
- The Windows virtual-drive prototype, cache behaviour and sync features have separate limitations documented in the product roadmap. CI compilation does not prove live provider interoperability or durable virtual-drive behaviour.

Do not describe unverified live-account behaviour as production-ready.
