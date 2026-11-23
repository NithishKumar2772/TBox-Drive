# Provider interface

Initial contract: `internal/core/provider`. Capabilities must be explicit. Unsupported operations return a typed unsupported error. Implementations should provide pagination, context cancellation, stable file IDs, normalized errors and integration tests. Avoid speculative interface expansion; add optional capability interfaces when a concrete provider requires them.
