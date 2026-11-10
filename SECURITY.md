# Security Policy

## Reporting

Report vulnerabilities privately through [GitHub security advisories](https://github.com/go-vikunja/vikunja/security/advisories/new) or the [contact information on our website](https://vikunja.io/contact/#security). Don't open public issues for them.

Please verify against the latest release or `main` before reporting.

## About Vikunja

Vikunja is a self-hosted to-do app: a Go API, a Vue frontend, and native desktop and mobile apps. One admin usually runs an instance for a small group of users who share projects with each other.

Trust model:

- **The admin is fully trusted.** They control the config, the database, the OIDC provider, and whatever the server sends to browsers.
- **Users are authenticated but don't trust each other.** Access to everything (tasks, comments, attachments, buckets, views) comes from project permissions (read, write, admin), granted through user shares, team shares, or link shares.
- **Anonymous visitors** can only reach login, registration, password reset, link-share authentication, and public instance info.

## What we care about

- One user reading or changing another user's projects, tasks, or anything attached to them.
- A read-only member gaining write or admin permissions.
- Bypassing authentication: passwords, TOTP, OIDC, API tokens, link shares, CalDAV/feed BasicAuth, session refresh.
- Stored XSS through user content (task titles, descriptions, comments, project descriptions) in the frontend or in emails.
- SSRF through outgoing requests (webhooks, avatar downloads, migration imports) while `outgoingrequests.allownonroutableips` is off.
- File handling: attachments, avatar uploads, and imports (path traversal, decompression bombs, size limit bypass).
- Resource exhaustion that one unauthenticated request or a cheap authenticated one can trigger.

## Recurring reports that are not vulnerabilities

These come in regularly. Don't report them unless you can show impact beyond what's described here.

**User or email enumeration through password reset or login.** Registration has to say whether an address is taken, so that information is already public. Reset and login give clear errors on purpose. All unauthenticated routes are rate-limited per IP (`ratelimit.noauthlimit`, which can't be disabled).

**The OAuth2 server accepts unregistered `client_id`s, has no client secret, or has no consent screen.** `/oauth/authorize` and `/oauth/token` exist only for Vikunja's own native apps. They implement the public-client flow from RFC 8252: PKCE (S256) is mandatory, and `redirect_uri` must use a `vikunja-*` scheme or be a loopback `http` URL. Public clients can't keep a secret, which is why PKCE exists. A consent screen wouldn't help either: any local app can claim to be Vikunja, and malware on the user's machine can read their tokens anyway.

**Things only an admin can do.** Examples: malicious config values ending up in the served frontend, enabling `outgoingrequests.allownonroutableips` and then reaching internal hosts, or enabling OIDC `emailfallback` with a provider that doesn't verify emails.

**SSRF through a configured proxy.** Requests sent via `outgoingrequests.proxyurl` aren't IP-checked, because the proxy resolves the target and filtering is its job. See [the webhook docs](https://vikunja.io/docs/webhooks/#hosting-security).

**CSRF on CalDAV or feeds.** These routes send no CORS headers, and every state-changing CalDAV method is non-simple, so the browser's preflight blocks cross-origin writes. GET requests change nothing, and their responses can't be read cross-origin.

**A permission check looks missing in a handler.** Permission checks (`CanRead`, `CanUpdate`, ...) run before the handler and derive from the parent project or task. Only report this with a request that succeeds against a running instance.

**EXIF metadata in attachments.** Files are stored exactly as uploaded, on purpose.

**Archived projects can still be changed in some ways.** Archiving hides projects; it isn't access control. Please open a regular bug report.

**Duplicating a project copies its shares.** The new owner can change them.

**Large `per_page` values.** They're capped by `service.maxitemsperpage`.

## Writing a good report

- Include the Vikunja version (from `/api/v1/info`).
- Include a bash or curl script that reproduces the issue against a fresh instance. Reports without one take much longer to triage.
- Keep it short. We get a lot of these reports and triaging means reading every one of them. Lead with an executive summary that explain the vulnerability and the impact in 1-2 senteces at most.
- If an AI wrote the report, check its claims against a running instance first.

Bugs without security impact belong in [public issues](https://github.com/go-vikunja/vikunja/issues).
