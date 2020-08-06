# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

All releases can be found on https://code.vikunja.io/frontend/releases.

The releases aim at the api versions which is why there are missing versions.

## [0.14.1 - 2020-08-06]

### Fixed

* Prevent html being rendered in tooltips

## [0.14.0 - 2020-07-01]

### Added

* Add border to colorpicker (fixes #146)
* Add changing list identifier
* Add changing the uid and gid in docker through env variables
* Add color picker to change task color to task detail view
* Add docker build pipelines for arm and amd64 (#164)
* Add docker multiarch manifest build step
* Add list duplicate (#172)
* Add mention of unsplash in the background settings
* Add option to hide the menu on desktop
* Add option to remove color from label, task, namespace or list (#157)
* Add repeating tasks from current date setting
* Add suffix for auto built docker images per arch
* Add todoist migrator to the frontend
* Add yarn timeout to build
* Custom backgrounds for lists (#144)
* Enable resetting search input
* List Background upload (#151)
* Namespaces & Lists Page (#160)
* Task Filters (#149)

### Fixed

* Always break kanban card titles
* Check if we have a service worker available before trying to communicate with it
* Don't disable the task add button if input is empty
* Don't try to fetch the initial unsplash results when unsplash backgrounds are disabled
* Don't try to make a request to get the totp status if its disabled
* Ensure consistent naming of title fields (#134)
* Fix changing task dates
* Fix Datetime Handling (#168)
* Fix docker arm build plugin
* Fix docker arm build tag
* Fix edit task repeat after being undefined (again)
* Fix error messages when trying to update tasks in kanban if kanban hasn't been opened yet
* Fix error when adding a background to a list which did not have one before
* Fix gantt chart not updating when navigating between lists
* Fix getting migration status
* Fix hamburger icon on mobile padding
* Fix kanban board height
* Fix kanban tasks with backgrounds
* Fix list title on mobile
* Fix login form on mobile
* Fix notifications not using task title
* Fix not sending the user to the view they came from when viewing task details
* Fix not showing changes in kanban when switching between views
* Fix redirect when not logged in
* Fix register
* Fix related tasks overflowing if a related task has a long name
* Fix related tasks search
* Fix repeat after value being undefined error in task edit panel
* Fix saving list view if not present in browser
* Fix search on mobile
* Fix task title not editable in edit task pane
* Fix trying to load kanban buckets if the kanban board is not in focus
* Fix typo when no upcoming tasks are available
* Fix user dropdown on mobile
* Only load tasks when the user is authenticated
* Remember list view when navigating between lists
* Remove old tasks when loading list view

### Changed

* Change logo primary color
* Color the whole card on kanban if the task has a color
* Don't show a success message if it is obvious the action has been successful
* Don't show the task id in list view
* Hide hints on start page if a user has tasks (#159)
* Hide totp settings if it is disabled server side
* Increase network timeout when building docker image
* Make sure the version includes the tag when building docker images
* #PrideMonth
* Only renew user token on tab focus events
* Redirect the user to login page if the token expired when the tab gets focus again
* Remove title length restrictions
* Rename routes to follow the same pattern
* Restructure components
* Save list view per list and not globally
* Show list background when viewing a link share
* Show namespace name in list search field
* Show task index instead of id on kanban
* Simplify pipeline
* Update dependency bulma to v0.9.0 (#150)
* Update dependency date-fns to v2.14.0 (#136)
* Update dependency eslint to v7.1.0 (#139)
* Update dependency eslint to v7.2.0 (#148)
* Update dependency eslint to v7.3.0 (#162)
* Update dependency eslint to v7.3.1 (#166)
* Update dependency @fortawesome/vue-fontawesome to v0.1.10 (#158)
* Update dependency vue-easymde to v1.2.1 (#145)
* Update dependency vue-router to v3.2.0 (#137)
* Update dependency vue-router to v3.3.1 (#141)
* Update dependency vue-router to v3.3.2 (#142)
* Update dependency vue-router to v3.3.4 (#156)
* Update dependency vuex to v3.5.0 (#170)
* Update dependency vuex to v3.5.1 (#171)
* Update Font Awesome (#161)
* Update vue monorepo (#153)
* Update vue monorepo to v4.4.1 (#140)
* Update vue monorepo to v4.4.4 (#154)
* Update vue monorepo to v4.4.5 (#165)
* Update vue monorepo to v4.4.6 (#167)
* Use the right Id when loading unsplash thumbnails

## [0.13] - 2020-05-12

#### Added 

* Add docker run script to change api url on startup
* Add github token for renovate (#89)
* Add input length validation for team names
* Add list title in overview page
* Add logging frontend version to console on startup
* Add moving tasks between lists
* Add scrolling for task table view
* Add telegram release notificiation (#98)
* Add user settings (#108)
* Better responsive layout for unauthenticated pages
* Change default api url to 3456 (Vikunja default)
* Configure Renovate (#80)
* Docker multistage build (#113)
* Don't open task detail in popup for list and table view
* Don't show the llama background when on mobile
* Highlight the current list when something list related is called
* Kanban (#118)
* Make api url configurable in index.html
* Make "Move task to different list" wording shorter
* Make sure the api url does not have a / at the end
* Show parent list and namespace for tasks in detail views
* Show the list of a related task if it belongs to a different list
* TOTP (#109)
* Open popup detail view when opening from task overview
* Vuex (#126)

#### Fixed

* Fetch tags when building in ci to display proper versions
* Fix attachment icon
* Fix avatar url
* Fix bucket spacing on kanban board
* Fix changing api url when releasing
* Fix closing of notifications by clicking on it not working
* Fix creating a new task on a list when in list view
* Fix date table cell getting wrong data
* Fix %done in table view
* Fix drone config
* Fix id params not being named correctly
* Fix listId not changing when switching between lists
* Fix listId not defined in list view switcher
* Fix loading state for kanban board
* Fix maintaining the current page for the list view when navigating back from another page
* Fix navigating back to list view after deleting a task
* Fix not all labels being shown
* Fix not redirecting to login page after logging out
* Fix not re-loading tasks when switching between overviews
* Fix opening link share list view
* Fix pagination for tasks
* Fix parsing nested array with non-objects when updating
* Fix parsing nested models
* Fix redirecting for unauthenticated pages to login
* Fix redirecting to list view from task detail
* Fix related tasks input size
* Fix related tasks list being too large
* Fix setting api url when building docker image
* Fix sharing rights not displayed correctly
* Fix task modal with when attachments are present
* Fix task relation kind dropdown
* Fix task sort parameters
* Fix task title overflowing in detail view
* Fix team managment (#121)
* Fix trying to load the current tasks even when not logged in (Fixes #133)
* Fix undefined getter for related tasks
* Fix uploading attachments
* Fix user search bar not hiding in edit team view
* Fix using filters for overview views
* Fix version console log when compiling for Docker
* Let labels take all available space on tasks

#### Changed

* Less explicit matching of api routes for service worker
* Make all api fields snake_case (#105)
* Make the task font size smaller for task cards
* Move conversion of snake_case to camelCase to model to make recursive models still work
* Only set fullpage state to false if the page is actually fullpage
* Only show undone tasks on task overview page
* Pin dependencies (#106)
* Pin dependencies (#81)
* Pin dependency vue-smooth-dnd to 0.8.1 (#120)
* Pin dependency vuex to 3.3.0 (#128)
* Pluralize related task kinds if there is more than one
* Remove debug log
* Remove debug logging
* Remove dependency in docker build step when releasing
* Remove dependency in docker build step when releasing latest
* Remove llama-upside-down.svg
* Remove task in kanban state when removing in task detail view
* Switch docker image to node for building
* Update dependency axios to v0.19.2 (#83)
* Update dependency babel-eslint to v10.1.0 (#84)
* Update dependency bulma to v0.8.1 (#85)
* Update dependency bulma to v0.8.2 (#104)
* Update dependency copy-to-clipboard to v3.3.1 (#100)
* Update dependency core-js to v3.6.4 (#101)
* Update dependency core-js to v3.6.5 (#102)
* Update dependency date-fns to v2.11.1 (#88)
* Update dependency date-fns to v2.12.0 (#103)
* Update dependency date-fns to v2.13.0 (#127)
* Update dependency eslint-plugin-vue to v6.2.2 (#91)
* Update dependency eslint to v6.8.0 (#90)
* Update dependency eslint to v7 (#129)
* Update dependency node-sass to v4.13.1 (#92)
* Update dependency node-sass to v4.14.0 (#119)
* Update dependency node-sass to v4.14.1 (#125)
* Update dependency register-service-worker to v1.7.1 (#93)
* Update dependency sass-loader to v8.0.2 (#94)
* Update dependency v-tooltip to v2.0.3 (#95)
* Update dependency vue-easymde to v1.2.0 (#116)
* Update dependency vue-router to v3.1.6 (#96)
* Update dependency vuex to v3.4.0 (#132)
* Update Font Awesome (#82)
* Update Node.js to v13.14.0 (#123)
* Update tasks in kanban board after editing them in task detail view (#130)
* Update vue-cli monorepo to v4.3.0 (#97)
* Update vue-cli monorepo to v4.3.1 (#99)
* Upgrade vue-cli

## [0.12] - 2020-04-04

#### Added

* Table View for tasks (#76)
* 404 page
* Add creating new related tasks
* Add getting the user avatar from the api (#68)
* Add support for archiving lists and namespaces (#73)
* Add task search term to query param to enable navigation
* Add undo button to notification when marking a task as done
* Add user to attachments list
* Colors for lists and namespaces (#74)
* Enable marking tasks as done from the task overview
* Ensure labels of a task get updated when updating them
* Input length validation for new tasks, lists and namespaces (#70)
* Pre/Suffix formatted dates with relative pronouns like "in [one day]" or "[two days] ago"

#### Fixed

* Fix avatar sizes
* Fix changing task dates (due/start/end/reminders)
* Fix comments not being loaded again when switching between tasks
* Fix error notification still being shown on password reset pages despite no error
* Fix gantt chart (#79)
* Fix icon overflowing in navigation
* Fix namespace model name showing wrong placeholder until the namespace was loaded
* Fix new related task not being visible in the search field
* Fix not highlighting the current list in menu when paginating
* Fix updating a task with repeat after interval from list view (Fixes #75)
* Use deep imports for importing lodash to make tree shaking easier
* Revert "Use deep imports for importing lodash to make tree shaking easier"
* Work around browsers preventing Vue bindings from working with autofill (Fixes #78)

#### Changed

* Schedule token renew every minute
* Swap moment.js with date-fns
* Change release bucket

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

