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

## Google / Google Workspace

Vikunja Config:

```yaml
openid:
    enabled: true
    redirecturl: https://vikunja.mydomain.com/auth/openid/  <---- slash at the end is important
    providers:
      - name: Google
        authurl: https://accounts.google.com
        clientid: <google-oauth-client-id>
        clientsecret: <google-oauth-client-secret>
```

Google config:

- Navigate to `https://console.cloud.google.com/apis/credentials` in the target project
- Create a new OAuth client ID
- Configure an authorized redirect URI of `https://vikunja.mydomain.com/auth/openid/google`

Note that there currently seems to be no way to stop creation of new users, even when `enableregistration` is `false` in the configuration. This means that this approach works well only with an "Internal Organization" app for Google Workspace, which limits the allowed users to organizational accounts only. External / public applications will potentially allow every Google user to register.

## Keycloak 

Vikunja Config:
```yaml
openid:
    enabled: true
    redirecturl: https://vikunja.mydomain.com/auth/openid/  <---- slash at the end is important
    providers:
      - name: Keycloak
        authurl: https://keycloak.mydomain.com/realms/<relam-name>
        logouturl: https://keycloak.mydomain.com/realms/<relam-name>/protocol/openid-connect/logout
        clientid: <vikunja-id>
        clientsecret: <vikunja secret>
```
Keycloak Config:
- Navigate to the keycloak instance
- Create a new client with the type `OpenID Connect` and a unique ID.
- Set `Client authentication` to On
- Set `Root Url` to `https://vikunja.mydomain.com`
- Set `Valid redirect URIs` to `/auth/openid/keycloak`
- Create the client the navigate to the credentials tab and copy the `Client secret`
