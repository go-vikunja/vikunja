---
date: "2022-08-09:00:00+02:00"
title: "OpenID example configurations"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "setup"
---

# OpenID example configurations

On this page you will find examples about how to set up Vikunja with a third-party OpenID provider.
To add another example, please [edit this document](https://kolaente.dev/vikunja/api/src/branch/main/docs/content/doc/setup/openid-examples.md) and send a PR.

{{< table_of_contents >}}

## Authelia

Vikunja Config:

```yaml
openid:
    enabled: true
    redirecturl: https://vikunja.mydomain.com/auth/openid/  <---- slash at the end is important
    providers:
      - name: Authelia
        authurl: https://login.mydomain.com
        clientid: <vikunja-id>
        clientsecret: <vikunja secret>
```

Authelia config:

```yaml
- id: <vikunja-id>
description: Vikunja
secret: <vikunja secret>
redirect_uris:
  - https://vikunja.mydomain.com/auth/openid/authelia
scopes:
  - openid
  - email
  - profile
```
