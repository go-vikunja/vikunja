---
date: "2023-03-09:00:00+02:00"
title: "Migration from third-party services"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "setup"
    weight: 5
---

# Migration from third-party services

There are several importers available for third-party services like Trello, Microsoft To Do or Todoist.
All available migration options can be found [here](https://kolaente.dev/vikunja/vikunja/src/branch/main/config.yml.sample#L218).

You can develop migrations for more services, see the [documentation]({{< ref "../development/migration.md">}}) for more info.

{{< table_of_contents >}}

## Trello

### Config Setup

Log into Trello and navigate to the [site](https://trello.com/app-key) to manage your API keys.
Save your `Personal Key` for later and add your Vikunja domain to the Allowed Origins list.

Create a `config.yml` file based on [default config file](https://kolaente.dev/vikunja/vikunja/src/branch/main/config.yml.sample) if you haven't already.

- Copy the [Trello options](https://kolaente.dev/vikunja/vikunja/src/branch/main/config.yml.sample#L233) from the default config file
- Set `enable` to true
- Set `key` to your [trello API key](https://trello.com/app-key) 
- Replace `<frontend url>` in `redirecturl` with your url

### Config Loading

To load the config with Vikunja, see the [installation]({{< ref "install.md">}}) documentation for instructions to load the `config.yml` file and start Vikunja.

### Config Loading with Docker Compose

In case you are using Docker Compose you need to edit `docker-compose.yml` to load `config.yml`.
Mount the `config.yml` file into the Vikunja container, by adding this line to the volumes of the Vikunja container and replacing the `./path/to/config.yml` with the relative path from the `docker-compose.yml` to your `config.yml`.
```yaml
volumes:
  - ./path/to/config.yml:/etc/vikunja/config.yml
```

After all the setup is done, start Vikunja as shown in the [Docker Compose setup]({{< ref "full-docker-example.md">}}). 

### Start the Migration Process

Log in, and navigate to Settings > Import from other services. In the list of available third-party services, there should be a Trello icon now.
If not, ensure that you are properly loading your config file. Refer to the Vikunja log to see if the config file was loaded or not.
In case the config file was loaded, and there is no Trello icon, make sure your [config setup](#config-setup) is correct.

Click on Trello and on Get Started. This will redirect you to Trello where you need to allow Vikunja Migration to access your account. In case there is an error when being directed to Trello, make sure that your Vikunja domain is in your Trello Allowed Origins list.
Once this is done, you will be redirected to Vikunja which should tell you that the migration is in progress now. Note that this can take up to several hours depending on the amount of boards in your Trello account.