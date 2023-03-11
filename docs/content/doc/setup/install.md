---
date: "2019-02-12:00:00+02:00"
title: "Installing"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "setup"
    weight: 10
---

# Installing

Vikunja consists of two parts: [API](https://code.vikunja.io/api) and [frontend](https://code.vikunja.io/frontend).

You will always need to install at least the API.
To actually use Vikunja you'll also need to somehow install a frontend to use it.
You can either:

* [Install the web frontend]({{< ref "install-frontend.md">}})
* Use the desktop app, which is essentially a web frontend packaged for easy installation on desktop devices
* Use the mobile app only, but as of right now it only supports the very basic features of Vikunja

Vikunja can be installed in various ways. 
This document provides an overview and instructions for the different methods.

* [API]({{< ref "install-backend.md">}})
  * [Installing from binary]({{< ref "install-backend.md#install-from-binary">}})
    * [Verify the GPG signature]({{< ref "install-backend.md#verify-the-gpg-signature">}})
    * [Set it up]({{< ref "install-backend.md#set-it-up">}})
    * [Systemd service]({{< ref "install-backend.md#systemd-service">}})
    * [Updating]({{< ref "install-backend.md#updating">}})
    * [Build from source]({{< ref "install-backend.md#build-from-source">}})
  * [Docker]({{< ref "install-backend.md#docker">}})
  * [Debian packages]({{< ref "install-backend.md#debian-packages">}})
  * [Configuration]({{< ref "config.md">}})
  * [UTF-8 Settings]({{< ref "utf-8.md">}})
* [Frontend]({{< ref "install-frontend.md">}})
  * [Docker]({{< ref "install-frontend.md#docker">}})
  * [NGINX]({{< ref "install-frontend.md#nginx">}})
  * [Apache]({{< ref "install-frontend.md#apache">}})
  * [Updating]({{< ref "install-frontend.md#updating">}})
* [Reverse proxies]({{< ref "reverse-proxies.md">}})
* [Full docker example]({{< ref "full-docker-example.md">}})
* [Backups]({{< ref "backups.md">}})

## Installation on kubernetes

A third-party Helm Chart is available from the k8s-at-home project [here](https://github.com/k8s-at-home/charts/tree/master/charts/stable/vikunja).

## Other installation resources

* [Docker Compose is MUCH Easier Than you Think - Let's Install Vikunja](https://www.youtube.com/watch?v=fGlz2PkXjuo) (Youtube)
* [Setup Vikunja using Docker Compose - Homelab Wiki](https://thehomelab.wiki/books/docker/page/setup-vikunja-using-docker-compose)
* [A Closer look at Vikunja - Email Notifications - Enable or Disable Registrations - Allow Attachments](https://www.youtube.com/watch?v=47wj9pRT6Gw) (Youtube)
* [Install Vikunja in Docker for self-hosted Task Tracking](https://smarthomepursuits.com/install-vikunja-in-docker-for-self-hosted-task-tracking/)
* [Self-Hosted To-Do List with Vikunja in Docker](https://www.youtube.com/watch?v=DqyqDWpEvKI) (Youtube)
* [Vikunja self-hosted (step by step)](https://nguyenminhhung.com/vikunja-self-hosted-step-by-step/)
* [How to Install Vikunja on Your Synology NAS](https://mariushosting.com/how-to-install-vikunja-on-your-synology-nas/)
