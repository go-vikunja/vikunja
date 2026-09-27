# WebSocket API

Vikunja exposes a WebSocket endpoint at `/api/v1/ws`. The connection is
unauthenticated at upgrade time; authentication happens via the first message
so browsers (which cannot set an `Authorization` header on a WebSocket) can
connect too.

## Protocol

All messages are JSON. See `messages.go` for the exact wire format.

### Authenticating

Send an `auth` message with a token. Both token types accepted by the REST API
work here:

- a JWT from the login flow, or
- an API token (the `tk_…` tokens from the API tokens endpoints).

```json
{"action": "auth", "token": "tk_…"}
```

The server replies `{"action": "auth.success", "success": true}` or closes the
connection with `{"error": "invalid_token"}`. Every rejected token produces the
same client-facing error, regardless of why it was rejected.

### Subscribing

```json
{"action": "subscribe", "event": "task.updated"}
```

Unsubscribe with `action: "unsubscribe"`. Subscribing before authenticating is
rejected with `{"error": "auth_required"}`.

### Events

| Event | Data |
|---|---|
| `notification.created` | the notification |
| `timer.created` / `timer.updated` / `timer.deleted` | the time entry |
| `task.created` / `task.updated` / `task.deleted` | the task |
| `task.comment.created` / `task.comment.edited` / `task.comment.deleted` | `{comment, mentions}` |

Task and comment events are only delivered to connections whose authenticated
user can read the task's project at delivery time; access is re-checked per
event, not at subscribe time.

Comment events carry mention metadata so clients can render mentions without
an extra API round-trip. `mentions` is a list of `{id, username, name}`
objects, sorted by user ID, of the users mentioned in the comment text:

```json
{
  "event": "task.comment.created",
  "data": {
    "comment": {…},
    "mentions": [{"id": 2, "username": "user2", "name": ""}]
  }
}
```
