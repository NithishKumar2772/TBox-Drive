# TeraBox integration

## Official Open Platform

TeraBox publishes its OAuth and basic capability documentation at https://www.terabox.com/integrations/docs?lang=en. Application developers must obtain their own `client_id`, `client_secret`, and `private_secret` from TeraBox. These are application credentials, not user login credentials, and must never be committed.

Implemented in `internal/providers/terabox`:
- OAuth device QR authorization, code exchange, token information and refresh
- Account-specific API and upload domains
- Paginated file listing and storage quota lookup
- Download-link lookup and streaming download
- Multipart upload: precreate, ordered shard upload, MD5 verification and create/finalize
- Resume by asking precreate which shard indexes are still missing

## Multipart upload behaviour

The official docs describe three stages: precreate with an ordered MD5 block list, upload the requested shards through the account's `upload_domain`, then call create with the same ordered MD5 list. For multi-shard uploads, each shard must be larger than 4 MiB. The planner balances shard sizes to avoid a too-small final part. Uploads are sequential in this first version; each uploaded shard's returned MD5 is checked before finalization.

If an upload is interrupted, rerun the same upload: precreate can report which parts remain. The client does not mark the file complete until the final create call returns a file ID.

## Remaining limitations

- Empty-file upload is not yet supported.
- Upload concurrency is sequential; bounded parallel upload is a later optimization.
- Folder creation, rename, delete and share management are not implemented.
- Transfer state is not yet stored in SQLite.
- End-to-end cloud tests require an approved TeraBox Open Platform application and a dedicated authorized account.

Do not substitute reverse-engineered PCS/NDUS web endpoints and label them official. Store credentials in an OS credential vault or deployment secret store. Never put secrets in frontend code, SQLite plaintext, logs, examples or GitHub Actions output.


## Ambiguous shard failures
When a shard request fails with a transient network or retryable HTTP error, the client re-runs precreate with the exact same ordered MD5 block list and upload path. If TeraBox no longer lists that shard as missing, the client treats the shard as accepted and does not send it twice. If TeraBox still lists it as missing, the client retries that shard with bounded attempts and delay. Permanent errors are not retried. This reconciliation behaviour is covered by an HTTP test server; it still requires live-account verification.
