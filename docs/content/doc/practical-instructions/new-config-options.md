---
date: "2019-02-12:00:00+02:00"
title: "Adding new   config options"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "practical instructions"
---

# Adding new config options

Vikunja uses [viper](https://github.com/spf13/viper) to handle configuration options.
It handles parsing all different configuration sources.

The configuration is done in sections. These are represented with a `.` in viper.
Take a look at `pkg/config/config.go` to understand how these are set.

To add a new config option, you should add a default value to `pkg/config/config.go`.
Default values should always enable the feature to work somehow, or turn it off completely if it always needs
additional configuration.

Make sure to add the new config option to [the config document]({{< ref "../setup/config.md">}}) and the default config file
(`config.yml.sample` at the root of the repository) to make sure it is well documented.

If you're using a computed value as a default, make sure to update the sample config file and debian
post-install scripts to reflect that.

To get a configured option, use `viper.Get("config.option")`.
Take a look at [viper's documentation](https://github.com/spf13/viper#getting-values-from-viper) to learn of the 
different ways available to get config options.