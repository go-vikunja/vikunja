# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

All releases can be found on https://code.vikunja.io/api/releases.

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

