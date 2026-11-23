# TBox Drive architecture — TeraBox only

TBox Drive is a single-provider application. TeraBox protocol logic lives in internal/providers/terabox; the Go transfer engine supplies streaming, chunk scheduling, retry, cancellation and durable local progress.

## Components
1. React UI: account, file browser, transfer queue and settings.
2. Go application services: authentication/session lifecycle, file operations and transfer orchestration.
3. Transfer engine: chunk plan, range reader, bounded workers, retry and local journal.
4. TeraBox adapter: verified service-specific requests and response parsing.
5. Local state: credential-vault references and durable transfer metadata.

No multi-cloud provider registry or other provider adapters are part of this product. A transfer is complete only after TeraBox finalization and remote verification.
