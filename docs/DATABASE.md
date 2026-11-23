# Database design (planned)

Planned tables: accounts (provider/display name/credential-vault reference); transfers (direction/status/byte counts/refs); transfer_parts (index/offset/length/status/provider reference/checksum); provider_sessions (provider session ref and expiry); events (redacted lifecycle events). Use versioned migrations and transactions. Never store raw authentication secrets.
