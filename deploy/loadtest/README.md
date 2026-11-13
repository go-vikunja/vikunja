# Cluster acceptance workload

`cluster.js` exercises a real Vikunja deployment through its public load balancer. The default target is **20,000 distinct, simultaneously connected users**, with 20% also reading or changing data. Run it against a disposable staging installation with the same ingress, ten or more application replicas, Redis, storage and database configuration intended for production. The script requires an explicit URL and `LOAD_CONFIRM=staging`; it creates comments and changes task titles.

This complements the in-process Go socket test in [pkg/websocket/load_test.go](../../pkg/websocket/load_test.go). That test isolates the hub and Redis transport. This workload includes JWT authentication, API authorization, database transactions, event listeners, notification generation, ingress and network behavior. Passing either test alone is not a capacity guarantee.

## Fixtures

Provision at least 20,000 staging users and export their **user JWTs**, each with its own task:

```json
[
  {"token": "USER_1_JWT", "taskId": 101, "projectId": 10},
  {"token": "USER_2_JWT", "taskId": 102, "projectId": 10}
]
```

Keep this file outside the repository, restrict access to it and delete it after the run. It contains credentials. The script does not create users, grant access, subscribe users, issue tokens or change authentication settings. It checks that JWT user IDs and task IDs are distinct, and that token expiry covers the whole run plus shutdown grace. JWT signature and session validation still happen on the application server. Obtain sufficiently long-lived staging sessions before testing; this harness does not refresh tokens.

For every entry:

1. Give that user read and write access to its project and task.
2. Subscribe that user explicitly to its own task using `POST /api/v2/subscriptions/task/{taskId}` during provisioning. Task creators can already be subscribed automatically; a `412` with code `12002` means an existing subscription. Verify its scope. Avoid subscribing every user to the entire shared project, which would turn each probe into 20,000 notifications.
3. Give the **next fixture user**, wrapping the final entry to the first, write access to this task. A shared project is convenient; several projects with paired permissions also work.
4. Enable task comments and database notifications. Disable outbound mail or route it to a staging mail sink, unless SMTP capacity is deliberately part of the test. Remove unrelated subscriptions, webhooks and integrations from fixture resources.

Each VU opens one socket as its own fixture user. For a delivery probe it submits a uniquely marked comment on its own task **using the next user's token**. Vikunja deliberately skips notifying the comment author, so this arrangement produces a real notification for the listening user. The VU tracks the expected marker and its deadline locally, which detects missing notifications without synchronizing load-generator VMs. Permissions and subscriptions must be prepared by the operator; a broken fixture fails the test.

The routes are taken from [tasks.go](../../pkg/routes/api/v2/tasks.go), [projects.go](../../pkg/routes/api/v2/projects.go), [task_comments.go](../../pkg/routes/api/v2/task_comments.go), and [subscriptions.go](../../pkg/routes/api/v2/subscriptions.go). Edits use the generated merge-patch route to avoid overwriting unrelated fields. WebSocket auth and subscription use the existing `/api/v1/ws` protocol in [connection.go](../../pkg/websocket/connection.go); there is no subscription acknowledgement, so work starts after a short delay.

## Run

