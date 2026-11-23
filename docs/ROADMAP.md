# Roadmap

## M0 Foundation
- [x] Repository conventions and docs
- [x] Provider interface skeleton
- [x] Chunk planner and tests
- [x] CI for Go and frontend
- [x] Retry policy and unit tests
- [x] SQLite account/transfer/part/session repository

## M1 TeraBox account and authentication
- [x] Official OAuth device QR request and polling
- [x] Authorization-code/device-code exchange and refresh
- [x] Token info and account-specific API domains
- [x] Account connection service
- [x] OS credential vault adapter (Windows/macOS/Linux)
- [ ] TeraBox application approval and real account verification
- [ ] Wails UI connection flow

## M2 TeraBox file operations
- [x] Paginated file listing
- [x] Storage quota lookup
- [x] Streaming download
- [x] Multipart upload and MD5 part verification
- [x] Resume discovery and retry after ambiguous part responses
- [x] Persist transfer task, session and completed part state
- [ ] Empty-file upload
- [ ] Folder create/rename/delete and file delete
- [ ] Real account integration tests

## M3 Transfer engine
- [x] Provider-aware shard planner
- [x] Exponential retry policy
- [x] SQLite transfer repository wired into TeraBox adapter
- [x] Resume by transfer ID
- [ ] Bounded parallel part upload
- [ ] Pause/resume UI and crash recovery experience

## M4 Desktop
- [ ] Wails integration
- [ ] React explorer connected to backend services
- [ ] Transfer dashboard and account settings

## M5 Multi-cloud
- [ ] Provider conformance suite
- [ ] Google Drive, OneDrive, Dropbox, S3 and WebDAV as separate milestones

## M6 Advanced
- [ ] Cloud-to-cloud transfer
- [ ] Sync and backup
- [ ] Virtual drive
- [ ] Mobile clients
