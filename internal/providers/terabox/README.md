# TeraBox adapter

Implemented using the official TeraBox Open Platform:
- OAuth device QR authorization, code exchange, token info and refresh
- Account-specific API domains
- File listing and storage quota
- Streaming file download
- Multipart upload with MD5 block planning, precreate, missing-part recovery, part verification and final create

The adapter supports listing, upload and download once configured with an authenticated API client. Folder creation, rename, delete, sharing, durable SQLite transfer records and parallel upload are not implemented yet.

Real account tests require an approved TeraBox Open Platform application and a dedicated authorized test account. Do not use reverse-engineered PCS/NDUS sessions as a substitute for the official OAuth contract.
