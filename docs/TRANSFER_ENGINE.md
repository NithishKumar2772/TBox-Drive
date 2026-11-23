# Transfer engine

- Stream from disk using bounded buffers.
- Bound concurrent workers and file descriptors.
- Persist task/session state around irreversible transitions.
- Retry transient errors with capped exponential backoff and jitter.
- Do not blindly retry permanent auth, quota or invalid-request errors.
- Use context cancellation.
- Verify remote completion before success.

The chunk planner is provider-neutral and does not imply TeraBox accepts a specific chunk size or concurrency. Adapter validates constraints. Recovery reconciles local sessions with provider state where supported; ambiguous/expired sessions must be surfaced, not marked complete.
