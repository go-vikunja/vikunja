---
title: "Cron Tasks"
date: 2021-07-13T23:21:52+02:00
draft: false
menu:
  sidebar:
    parent: "development"
---

# How to add a cron job task

Cron jobs are tasks which run on a predefined schedule.
Vikunja uses these through a light wrapper package around the excellent [github.com/robfig/cron](https://github.com/robfig/cron) package.

The package exposes a `cron.Schedule` method with two arguments: The first one to define the schedule when the cron task 
should run, and the second one with the actual function to run at the schedule.
You would then create a new function to register your the actual cron task in your package.

A basic function to register a cron task looks like this:

{{< highlight golang >}}
func RegisterSomeCronTask() {
	err := cron.Schedule("0 * * * *", func() {
		// Do something every hour
	}
}
{{< /highlight >}}

Call the register method in the `FullInit()` method of the `init` package to actually register it.

## Schedule Syntax

The cron syntax uses the same on you may know from unix systems.

It is described in detail [here](https://pkg.go.dev/github.com/robfig/cron#hdr-CRON_Expression_Format).
