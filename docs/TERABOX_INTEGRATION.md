# TeraBox integration

## Official Open Platform

TeraBox publishes OAuth and basic capability documentation at https://www.terabox.com/integrations/docs?lang=en. Application developers must obtain their own `client_id`, `client_secret`, and `private_secret` from TeraBox. These are application credentials, not user login credentials, and must never be committed.

Implemented in `internal/providers/terabox`:
- OAuth device QR authorization, code exchange, interval-respecting polling, token information and refresh
- Account-specific API/upload domains
- Paginated file listing and storage quota lookup
- Streaming download URL lookup and file streaming
- Multipart upload: precreate, balanced shard planning, sequential shard transfer, MD5 verification and create/finalize
- Resume discovery and retry after ambiguous shard failures
- Verification of upload ID, part sequence, part MD5 and source-file immutability
- SQLite task/part/session persistence and resumable transfer IDs

Implemented in `internal/app/TeraBoxService`:
- Start and complete the device authorization flow
- Save account metadata to SQLite
- Store user OAuth tokens in the OS credential vault, never SQLite
- Refresh single-use refresh tokens under a mutex and refresh account-specific API domain information
- Construct an authenticated provider adapter with the durable transfer repository

## OAuth flow

The official integration uses the device authorization endpoint to start QR/device approval, exchanges the approved device code for tokens, obtains account-specific API and upload domains, and refreshes rotating refresh tokens. Polling respects the interval and expiry returned by TeraBox. Refresh-token replacement must be persisted atomically because refresh tokens are single-use.

## Multipart upload behaviour

The official docs describe three stages: precreate with an ordered MD5 block list, upload the requested shards through the account's `upload_domain`, then call create with the same ordered MD5 list. For multi-shard uploads, each shard must be larger than 4 MiB. The planner balances shard sizes to avoid a too-small final part. Uploads are sequential in this first version. If a part response is ambiguous, the client asks precreate for the authoritative missing-part list before retrying. Completion requires the final create call.

## Remaining limitations

- Empty-file upload is not supported.
- Upload concurrency is sequential.
- Rename, move, copy and delete methods are implemented against the documented Open Platform file-manager operation set; verify behaviour with an authorized live account before release.
- The current Open Platform file-manager documentation lists copy, move, rename and delete, but does not document directory creation. Do not add an undocumented mkdir endpoint; folder creation and folder-tree upload remain blocked on an approved documented API.
- The Wails desktop UI is connected for account connection, file listing, multi-file selection/upload, download and transfer history; native virtual-drive integration and synchronization are not implemented yet.
- Real account integration tests require an approved TeraBox Open Platform application.

Do not substitute reverse-engineered PCS/NDUS web endpoints and label them official. Never put user tokens in frontend code, SQLite plaintext, logs, examples or GitHub Actions output.
