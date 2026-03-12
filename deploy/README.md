# Deploy

Deployment assets will live here as the project moves beyond local development.

Current contents:

- `docker-compose.yml` for local gateway + Postgres development
- `migrations/` SQL schema files for Postgres

Local Postgres notes:

- The gateway runs migrations automatically when `AGENTFENCE_POSTGRES_DSN` is set.
- The CLI uses the same DSN for approval commands when present.
- Without a Postgres DSN, approvals continue to fall back to the local JSON file store.
