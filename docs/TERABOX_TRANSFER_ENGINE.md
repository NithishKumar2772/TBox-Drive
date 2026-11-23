# TeraBox transfer engine — implementation status

The transfer package provides deterministic bounded chunk planning, file-range streaming, bounded workers, context cancellation, a retry policy, and a local JSON journal tagged specifically for TeraBox.

## Protocol boundary
This does not claim live TeraBox multipart upload works. The adapter still needs verified session creation, accepted part size/count, part request format, retry idempotency, session reconciliation, finalization and remote verification. Do not enable uploads until verified against a permitted current interface and tested with an authorized account.

## Resume safety
The journal records source path, size, modification timestamp, chunk layout, remote session reference and completed-part metadata. Before resume, the orchestrator must re-stat the source and compare size and modification time; mismatch means start a new task. Local completion is not proof of remote persistence until TeraBox confirms it during reconciliation.

## Resource controls
Use OpenChunk to stream ranges. RunWorkers bounds concurrent jobs. Retry Do only for operations proven safe to repeat by TeraBox protocol semantics.
