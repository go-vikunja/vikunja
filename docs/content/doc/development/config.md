---
date: "2019-02-12:00:00+02:00"
title: "Configuration Options"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "development"
---

# Configuration options

All configuration variables are declared in the `config` package.
It uses [viper](https://github.com/spf13/viper) under the hood to handle setting defaults and parsing config files.
Viper handles parsing all different configuration sources.

## Adding new config options

To make handling configuration parameters a bit easier, we introduced a `Key` string type in the `config` package which 
you can call directly to get a config value.

To add a new config option, you should add a new key const to `pkg/config/config.go` and possibly a default value.
Default values should always enable the feature to work or turn it off completely if it always needs
additional configuration.

Make sure to also add the new config option to the default config file (`config.yml.sample` at the root of the repository) 
with an explanatory comment to make sure it is well documented.
Then run `mage generate-docs` to generate the configuration docs from the sample file.

## Getting Configuration Values

To retreive a configured value call the key with a getter for the type you need.
For example:

{{< highlight golang >}}
if config.CacheEnabled.GetBool() {
	// Do something with enabled caches
}
{{< /highlight >}}

Take a look at the methods declared on the type to see what's available.
