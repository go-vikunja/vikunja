---
title: "Releasing a new Vikunja version"
date: 2022-10-28T13:06:05+02:00
draft: false
menu:
  sidebar:
    parent: "development"
---

# Releasing a new Vikunja version

This checklist is a collection of all steps usually involved when releasing a new version of Vikunja.
Not all steps are necessary for every release.

* Website update
	* New Features: If there are new features worth mentioning the feature page should be updated.
	* New Screenshots: If an overhaul of an existing feature happened so that it now looks different from the existing screenshot, a new one is required.
* Generate changelogs (with git-cliff)
* Tag a new version: Include the changelog for that version as the tag message
	* Once built: Prune the cloudflare cache so that the new versions show up at [dl.vikunja.io](https://dl.vikunja.io/)
    * Update the [Flathub desktop package](https://github.com/flathub/io.vikunja.Vikunja)
* Release Highlights Blogpost
	* Include a section about Vikunja in general (totally fine to copy one from the earlier blog posts)
	* New Features & Improvements: Mention bigger features, potentially with screenshots. Things like refactoring are sometimes also worth mentioning.
* Publish
	* Reddit
	* Twitter
	* Mastodon
	* Chat
	* Newsletter
	* Forum
	* If features in the release were sponsored, send an email to relevant stakeholders
* Update Vikunja Cloud version and other instances
