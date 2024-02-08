---
title: "Webhooks"
date: 2023-10-17T19:51:32+02:00
draft: false
type: doc
menu:
  sidebar:
    parent: "usage"
---

# Webhooks

Starting with version 0.22.0, Vikunja allows you to define webhooks to notify other services of events happening within Vikunja.

{{< table_of_contents >}}

## How to create webhooks

To create a webhook, in the project options select "Webhooks". The form will allow you to create and modify webhooks.

Check out [the api docs](https://try.vikunja.io/api/v1/docs#tag/webhooks) for information about how to create webhooks programatically.

## Available events and their payload

All events registered as webhook events in [the event listeners definition](https://kolaente.dev/vikunja/vikunja/src/branch/main/pkg/models/listeners.go#L69) can be used as webhook target.

A webhook payload will look similar to this:

```json
{
	"event_name": "task.created",
	"time": "2023-10-17T19:39:32.924194436+02:00",
	"data": {}
}
```

The `data` property will contain the raw event data as it was registered in the `listeners.go` file.

The `time` property holds the time when the webhook payload data was sent.
It always uses the ISO 8601 format with date, time and time zone offset.

## Security considerations

### Signing

Vikunja allows you to provide a secret when creating the webhook.
If you set a secret, all outgoing webhook requests will contain an `X-Vikunja-Signature` header with an HMAC signature over the webhook json payload.

Check out [webhooks.fyi](https://webhooks.fyi/security/hmac) for more information about how to validate the HMAC signature.

### Hosting webhook infrastructure

Vikunja has support to use [mole](https://github.com/frain-dev/mole) as a proxy for outgoing webhook requests.
This allows you to prevent SSRF attacts on your own infrastructure.

You should use this and [configure it appropriately]({{< ref "../setup/config.md">}}#webhooks) if you're not the only one using your Vikunja instance.

Check out [webhooks.fyi](https://webhooks.fyi/best-practices/webhook-providers#implement-security-on-egress-communication) for more information about the attack vector and reasoning to prevent this.
