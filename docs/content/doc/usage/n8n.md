---
title: "n8n"
date: 2023-10-24T19:31:35+02:00
draft: false
menu:
    sidebar:
        parent: "usage"
---

# Using Vikunja with n8n

Vikunja maintains a [community node](https://github.com/go-vikunja/n8n-vikunja-nodes) for [n8n](https://n8n.io),
allowing you to easily integrate Vikunja with all kinds of other tools and services.

{{< table_of_contents >}}

## Installation

To install the node in your n8n installation:

1. In your n8n instance, go to **Settings > Community Nodes**.
2. Select Install.
3. Enter `n8n-nodes-vikunja` as the npm Package Name
4. Agree to the risks of using community nodes: select I understand the risks of installing unverified code from a
   public source.
5. Select Install. n8n installs the node, and returns to the Community Nodes list in Settings.
6. Vikunja actions and triggers are now available in n8n.

[Official n8n docs about the installation](https://docs.n8n.io/integrations/community-nodes/installation/)

## Authentication

To authenticate your automation against Vikunja:

1. In Vikunja, go to **Settings > API Tokens** and create a new token. Use all scopes for the kind of task you want to
   do. \
   *Note:* If you want to use the webhook trigger node, the api token should have permissions to create, read and delete
   webhooks.
2. Now in n8n, go to **Credentials** and then click on **Add Credential**.
3. Search for `Vikunja API` and click *Continue*
4. Enter the API key you created in step 1.
5. Enter the API URL of your Vikunja instance, with `/api/v1` suffix.
6. When you now create a Vikunja node, select the created credentials.
