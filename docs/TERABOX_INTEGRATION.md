# TeraBox official Open Platform integration notes

Primary reference: https://www.terabox.com/integrations/docs/hindi (official TeraBox integration documentation; English-language pages/locales may also be published by TeraBox).

## Authentication requirements
The official Open Platform requires an application registration and issued client_id, client_secret and private_secret. Authorization-code and device-code flows are documented. The token exchange signature is MD5 over the documented concatenated values client_id, timestamp, client_secret and private_secret. Access tokens and refresh tokens have finite lifetimes; refresh tokens are single-use and rotate. Do not embed app secrets in the desktop binary. A secure architecture should perform confidential-client token exchange in a controlled backend, or follow TeraBox's approved desktop-app arrangement if provided by the platform owner.

## Regional API domains
After obtaining an access token, call /oauth/tokeninfo and use the returned api_domain for normal API requests and upload_domain for the sharded upload endpoint. Cache this mapping for no more than one hour as recommended in the docs.

## Official large-file upload flow
1. Calculate MD5 for each file shard and submit the ordered block_list to /openapi/api/precreate with the destination path and autoinit=1.
2. Upload each shard to /rest/2.0/pcs/superfile2 on the tokeninfo upload_domain. The documented request uses method=upload, app_id=250528, path, uploadid and zero-based partseq, with the shard bytes as the file body.
3. Submit /openapi/api/create with the original path, file size, uploadid and the same ordered block_list to finalize the logical file.

The official docs state that for multi-shard uploads each shard must be larger than 4 MB, and that shard order must match the block_list order. This repository must not assume an undocumented maximum part size, max part count, parallelism allowance, retry idempotency, or session-resume behavior. Verify those details with TeraBox before enabling concurrent uploads or retrying ambiguous part requests.

## Current implementation boundary
The transfer engine currently plans chunks, streams byte ranges, bounds worker concurrency, provides retry primitives, and persists a local part journal. It does not yet call these TeraBox endpoints. The API adapter, OAuth flow, secure secret storage, token refresh, remote reconciliation, finalization and live tests remain future work.
