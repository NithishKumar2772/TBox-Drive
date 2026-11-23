# TBox Drive roadmap — TeraBox only

## M0 Foundation
- [x] Repository conventions and documentation
- [x] Go application skeleton
- [x] Initial React dashboard

## M1 Transfer engine foundation
- [x] Overflow-safe chunk planner with allocation limit
- [x] File-range streaming reader
- [x] Bounded worker pool and cancellation
- [x] Retry policy with capped exponential backoff and jitter
- [x] Durable local part journal
- [ ] Go tests and vet verified in CI
- [ ] Source-file identity validation and remote session reconciliation

## M2 TeraBox account and file access
- [ ] Verify permitted current authentication/API flow
- [ ] Secure session storage and validation
- [ ] Account profile and paginated file listing
- [ ] Folder and metadata operations

## M3 Real TeraBox transfers
- [ ] Verify upload session/part/finalize protocol
- [ ] Implement resumable multipart upload adapter
- [ ] Reconcile remote parts on restart
- [ ] Streaming download and integrity verification
- [ ] Authorized live integration tests

## M4 Desktop application
- [ ] Wails shell and backend bindings
- [ ] TeraBox file explorer
- [ ] Upload/download queue and settings
- [ ] Installers and release workflow

No other cloud providers, multi-cloud abstractions, cloud-to-cloud transfers or virtual-drive features are in scope.
