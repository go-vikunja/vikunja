# Web Push notifications

Vikunja can deliver task activity, explicit task reminders, API token expiry notices and daily overdue summaries to installed Progressive Web Apps. Task data remains online-only. This feature does not add offline task editing or background synchronization.

## Requirements

- Serve Vikunja over HTTPS. Plain HTTP is accepted only for localhost development.
- Configure one stable VAPID P-256 key pair.
- On iPhone and iPad, use iOS or iPadOS 16.4 or later and install Vikunja on the Home Screen before enabling notifications.
- On Android, install Vikunja from a browser that supports the Push API, such as current Chrome.

Generate the key pair once:

```shell
vikunja webpush generate-keys
```

The command prints a ready-to-copy configuration block containing the public and private keys. Treat its output as a secret.

Configure the server:

```yaml
webpush:
  enabled: true
  publickey: "base64url-public-key"
  privatekey: "base64url-private-key"
```

`service.publicurl` is used as the VAPID contact subject. Web Push uses the common `outgoingrequests.proxyurl`, `outgoingrequests.proxypassword`, `outgoingrequests.timeoutseconds` and `outgoingrequests.allownonroutableips` settings.

Back up both VAPID keys and reuse them across upgrades and deployments. Replacing the pair invalidates existing browser subscriptions, so every device must enable notifications again.

## User behavior

Each user enables Push notifications on this device from General Settings. The browser permission prompt appears only after the user presses Enable. A subscription belongs to the current device, user and Vikunja session. Logout, remote session revocation, password changes, session expiry and account deletion invalidate it.

When Vikunja is visible, the open app refreshes its notification list instead of showing a system notification. Test notifications are always visible. When the app is hidden or closed, the operating system shows the notification. Clicking it opens the relevant task or settings page and marks the related in-app notification read when possible.

The existing overdue reminder enabled and time settings control both email and Push overdue summaries. The task email checkbox remains email-only, so explicit task reminders can still reach a Push-enabled device when reminder emails are off.

## Reliability and rollback

The server stores pending deliveries until the push service accepts them or their expiry window ends. Network failures and temporary push-service errors are retried across server restarts. Apple, Google and the operating system still control final display timing, including battery restrictions, connectivity, permission changes and Focus modes.

To roll back, set `webpush.enabled` to `false` and revert the application version. Keep the additive database tables and VAPID keys so existing subscriptions remain usable when the feature is enabled again.
