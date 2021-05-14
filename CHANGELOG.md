# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

All releases can be found on https://code.vikunja.io/api/releases.

## [0.17.0] - 2021-05-14

### Added

* Add a "done" option to kanban buckets (#821)
* Add arm64 builds
* Add basic auth for metrics endpoint
* Add bucket limit validation
* Add crud endpoints for notifications (#801)
* Add endpoint to remove a list background
* Add events (#777)
* Add github funding link
* Add link share password authentication (#831)
* Add names for link shares (#829)
* Add notifications package for easy sending of notifications (#779)
* Add reminders for overdue tasks (#832)
* Add repeat monthly setting for tasks (#834)
* Add security information to readme
* Add separate docker manifest file for latest docker images
* Add systemd service file to linux packages
* Add test for moving a task to another list
* Enable searching users by full email or name
* Expose tls parameter of Go MySQL driver to config file (#855)
* Pagingation for tasks in kanban buckets (#805)

### Changed

* Change keyvalue.Get to return if a value exists or not instead of an error
* Change main branch to main
* Change test file names to unstable
* Change the name of the newly created bucket from "New Bucket" to "Backlog"
* Change unstable versions in migration tests
* Check if we're on main and change the version name accordingly if that's the case
* Cleanup listener names
* Cleanup old docs themes submodule
* Disable deb repo in drone
* Don't keep old releases from os packages when releasing for master
* Don't try to get users for tasks if no tasks were found when looking for reminders
* Explicitly add docker build step for latest
* Explicitly check if there are Ids before trying to get items by a list of Ids
* Improve duration format of overdue tasks in reminders
* Improve loading labels performance (#824)
* Improve sending overdue task reminders by only sending one for all overdue tasks
* Make sure all tables are properly pluralized
* Only send reminders for undone tasks
* Re-Enable migration test steps in pipeline
* Refactor getting all namespaces
* Remove unused tools from tools.go
* Run all lint checks at once
* Send a notification to the user when they are added to the list
* Show empty avatar when the user was not found
* Subscribe a user to a task when they are assigned to it
* Subscriptions and notifications for namespaces, tasks and lists (#786)
* Switch building the docs to download the theme instead of building
* Switch telegram notifications to matrix notifications
* Temporarily disable migration step
* Temporary build fix
* Update changelog
* Update copyright year
* Update README (#858)
* Use golang's tzdata package to handle time zones

### Fixed

* Explicitly set darwin-10.15 when building binaries
* Fix build
* Fix checking list rights when accessing a bucket
* Fix /dav/principals/*/ throwing a server error when accessed with GET instead of PROPFIND (#769)
* Fix deleting task relations
* Fix docs
* Fix drone file
* Fix due dates with times when migrating from todoist
* Fix event error handler retrying infinitely
* Fix filter for task index
* Fix getting lists for shared, favorite and saved lists namespace
* Fix getting user info from /user endpoint for link shares
* Fix IncrBy and DecrBy in memory keyvalue implementation if there was no value set previously
* Fix lint
* Fix matrix notify room id
* Fix moving repeating tasks to the done bucket
* Fix multiarch docker image building
* Fix not able to make saved filters favorite
* Fix notifications table not being created on initial setup
* Fix resetting the bucket limit
* Fix retrieving over openid providers if there are none
* Fix sending notifications to users if the user object didn't have an email
* Fix setting the user in created_by when uploading an attachment
* Fix shared lists showing up twice
* Fix tests
* Fix the shared lists pseudo namespace containing owned lists
* Fix unstable version build file names
* Fix user uploaded avatars
* Pin golang alpine builder image to 3.12 to fix builds on arm
* Revert "Update alpine Docker tag to v3.13 (#768)"

### Dependency Updates

* Update alpine Docker tag to v3.13 (#768)
* Update github.com/gordonklaus/ineffassign commit hash to 2e10b26 (#803)
* Update github.com/gordonklaus/ineffassign commit hash to d0e41b2 (#780)
* Update golang.org/x/crypto commit hash to 0c34fe9 (#822)
* Update golang.org/x/crypto commit hash to 3497b51 (#853)
* Update golang.org/x/crypto commit hash to 38f3c27 (#854)
* Update golang.org/x/crypto commit hash to 4f45737 (#836)
* Update golang.org/x/crypto commit hash to 513c2a4 (#817)
* Update golang.org/x/crypto commit hash to 5bf0f12 (#839)
* Update golang.org/x/crypto commit hash to 5ea612d (#797)
* Update golang.org/x/crypto commit hash to 83a5a9b (#840)
* Update golang.org/x/crypto commit hash to b8e89b7 (#793)
* Update golang.org/x/crypto commit hash to c07d793 (#861)
* Update golang.org/x/crypto commit hash to cd7d49e (#860)
* Update golang.org/x/crypto commit hash to e6e6c4f (#816)
* Update golang.org/x/crypto commit hash to e9a3299 (#851)
* Update golang.org/x/image commit hash to 4410531 (#788)
* Update golang.org/x/image commit hash to 55ae14f (#787)
* Update golang.org/x/image commit hash to 7319ad4 (#852)
* Update golang.org/x/image commit hash to ac19c3e (#798)
* Update golang.org/x/oauth2 commit hash to 0101308 (#776)
* Update golang.org/x/oauth2 commit hash to 01de73c (#762)
* Update golang.org/x/oauth2 commit hash to 16ff188 (#789)
* Update golang.org/x/oauth2 commit hash to 22b0ada (#823)
* Update golang.org/x/oauth2 commit hash to 2e8d934 (#827)
* Update golang.org/x/oauth2 commit hash to 5366d9d (#813)
* Update golang.org/x/oauth2 commit hash to 5e61552 (#833)
* Update golang.org/x/oauth2 commit hash to 6667018 (#783)
* Update golang.org/x/oauth2 commit hash to 81ed05c (#848)
* Update golang.org/x/oauth2 commit hash to 8b1d76f (#764)
* Update golang.org/x/oauth2 commit hash to 9bb9049 (#796)
* Update golang.org/x/oauth2 commit hash to af13f52 (#773)
* Update golang.org/x/oauth2 commit hash to ba52d33 (#794)
* Update golang.org/x/oauth2 commit hash to cd4f82c (#815)
* Update golang.org/x/oauth2 commit hash to d3ed898 (#765)
* Update golang.org/x/oauth2 commit hash to f9ce19e (#775)
* Update golang.org/x/sync commit hash to 036812b (#799)
* Update golang.org/x/term commit hash to 6a3ed07 (#800)
* Update golang.org/x/term commit hash to 72f3dc4 (#828)
* Update golang.org/x/term commit hash to a79de54 (#850)
* Update golang.org/x/term commit hash to b80969c (#843)
* Update golang.org/x/term commit hash to c04ba85 (#849)
* Update golang.org/x/term commit hash to de623e6 (#818)
* Update golang.org/x/term commit hash to f5beecf (#845)
* Update module adlio/trello to v1.9.0 (#825)
* Update module coreos/go-oidc to v3 (#760)
* Update module gabriel-vasile/mimetype to v1.2.0 (#812)
* Update module gabriel-vasile/mimetype to v1.3.0 (#857)
* Update module getsentry/sentry-go to v0.10.0 (#792)
* Update module go-redis/redis/v8 to v8.4.10 (#771)
* Update module go-redis/redis/v8 to v8.4.11 (#774)
* Update module go-redis/redis/v8 to v8.4.9 (#770)
* Update module go-redis/redis/v8 to v8.5.0 (#778)
* Update module go-redis/redis/v8 to v8.6.0 (#795)
* Update module go-sql-driver/mysql to v1.6.0 (#826)
* Update module go-testfixtures/testfixtures/v3 to v3.5.0 (#761)
* Update module go-testfixtures/testfixtures/v3 to v3.6.0 (#838)
* Update module iancoleman/strcase to v0.1.3 (#766)
* Update module imdario/mergo to v0.3.12 (#811)
* Update module jgautheron/goconst to v1 (#804)
* Update module labstack/echo/v4 to v4.2.0 (#785)
* Update module labstack/echo/v4 to v4.2.1 (#810)
* Update module labstack/echo/v4 to v4.2.2 (#830)
* Update module labstack/echo/v4 to v4.3.0 (#856)
* Update module lib/pq to v1.10.0 (#809)
* Update module lib/pq to v1.10.1 (#841)
* Update module mattn/go-sqlite3 to v1.14.7 (#835)
* Update module olekukonko/tablewriter to v0.0.5 (#782)
* Update module prometheus/client_golang to v1.10.0 (#819)
* Update module spf13/afero to v1.6.0 (#820)
* Update module spf13/cobra to v1.1.2 (#781)
* Update module spf13/cobra to v1.1.3 (#784)
* Update module src.techknowlogick.com/xgo to v1.3.0+1.16.0 (#791)
* Update module src.techknowlogick.com/xgo to v1.4.0+1.16.2 (#814)
* Update module stretchr/testify to v1.7.0 (#763)

## [0.16.1] - 2021-04-22

### Fixed

* Fix checking list rights when accessing a bucket
* Remove old deb-structure ci step
* Fix docker from

## [0.16.0] - 2021-01-10

### Added

* Add colors for caldav (#738)
* Add email reminders (#743)
* Add "like" filter comparator
* Add login via email (#740)
* Add Microsoft Todo migration (#737)
* Add name field to users
* Add support for migrating todoist boards (#732)
* Add task filter for assignees (#746)
* Add task filter for labels (#747)
* Add task filter for lists and namespaces (#748)
* Add task filter for reminders (#745)
* Add task filters for kanban
* Add testing endpoint to reset db tables (#716)
* Add tests for sending task reminders (#757)
* Add trello migration (#734)
* Authentication with OpenID Connect providers (#713)

### Fixed

* Fix completion status in DAV for OpenTasks and multiline descriptions (#697)
* Fix docs about caldav tasks.org
* Fix drone badge in README
* Fix getting current user when updating avatar or user name
* Fix go header lint
* Fix /info endpoint 500 error when no openid providers were configured
* Fix missing auto increments from b0d4902406 on mysql
* Fix not possible to create tasks if metrics were enabled
* Fix password reset without a reseet token
* Fix task updated timestamp not being updated in the response after updating a task

### Changed

* Change avatar endpoint
* Change license to AGPLv3
* Clarify docs about cors configuration
* Don't create a list identifier by default
* Make sure all int64 db fields are using bigint when actually storing the data (#741)
* Make sure a password reset token can be used only once
* Make the debian repo structure for buster instead of strech
* Refactor adding more details to tasks (#739)
* Simplify updating task reminders
* Update code header template
* Update github.com/gordonklaus/ineffassign commit hash to 3b93a88 (#701)
* Update github.com/gordonklaus/ineffassign commit hash to 8eed68e (#755)
* Update github.com/jgautheron/goconst commit hash to b58d7cf (#702)
* Update github.com/jgautheron/goconst commit hash to ccae5bf (#712)
* Update github.com/jgautheron/goconst commit hash to f8e4fe8 (#703)
* Update golang.org/x/crypto commit hash to 0c6587e (#706)
* Update golang.org/x/crypto commit hash to 5f87f34 (#729)
* Update golang.org/x/crypto commit hash to 8b5274c (#733)
* Update golang.org/x/crypto commit hash to 9d13527 (#736)
* Update golang.org/x/crypto commit hash to be400ae (#719)
* Update golang.org/x/crypto commit hash to c8d3bf9 (#710)
* Update golang.org/x/crypto commit hash to eec23a3 (#749)
* Update golang.org/x/image commit hash to 35266b9 (#727)
* Update golang.org/x/lint commit hash to 83fdc39 (#728)
* Update golang.org/x/oauth2 commit hash to 08078c5 (#722)
* Update golang.org/x/oauth2 commit hash to 0b49973 (#718)
* Update golang.org/x/oauth2 commit hash to 9fd6049 (#714)
* Update golang.org/x/sync commit hash to 09787c9 (#725)
* Update golang.org/x/sync commit hash to 67f06af (#695)
* Update golang.org/x/term commit hash to 2321bbc (#731)
* Update golang.org/x/term commit hash to ee85cb9 (#726)
* Update module cweill/gotests to v1.6.0 (#752)
* Update module fzipp/gocyclo to v0.3.1 (#696)
* Update module gabriel-vasile/mimetype to v1.1.2 (#708)
* Update module getsentry/sentry-go to v0.8.0 (#709)
* Update module getsentry/sentry-go to v0.9.0 (#723)
* Update module go-redis/redis/v8 to v8.4.4 (#742)
* Update module go-redis/redis/v8 to v8.4.6 (#756)
* Update module go-redis/redis/v8 to v8.4.7 (#758)
* Update module go-redis/redis/v8 to v8.4.8 (#759)
* Update module lib/pq to v1.9.0 (#717)
* Update module magefile/mage to v1.11.0 (#754)
* Update module mattn/go-sqlite3 to v1.14.5 (#711)
* Update module mattn/go-sqlite3 to v1.14.6 (#751)
* Update module pquerna/otp to v1.3.0 (#705)
* Update module prometheus/client_golang to v1.9.0 (#735)
* Update module spf13/afero to v1.5.0 (#724)
* Update module spf13/afero to v1.5.1 (#730)
* Update module src.techknowlogick.com/xgo to v1.2.0+1.15.6 (#720)
* Update module src.techknowlogick.com/xormigrate to v1.4.0 (#700)
* Update module swaggo/swag to v1.6.9 (#694)
* Update module swaggo/swag to v1.7.0 (#721)
* Update module ulule/limiter/v3 to v3.8.0 (#699)
* Update nfpm config for nfpm v2
* Use db sessions everywere (#750)

## [0.15.1] - 2020-10-20

### Fixed

* Fix not possible to create tasks if metrics were enabled

## [0.15.0] - 2020-10-19

### Added

* Add app support info for DAV (#692)
* Add better tests for namespaces
* Add caldav enabled/disabled to /info endpoint
* Add checks if tasks exist in maps before trying to access them
* Add config option to force ssl connections to connect with the mailer
* Add dav proxy directions to example proxy configurations
* Add docs about using vikunja with utf-8 characters
* Add FreeBSD guide to installation docs
* Add github sponsor link
* Add Golangci Lint (#676)
* Add mage command to create a new migration
* Add option to configure legal urls
* Add rootpath to deb command to not include everything in the deb file
* Add toc to docs
* Add update route to toggle team member admin status
* Add util function to move files
* Disable gocyclo for migration modules
* Favorite lists (#654)
* Favorite tasks (#653)
* Generate config docs from sample config (#684)
* Kanban bucket limits (#652)
* Key-Value Storages (#674)
* Manage users via cli (#632)
* Mention client_max_body_size in nginx proxy settings
* More avatar providers (#622)
* Return rights when reading a single item (#626)
* Saved filters (#655)

### Fixed

* Cleanup references to make
* Don't add a subtask to the top level of tasks to not add it twice in the list
* Fetch tasks for caldav lists (#641)
* Fix building for darwin with mage
* Fix creating lists with non ascii characters (#607)
* Fix decoding active users from redis
* Fix dockerimage build
* Fix docs index links
* Fix duplicating a list with background
* "Fix" gocyclo
* Fix loading list background information for uploaded backgrounds
* Fix migrating items with large items from todoist
* Fix nfpm command in drone
* Fix parsing todoist reminder dates
* Fix reading passwords on windows
* Fix release commands in drone
* Fix release trigger
* Fix release trigger in drone
* Fix token renew for link shares
* Fix trigger for pushing release artifacts to drone
* Fix updating team admin status
* Fix upload avatar not working
* Fix users with disabled totp but not enrolled being unable to login
* Makefile: make add EXTRA_GOFLAG to GOFLAGS (#605)
* Make sure built binary files are executable when compressing with upx
* Make sure lists which would have a duplicate identifier can still be duplicated
* Make sure the metrics map accesses only happen explicitly
* Make sure to copy the permissions as well when moving files
* Make sure to only initialize all variables when needed
* Make sure to require admin rights when modifying list/namespace users to be consistent with teams
* Make sure we have git installed when building os packages
* Make sure we have go installed when building os packages (for build step dependencies)
* Only check if a bucket limit is exceeded when moving a task between buckets
* Only try to download attachments from todoist when there is a url
* Pin telegram notification plugin in drone
* Regenerate swagger docs
* Skip directories when moving build release artefacts in drone
* Support absolute iCal timestamps in CalDAV requests (#691)
* Work around tasks with attachments not being duplicated

### Changed

* Replace renovate tokens with env
* Switch s3 release bucket to scaleway
* Switch to mage (#651)
* Testing improvements (#666)
* Update docs with testmail command + reorder
* Update github.com/asaskevich/govalidator commit hash to 29e1ff8 (#639)
* Update github.com/asaskevich/govalidator commit hash to 50839af (#637)
* Update github.com/asaskevich/govalidator commit hash to 7a23bdc (#657)
* Update github.com/asaskevich/govalidator commit hash to df4adff (#552)
* Update github.com/c2h5oh/datasize commit hash to 48ed595 (#644)
* Update github.com/gordonklaus/ineffassign commit hash to e36bfde (#625)
* Update github.com/jgautheron/goconst commit hash to 8f5268c (#658)
* Update github.com/shurcooL/vfsgen commit hash to 0d455de (#642)
* Update golang.org/x/crypto commit hash to 123391f (#619)
* Update golang.org/x/crypto commit hash to 5c72a88 (#640)
* Update golang.org/x/crypto commit hash to 7f63de1 (#672)
* Update golang.org/x/crypto commit hash to 84dcc77 (#678)
* Update golang.org/x/crypto commit hash to 948cd5f (#609)
* Update golang.org/x/crypto commit hash to 9e8e0b3 (#685)
* Update golang.org/x/crypto commit hash to ab33eee (#608)
* Update golang.org/x/crypto commit hash to afb6bcd (#668)
* Update golang.org/x/crypto commit hash to c90954c (#671)
* Update golang.org/x/crypto commit hash to eb9a90e (#669)
* Update golang.org/x/image commit hash to 4578eab (#663)
* Update golang.org/x/image commit hash to a67d67e (#664)
* Update golang.org/x/image commit hash to e162460 (#665)
* Update golang.org/x/image commit hash to e59bae6 (#659)
* Update golang.org/x/sync commit hash to 3042136 (#667)
* Update golang.org/x/sync commit hash to b3e1573 (#675)
* Update module 4d63.com/tz to v1.2.0 (#631)
* Update module fzipp/gocyclo to v0.2.0 (#686)
* Update module fzipp/gocyclo to v0.3.0 (#687)
* Update module getsentry/sentry-go to v0.7.0 (#617)
* Update module go-errors/errors to v1.1.1 (#677)
* Update module go-testfixtures/testfixtures/v3 to v3.4.0 (#627)
* Update module go-testfixtures/testfixtures/v3 to v3.4.1 (#693)
* Update module iancoleman/strcase to v0.1.0 (#636)
* Update module iancoleman/strcase to v0.1.1 (#645)
* Update module iancoleman/strcase to v0.1.2 (#660)
* Update module imdario/mergo to v0.3.10 (#615)
* Update module imdario/mergo to v0.3.11 (#629)
* Update module labstack/echo/v4 to v4.1.17 (#646)
* Update module lib/pq to v1.7.1 (#616)
* Update module lib/pq to v1.8.0 (#618)
* Update module mattn/go-sqlite3 to v1.14.1 (#638)
* Update module mattn/go-sqlite3 to v1.14.2 (#647)
* Update module mattn/go-sqlite3 to v1.14.3 (#661)
* Update module mattn/go-sqlite3 to v1.14.4 (#670)
* Update module prometheus/client_golang to v1.8.0 (#681)
* Update module spf13/afero to v1.3.2 (#610)
* Update module spf13/afero to v1.3.3 (#623)
* Update module spf13/afero to v1.3.4 (#628)
* Update module spf13/afero to v1.3.5 (#650)
* Update module spf13/afero to v1.4.0 (#662)
* Update module spf13/afero to v1.4.1 (#673)
* Update module spf13/cobra to v1.1.0 (#679)
* Update module spf13/cobra to v1.1.1 (#690)
* Update module spf13/viper to v1.7.1 (#620)
* Update module src.techknowlogick.com/xgo to v1.1.0+1.15.0 (#630)
* Update module src.techknowlogick.com/xgo to v1 (#613)
* Update module swaggo/swag to v1.6.8 (#680)
* Update renovate token
* Update src.techknowlogick.com/xgo commit hash to 7c2e3c9 (#611)
* Update src.techknowlogick.com/xgo commit hash to 96de19c (#612)
* update theme
* Update xgo to v1.0.0+1.14.6
* Use db sessions for task-related things (#621)
* Use nfpm to build deb, rpm and apk packages (#689)

## [0.14.1] - 2020-07-07

### Fixed

* Fix creating lists with non ascii characters (#607)
* Fix decoding active users from redis
* Fix parsing todoist reminder dates
* Make sure the metrics map accesses only happen explicitly

### Changed

* Update docs theme

## [0.14.0] - 2020-07-01

### Added 

* Add ability to run the docker container with configurable user and group ids
* Add better errors if the sqlite db file is not writable
* Add cache for initial unsplash collection
* Add docker setup guide from start to finish
* Add docs for restore
* Add dump command (#592)
* Add section to full-docker-example.md for Caddy v2 (#595)
* Add go version to version command
* Add list background information when getting all lists
* Add logging if downloading an image from unsplash fails
* Add migration test in drone (#585)
* Add option to disable totp for everyone
* Add plausible to docs
* Add restarting commands to all example docker compose files
* Add seperate docker pipeline for amd64 and arm
* Add test mail command (#571)
* Add todoist migrator to available migrators in info endpoint if it is enabled
* Add unsplash image proxy for images and thumbnails
* Add returning unsplash info when searching
* Don't return all tasks when a user has no lists
* Duplicate Lists (#603)
* Enable upload backgrounds by default
* Generate a random list identifier based on the list title
* List Backgrounds (#568)
* List Background upload (#582)
* Repeat tasks after completion (#587)
* Restore command (#593)
* Sentry integration (#591)
* Todoist Migration (#566)

### Fixed

* Ensure consistent naming of title fields (#528)
* Ensure task dates are in the future if a task has a repeating interval (#586)
* Fix caching of initial unsplash results per page
* Fix case-insensitive task search for postgresql (#524)
* Fix docker manifest build
* Fix docker multiarch build
* Fix docs theme build
* Fix getting unsplash thumbnails for non "photo-*" urls
* Fix migration 20200425182634
* Fix migration 20200516123847
* Fix migration to add position to task
* Fix misspell
* Fix namespace title not being updated
* Fix not loading timezones on all operating systems
* Fix proxying unsplash images (security)
* Fix removing existing sqlite files
* Fix resetting list, label & namespace colors
* Fix searching for unsplash pictures with words that contain a space
* Fix setting a list identifier to empty
* Fix sqlite db not working when creating a new one
* Fix sqlite path in default config
* Fix swagger docs
* Fix updating the index when moving a task
* Prevent crashing when trying to register with an empty payload
* Properly ping unsplash when using unsplash images
* Return errors when dumping
* Set the list identifier when creating a new task

### Changed

* Expose namespace id when querying lists
* Improve getting all namespaces performance (#526)
* Improve memory usage of dump by not loading all files in memory prior to adding them to the zip
* Improve metrics performance
* Load the list when setting a background
* Make the db timezone migration mysql compatible
* Make the `_unix` suffix optional when sorting tasks
* Migrate all timestamps to real iso dates (#594)
* Make sure docker images are only built when tests pass
* Remove build date from binary
* Remove dependencies on build step to speed up test pipeline (#521)
* Remove go mod vendor todo from pr template now that we don't keep dependencies in the repo anymore
* Remove migration dependency to models
* Remove min length for labels, lists, namespaces, tasks and teams
* Remove vendored dependencies
* Reorganize cmd init functions
* Set unsplash empty collection caching to one hour
* Simplify pipeline & add docker manifest step
* Update alpine Docker tag to v3.12 (#573)
* Update and fix staticcheck
* Update dependency github.com/mattn/go-sqlite3 to v1.14.0
* Update github.com/shurcooL/vfsgen commit hash to 92b8a71 (#599)
* Update golang.org/x/crypto commit hash to 279210d (#577)
* Update golang.org/x/crypto commit hash to 70a84ac (#578)
* Update golang.org/x/crypto commit hash to 75b2880 (#596)
* Update module go-redis/redis/v7 to v7.3.0 (#565)
* Update module go-redis/redis/v7 to v7.4.0 (#579)
* Update module go-testfixtures/testfixtures/v3 to v3.3.0 (#600)
* Update module lib/pq to v1.6.0 (#572)
* Update module lib/pq to v1.7.0 (#581)
* Update module prometheus/client_golang to v1.7.0 (#589)
* Update module prometheus/client_golang to v1.7.1 (#597)
* Update module spf13/afero to v1.3.0 (#588)
* Update module spf13/afero to v1.3.1 (#602)
* Update module spf13/cobra to v1 (#511)
* Update module src.techknowlogick.com/xormigrate to v1.2.1 (#574)
* Update module src.techknowlogick.com/xormigrate to v1.3.0 (#590)
* Update module stretchr/testify to v1.6.0 (#570)
* Update module stretchr/testify to v1.6.1 (#580)
* Update module swaggo/swag to v1.6.7 (#601)
* Update src.techknowlogick.com/xgo commit hash to 209a5cf (#523)
* Update src.techknowlogick.com/xgo commit hash to a09175e (#576)
* Update src.techknowlogick.com/xgo commit hash to eeb7c0a (#575)
* update theme
* Update theme
* Update web handler
* Update xorm.io/xorm 1.0.1 -> 1.0.2
* Use the db logger instance for logging migration related stuff

## [0.13.1] - 2020-05-19

### Fixed

* Don't get all tasks if a user has no lists

## [0.13] - 2020-05-12

#### Added

* Add 2fa for authentification (#383)
* Add categories to error docs
* Add changing email for users
* Add community link
* Add configuration options for log level
* Add creating a new first bucket when creating a new list
* Add docs for changing frontend url
* Add endpoint to disable totp auth
* Add endpoint to get the current users totp status
* Add explanation to docs about cors
* Add github token for renovate (#164)
* Add gosec static analysis
* Add moving tasks between lists (#389)
* Add real buckets for tasks which don't have one (#446)
* Add traefik 2 example configuration
* Configure Renovate (#159)
* Kanban (#393)
* Task filters (#243)
* Task Position (#412)

#### Fixed

* Add checking and logging when trying to put a task into a nonexisting bucket
* Fix bucket ID being reset with no need to do so
* Fix creating new things with a link share auth
* Fix dependencies
* Fix gosec in drone
* Fix link share creation & creating admin link shares without admin rights
* Fix moving tasks back into the empty (ID: 0) bucket
* Fix moving tasks in buckets
* Fix not moving its bucket when moving a task between lists
* Fix pagination count for task collection
* Fix parsing array style comparators by query param
* Fix reference to reverse proxies in docs
* Fix removing the last bucket
* Fix replace statements for tail
* Fix team rights not updating for namespace rights
* Fix tests after renaming json fields to snake_case
* Fix total label count when getting all labels (#477)
* Remove setting task bucket to 0
* Task Filter Fixes (#495)

#### Changed

* Change all json fields to snake_case
* Change totp secret datatype from varchar to text
* Update alpine Docker tag to v3.11 (#160)
* Update docs theme
* Update github.com/c2h5oh/datasize commit hash to 28bbd47 (#212)
* Update github.com/gordonklaus/ineffassign commit hash to 7953dde (#233)
* Update github.com/jgautheron/goconst commit hash to cda7ea3 (#228)
* Update github.com/shurcooL/httpfs commit hash to 8d4bc4b (#229)
* Update golang.org/x/crypto commit hash to 056763e (#222)
* Update golang.org/x/crypto commit hash to 06a226f (#504)
* Update golang.org/x/crypto commit hash to 0848c95 (#371)
* Update golang.org/x/crypto commit hash to 3c4aac8 (#419)
* Update golang.org/x/crypto commit hash to 44a6062 (#429)
* Update golang.org/x/crypto commit hash to 4b2356b (#475)
* Update golang.org/x/crypto commit hash to 4bdfaf4 (#438)
* Update golang.org/x/crypto commit hash to 729f1e8 (#458)
* Update golang.org/x/crypto commit hash to a76a400 (#411)
* Update golang.org/x/lint commit hash to 738671d (#223)
* Update module go-redis/redis to v6.15.7 (#234)
* Update module go-redis/redis to v6.15.7 (#290)
* Update module go-redis/redis to v7 (#277)
* Update module go-redis/redis to v7 (#309)
* Update module go-testfixtures/testfixtures/v3 to v3.1.2 (#457)
* Update module go-testfixtures/testfixtures/v3 to v3.2.0 (#505)
* Update module imdario/mergo to v0.3.9 (#238)
* Update module labstack/echo/v4 to v4.1.16 (#241)
* Update module lib/pq to v1.4.0 (#428)
* Update module lib/pq to v1.5.0 (#476)
* Update module lib/pq to v1.5.1 (#485)
* Update module lib/pq to v1.5.2 (#491)
* Update module olekukonko/tablewriter to v0.0.4 (#240)
* Update module prometheus/client_golang to v0.9.4 (#245)
* Update module prometheus/client_golang to v1
* Update module prometheus/client_golang to v1.6.0 (#463)
* Update module spf13/cobra to v0.0.7 (#271)
* Update module spf13/viper to v1.6.2 (#272)
* Update module spf13/viper to v1.6.3 (#291)
* Update module spf13/viper to v1.7.0 (#494)
* Update module stretchr/testify to v1.5.1 (#274)
* Update Renovate Configuration (#161)
* Update src.techknowlogick.com/xgo commit hash to bb0faa3 (#279)
* Update src.techknowlogick.com/xgo commit hash to c43d4c4 (#224)
* Update xorm redis cacher to use the xorm logger instead of a special seperate one
* Update xorm to v1 (#323)

## [0.12] - 2020-04-04

#### Added 

* Add support for archiving lists and namespaces (#152)
* Colors for lists and namespaces (#155)
* Add build time to compile flags
* Add proxying gravatar requests for user avatars (#148)
* Add empty avatar provider (#149)
* expand relative path ~/.config/vikunja to $HOME/.config/vikunja **WINDOWS** (#147)
* Show lists as archived if their namespace is archived

#### Fixed

* Workaround for timezones on windows (#151)
* Fix getting one namespace
* Fix getting the authenticated user with caldav
* Fix searching for config in home directories
* Fix updating lists with an identifier

#### Changed

* Change release bucket

## [0.11] - 2020-03-01

### Added 

* Add config options for cors handling (#124)
* Add config options for task attachments (#125)
* Add generate as a make dependency for make build
* Add logging for invalid model errors (#126)
* Add more logging to web handler methods
* Add postgres support (#135)
* Add rate limit by ip for non-authenticated routes (#127)
* Better efficency for loading teams (#128)
* Expand relative path ~/.config/vikunja to $HOME/.config/vikunja (#146)
* Task Comments (#138)

### Fixed

* Fix typo in docker-compose example (#140)
* Fix frontend url for wunderlist migration in docs
* Fix inserting task structure with related tasks (#142)
* Fix time zone settings not working in Docker
* Fix updating dates when marking a task as done (#145)
* Make sure the author is returned when creating a new comment
* Remove double user field

### Changed

* Explicitly disable wunderlist migration by default (#141)
* Migration Improvements (#122)
* Refactor User and DB handling (#123)
* Return iso dates for everything date related from the api (#130)
* Update copyright header
* Update theme
* Update xorm to use the new import path (#133)
* Use relative url in .gitmodules (#132)

## [0.10] - 2020-01-19

### Added

* Migration (#120)
* Endpoint to get tasks on a list (#108)
* Sort Order for tasks (#110)
* Add files volume to docker compose docs
* Add motd config option to docs
* Add option to disable registration (#117)
* Add task identifier (#115)
* Add tests for md5 generation (#111)
* Add user token renew (#113)

### Fixed

* Fix new tasks not getting a new task index (#116)
* Fix owner field being null for user shared namespaces (#119)
* Fix passing sort_by and order_by as query path arrays
* Fix sorting tasks by bool values
* Fix task collection tests
* Consistent copyright text in file headers (#112)

### Changed

* Task collection improvements (#109)
* Update copyright year (#118)
* Update docs with a traefik configuration
* Use redis INCRBY and DECRBY when updating metrics values (#121)
* Use utf8mb4 instead of plain utf8 (#114)
* Update docs theme

## [0.9] - 2019-11-24

### Added 

* Task Attachments (#104)
* Task Relations (#103)
* Add endpoint to get a single task (#106)
* Add file volume to the docker image
* Added extra depth to logging to correctly show the functions calling the logger in logs
* Added more infos to a link share auth (#98)
* Added percent done to tasks (#102)

### Fixed

* Fix default logging settings (#107)
* Fixed a bug where adding assignees or reminders via an update would re-create them and not respect already inserted ones, leaving a lot of garbage
* Fixed a bug where deleting an attachment would cause a nil panic
* Fixed building docs theme
* Fixed error when setting max file size on 32-Bit systems
* Fixed labels being displayed multiple times if they were associated with more than one task (#99)
* Fixed metrics on/off setting
* Fixed migration for task relations
* Fixed not getting all labels when retrieving a list with all tasks
* Fixed panic when using link share and metrics
* Fixed rate limit panic when authenticating with a link share auth token (#97)
* Fixed removing reminders
* Small link share fixes (#96)

### Changed 

* Improve pagination (#105)
* Moved `teams_{namespace|list}_*` to `{namespace|list}_teams_*` for better consistency (#101)
* Refactored getting all lists for a namespace (#100)
* Refactored getting task IDs for labels
* Switched default logger to stdout instead of stderr
* update docs theme

### Misc

* Move from markdown lists to Vikunja for roadmap

## [0.8] - 2019-09-01

### Added

* Better Caldav support (#73)
* Added settings for max open/idle connections and max connection lifetime (#74)
* /info endpoint (#85)
* Added http endpoint to list all users on a list (#87)
* Rate limits (#91)
* Sharing of lists via public links (#94)

### Changed

* Reminders now use an extra table (#75)
* Use the username instead of a full user object when adding a user to a team or giving it rights (#76)
* Add the md5-hashed user email to user objects for use with gravatar (#78)
* Use the auth methods to get IDs to avoid unneeded casts
* Better config handling with constants (#83)
* Statically compile templates in the final binary (#84)
* Use longtext instead of varchar(1000) on description fields (#88)
* Logger refactoring (#90)

### Fixed

* Fixed `listID` not being returned in tasks
* Fixed tests (#72)
* Fixed metrics endpoint not working
* Fixed check if the user really exists before updating/deleting its rights (#77)
* Fixed duedate spelling issue (#79)

### Misc

* Integration tests (#71)
* Make sure the version works when building in drone
* Switched to another version of xgo
* Simplified the docker image (#80)
* Update echo (#82)
* Compress binaries after building them (#81)
* Simplify structure by having less files (#86)
* Limit the test pipeline to run only on pull requests (#89)
* GetUser now returns a pointer (#93)
* Refactor ListTask to Task (#92)

## [0.7] - 2019-04-05

### Added

* DB migrations (#67)
* More cli options for Vikunja (#66 #68)
* Use query params to sort tasks instead of url params (#61)
* More config paths (#55)

### Fixed

* Fixed Priority not updating when setting it to 0
* Fixed getting lists by namespace
* Fixed rights check (#70 #62)
* Fixed labels not being queried correctly on tasks
* Fixed bulk update label tasks

### Changed

* Hide a user's email address everywhere (#69)
* Refactored `canRead()` to get the list before checking rights #65
* Let rights methods return errors (#64 #63)
* Improved Swagger docs for label tasks
* Docs improvements (#58)
* Logging Handling (#57)
* Rights performance improvements (#54)

### Misc

* Releases also as Debian packages (#56)

## [0.6] - 2019-01-16

### Added

* Added prometheus endpoint to get metrics (#33)
* More unit tests (#34)
* Tests can now use config files (#36)
* Redoc for swagger ui (#39, #46)
* Start and end dates for tasks (#40)
* Get tasks between a date range (#41)
* Bulk edit for tasks (#42)
* More ci checks (#43)
* Task assignees (#44, #47)
* Task labels (#45, #48)

### Fixed

* Fixed path to get all tasks (echo bug)
* Explicitly get the peudonamespace with all shared lists (#32)
* Properly init tabels Redis
* unexpected EOF when using metrics (#35)
* Task sorting in lists (#36)
* Various user fixes (#38)
* Fixed a bug where updating a list would update it with the same values it had

### Changed

* Simplified list rights check (#50)
* Refactored some structs to not expose unneded values via json (#52)

### Misc

* Updated libraries
* Updated drone to version 1
* Releases are now signed with our pgp key (more info about this on [the download page](https://vikunja.io/en/download/)).

## [0.5] - 2018-12-02

### Added

* Shared lists are now shown in a pseudonamespace with all other namespaces, has the ID -1
* Tasks can have multiple reminders
* Tasks can have subtasks. Subtasks are fully-fleged tasks, but not shown in the task list of a list.
* Tasks can have priorities

### Changed

* Validation not so verbose anymore
* [License](https://git.kolaente.de/vikunja/api/src/branch/master/LICENSE) is now GPLv3
* The crudhandler now has its [own repo](https://git.kolaente.de/vikunja/web) - you can use it in your own projects!

## [0.4] - 2018-11-16

#### Added

* Get all tasks for the authenticated user sorted by their due date
* CalDAV support
* Pagination for everything which returns an array
* Search all the things
* More validation for most of the structs
* Improved Swagger docs (available on `/api/v1/swagger`)

## [0.3] - 2018-11-02

### Added 

* Password reset
* Email verification when registering

Misc bugfixes and improvements to the build process

## [0.2] - 2018-10-17

## [0.1] - 2018-09-20

