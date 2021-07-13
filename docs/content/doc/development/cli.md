---
date: "2019-03-31:00:00+01:00"
title: "Cli Commands"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "development"
---

# Adding new cli commands

All cli-related functions are located in `pkg/cmd`.
Each cli command usually calls a function in another package.
For example, the `vikunja migrate` command calls `migration.Migrate()`. 

Vikunja uses the amazing [cobra](https://github.com/spf13/cobra) library for its cli.
Please refer to its documentation for informations about how to use flags etc.

To add a new cli command, add something like the following:

{{< highlight golang >}}
func init() {
	rootCmd.AddCommand(myCmd)
}

var myCmd = &cobra.Command{
	Use:   "My-command",
	Short: "A short description about your command.",
	Run: func(cmd *cobra.Command, args []string) {
		// Call other functions
	},
}
{{</ highlight >}}