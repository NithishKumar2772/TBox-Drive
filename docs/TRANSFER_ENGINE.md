# Transfer engine

- Stream from disk with bounded buffers.
- Bound concurrency and file descriptors.
- Persist task/session state around irreversible transitions.
- Retry transient errors with capped exponential backoff and jitter.
- Do not blindly retry auth, quota or invalid-request errors.
- Use context cancellation.
- Verify remote completion before success.

The chunk planner is provider-neutral and does not imply TeraBox accepts a particular part size or concurrency. The adapter validates constraints. Recovery reconciles local sessions with provider state where supported; ambiguous/expired sessions must be surfaced, not marked complete.

The retry helper requires the caller to classify transient errors and the operation to be safe/idempotent. MaxAttempts includes the initial attempt.
