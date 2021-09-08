# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

All releases can be found on https://code.vikunja.io/frontend/releases.

The releases aim at the api versions which is why there are missing versions.

## [0.18.1] - 2021-09-08

### Added

* feat: make it possible to fake online state via dev env (#720)

### Fixed

* fix: call to /null from background image (#714)
* Fix data export download progress
* fix: kanban-card mutatation violation (#712)
* Fix missing translation when creating a new task on the kanban board
* Fix rearranging tasks in a kanban bucket when its limit was reached
* Fix sort order for table view
* Fix task attributes overridden when saving the task title with enter
* Fix translation badge

### Dependency Updates

* Update dependency @4tw/cypress-drag-drop to v2 (#711)
* Update dependency axios to v0.21.4 (#705)
* Update dependency jest to v27.1.1 (#716)
* Update dependency vite-plugin-vue2 to v1.8.2 (#707)
* Update dependency vite to v2.5.4 (#708)
* Update dependency vite to v2.5.5 (#709)
* Update typescript-eslint monorepo to v4.31.0 (#706)


## [0.18.0] - 2021-09-05

### Added

* Add a button to copy an attachment url from the attachment overview
* Add collapsing kanban buckets
* Add confirm with enter when setting a new password
* Add default list setting & creating tasks from home (#520)
* Add depends_on for push step
* Add depends_on for upload step
* Add drag delay on mobile
* Add express for serve:dev
* Add filters for quick action bar
* Add frontend tests for list history
* Add making tasks favorite from the task detail view
* Add missing position property to list and bucket models
* Add more debug logs for gantt charts
* Add more global state tests (#521)
* Add proofread languages to available languages
* Add quick action bar shortcut to shortcut overview
* Add setting for the first day of the week
* Add showing version info in GUI
* Add syncing translations to crowdin
* Add timeout to fix race condition when authenticating as a link share and renewing the token simultaneously
* Add translations (#562)
* Add typescript support for helper functions (#598)
* Add vite (#416)
* Allow failure of the weblate update step
* Always set the kanban board to full width for share links
* Another day, another js date edge-case
* Automatically update approved translations from crowdin
* Break long list titles in list overview
* Preload labels and use locally stored in vuex
* PWA improvments (#622)
* Quick Actions & global search (#528)
* Quick add magic for tasks (#570)
* Reorder tasks, lists and kanban buckets (#620)
* Show last visited list on home page
* Show recently visited lists in quick actions
* Show salutation based on the time of day
* Sort labels alphabetically on tasks
* Switch the :latest docker image tag to contain the latest release instead of the latest unstable

### Changed

* Change building latest docker image
* Change desktop downstream trigger plugin with our own debug build
* Change menu hamburger icon
* Change quick add magic characters to be more familiar with the todoist ones
* Change the docker builder image to a working one on arm
* chore: discard old font file formats (#673)
* chore: only import common languages (#671)
* Cleanup broken sw functions
* Cleanup drone pipeline
* Cleanup old vue cli config
* Configure tests retries
* Decrease page padding on task detail page
* Directly redirect to the openid auth provider if that's the only auth method
* Don't allow dragging a list when the user does not have the rights
* Don't load already loaded task attachments again when saving an edited task description
* Don't prefetch all i18n files
* Don't show archived lists/namespaces in quick actions
* feat: provide global variables in all components (#669)
* Hide favorite list edit menu
* Hide keyboard shortcuts indicator on mobile
* Improve chunk size
* Improve some translations (#581)
* Improve tests
* Indicate done tasks in quick actions
* Load list background in list card
* Make editor edit button at the bottom the default and make sure the done button stands out more
* Make saving a text edit a button
* Make sure highlight.js is always lazy-loaded
* Make sure the task popup view takes up all the space it can on mobile
* Make tests less flaky
* Make the logo smaller on link shared lists
* Make the progress bar color lighter
* Move creation of new items to the bottom of the multiselect list
* Move general settings to the top
* Move translated files after downloading them
* Move weblate ping to shell script
* Only add a drag delay if on mobile instead of setting it to 0
* Only build a bundle for modern browsers
* Refactor success and error messages
* Refactor success and error notifications to prevent html in them
* Remove logout button for link shares
* Run frontend-tests with dist in ci (#605)
* Save auth tokens from link shares only in memory, don't persist them to localStorage
* Search namespaces locally only when duplicating a list
* Show errors from openid provider
* Show labels alphabetically sorted in the overview
* Small cleanups & code improvements
* TOTP UX improvements & translation fixes

### Fixed

* Fix changing the repeat mode of a task when no value is entered yet
* Fix comment on different task after clicking on a task notification
* Fix CTA spacings
* Fix date parsing parsing words with weekdays in them (#607)
* fix(deps): update dependency marked to v3.0.1 (#677)
* fix(deps): update dependency marked to v3.0.2 (#682)
* Fix error property already defined as a function
* Fix flickering pre-loaded search results when focusing the search input
* Fix Gantt layout overflowsing on mobile
* Fix gantt months being wrong
* Fix git push remote to update crowdin translations
* Fix global mutation of has tasks state
* Fix header layout for long list titles
* Fix highlight.js in editor
* Fix home page tests
* Fix keyboard shortcuts not working on the task detail page
* Fix label changes appearing to be saved immediately when editing them
* Fix labels list in saved filter spacing
* Fix lint
* Fix list archived notification mobile layout
* Fix list settings not being available when list backgrounds are disabled
* Fix lists showing up multiple times in history
* Fix llama background url
* Fix loading a list when it was already partially saved in vuex
* Fix loading & disabled state on inputs when creating a new task
* Fix loading labels when editing a saved filter
* Fix menu styles
* Fix missing background for tasks on a shared list with a background
* Fix multiselect search padding
* Fix new lists created with quick actions not showing up in the menu
* fix: non unique ids (#672)
* Fix not reloading tasks of a saved filter after editing it
* Fix not updating list name in store when changing it
* Fix other values getting pushed away when creating a new one through multiselect
* Fix padding for kanban cards
* Fix parsing dates on the last day of the month
* Fix populating task details ater updating the description
* Fix quick actions not opening
* Fix quick actions not working when nonexisting lists where left over in history
* Fix redirecting to /login for some routes
* Fix removing a namespace from state after it was deleted
* Fix resetting date filters from upcoming after viewing a task detail page (popup)
* Fix sass division
* Fix saving showing archived setting
* Fix selecting a single value from multiselect
* Fix sending openid scopes when authenticating
* Fix sending the user back to the list view they came from when opening a task in detail view
* Fix setting a task as favorite button
* Fix setting delete button for newly created task comments
* Fix setting filters for reminders
* Fix setting secret for updating translations
* Fix setting task favorite status in test fixtures
* Fix showing an editor save button in cases where it wasn't required
* Fix showing edit buttons when the user does not have the rights to use them
* Fix showing import tasks cta when tasks are loading
* Fix some translation strings
* Fix sorting labels
* Fix spacing for task detail view in lists with a background
* Fix table headers wrapping in table view
* Fix table text alignment in task detail page
* Fix table view scrolling on mobile
* Fix test for saving a task description
* Fix tests failing on thursdays
* Fix token in storage not getting renewed
* Fix translating dates
* Fix usage of / in sass
* Fix user name and avatar alignment in navbar
* Fix users not removed from the list in settings when unshared
* Fix user test fixtures
* fix: vuex mutation violation from draggable (#674)

### Dependency Updates

* chore(deps): update dependency @4tw/cypress-drag-drop to v1.8.1 (#693)
* chore(deps): update dependency autoprefixer to v10.3.3 (#684)
* chore(deps): update dependency autoprefixer to v10.3.4 (#697)
* chore(deps): update dependency axios to v0.21.2 (#698)
* chore(deps): update dependency axios to v0.21.3 (#700)
* chore(deps): update dependency cypress to v8.3.1 (#689)
* chore(deps): update dependency esbuild to v0.12.23 (#683)
* chore(deps): update dependency esbuild to v0.12.24 (#688)
* chore(deps): update dependency esbuild to v0.12.25 (#696)
* chore(deps): update dependency eslint-plugin-vue to v7.17.0 (#686)
* chore(deps): update dependency jest to v27.1.0 (#687)
* chore(deps): update dependency sass to v1.38.1 (#679)
* chore(deps): update dependency sass to v1.38.2 (#690)
* chore(deps): update dependency sass to v1.39.0 (#695)
* chore(deps): update dependency typescript to v4.4.2 (#685)
* chore(deps): update dependency vite-plugin-pwa to v0.11.2 (#681)
* chore(deps): update dependency vite to v2.5.1 (#680)
* chore(deps): update dependency vite to v2.5.2 (#692)
* chore(deps): update dependency vite to v2.5.3 (#694)
* chore(deps): update typescript-eslint monorepo to v4.29.3 (#676)
* chore(deps): update typescript-eslint monorepo to v4.30.0 (#691)
* Update dependency autoprefixer to v10.3.2 (#670)
* Update dependency browserslist to v4.16.7 (#634)
* Update dependency browserslist to v4.16.8 (#664)
* Update dependency browserslist to v4.17.0 (#701)
* Update dependency bulma to v0.9.3 (#554)
* Update dependency cypress-file-upload to v5.0.8 (#556)
* Update dependency cypress to v7.3.0 (#507)
* Update dependency cypress to v7.4.0 (#517)
* Update dependency cypress to v7.5.0 (#541)
* Update dependency cypress to v7.6.0 (#561)
* Update dependency cypress to v7.7.0 (#577)
* Update dependency cypress to v8.1.0 (#624)
* Update dependency cypress to v8.2.0 (#637)
* Update dependency cypress to v8.3.0 (#660)
* Update dependency cypress to v8 (#601)
* Update dependency date-fns to v2.22.0 (#523)
* Update dependency date-fns to v2.22.1 (#524)
* Update dependency date-fns to v2.23.0 (#604)
* Update dependency dompurify to v2.2.9 (#529)
* Update dependency dompurify to v2.3.0 (#573)
* Update dependency dompurify to v2.3.1 (#655)
* Update dependency esbuild to v0.12.15 (#610)
* Update dependency esbuild to v0.12.16 (#614)
* Update dependency esbuild to v0.12.17 (#623)
* Update dependency esbuild to v0.12.18 (#638)
* Update dependency esbuild to v0.12.19 (#643)
* Update dependency esbuild to v0.12.20 (#654)
* Update dependency esbuild to v0.12.21 (#666)
* Update dependency esbuild to v0.12.22 (#668)
* Update dependency eslint-plugin-vue to v7.10.0 (#525)
* Update dependency eslint-plugin-vue to v7.11.0 (#547)
* Update dependency eslint-plugin-vue to v7.11.1 (#548)
* Update dependency eslint-plugin-vue to v7.12.1 (#565)
* Update dependency eslint-plugin-vue to v7.13.0 (#574)
* Update dependency eslint-plugin-vue to v7.14.0 (#597)
* Update dependency eslint-plugin-vue to v7.15.0 (#625)
* Update dependency eslint-plugin-vue to v7.15.1 (#633)
* Update dependency eslint-plugin-vue to v7.16.0 (#648)
* Update dependency eslint to v7.27.0 (#514)
* Update dependency eslint to v7.28.0 (#539)
* Update dependency eslint to v7.29.0 (#555)
* Update dependency eslint to v7.30.0 (#571)
* Update dependency eslint to v7.31.0 (#596)
* Update dependency eslint to v7.32.0 (#627)
* Update dependency highlight.js to v11.0.1 (#538)
* Update dependency highlight.js to v11.1.0 (#582)
* Update dependency highlight.js to v11.2.0 (#630)
* Update dependency highlight.js to v11 (#527)
* Update dependency jest to v27.0.3 (#526)
* Update dependency jest to v27.0.4 (#535)
* Update dependency jest to v27.0.5 (#558)
* Update dependency jest to v27.0.6 (#569)
* Update dependency jest to v27 (#519)
* Update dependency marked to v2.0.4 (#510)
* Update dependency marked to v2.0.5 (#513)
* Update dependency marked to v2.0.6 (#522)
* Update dependency marked to v2.0.7 (#532)
* Update dependency marked to v2.1.0 (#552)
* Update dependency marked to v2.1.1 (#553)
* Update dependency marked to v2.1.2 (#559)
* Update dependency marked to v2.1.3 (#567)
* Update dependency marked to v3 (#657)
* Update dependency @rollup/plugin-commonjs to v19.0.2 (#617)
* Update dependency sass to v1.33.0 (#512)
* Update dependency sass to v1.34.0 (#515)
* Update dependency sass to v1.34.1 (#534)
* Update dependency sass to v1.35.0 (#550)
* Update dependency sass to v1.35.1 (#551)
* Update dependency sass to v1.35.2 (#579)
* Update dependency sass to v1.36.0 (#606)
* Update dependency sass to v1.37.0 (#628)
* Update dependency sass to v1.37.2 (#632)
* Update dependency sass to v1.37.5 (#635)
* Update dependency sass to v1.38.0 (#661)
* Update dependency ts-jest to v27.0.4 (#602)
* Update dependency ts-jest to v27.0.5 (#662)
* Update dependency @types/jest to v27.0.1 (#653)
* Update dependency @types/jest to v27 (#650)
* Update dependency vite-plugin-pwa to v0.10.0 (#644)
* Update dependency vite-plugin-pwa to v0.11.0 (#667)
* Update dependency vite-plugin-pwa to v0.8.2 (#612)
* Update dependency vite-plugin-pwa to v0.9.3 (#629)
* Update dependency vite-plugin-vue2 to v1.7.3 (#613)
* Update dependency vite-plugin-vue2 to v1.8.0 (#646)
* Update dependency vite-plugin-vue2 to v1.8.1 (#656)
* Update dependency vite to v2.4.3 (#611)
* Update dependency vite to v2.4.4 (#619)
* Update dependency vite to v2.5.0 (#658)
* Update dependency vue-advanced-cropper to v1.6.0 (#516)
* Update dependency vue-advanced-cropper to v1.7.0 (#543)
* Update dependency vue-advanced-cropper to v1.8.0 (#641)
* Update dependency vue-advanced-cropper to v1.8.1 (#642)
* Update dependency vue-advanced-cropper to v1.8.2 (#645)
* Update dependency vue-flatpickr-component to v8.1.7 (#572)
* Update dependency vue-i18n to v8.24.5 (#564)
* Update dependency vue-i18n to v8.25.0 (#595)
* Update dependency vue-router to v3.5.2 (#557)
* Update dependency wait-on to v6 (#568)
* Update dependency workbox-cli to v6.1.5 (#609)
* Update Font Awesome (#636)
* Update Node.js (#549)
* Update Node.js to v16.4.1 (#576)
* Update Node.js to v16.4.2 (#578)
* Update typescript-eslint monorepo to v4.28.4 (#600)
* Update typescript-eslint monorepo to v4.28.5 (#618)
* Update typescript-eslint monorepo to v4.29.0 (#631)
* Update typescript-eslint monorepo to v4.29.1 (#647)
* Update typescript-eslint monorepo to v4.29.2 (#659)
* Update vue monorepo to v2.6.13 (#530)
* Update vue monorepo to v2.6.14 (#540)
* Update workbox monorepo to v6.2.0 (#639)
* Update workbox monorepo to v6.2.2 (#640)
* Update workbox monorepo to v6.2.4 (#649)
* User account deletion (#651)
* User Data Export and import (#699)

## [0.17.0 - 2021-05-14]

### Added

* Add a "done" option to kanban buckets (#440)
* Add arm64 builds
* Add button to un-archive a namespace
* Add clearer call to action when no lists are available yet
* Add code highlighting for rendered user input text
* Add github sponsoring
* Add link share password authentication (#466)
* Add names to link shares when creating them (#456)
* Add notifications overview (#414)
* Add option to remove a list background
* Add overdue task reminder notification setting
* Add repeat after one-click intervals
* Add repeat mode setting for tasks
* Add security information to readme
* Add separate manifest template for latest
* Add settings for user search (#458)
* Add success message when modifying buckets
* Add "today" task filter
* Add view image modal for image attachments
* Pagingation for tasks in kanban buckets (#419)
* Persist show archived state
* Play a sound when marking a task as done

### Fixed

* Fix adding a label twice when selecting it and pressing enter
* Fix attachment hover
* Fix attachment not being added if the task was not a kanban task
* Fix attachments being added mutliple times
* Fix bucket test fixture when moving tasks between lists test
* Fix button height
* Fix caldav url not containing the api url if the frontend and api are on the same domain
* Fix checking for undefined behaviour when viewing a task
* Fix closing popups when clicking outside of them (#378)
* Fix "create new list" and import buttons on home page
* Fix create new list test
* Fix create new namespace test
* Fix current password id being available twice
* Fix datepicker popup not fully aligned on mobile
* Fix defer due date popup
* Fix delete buttons in forms
* Fix deleting task relations
* Fix editor buttons alignment
* Fix editor placeholder color
* Fix edit task description test
* Fix empty call to actions
* Fix filter container positioning
* Fix filter container positioning in link shares
* Fix flaky test
* Fix flaky test part 2
* Fix font caching in docker image
* Fix formatting invalid dates
* Fix getting back to the default task view when navigating back from a task modal
* Fix getting back to the kanban board after closing a task popup
* Fix iterating over check boxes and attachment images in the editor rendering
* Fix kanban board slightly scrolling
* Fix kanban height on mobile
* Fix kanban infinite scrolling on chrome
* Fix label spacing
* Fix labels randomly changing color after saving
* Fix list counter in the navigation counting archived lists
* Fix list layout when the list has no background for link shares
* Fix login or register not working when pressing enter
* Fix logout test
* Fix map_hash_max_size for docker images
* Fix misspelling (#415)
* Fix multiselect on mobile
* Fix namespace actions alignment in the menu
* Fix no color selected in the color picket
* Fix notification parsing for team memeber added
* Fix notification styling
* Fix pasting text into task comments or task descriptions
* Fix priority label width in task list
* Fix release pipeline steps
* Fix reloading the task list after changing a filter
* Fix removing dates from a filter
* Fix resetting colors from the color picker
* Fix setting a default color when none was saved
* Fix setting dates in safari
* Fix showing and hiding lists in the menu
* Fix sorting task by due date on task overview
* Fix spacing for lists with no rights to add new tasks
* Fix table names in test fixtures
* Fix task detail view spacings
* Fix task filter toggle button if the list has a background
* Fix task icon size
* Fix task icons on kanban if there were multiple different ones
* Fix task id spacing
* Fix task pagination
* Fix task relation search test
* Fix tasks moving infinitely in gantt chart (#493)
* Fix tasks not disappearing from the kanban board when moving them between lists
* Fix task title heading ux
* Fix team edit test
* Fix team edit test (#382)
* Fix team name in team member added notification
* Fix test
* Fix tests after changing button classes
* Fix text color
* Fix transition between pages
* Fix undo when marking a task as done
* Fix waiting for dependency step when building
* Fix yarn.lock
* Only check for token renew when the user is authenticated
* Only show the llama background for unauthenticated users
* Only use dark shadows for buttons
* Prevent setting a bucket limit < 0

### Changed

* Automatically go back after saving from a popup
* Better wording of new namespace and list buttons
* Bring up the keyboard shortcuts when pressing ?
* Change bucket background color
* Change main branch to main
* Cleanup font caching and requesting
* Don't hide all lists of namespaces when loosing network connectivity
* Don't save the editor text when it is loaded
* Don't show the list color in the list view
* Don't show the "new bucket" button when buckets are still loading
* Focus task detail elements when they show up
* Hide new related tasks form when related tasks exist
* Hide task elements while the task is loading
* Hide the bucket limit input when clicked away
* Hide the login form if no api url is configured
* Improve consistency of the layout (#386)
* Inline mutliselect search input for multiple elements
* Make filter buttons look better on mobile
* Make full task in task list clickable
* Make hidden lists in the menu more compact
* Make message undo button secondary
* Make release steps on master depend on building/testing
* Make sure all arm64 build steps run in parallel
* Make sure all empty pages have a call to action
* Make sure all popups & dropdowns are animated
* Make sure attachements are only added once to the list after uploading + Make sure the attachment list shows up every
  time after adding an attachment
* Make sure no cta's are visible while the page is loading
* Make sure the loading spinner is always visible at the end of the page
* Make the button shadow lighter
* Make the icons in the menu light grey
* Make the input full width by default
* Make the scrollbars a lighter grey (#394)
* Make the "upload attachment" button less obvious
* Move all content to cards (#387)
* Move all create views to better looking popups (#383)
* Move buttons to separate component (#380)
* Move list edit/namespace to separate pages and in a menu (#397)
* Move the search input to filters
* Open links to external sites in a new window
* Rearrange task actions
* Reduce quick task edit fields
* Remove the shadow at the "+" button for related tasks
* Rename .noshadow to .has-no-shadow
* Rework attachments list to look great everywhere
* Set user info from api instead of only relying on the info encoded in the jwt token
* Show call to action for task description if there is none
* Show label colors when searching for labels
* Show list if the search result for a task belongs to a different list
* Show "powered by Vikunja" in link shares
* Subscriptions and notifications for namespaces, tasks and lists (#410)
* Switch node-sass to sass
* Switch telegram notifications to matrix
* Update ShowTasks view to sort tasks by ascending (#406)
* Use a lighter grey for comment created dates
* Use buttons more consistently
* Use mousedown instead of click event to close modals
* Work around auto tag for main branch

### Dependency Updates

* Pin dependency browserslist to 4.16.6 (#500)
* Pin dependency highlight.js to 10.5.0 (#371)
* Update browserlist and caniuse-lite db
* Update dependency bulma to v0.9.2 (#392)
* Update dependency cypress-file-upload to v5.0.3 (#437)
* Update dependency cypress-file-upload to v5.0.4 (#455)
* Update dependency cypress-file-upload to v5.0.5 (#461)
* Update dependency cypress-file-upload to v5.0.6 (#481)
* Update dependency cypress-file-upload to v5.0.7 (#498)
* Update dependency cypress-file-upload to v5 (#379)
* Update dependency cypress to v6.3.0 (#381)
* Update dependency cypress to v6.4.0 (#399)
* Update dependency cypress to v6.5.0 (#412)
* Update dependency cypress to v6.6.0 (#421)
* Update dependency cypress to v6.7.1 (#430)
* Update dependency cypress to v6.8.0 (#435)
* Update dependency cypress to v6.9.1 (#452)
* Update dependency cypress to v7.1.0 (#472)
* Update dependency cypress to v7.2.0 (#494)
* Update dependency cypress to v7 (#453)
* Update dependency date-fns to v2.17.0 (#403)
* Update dependency date-fns to v2.18.0 (#420)
* Update dependency date-fns to v2.19.0 (#423)
* Update dependency date-fns to v2.20.0 (#459)
* Update dependency date-fns to v2.20.1 (#463)
* Update dependency date-fns to v2.20.2 (#470)
* Update dependency date-fns to v2.20.3 (#473)
* Update dependency date-fns to v2.21.0 (#477)
* Update dependency date-fns to v2.21.1 (#482)
* Update dependency date-fns to v2.21.2 (#499)
* Update dependency date-fns to v2.21.3 (#505)
* Update dependency dompurify to v2.2.7 (#426)
* Update dependency dompurify to v2.2.8 (#496)
* Update dependency eslint-plugin-vue to v7.5.0 (#384)
* Update dependency eslint-plugin-vue to v7.6.0 (#411)
* Update dependency eslint-plugin-vue to v7.7.0 (#422)
* Update dependency eslint-plugin-vue to v7.8.0 (#438)
* Update dependency eslint-plugin-vue to v7.9.0 (#469)
* Update dependency eslint to v7.18.0 (#376)
* Update dependency eslint to v7.19.0 (#398)
* Update dependency eslint to v7.20.0 (#409)
* Update dependency eslint to v7.21.0 (#418)
* Update dependency eslint to v7.22.0 (#427)
* Update dependency eslint to v7.23.0 (#443)
* Update dependency eslint to v7.24.0 (#464)
* Update dependency eslint to v7.25.0 (#490)
* Update dependency eslint to v7.26.0 (#504)
* Update dependency faker to v5.2.0 (#389)
* Update dependency faker to v5.3.1 (#400)
* Update dependency faker to v5.4.0 (#408)
* Update dependency faker to v5.5.0 (#442)
* Update dependency faker to v5.5.1 (#444)
* Update dependency faker to v5.5.2 (#450)
* Update dependency faker to v5.5.3 (#462)
* Update dependency highlight.js to v10.6.0 (#407)
* Update dependency highlight.js to v10.7.1 (#436)
* Update dependency highlight.js to v10.7.2 (#451)
* Update dependency lodash to v4.17.21 (#413)
* Update dependency marked to v1.2.8 (#391)
* Update dependency marked to v1.2.9 (#401)
* Update dependency marked to v2.0.1 (#417)
* Update dependency marked to v2.0.2 (#465)
* Update dependency marked to v2.0.3 (#468)
* Update dependency marked to v2 (#405)
* Update dependency sass-loader to v10.1.1 (#372)
* Update dependency sass-loader to v10.2.0 (#506)
* Update dependency sass to v1.32.13 (#509)
* Update dependency vue-advanced-cropper to v1.3.0 (#404)
* Update dependency vue-advanced-cropper to v1.3.1 (#424)
* Update dependency vue-advanced-cropper to v1.3.2 (#425)
* Update dependency vue-advanced-cropper to v1.3.3 (#439)
* Update dependency vue-advanced-cropper to v1.3.4 (#441)
* Update dependency vue-advanced-cropper to v1 (#393)
* Update dependency vue-advanced-cropper to v1.4.0 (#454)
* Update dependency vue-advanced-cropper to v1.4.1 (#460)
* Update dependency vue-advanced-cropper to v1.5.0 (#471)
* Update dependency vue-advanced-cropper to v1.5.1 (#495)
* Update dependency vue-advanced-cropper to v1.5.2 (#497)
* Update dependency vue-drag-resize to v1.5.1 (#457)
* Update dependency vue-drag-resize to v1.5.2 (#501)
* Update dependency vue-drag-resize to v1.5.4 (#502)
* Update dependency vue-easymde to v1.4.0 (#449)
* Update dependency vue-router to v3.5.0 (#388)
* Update dependency wait-on to v5.3.0 (#434)
* Update Font Awesome (#374)
* Update Font Awesome (#432)
* Update vue monorepo (#390)
* Update vue monorepo to v4.5.11 (#385)
* Update vue monorepo to v4.5.12 (#433)
* Update vue monorepo to v4.5.13 (#503)

## [0.16.0 - 2021-01-10]

### Added

* Add autocomplete attributes to login and register forms
* Add color indicators to task list (#321)
* Add default color palette to picker
* Add disabled state for task titles
* Add downloading assets when building docker images
* Add filters to gantt chart
* Add login via email
* Add maskable icon
* Add Microsoft Todo migration (#339)
* Add more spacing for checkboxes in the editor
* Add more spacing to the "Archived" badge in namespace overview
* Add "new label" button to label management (#359)
* Add openid scope when redirecting to external openid provider
* Add proper focus styles
* Add setting for sending reminder emails (#343)
* Add showing and modifying user name (#306)
* Add task filter for assignees (#349)
* Add task filter for kanban
* Add task filter for labels (#350)
* Add task filter for lists and namespaces (#351)
* Add task filter for reminders (#347)
* Add trello migration (#336)
* Add wait in cypress test for user settings
* Add yarn cache to drone (#312)
* Authentication with OpenID Connect providers (#305)
* Better reminders (#308)
* Better save messages for tasks (#307)
* Build custom v-tooltip (#290)
* Build modern build for modern browsers
* Frontend Testing With Cypress (#313)

### Fixed

* Don't hide the "new bucket" when updating tasks
* Don't reset task relation kind after adding a task relation
* Don't show filter and search buttons for saved filter lists
* Don't show the "next week/month" buttons on the start page
* Fix avatar icon of attachments created by
* Fix deleting a saved filter
* Fixed squishy color bubble (#358)
* Fix list not added to lists when duplicating
* Fix list not being removed from the menu list when deleting it
* Fix loading states for unrelated components (#370)
* Fix logging out after reloading the page
* Fix logging the user out when renewing the token while the api is not reachable
* Fix non-release docker builds (#357)
* Fix parsing task done at date
* Fix password reset
* Fix related tasks width when the task is opened in a modal
* Fix reminder inputs and the close buttons not properly aligned
* Fix removing a kanban bucket
* Fix removing a namespace not removing it from the list
* Fix renewing token on focus
* Fix repeat after layout
* Fix resetting list rights after updating the list
* Fix showing the keyboard shortcuts from the menu
* Fix task background color for link shares
* Fix tooltip still existing in viewport after hiding them
* Get rid of the null reminder to fix jumping inputs when updating reminders
* Hide menu on mobile after navigating
* Hide share links table header when no share links are available yet
* Make sure task title and task id are properly shown on mobile (#334)
* Make sure the editor does not break if the text has checkboxes
* Make the menu have a fixed width
* Mobile Menu Fixes (#332)
* Only show a loading spinner per task when updating a task on the kanban board
* Only show attachments table header when there are attachments
* Only show loading spinner over menu when loading namespaces
* Only show the list with teams if there are any teams
* Performance improvements (#288)
* Properly cache html files
* Refactor app component (#283)

### Changed

* Bump ini from 1.3.5 to 1.3.8
* Change avatar endpoint
* Change cache key for dependencies
* Change license to AGPLv3
* Change test waits (I wish I wouldn't need them)
* Create list through store to make sure it is updated everywhere
* Improve comment avatars on mobile
* Improve editor buttons UX (#361)
* Log the user out if the token could not be renewed
* Make adding fields to tasks more intuitive (#365)
* Make keyboard shortcuts single keys
* Move focus directive to seperate file
* Move next week/next month task overview pages into a single "Upcoming" page and allow toggle
* Move "Teams" menu further down the list
* Pin dependencies (#324)
* Pin dependency jest to 26.6.3 (#311)
* Remove "collapse menu button" and make the hamburger button always visible
* Remove core-js from direct dependencies
* Remove leftover '.only' modifier
* Remove the drone cache image since there is no arm compatible image available
* Remove the focus of the bucket title element after saving the title
* Replace vue-multiselect with a custom component (#366)
* Show all available shortcuts everywhere but indicate which work on the current page
* Show a loading spinner when creating a new kanban task
* Show an icon if a task has non-empty description (Kanban view and List view) (#360)
* Show created/updated by for tasks
* Show done at in task detail view
* Show loading spinner when loading namespaces & lists
* Show task progress on task (#354)
* Update browserlist db
* Update dependency axios to v0.21.0 (#278)
* Update dependency axios to v0.21.1 (#353)
* Update dependency camel-case to v4.1.2 (#315)
* Update dependency cypress to v6.1.0 (#325)
* Update dependency cypress to v6.2.0 (#352)
* Update dependency cypress to v6.2.1 (#367)
* Update dependency dompurify to v2.2.0 (#274)
* Update dependency dompurify to v2.2.1 (#287)
* Update dependency dompurify to v2.2.2 (#289)
* Update dependency dompurify to v2.2.3 (#320)
* Update dependency dompurify to v2.2.4 (#330)
* Update dependency dompurify to v2.2.5 (#340)
* Update dependency dompurify to v2.2.6 (#342)
* Update dependency eslint-plugin-vue to v7.2.0 (#319)
* Update dependency eslint-plugin-vue to v7.3.0 (#333)
* Update dependency eslint-plugin-vue to v7.4.0 (#356)
* Update dependency eslint-plugin-vue to v7.4.1 (#368)
* Update dependency eslint to v7.12.0 (#279)
* Update dependency eslint to v7.12.1 (#281)
* Update dependency eslint to v7.13.0 (#293)
* Update dependency eslint to v7.14.0 (#303)
* Update dependency eslint to v7.15.0 (#318)
* Update dependency eslint to v7.16.0 (#344)
* Update dependency eslint to v7.17.0 (#364)
* Update dependency @fortawesome/vue-fontawesome to v2.0.2 (#337)
* Update dependency marked to v1.2.2 (#275)
* Update dependency marked to v1.2.3 (#291)
* Update dependency marked to v1.2.4 (#299)
* Update dependency marked to v1.2.5 (#302)
* Update dependency marked to v1.2.6 (#326)
* Update dependency marked to v1.2.7 (#331)
* Update dependency node-sass to v5 (#282)
* Update dependency register-service-worker to v1.7.2 (#323)
* Update dependency sass-loader to v10.0.4 (#276)
* Update dependency sass-loader to v10.0.5 (#286)
* Update dependency sass-loader to v10.1.0 (#295)
* Update dependency snake-case to v3.0.4 (#316)
* Update dependency vue-advanced-cropper to v0.17.4 (#273)
* Update dependency vue-advanced-cropper to v0.17.6 (#277)
* Update dependency vue-advanced-cropper to v0.17.7 (#284)
* Update dependency vue-advanced-cropper to v0.17.8 (#294)
* Update dependency vue-advanced-cropper to v0.17.9 (#300)
* Update dependency vue-advanced-cropper to v0.18.1 (#322)
* Update dependency vue-advanced-cropper to v0.19.1 (#327)
* Update dependency vue-advanced-cropper to v0.19.2 (#328)
* Update dependency vue-advanced-cropper to v0.19.3 (#338)
* Update dependency vue-advanced-cropper to v0.20.0 (#346)
* Update dependency vue-advanced-cropper to v0.20.1 (#348)
* Update dependency vue-easymde to v1.3.1 (#298)
* Update dependency vue-easymde to v1.3.2 (#335)
* Update dependency vue-router to v3.4.8 (#280)
* Update dependency vue-router to v3.4.9 (#292)
* Update dependency vuex to v3.6.0 (#309)
* Update dependency wait-on to v5.2.1 (#355)
* Update vue monorepo to v4.5.10 (#369)
* Update vue monorepo to v4.5.9 (#301)
* Use yarn caches when building docker images

## [0.15.0 - 2020-10-19]

### Added

* Add app shortcuts when using vikunja as pwa
* Add build hash as meta tag to index.html to ensure always loading the new index file
* Add checkbox to show only tasks which have a due date
* Add creating labels when creating a task (#192)
* Add debug logs for loading list + kanban buckets
* Add deferring task's due dates directly from the overview (#199)
* Add easymde & markdown preview for editing descriptions and comments (#183)
* Add github sponsor link
* Add limits for kanban boards (#234)
* Add loading spinner when duplicating a list
* Add more debugging when loading lists or buckets
* Add more prefetching of components
* Add notice to a list if it has no tasks
* Add options to show tasks in range on the overview pages
* Add Page Titles Everywhere (#177)
* Allow setting api url from the login screen (#264)
* Favorite lists (#237)
* Favorite tasks (#236)
* Keyboard Shortcuts (#193)
* Saved filters (#239)
* Show caldav url in settings if it's enabled server side
* Show legal links from api if configured

### Fixed

* Fix archived lists still showing up in the side menu
* Fix Assignees being deleted when adding a due date (#254)
* Fix bottom padding on kanban
* Fix bottom white margin
* Fix checking for existing migration from other services
* Fix comparing the currently loaded list with the current list to make sure to only load the list if needed
* Fix create new bucket button having no margin to the right
* Fix due date changes not saved on mobile
* Fix editor spacing
* Fix long text overflowing in task comments
* Fix pagination button hover color
* Fix pwa icon for iOS
* Fix related tasks list spacing
* Fix sort order when marking a task as done from the overview
* Fix task in list style for tasks with assignees
* Fix task layout in kanban
* Fix task list if it has tasks with a long unbreakable title
* Fix task title input taking up almost no space if empty
* Fix update available breaking the navbar position
* Make sure to always load the home route when starting the app
* Make sure to make the list id from the route an int to not fail the comparison
* More avatar providers (#200)
* Only show the list at the end of the task if it was not specially required to show the list
* Only trigger desktop rebuilds on pushes to master
* Pin dependencies (#184)
* Pin dependency vue-advanced-cropper to 0.16.10 (#201)
* Pin dependency vue-shortkey to 3.1.7 (#194)
* Pin telegram notify in drone
* Prevent loading the list + kanban board again when closing the task popup
* Prevent rendering html in tooltips
* Release preparations
* Remove html from tooltip
* Replace renovate tokens with env

### Changed

* Always focus inputs on kanban when adding a new task or bucket
* Automatically scroll to the bottom of a bucket after adding a new task to it
* Bump http-proxy from 1.18.0 to 1.18.1
* Cleanup code & make sure it has a common code style
* Disabele spellcheck on bucket titles
* Don't cache everything in the service worker, only explicitly assets
* Don't create a label through quick add if the title is empty
* Don't show a confusing message if no options are available
* Hide the user menu if clicked outside of it
* Hide UI elements if the user does not have the right to use them (#211)
* Include fonts css in the main css bundle
* Make task list, teams and settings pages max width of $desktop and centered
* Make the task view full width for shares if the list has a background
* Mark tasks as done from the kanban board with ctrl+click
* Open unsplash author links in a new window
* Put the editor container higher up for task description
* Redirect to current list view on click on list in menu again
* Switch release bucket to scaleway s3
* Trigger a rebuild of the desktop app on builds to master for the frontend
* Trigger @change when pasting content into editor
* Update dependency axios to v0.20.0 (#216)
* Update dependency bulma to v0.9.1 (#252)
* Update dependency date-fns to v2.15.0 (#190)
* Update dependency date-fns to v2.16.0 (#220)
* Update dependency date-fns to v2.16.1 (#223)
* Update dependency dompurify to v2.0.14 (#221)
* Update dependency dompurify to v2.0.15 (#229)
* Update dependency dompurify to v2.0.17 (#241)
* Update dependency dompurify to v2.1.0 (#245)
* Update dependency dompurify to v2.1.1 (#248)
* Update dependency eslint-plugin-vue to v7.0.1 (#257)
* Update dependency eslint-plugin-vue to v7.1.0 (#271)
* Update dependency eslint-plugin-vue to v7 (#255)
* Update dependency eslint to v7.10.0 (#250)
* Update dependency eslint to v7.11.0 (#263)
* Update dependency eslint to v7.4.0 (#175)
* Update dependency eslint to v7.5.0 (#191)
* Update dependency eslint to v7.6.0 (#198)
* Update dependency eslint to v7.7.0 (#213)
* Update dependency eslint to v7.8.0 (#225)
* Update dependency eslint to v7.8.1 (#228)
* Update dependency eslint to v7.9.0 (#242)
* Update dependency @fortawesome/vue-fontawesome to v2 (#226)
* Update dependency http-proxy from 1.18.0 to 1.18.1
* Update dependency lodash to v4.17.16 (#178)
* Update dependency lodash to v4.17.17 (#179)
* Update dependency lodash to v4.17.18 (#180)
* Update dependency lodash to v4.17.19 (#181)
* Update dependency lodash to v4.17.20 (#212)
* Update dependency marked to v1.1.1 (#185)
* Update dependency marked to v1.2.0 (#251)
* Update dependency sass-loader to v10.0.1 (#219)
* Update dependency sass-loader to v10.0.2 (#230)
* Update dependency sass-loader to v10.0.3 (#262)
* Update dependency sass-loader to v10 (#217)
* Update dependency sass-loader to v9.0.1 (#174)
* Update dependency sass-loader to v9.0.2 (#176)
* Update dependency sass-loader to v9.0.3 (#203)
* Update dependency sass-loader to v9 (#173)
* Update dependency vue-advanced-cropper to v0.17.0 (#231)
* Update dependency vue-advanced-cropper to v0.17.1 (#232)
* Update dependency vue-advanced-cropper to v0.17.2 (#238)
* Update dependency vue-advanced-cropper to v0.17.3 (#243)
* Update dependency vue-drag-resize to v1.4.1 (#182)
* Update dependency vue-drag-resize to v1.4.2 (#197)
* Update dependency vue-easymde to v1.2.2 (#187)
* Update dependency vue-easymde to v1.3.0 (#256)
* Update dependency vue-flatpickr-component to v8.1.6 (#222)
* Update dependency vue-router to v3.4.0 (#202)
* Update dependency vue-router to v3.4.1 (#204)
* Update dependency vue-router to v3.4.2 (#205)
* Update dependency vue-router to v3.4.3 (#210)
* Update dependency vue-router to v3.4.4 (#247)
* Update dependency vue-router to v3.4.5 (#249)
* Update dependency vue-router to v3.4.6 (#260)
* Update dependency vue-router to v3.4.7 (#269)
* Update Font Awesome (#188)
* Update Font Awesome (#253)
* Update Font Awesome (#258)
* Update renovate token
* Update vue monorepo to v2.6.12 (#215)
* Update vue monorepo to v4.5.2 (#208)
* Update vue monorepo to v4.5.3 (#209)
* Update vue monorepo to v4.5.4 (#214)
* Update vue monorepo to v4.5.6 (#244)
* Update vue monorepo to v4.5.7 (#259)
* Update vue monorepo to v4.5.8 (#272)
* Use team update route to update a team member's admin status

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
* # PrideMonth
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

