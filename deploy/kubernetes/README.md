# Running multiple Vikunja replicas

The normal binary runs the API, scheduler and durable workers together. SQLite, MySQL and PostgreSQL remain supported for a single process. Redis and Kubernetes are optional for that installation.

A cluster requires PostgreSQL, shared Redis and shared file storage. Every replica may run schedules and workers. PostgreSQL row locks and transactional watermarks ensure one run per shared schedule window. `scalable.yaml` starts ten replicas; its resource settings are starting points to measure.

## Deployment

1. Build the image and replace every image placeholder in both manifests with the same immutable digest. Keep plugins and configuration identical across replicas.
2. Create `vikunja-secrets` with the persistent shared `VIKUNJA_SERVICE_SECRET`, PostgreSQL credentials/TLS configuration, Redis credentials and S3 configuration. Supply the existing license configuration unchanged.
3. Apply the ConfigMap, then run `migrate.yaml` once. Wait for `kubectl wait --for=condition=complete job/vikunja-migrate --timeout=15m` before starting replicas. This creates core/plugin schemas and the shared license instance identity. Keep automatic migrations disabled on cluster replicas.
4. Apply the workloads. Every API pod schedules by default; there is no singleton scheduler to fence or replace. Operators can disable `service.enablescheduler` on selected pods.
5. Expose the Service through an ingress with WebSocket upgrades and an idle timeout above the 30-second heartbeat (for example 120 seconds). Sticky sessions are unnecessary. Configure trusted proxy CIDRs and terminate TLS at ingress.
6. Configure backups, S3 lifecycle and monitoring. Local file storage requires `files.shared=true` and the same read/write filesystem mounted everywhere. S3 temporary files remain local.

Use a dedicated Redis database per installation. `redis.prefix` namespaces WebSocket channels, together with the Redis database number; it does not namespace existing key-value data. Keep security keys free from eviction. Redis Pub/Sub is transient; visible clients reconcile authoritative notifications and timer queries every five to six minutes with jitter.

## Scheduling and failure recovery

`cron_state` stores one watermark per named schedule. Reminders use `(from, to]` windows and catch up at most 24 hours. Overdue reminders use each recipient's local calendar date. State-driven cleanups inspect current state and retain their per-item transactions.

PostgreSQL lock holders disable `idle_in_transaction_session_timeout` locally so a cleanup using other sessions cannot lose exclusion halfway through. Cluster connections set TCP keepalives to 5s idle / 2s interval / 2 probes. Configure equivalent server-side behavior when using a pooler. Network failure detection is bounded by PostgreSQL/pooler and operating-system behavior; the liveness probe must restart a frozen process.

Watch `vikunja_cron_lag_seconds{job=...}`. A rising value catches failed ticks, held locks and absent schedulers. Hourly jobs normally show up to an hour of lag. Long transactions can delay vacuum; investigate schedules lasting several minutes.

## Durable workers

Mail, reminder channels, webhook expansion/delivery and notification creation use the application database's `jobs` table. Jobs are inserted with business changes. Database effects commit with completion, guarded by current owner and attempt. External SMTP and HTTP delivery are at-least-once: a process may die after the remote server accepts the effect and before completion reaches SQL. Webhooks carry a stable `X-Vikunja-Delivery` header for recipient deduplication.

The queues `mail`, `bulk-mail`, `notifications` and `webhooks` each default to two handlers per process. Bulk reminder mail cannot occupy interactive mail slots. Set `jobs.queues.<queue>.concurrency`, `.lease` and `.timeout`; defaults are 2, 60s and 5m. Heartbeats renew active claims every 15s. A handler deadline stops renewal; expired claims are rescued by the next minute's maintenance tick. Uncooperative cancelled handlers keep their local slot until they return, bounding local work.

Set `jobs.enabled=false` on API-only pods and run `vikunja worker` elsewhere. Disable individual queues with `jobs.queues.<queue>.enabled=false` to dedicate pools. `vikunja scheduler` remains an optional headless role and participates in the same locking protocol. At least one process must run shared schedules for rescue and cleanup.

Retries use persisted exponential backoff. Five failed attempts park a job with its last error. Inspect with `vikunja jobs failed` and explicitly retry with `vikunja jobs replay <id>`. Payloads may contain rendered emails and private application data; restrict database access. Unknown payload versions are retried then parked, not silently discarded.

Completed jobs retain their deduplication keys for `jobs.retention` (default 48h, required to exceed 24h), then are deleted in batches of 1,000. Parked jobs never expire automatically. SQL storage capacity is part of the queue budget. Imports remain on the local bus with their existing no-generic-retry behavior; other unmigrated listeners remain volatile.

## Shutdown, probes and capacity

SIGTERM drains HTTP producers, stops schedules and local listeners, and stops durable workers. Interrupted durable claims are rescued. `service.shutdowntimeout` defaults to 30s; the manifest uses 60s and a 90s Kubernetes grace period.

`/api/v2/health` checks serving dependencies, including worker readiness when workers are enabled. TCP liveness avoids restarting every pod during a shared dependency outage. Spread pods across machines and zones. Slow scale-down reduces socket reconnect spikes; adding replicas does not move established sockets.

Budget database connections across all replicas, handlers, schedules and migrations. Twenty connections per pod across ten pods already means 200 connections. Monitor connection waits, queue state/age, retained storage, cron lag, SMTP failures, HTTP latency, Redis latency and per-pod memory/sockets. Send logs to a collector rather than sharing rotating log files.

Use the [load-test guide](../loadtest/README.md) for later staging verification. The prior socket measurements do not establish mixed-workload capacity. The implementation's deployment/failure testing is a separate follow-up.
