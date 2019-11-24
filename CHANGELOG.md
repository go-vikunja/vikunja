# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

All releases can be found on https://code.vikunja.io/frontend/releases.

The releases aim at the api versions which is why there are missing versions.

## [0.9] - 2019-11-24

### Added

* Add minimal PWA (#34)
* Added caching to the docker image
* Added changing %Done on a task
* Added global api config (#31)
* Added handling if the user is offline (#35)
* Added labels for login and register inputs
* Added link sharing (#30)
* Added meta description tag
* Added support for HTTP/2 to the docker image
* Added the function to collapse all lists in a namespace in the sidebar menu

### Changed

* Correctly preload fonts
* Different edit icon
* Improved font handling
* Load the offline image quietly in the background
* Moved non-theme stuff in general.scss
* Removed rancher configuration
* Removed unused preload fonts tags
* Replace all spaces with tabs
* Show avatars of assigned users
* Sort tasks by done/undone first and then newest
* Task Detail View (#37)
* Update vue/cli-service
* Updated axios
* Updated dependencies
* Updated packages
* Updated packages to their latest versiosn
* Use the new listuser endpoint to search for users

### Fixed

* Fix edit label pane not closing when clicking on it
* Fixed gzip compression in docker
* Fixed label edit still opening when deleting a label
* Fixed menu not being visible on mobile
* Fixed namespace loading (#32)
* Fixed new task field not being reset after adding a new task
* Fixed redirect to login page (#33)
* Fixed scroll behaviour
* Fixed shared lists overflowing
* Fixed sharing with a user not working
* Fixed task update not working
* Fixed task update not working (again)
* Fixed team creating not working
* Handle task relations the right way (#36)

### Misc

* Moved markdown-based todo list to Vikunja [skip ci]
* Use yarn image instead of installing it every time

## [0.7] - 2019-04-30

### Added

* Design overhaul (#28)
* Gantt charts (#29)
* Pretty Scrollbars
* Task colors

### Fixed

* Fixed getting tasks (#27)

## [0.6] - 2019-03-08

### Added

* Labels (#25)
* Task priorites (#19)
* Task assingees (#21)

### Changed

* All requests are now using models and services, improving the development experience
* Team managing (#18)

## [0.5] - 2018-12-29

### Added

* User email verification when registering
* password reset
* Task overview
* Multiple reminders
* Repeating tasks
* Subtasks
* Task duration
* All new design
* Week and month view for tasks

### Changed

* Go to overview when clicking on the logo
* CSS improvements
* Don't show options to edit pseudonamespace
* Delay loading animation to not show it when the request finishes in < 100ms
* Use email instead of username when resetting a password

### Fixed
* Fixed trying to verify an email when there was none
* Fixed loading tasks when the user was not authenticated

## [0.1] - 2018-09-20