Use k6 v2.3.0 or a compatible version with the stable `k6/websockets` module. It uses the shared event loop; asynchronous HTTP keeps notification reads running while requests are in flight. The fixture array is shared across VUs to avoid copying all JWTs into each VM. These choices follow the current [WebSocket API](https://grafana.com/docs/k6/latest/javascript-api/k6-websockets/), [async HTTP API](https://grafana.com/docs/k6/latest/javascript-api/k6-http/asyncrequest/) and [SharedArray documentation](https://grafana.com/docs/k6/latest/javascript-api/k6-data/sharedarray/).

Start with a two-user smoke run before increasing load:

```sh
LOAD_CONFIRM=staging \
BASE_URL=https://vikunja-staging.example.com \
FIXTURES=/secure/vikunja/fixtures.json \
RUN_ID=smoke-001 \
USERS=2 ACTIVE_RATIO=1 PROBE_RATIO=0.5 EDIT_RATIO=0.2 \
RAMP_SECONDS=10 HOLD_SECONDS=60 DOWN_SECONDS=10 SESSION_SECONDS=30 \
k6 run deploy/loadtest/cluster.js
```

Then run the full workload, retaining the output and monitoring series:

```sh
LOAD_CONFIRM=staging \
BASE_URL=https://vikunja-staging.example.com \
FIXTURES=/secure/vikunja/fixtures.json \
RUN_ID=capacity-001 \
USERS=20000 \
k6 run deploy/loadtest/cluster.js
```

`BASE_URL` is the application root, without `/api/v2`. No production URL is supplied by default. Use a unique `RUN_ID` for every run, shared by all generators. Never put JWTs in command arguments or metric tags. Comment and title markers identify the run; reset the disposable database snapshot afterwards. The harness deliberately leaves generated data for inspection.

| Setting | Default | Meaning |
| --- | --- | --- |
| `USERS` | `20000` | Peak VUs, each with a distinct authenticated socket |
| `RAMP_SECONDS`, `HOLD_SECONDS`, `DOWN_SECONDS` | `300`, `600`, `60` | Ramp, sustained peak and ramp down |
| `ACTIVE_RATIO` | `0.2` | Deterministically selected users doing HTTP work; `0` tests idle connections |
| `ACTIONS_PER_SECOND` | `0.2` | Offered action rate per active user |
| `EDIT_RATIO`, `PROBE_RATIO` | `0.2`, `0.1` | Fractions of actions editing a title or posting a notification probe; the remainder reads tasks/projects equally |
| `SESSION_SECONDS` | `300` | Socket session length, randomized ±20%; reconnect with the same principal afterwards |
| `PROBE_TIMEOUT_SECONDS` | `10` | Deadline from starting the comment POST to receiving its notification |
| `HTTP_TIMEOUT_SECONDS`, `AUTH_TIMEOUT_SECONDS` | `10`, `5` | HTTP request and socket authentication deadlines |
| `MAX_UNEXPECTED_CLOSE_RATE` | `0` | Allowed fraction of sessions closed unexpectedly; baseline requires none |

At the default peak, the offered rate is about 800 HTTP actions/second: 560 reads, 160 title edits and 80 comment probes. Actions stop shortly before each session ends to drain outstanding work. Authentication, reconnection and this drain window reduce achieved throughput. A VU permits one HTTP request at a time; overlapping scheduled actions increment `actions_skipped` and fail the baseline threshold. Compare observed rates with the offered workload instead of interpreting VU count as throughput.

The executor allows sessions to complete during ramp down, so connections can remain after the nominal ramp ends. It follows the [ramping-VU graceful-stop model](https://grafana.com/docs/k6/latest/using-k6/scenarios/executors/ramping-vus/). Reserve enough time for the configured session grace.

## Multiple load generators

20,000 k6 VMs and sockets require substantial generator memory, CPU, file descriptors and network capacity. Use several generators on machines separate from the application nodes. Monitor generators as well as Vikunja; generator saturation invalidates latency and capacity conclusions.

Use the [k6 Operator TestRun](https://grafana.com/docs/k6/latest/set-up/set-up-distributed-k6/usage/executing-k6-scripts-with-testrun-crd/) with `parallelism`, a pinned k6 image and the identical script, fixture file, `RUN_ID`, `USERS` and environment on every runner. Its initializer also needs the environment and fixture mount for script initialization. A 20,000-JWT fixture exceeds the 1 MiB ConfigMap limit; use a restricted fixture volume or an appropriate secret delivery mechanism, not a public ConfigMap. Keep load generators off the application worker pool.

For manually coordinated runners, divide execution with matching segment sequences, for example `--execution-segment '0:1/4' --execution-segment-sequence '0,1/4,1/2,3/4,1'` on the first of four runners, followed by the other three non-overlapping segments. Start them together. Keep `USERS=20000` on all four, not 5000. The script uses [`exec.vu.idInTest`](https://grafana.com/docs/k6/latest/javascript-api/k6-execution/) to select unique principals across correctly segmented runners. Starting four unrelated, unsegmented runs duplicates user identities and does not test 20,000 distinct connected users.

Thresholds are evaluated per runner. Aggregate time series for the cluster and retain every runner's exit status; a successful runner must not hide a failed one. Supply enough fixtures for the largest global VU ID if using a custom allocation scheme.

## Acceptance and failure drills

The built-in baseline thresholds require socket authentication and API success above 99.9%, auth p95 below 1 second, API p95 below 500 ms, notification delivery success above 99.9%, and notification p95 below 1 second. They also require at least one authenticated session and, when enabled, at least one probe; a test with no work is not evidence of delivery. Tune SLO values in the script deliberately before a benchmark, then keep them fixed when comparing configurations.

`delivery_success` covers every attempted probe, including rejected HTTP requests, missing messages, timeouts and socket loss. `delivery_latency` covers successful probes from just before HTTP submission until socket receipt. It therefore includes the write, event processing and fanout, not just Redis transit. Markers for another fixture user increment `wrong_recipient_probes` and fail the run. Ordinary unrelated notifications are ignored.

Use these additional acceptance gates; they cannot be inferred from a successful handshake counter:

- During the full hold window, confirm approximately 20,000 live authenticated clients. Compare client auth failures with `sum(vikunja_websocket_connections)`; the server gauge includes sockets awaiting authentication. Inspect connections by pod to detect skew. Existing long-lived sockets do not move when replicas are added.
- Keep all application pods ready, with no OOM kills, restarts, growing event backlog, growing memory, connection exhaustion or storage failures. Record per-pod CPU, RSS, goroutines, descriptors and event-handler latency, plus Redis connections/network and ingress errors. Check `vikunja_websocket_resyncs_total` for unexpected client eviction.
- Verify cross-pod delivery using ingress/upstream observability or a staging routing rule that directs probe HTTP traffic to a different pod from its recipient's socket. With ordinary load balancing most probes should cross pods, but the harness does not assert pod identity.
- Repeat at increasing active ratios and action rates. Add representative searches, large projects, attachments, imports and webhooks for the enterprise's real workload. Idle sockets and small fixture tasks do not exercise those bottlenecks.
- During separate fault runs, terminate one app pod, roll the deployment, restart/fail over Redis, interrupt a publishing connection, and remove a scheduler leader. Record reconnection time, readiness, duplicates and missing notifications. Baseline thresholds should fail when the drill causes loss; do not call the drill a clean baseline by hiding those failures.
- Verify the real frontend reloads persisted notifications after reconnect and eventually after a publisher-only failure. This k6 client measures live delivery; it does not implement the browser's authoritative refresh or claim Pub/Sub replays missed events. Inspect the database/API to distinguish persisted notifications from lost transient delivery.

Repeat sustained runs long enough to expose leaks and queue growth. Preserve the workload parameters, server commit, hardware, limits, pod placement, dependency versions and time-series evidence with the result. A single pass does not establish safe headroom for every deployment.

## Verification status

Verified locally with k6 v2.3.0: syntax/imports, a mixed read/edit/probe protocol mock, and a negative mock that deliberately dropped notifications and failed the delivery threshold. A real two-process smoke then used two registered users, a shared SQLite fixture database and real Redis. A proxy forced every socket onto process one and every HTTP request onto process two. All 32 API actions and 16 cross-process notification probes succeeded; notification p95 was 186 ms. This validates the harness routes and event integration, not ten-machine capacity. See the [verification record](../../plans/horizontal-scaling/verification.md) and record real cluster results there separately.
