# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

All releases can be found on https://code.vikunja.io/frontend/releases.

The releases aim at the api versions which is why there are missing versions.

## [0.11] - 2020-03-01

### Added

* Add a button to the task detail page to mark a task as done
* Add a link to vikunja.io (#56)
* Add automatic user token renew (#43)
* Add auto save for task edit sidebar
* Add moment.js for date related things (#50)
* Add removing of tasks (#48)
* Add saving task title with ctrl+enter
* Add saving the description with ctrl+enter
* Add slight background change when hovering over a task in the list
* Add Wunderlist migration (#46)
* Task Comments (#66)
* Task Pagination (#38)
* Task Search (#52)
* Task sorting (#39)
* Notifications for task reminders (#57)
* PWA update available notification (#42)
* Set the end date to the same as the due date if a start date was set but no end date
* Show parent tasks in task overview list (#41)

### Fixed

* Fix textarea in task detail view not having a background when focused (#937 in Vikunja)
* Fix "Add a reminder" being shown
* Fix adding a task to an empty list
* Fix a typo (#64)
* Fix changelog version
* Fix changing the right of a list shared with a user
* Fix date handling on task detail page
* Fix drone testing pipeline triggering only when pushing to master and not on prs
* Fix email field type (#58)
* Fix error container at registration page always being displayed
* Fix gravatar url
* Fix height of task add button
* Fix initial dates on task edit sidebar
* Fix label input field breaking in a new line on task detail page
* Fix loading tasks for the first page after navigating to a new list
* Fix not using router links for previous and back buttons
* Fix priority label styling
* Fix reminders not being shown on task detail view on mobile
* Fix task text breaking on list home on mobile
* Fix task title on mobile (#54)
* Fix update notification layout on mobile (#44)
* Fix using the error data prop in components (#53)
* Don't schedule a reminder if the reminder date is in the past
* Don't try to cancel notifications if the browser does not support it
* Only focus inputs if the viewport is large enough (#55)
* Set user menu inactive when logging out
* Show if a related task is done (#49)

### Changed

* Always schedule notification
* Hide the llama from the top on the task detail page
* Improve link share layout
* Load Fonts directly
* Make sure to use date objects everywhere where dealing with dates
* Migration Improvements (#47)
* Move "Next Week" section in menu below "Next Month"
* Move the Vikunja logo to the hamburger menu on mobile
* Preload fonts css
* Rearrange button order on task detail view
* Reorganize Styles (#45)
* Show motd everywhere
* Sort tasks on start page by due date desc and id desc
* Update dependencies (#40)
* Use message mixin for handling success and error messages (#51)
* Use the same method everywhere to calculate the avatar url
* Better default profile image
* Better wording for shared settings
* Bump npm to 6.13
* Put the add reminders button on the task detail page higher up
* Directly link to the task for tasks on the start page
* Disable production source maps

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

