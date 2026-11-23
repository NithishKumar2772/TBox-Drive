# Roadmap

## M0 Foundation
- [x] Repository conventions and docs
- [x] Provider interface skeleton
- [x] Chunk planner and tests
- [x] CI for Go and frontend
- [x] Retry policy and unit tests

## M1 TeraBox authentication and account
- [x] Official OAuth device QR request
- [x] Authorization-code/device-code exchange
- [x] Device-code polling and refresh
- [x] Token info and account-specific API domains
- [ ] TeraBox app credentials approval and real account verification
- [ ] OS credential vault integration

## M2 TeraBox file operations
- [x] Paginated file listing
- [x] Storage quota lookup
- [x] Streaming download
- [x] Multipart upload and server-side resume discovery
- [ ] Empty-file upload
- [ ] Folder create/rename/delete and file delete
- [ ] Real account integration tests

## M3 Transfer engine and persistence
- [x] Chunk planning and provider-aware minimum part size
- [x] MD5 part verification and finalization
- [x] Exponential retry policy
- [x] SQLite task/part/session schema and repository
- [ ] Wire durable repository into transfer orchestration
- [ ] Bounded parallel part upload
- [ ] Pause/resume and crash recovery end-to-end

## M4 Desktop
- [ ] Wails integration
- [ ] React explorer connected to backend
- [ ] Transfer dashboard and account settings

## M5 Multi-cloud
- [ ] Provider conformance suite
- [ ] Google Drive, OneDrive, Dropbox, S3 and WebDAV as separate milestones

## M6 Advanced
- [ ] Cloud-to-cloud transfer
- [ ] Sync and backup
- [ ] Virtual drive
- [ ] Mobile clients
