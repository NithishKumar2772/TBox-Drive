# Security design

Use OS credential vault for secrets. Never persist raw cookies/tokens in SQLite or logs. Redact Authorization, Cookie, Set-Cookie and token query parameters. Validate local paths and prevent traversal. Use TLS with certificate validation. Keep test fixtures synthetic and avoid logging personal filenames or file contents.
