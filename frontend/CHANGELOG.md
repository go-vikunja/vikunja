# Changelog

THIS CHANGELOG ONLY EXISTS FOR HISTORICAL REASONS.
Starting with version 0.23.0, all changes are logged in the CHANGELOG.md in the root of this repository since the repos were merged.

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

All releases can be found on https://code.vikunja.io/frontend/releases.

The releases aim at the api versions which is why there are missing versions.

## [0.22.1] - 2024-01-28

### Bug Fixes

* *(auth)* Correctly construct redirect url from current window href
* *(ci)* Use working crowdin image
* *(ci)* Use working image for crowdin update step
* *(ci)* Use working crowdin image
* *(color picker)* When picking a color, the color picker should not be black afterwards
* *(editor)* List icons
* *(editor)* Use higher-contrast colors for links and code
* *(editor)* Don't bubble up changes when no changes were made
* *(editor)* Focus the editor when clicking on the whole edit container
* *(editor)* Render images without crashing
* *(editor)* Use a stable image id to prevent constant re-rendering
* *(editor)* Use manual input prompt instead of window.prompt
* *(filter)* Validate filter title field after loading a filter for edit
* *(kanban)* Ensure text and icon color only depends on the card background, not on the color scheme
* *(kanban)* Make sure the checklist summary uses the correct text color
* *(kanban)* Make sure spacing between assignees and other task details works out evenly
* *(labels)* Make color reset work
* *(labels)* Text and background combination in dark mode
* *(notifications)* Unread indicator spacing
* *(notifications)* Always left-align notification text
* *(notifications)* Read indicator size
* *(openid)* Use the full path when building the redirect url, not only the host
* *(openid)* Use the calculated redirect url when authenticating with openid providers
* *(project)* Always use the appropriate color for task estimate during deletion dialog
* *(quick add magic)* Ensure month is removed from task text
* *(table view)* Make sure popup does not overlap
* *(task)* Don't immediately re-trigger date change when nothing changed
* *(task)* Bubble date changes from the picker up
* *(task)* Update due date when marking a task done
* *(task)* Don't show edit button when the user does not have permission to edit the task
* *(task)* Don't show assignee edit buttons and input when the user does not have the permission to edit
* *(tasks)* Make sure tasks show up if their parent task is not available in the current view
* *(tasks)* Don't load tasks multiple times when viewing list or gantt view
* *(test)* Make date assertion not brittle
* Lint ([5e991f3](5e991f3024f7856420614171ec66468eb2e2df63))


### Dependencies

* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v2 (#3862)
* *(deps)* Update pnpm to v8.14.0
* *(deps)* Update dependency vue to v3.4.7 (#3873)
* *(deps)* Update dependency axios to v1.6.5 (#3871)
* *(deps)* Update dependency date-fns to v3 (#3857)
* *(deps)* Update dev-dependencies (#3861)
* *(deps)* Update dependency @kyvg/vue3-notification to v3.1.3 (#3864)
* *(deps)* Update dependency node to v20.11.0
* *(deps)* Update dependency vue-i18n to v9.9.0 (#3880)
* *(deps)* Update dependency dompurify to v3.0.8 (#3881)
* *(deps)* Update dependency floating-vue to v2.0.0 (#3883)
* *(deps)* Update tiptap to v2.1.15 (#3884)
* *(deps)* Update vueuse to v10.7.1 (#3872)
* *(deps)* Update pnpm to v8.14.1 (#3885)
* *(deps)* Update sentry-javascript monorepo to v7.93.0 (#3859)
* *(deps)* Update dependency floating-vue to v5 (#3887)
* *(deps)* Update dependency vue to v3.4.8 (#3886)
* *(deps)* Update node.js to v20.11 (#3888)
* *(deps)* Increase renovate timeout
* *(deps)* Update tiptap to v2.1.16 (#3892)
* *(deps)* Pin node.js (#3895)
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency sortablejs to v1.15.2
* *(deps)* Update vueuse to v10.7.2
* *(deps)* Update dependency floating-vue to v5.1.0
* *(deps)* Update dependency vue to v3.4.14
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies (major) (#3890)
* *(deps)* Update dependency floating-vue to v5.1.1
* *(deps)* Update dependency floating-vue to v5.2.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency vue to v3.4.15
* *(deps)* Update dependency happy-dom to v13.2.0
* *(deps)* Update sentry-javascript monorepo to v7.94.1
* *(deps)* Update dependency vite to v5.0.12
* *(deps)* Update dependency date-fns to v3.3.0
* *(deps)* Update dev-dependencies
* *(deps)* Update pnpm to v8.14.2
* *(deps)* Update dependency date-fns to v3.3.1
* *(deps)* Update dev-dependencies to v6.19.1
* *(deps)* Update pnpm to v8.14.3
* *(deps)* Update sentry-javascript monorepo to v7.95.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency axios to v1.6.6
* *(deps)* Update dev-dependencies
* *(deps)* Update sentry-javascript monorepo to v7.97.0
* *(deps)* Update sentry-javascript monorepo to v7.98.0
* *(deps)* Update dependency axios to v1.6.7
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies

### Features

* *(reminders)* Show reminders in notifications bar
* Datepicker locale support (#3878) ([92f7d9d](92f7d9ded5d56b95ba7d647eba01372f6ef682ad))


### Miscellaneous Tasks

* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(perf)* Import some modules dynamically (#3179)
* Only show webhooks overview table when there are webhooks ([326bfb5](326bfb557ab359fa154b163f5dd957928f46d3ec))
* Only show webhooks overview table when there are webhooks ([631b02d](631b02d2eedc4a403b7c55f1c56ceaeca5379bf5))

## [0.22.0] - 2023-12-19

### Bug Fixes

* *(api tokens)* Expiry of tokens in a number of days
* *(api tokens)* Lint
* *(api tokens)* Make deletion of old tokens work
* *(api tokens)* Show a token after it was created
* *(attachments)* Layout and coloring in dark mode
* *(auth)* Correctly redirect the user to the last visited page after login
* *(auth)* Silently discard invalid auth tokens and log the user out
* *(background)* Unsplash author credit in dark mode
* *(build)* Don't download Puppeteer when building for prod
* *(ci)* Pin used node version to 20.5 to avoid build issues
* *(ci)* Use correct secret key to push
* *(docker)* Set correct default value for custom logo url
* *(editor)* Actions styling
* *(editor)* Actually populate loaded data into the editor
* *(editor)* Add icons for clearing marks and nodes
* *(editor)* Add missing dependencies for commands
* *(editor)* Add missing dependency
* *(editor)* Add workaround for checklist tiptap bug
* *(editor)* Alignment and focus states
* *(editor)* Allow checking a checkbox even when the editor is set to read only
* *(editor)* Always set mode to preview after save
* *(editor)* Always show placeholder when empty
* *(editor)* Change description when switching between tasks
* *(editor)* Check for almost empty editor value
* *(editor)* Check for empty content
* *(editor)* Checklist button icon
* *(editor)* Commands list in dark mode
* *(editor)* Correctly resolve images in descriptions
* *(editor)* Don't check parent checkbox when child label was clicked
* *(editor)* Don't crash when the component isn't completely mounted
* *(editor)* Don't create empty "blob" files when pasting images
* *(editor)* Don't prevent typing editor focus shortcut when other instance of an editor is focused already
* *(editor)* Don't use global shortcut when anything is focused
* *(editor)* Duplicate name
* *(editor)* Duplicate name for extension
* *(editor)* Focus state
* *(editor)* Image button icon
* *(editor)* Image paste handling
* *(editor)* Keep editor open when emptying content from the outside
* *(editor)* Keep editor open when emptying content from the outside (#3852)
* *(editor)* Lint
* *(editor)* Lint
* *(editor)* List styling
* *(editor)* Make checklist indicator work again
* *(editor)* Make initial editor mode (preview/edit) work
* *(editor)* Make tests work with changed structure
* *(editor)* Permission check for table editing
* *(editor)* Placeholder showing or not showing
* *(editor)* Reset on empty
* *(editor)* Show editor if there is no content initially
* *(editor)* Use edit enable
* *(editor)* Use modelValue directly to update values in the editor
* *(filter)* Don't immediately re-trigger prepareFilter
* *(filter)* Don't prevent entering date math strings
* *(filter)* Don't show other filters in project selection in saved filter
* *(filter)* Make other filters are not available for project selection
* *(filters)* Don't allow marking a filter as favorite
* *(filters)* Incorrect translation string
* *(filters)* Infinite loop when creating filters with dates (#3061)
* *(gantt)* Open task with double click from the gantt chart
* *(gantt)* Update the gantt view when switching between projects
* *(i18n)* Add upload files config
* *(i18n)* Fall back to browser language if the configured user language is invalid
* *(i18n)* Hungarian translation
* *(kanban)* Check if doneBucketId is set
* *(kanban)* Make sure kanban cards always have text color matching their background
* *(kanban)* Opening a task from the kanban board and then reloading the page should not crash everything when then navigating back
* *(list view)* Align nested subtasks with the parent text
* *(menu)* Separate favorite and saved filter projects from other projects
* *(navigation)* Don't hide color bubble in navigation on touch devices
* *(navigation)* Show filter settings dropdown
* *(project)* Correctly show project color next to project title in list view
* *(projects)* Don't suggest to create a new task in an empty filter
* *(quick actions)* Always open quick actions with hotkey, even if other inputs are focused
* *(quick actions)* Always search for projects
* *(quick actions)* Don't show projects when searching for labels or tasks
* *(quick actions)* Invalid class prop
* *(quick actions)* Project filter
* *(quick actions)* Project search
* *(quick actions)* Search for tasks within a project when specifying a project with quick add magic
* *(quick add magic)* Annually and variants spelling
* *(quick add magic)* Headline
* *(quick add magic)* Ignore common task indentation when adding multiple tasks at once
* *(quick add magic)* Repeating intervals in words
* *(settings)* Allow removing the default project via settings
* *(settings)* Move overdue remindeer time below
* *(sw)* Remove debug option via env as it would not be replaced correctly in prod builds
* *(task)* Correct spacing to task and project title
* *(task)* Correctly build task identifier
* *(task)* Don't reload the kanban board when opening a task
* *(task)* Don't reload the kanban board when opening a task
* *(task)* Duplicate attribute
* *(task)* Make sure the modal close button is not overlapped with the title field (#3256)
* *(task)* Priority label sizing and positioning in different environments
* *(task)* Priority label spacing
* *(task)* Remove wrong repeat types
* *(task)* Show related tasks form with shortcut even when there are already other related tasks
* *(task)* Use editor as preview first, then check for edit
* *(task)* Use empty description helper everywhere
* *(tasks)* Don't use the filter for upcoming when one is set for the home page
* *(tasks)* Favorited sub tasks are not shown in favorites pseudo list
* *(tasks)* Ignore empty lines when adding multiple tasks at once
* *(tasks)* Make sure tasks are fully clickable
* *(tasks)* Play pop sound directly and not from store
* *(tasks)* Prevent endless references
* *(tasks)* Reset page number when applying filters
* *(tasks)* Update api route
* *(tasks)* Update sub task relations in list view after they were created
* *(tasks)* Use mousedown event instead of click to close the task popup
* *(test)* Use correct file input
* *(user)* Allow openid users to request their deletion
* *(webhooks)* Styling* Correctly resolve kanban board in the background when moving a task ([8902c15](8902c15f7e9590da075e860f3d35939169ee246a))
* Don't render route modal when no properties are defined ([b1fe3fe](b1fe3fe29b3f7c8e3f1fa279b74f674bc63db232))
* Don't try to load buckets for project id 0 ([15ecafd](15ecafdf04391139da27f38dac9ed915d6220a9a))
* Lint ([218d724](218d72494a088b612e720ca2e9b566c0d3446579))
* Lint ([337c3e5](337c3e5e3e06a9e4928bebffda2e2f223fef398b))
* Lint ([7f2d921](7f2d92138e302188d6000632b4bc9bf053194dee))
* Lint ([99e2161](99e2161c09b1e2b08f3a907bd2e3ad2c71da87d3))
* Lint ([c01957a](c01957aae24696812c80b18c77137b5030fc757a))
* Tests ([f6d1db3](f6d1db35957c4c2fda7a58539a0a39db1b683ccb))


### Dependencies

* *(deps)* Remove unused dependencies
* *(deps)* Update dependencies
* *(deps)* Update dependencies
* *(deps)* Update dependency @fortawesome/vue-fontawesome to v3.0.5 (#3815)
* *(deps)* Update dependency @github/hotkey to v2.1.0 (#3766)
* *(deps)* Update dependency @github/hotkey to v2.1.1 (#3770)
* *(deps)* Update dependency @github/hotkey to v2.2.0 (#3809)
* *(deps)* Update dependency @github/hotkey to v2.3.0 (#3810)
* *(deps)* Update dependency @github/hotkey to v2.3.1 (#3845)
* *(deps)* Update dependency @github/hotkey to v3
* *(deps)* Update dependency @infectoone/vue-ganttastic to v2.2.0
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v0.12.2
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v1
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v1.5.0 (#3812)
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v1.6.0
* *(deps)* Update dependency @kyvg/vue3-notification to v3
* *(deps)* Update dependency @kyvg/vue3-notification to v3.1.2
* *(deps)* Update dependency @types/is-touch-device to v1.0.1 (#3786)
* *(deps)* Update dependency @types/is-touch-device to v1.0.2 (#3816)
* *(deps)* Update dependency @types/lodash.clonedeep to v4.5.8 (#3787)
* *(deps)* Update dependency @types/lodash.clonedeep to v4.5.9 (#3817)
* *(deps)* Update dependency @types/node to v18.17.0
* *(deps)* Update dependency @types/node to v20 (#3796)
* *(deps)* Update dependency @types/sortablejs to v1.15.4 (#3788)
* *(deps)* Update dependency @types/sortablejs to v1.15.5 (#3818)
* *(deps)* Update dependency @vueuse/core to v10.3.0
* *(deps)* Update dependency @vueuse/core to v10.4.0 (#3723)
* *(deps)* Update dependency axios to v1.5.1
* *(deps)* Update dependency axios to v1.6.0 (#3801)
* *(deps)* Update dependency axios to v1.6.2 (#3820)
* *(deps)* Update dependency caniuse-lite to v1.0.30001514
* *(deps)* Update dependency codemirror to v5.65.14
* *(deps)* Update dependency dayjs to v1.11.10 (#3753)
* *(deps)* Update dependency dompurify to v3.0.5
* *(deps)* Update dependency dompurify to v3.0.6 (#3754)
* *(deps)* Update dependency eslint to v8.52.0 (#3785)
* *(deps)* Update dependency highlight.js to v11.9.0 (#3763)
* *(deps)* Update dependency lowlight to v2.9.0 (#3789)
* *(deps)* Update dependency marked to v5.1.1
* *(deps)* Update dependency marked to v9
* *(deps)* Update dependency marked to v9.1.0 (#3760)
* *(deps)* Update dependency marked to v9.1.1 (#3768)
* *(deps)* Update dependency marked to v9.1.2 (#3774)
* *(deps)* Update dependency node (#3797)
* *(deps)* Update dependency node (#3834)
* *(deps)* Update dependency node to v18.18.0
* *(deps)* Update dependency node to v18.18.1
* *(deps)* Update dependency node to v18.18.2
* *(deps)* Update dependency pinia to v2.1.6
* *(deps)* Update dependency pinia to v2.1.7 (#3771)
* *(deps)* Update dependency sass to v1.69.2 (#3767)
* *(deps)* Update dependency sortablejs to v1.15.1 (#3841)
* *(deps)* Update dependency ufo to v1.2.0
* *(deps)* Update dependency ufo to v1.3.1
* *(deps)* Update dependency ufo to v1.3.2 (#3824)
* *(deps)* Update dependency vite to v4.4.2
* *(deps)* Update dependency vite to v4.4.3
* *(deps)* Update dependency vue to v3.3.10 (#3843)
* *(deps)* Update dependency vue to v3.3.13
* *(deps)* Update dependency vue to v3.3.5 (#3782)
* *(deps)* Update dependency vue to v3.3.6 (#3784)
* *(deps)* Update dependency vue to v3.3.7 (#3799)
* *(deps)* Update dependency vue to v3.3.8 (#3814)
* *(deps)* Update dependency vue to v3.3.9 (#3837)
* *(deps)* Update dependency vue-i18n to v9.5.0
* *(deps)* Update dependency vue-i18n to v9.6.0 (#3800)
* *(deps)* Update dependency vue-i18n to v9.6.1 (#3803)
* *(deps)* Update dependency vue-i18n to v9.6.5 (#3807)
* *(deps)* Update dependency vue-i18n to v9.7.0 (#3825)
* *(deps)* Update dependency vue-i18n to v9.8.0 (#3833)
* *(deps)* Update dependency vue-router to v4.2.5 (#3755)
* *(deps)* Update dessant/repo-lockdown action to v4
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies (#3721)
* *(deps)* Update dev-dependencies (#3726)
* *(deps)* Update dev-dependencies (#3740)
* *(deps)* Update dev-dependencies (#3746)
* *(deps)* Update dev-dependencies (#3747)
* *(deps)* Update dev-dependencies (#3757)
* *(deps)* Update dev-dependencies (#3761)
* *(deps)* Update dev-dependencies (#3769)
* *(deps)* Update dev-dependencies (#3776)
* *(deps)* Update dev-dependencies (#3780)
* *(deps)* Update dev-dependencies (#3793)
* *(deps)* Update dev-dependencies (#3802)
* *(deps)* Update dev-dependencies (#3806)
* *(deps)* Update dev-dependencies (#3811)
* *(deps)* Update dev-dependencies (#3813)
* *(deps)* Update dev-dependencies (#3821)
* *(deps)* Update dev-dependencies (#3826)
* *(deps)* Update dev-dependencies (#3828)
* *(deps)* Update dev-dependencies (#3829)
* *(deps)* Update dev-dependencies (#3835)
* *(deps)* Update dev-dependencies (#3842)
* *(deps)* Update dev-dependencies (#3846)
* *(deps)* Update dev-dependencies (#3856)
* *(deps)* Update dev-dependencies (major) (#3741)
* *(deps)* Update dev-dependencies (major) (#3827)
* *(deps)* Update dev-dependencies to v6
* *(deps)* Update flake
* *(deps)* Update font awesome to v6.4.2
* *(deps)* Update font awesome to v6.5.1 (#3839)
* *(deps)* Update lockfile
* *(deps)* Update lockfile
* *(deps)* Update lockfile
* *(deps)* Update lockfile
* *(deps)* Update lockfile
* *(deps)* Update lockfile
* *(deps)* Update node.js to v18.17.0
* *(deps)* Update node.js to v18.17.1
* *(deps)* Update node.js to v20.7 (#3736)
* *(deps)* Update node.js to v20.8 (#3756)
* *(deps)* Update pnpm to v8.10.2
* *(deps)* Update pnpm to v8.10.5
* *(deps)* Update pnpm to v8.11.0
* *(deps)* Update pnpm to v8.12.1
* *(deps)* Update pnpm to v8.6.12
* *(deps)* Update pnpm to v8.6.7
* *(deps)* Update pnpm to v8.6.8
* *(deps)* Update pnpm to v8.6.9
* *(deps)* Update pnpm to v8.7.0
* *(deps)* Update pnpm to v8.8.0
* *(deps)* Update pnpm to v8.9.0
* *(deps)* Update pnpm to v8.9.2
* *(deps)* Update sentry-javascript monorepo to v7.58.0
* *(deps)* Update sentry-javascript monorepo to v7.58.1
* *(deps)* Update sentry-javascript monorepo to v7.59.1
* *(deps)* Update sentry-javascript monorepo to v7.59.2
* *(deps)* Update sentry-javascript monorepo to v7.59.3
* *(deps)* Update sentry-javascript monorepo to v7.60.0
* *(deps)* Update sentry-javascript monorepo to v7.73.0
* *(deps)* Update sentry-javascript monorepo to v7.74.0 (#3772)
* *(deps)* Update sentry-javascript monorepo to v7.74.1 (#3778)
* *(deps)* Update sentry-javascript monorepo to v7.75.1 (#3798)
* *(deps)* Update sentry-javascript monorepo to v7.77.0 (#3805)
* *(deps)* Update sentry-javascript monorepo to v7.80.1 (#3819)
* *(deps)* Update sentry-javascript monorepo to v7.85.0 (#3831)
* *(deps)* Update sentry-javascript monorepo to v7.88.0
* *(deps)* Update sub-dependencies
* *(deps)* Update tiptap to v2.1.12 (#3790)
* *(deps)* Update tiptap to v2.1.13 (#3840)
* *(deps)* Update vueuse to v10.5.0 (#3762)
* *(deps)* Update vueuse to v10.6.1 (#3822)
* *(deps)* Update vueuse to v10.7.0 (#3844)

### Features

* *(api tokens)* Add basic api token overview
* *(api tokens)* Add deleting api tokens
* *(api tokens)* Add token creation form
* *(api tokens)* Allow custom selection of expiry dates
* *(api tokens)* Allow selecting all permissions
* *(api tokens)* Format permissions and groups human-readable
* *(api tokens)* Show warning if token has expired
* *(api tokens)* Validate title field when creating a new token
* *(assignees)* Improve avatar list consistency
* *(editor)* Add all slash commands
* *(editor)* Add bubble menu
* *(editor)* Add code highlighting
* *(editor)* Add command list example
* *(editor)* Add comment when pressing ctrl enter
* *(editor)* Add placeholder
* *(editor)* Add proper description for all buttons
* *(editor)* Add tests to check rendering of task description
* *(editor)* Add tooltips for everything
* *(editor)* Add uploading an image on save
* *(editor)* Allow passing placeholder down
* *(editor)* Edit mode
* *(editor)* Edit shortcut to set focus into the editor
* *(editor)* Enable table
* *(editor)* Image upload
* *(editor)* Improve overall styling
* *(editor)* Make image upload work via slash command
* *(editor)* Make task list work
* *(editor)* Mark a checkbox item as done when clicking on its text
* *(editor)* Move all editor related components into one folder
* *(editor)* Only load attachment images when rendering is done
* *(editor)* Open links when clicking on them
* *(editor)* Properly bubble changes when they are made
* *(editor)* Resolve and load attachment images from content
* *(editor)* Save when pressing ctrl enter
* *(gantt)* Implement dynamic sizing on small date ranges (#3750)
* *(i18n)* Add Slovene language for selection in the ui
* *(i18n)* Add arabic to list of selectable languages
* *(i18n)* Add hungarian translation for selection
* *(i18n)* Run translation update directly
* *(i18n)* Update crowdin sync to use v2 api
* *(i18n)* Update translations only once a day
* *(kanban)* Add icon for bucket collapse
* *(kanban)* Add setting for default bucket
* *(kanban)* Save done bucket with project instead of bucket
* *(labels)* Assign random color when creating labels
* *(list view)* Show subtasks nested
* *(migration)* Proper wording for async migration
* *(notifications)* Add option to mark all as read
* *(quick actions)* Show done tasks last
* *(quick actions)* Show labels as labels and tasks with all of their details
* *(quick actions)* Show task identifier
* *(quick actions)* Show tasks for a label when selecting it
* *(quick add magic)* Allow using the project identifier via quick add magic
* *(task)* Add more tests
* *(task)* Group related task action buttons
* *(task)* Immediately set focus on the task search input when opening the related tasks menu
* *(task)* Move task priority to the front when showing tasks inline
* *(task)* Save currently opened task with control/meta + s
* *(tasks)* Make the whole task in list view clickable
* *(tasks)* Update due date text every minute
* *(webhooks)* Add form validation* Allow custom logo via environment variable (#3685) ([cade3df](cade3df3e9a7eca8e0aa9d1553dd5597f0f5a8a2))
* *(webhooks)* Add webhook management form
* Add demo mode warning message ([ed8fb71](ed8fb71ff0b05860f320e2a1fe6c3cb29ed2889a))
* Add setting for default bucket ([04ba101](04ba1011cc3042f657ddb40ee727caf455db8b64))
* Api tokens ([28f2551](28f2551d87b99c59055a4909195e435dbd9794b6))
* Improve error message for invalid API url ([725fd1a](725fd1ad467fb988810cb23f12d372af236bd21d))
* Move from easymde to tiptap editor (#2222) ([26fc9b4](26fc9b4e4f8b96616385f4ca0a77a0ff7ee5eee5))
* Quick actions improvements ([47d5890](47d589002ccef5047a25ea3ad8ebe582c3b0bbc6))
* Webhooks (#3783) ([5d991e5](5d991e539bb3a249447847c13c92ee35d356b902))

### Miscellaneous Tasks

* *(ci)* Sign drone config
* *(editor)* Add break icon
* *(editor)* Add horizontal line icon
* *(editor)* Cleanup
* *(editor)* Cleanup unused options
* *(editor)* Format
* *(editor)* Make sure all tiptap dependencies are updated as one
* *(editor)* Move checklist to the other lists
* *(editor)* Remove converting markdown
* *(editor)* Remove marked usages
* *(editor)* Remove old editor component
* *(editor)* Remove unused components
* *(editor)* Use typed props definition
* *(filter)* Remove debug log
* *(quick actions)* Format* Provide better error messages when refreshing user info fails ([d535879](d5358793de7fc53795329382222e5f3bafc7fba1))
* Add pr lockdown ([07b1e9a](07b1e9a6b76eb7d92640e00a1dec4294efd2947b))
* Cleanup ([a4a2b95](a4a2b95dc7eaad5fe313884eec0d22d7ae5f85c1))
* Debug ([3cb1e7d](3cb1e7dede659acd19410e0611346e0f582f2ff3))
* Format ([c3f85fc](c3f85fcb1988603a58104552b35101b13e93b06e))
* Improve checking for API url '/' suffix  (#121) ([311b1d7](311b1d7594cfd03be4d998f4aead041a8ca63f8c))
* Include version json string in release zip ([c4adcf4](c4adcf4655550214ae795d941eb51878f34cedeb))
* Update flake ([64c90c7](64c90c7fe8a77ded21778a798f18862fe966bd1a))
* Update lockfile ([9f82ec4](9f82ec4162151ba32f329cb8e335eff6b21cebd4))

### Other

* *(other)* [skip ci] Updated translations via Crowdin

## [0.21.0] - 2023-07-07

### Bug Fixes

* *(Expandable)* Spelling
* *(building)* Let the compiler ignore props interface
* *(ci)* Always pull latest unstable api image for testing
* *(ci)* Directly build docker images and not use releases to avoid caching issues
* *(ci)* Disable puppeteer chrome download
* *(docker)* Copy patches prior to installing dependencies so that the installation actually works
* *(docker)* Don't set nginx worker rlimit
* *(filters)* Load projects after creating a filter
* *(filters)* Load projects after deleting a filter
* *(filters)* Load projects after updating a filter
* *(gantt)* Only update today value when changing to the gantt chart view
* *(i18n)* OrderedList translationid
* *(i18n)* Typo
* *(kanban)* Decrease task count per bucket when deleting a task
* *(kanban)* Don't export buckets as readonly because that makes it impossible to update them, even from within the store
* *(link share)* Default share view should be list, not project
* *(link share)* Redirect to list view after authenticating
* *(navigation)* Favorites project
* *(navigation)* Hide archived subprojects
* *(navigation)* Hide left ul border
* *(navigation)* Highlight saved filters in project view and prevent them from being dragged around
* *(navigation)* Hover state of other menu items
* *(navigation)* Make marking a project as favorite work
* *(navigation)* Make sure the Favorites project shows up when marking or unmarking a task as favorite
* *(navigation)* Make sure updating a project's state works for sub projects as well.
* *(navigation)* Make the styles work again
* *(navigation)* Menu item overflow
* *(navigation)* Nav item width for items without sub projects
* *(navigation)* Show text ellipsis for very long project titles
* *(navigation)* Sidebar top spacing
* *(navigation)* Watcher
* *(project)* Correctly load background when switching from or to a project view
* *(project)* Don't try to read title of undefined project
* *(project)* Duplicate a project without new parent
* *(project)* Make sure the correct tasks are loaded when switching between projects
* *(project)* Set maxRight on projects after opening a task
* *(projects)* Make sure the project hierarchy is properly updated when moving projects between parents
* *(projects)* Update project duplicate api definitions
* *(quick add magic)* Cleanup all assignee properties
* *(quick add magic)* Date parsing with a date at the beginning
* *(quick add magic)* Don't replace the prefix in every occurrence when it is present in the matched part
* *(quick add magic)* Use the project user service to find assignees for quick add magic
* *(reminders)* Align remove icon with the rest
* *(reminders)* Assignment to const when changing a reminder
* *(reminders)* Custom relative highlight now only when a custom relative reminder was actually selected
* *(reminders)* Don't assign the task
* *(reminders)* Don't assume 30 days are always a month
* *(reminders)* Don't sync negative relative reminder amounts in ui
* *(reminders)* Duplicate reminder for each change
* *(reminders)* Flatpickr styling improvements
* *(reminders)* Properly parse relative reminders which don't have an amount
* *(reminders)* Set date over relative reminder
* *(reminders)* Style flatpickr so that it blends in more
* *(repeat)* Prevent disappearing repeat mode settings when modes other than default repeat mode were selected
* *(sentry)* Don't fail the build when sentry upload fails
* *(sentry)* Use correct environment from vite env mode
* *(settings)* Don't try to sort timezones if there are none
* *(task detail view)* Make project display show the task's project
* *(task)* Break long task titles after 4 lines only
* *(task)* Call getting task identifier directly instead of using model function
* *(task)* Make an attachment cover image
* *(task)* Repeat mode now saves correctly
* *(tests)* Make sure the task is created with a bucket
* *(tests)* New project input field
* *(tests)* Project archived filter checkbox selector
* *(tests)* Wait for request instead of fixed time
* *(user)* Fix flickering of default settings
* *(user)* Lint* Fix comment
* *(user)* Set the language when saving
* Add await ([9d9fb95](9d9fb959d8f1c4a12110f1a988115116085b6aaf))
* Add default for level ([9402344](9402344b7ea70359c592412b6c341897e45c6069))
* Add interval to uses of useNow so that it uses less resources ([b77c7c2](b77c7c2f45495a0fe6d132b5f569e807074c6d12))
* Add more padding to the textarea ([dfa6cd7](dfa6cd777bc5d03cf88d62db9008aa0b366aa806))
* Add spacing between checkbox and title of related task ([62825d2](62825d2e6409e08ab3229bf693ed068198e18085))
* Allow icon changes configuration via env (#3567) ([57218d1](57218d14548bf1d4cd59f6976e84cf178023305d))
* Avoid crashing browser processes during tests ([7b05ed9](7b05ed9d3d24e07a6535f2462d215c47b6650be1))
* Bottom margin of project header ([1a94496](1a9449680114212eeb93be2aba3f10c416f67e78))
* Bubble changes from the editor immediately and move the delay to callers ([f4a7943](f4a79436809d13e1d2c5337f79358c15310d08d2))
* Checkbox label size based on icon ([fd699ad](fd699ad777c47764b35345b7ec18a854957ff5d1))
* Clarify user search setting ([ae025e3](ae025e30c659d43cce1e3f8361bd1c4c7cb860da))
* Cleanup unused translation strings ([aaa9d55](aaa9d553d080a83a9fd1bcdece366fb5832831f1))
* Collapsing child projects ([2250918](225091864f9088a07120cd3d36918f3060d57d30))
* Correctly sync filters on upcoming tasks page ([faa6298](faa62985dff877afc54c3510be8d27d493717780))
* Disable autocomplete in assignee search ([64f9f4f](64f9f4fd88a513cbc401aacbeab87695bb9f55bf))
* Don't allow creating a new label from filter view ([4c969f0](4c969f0a427e98b491c49646aaf19e19cf9ec924))
* Don't require variant prop on loading component as it already has a default one set ([01ac84c](01ac84ce1eda1de79fd752792115a71cb5c15698))
* Don't set the current project when setting a project ([31b7c1f](31b7c1f217532bf388ba95a03f469508bee46f6a))
* Don't show > for top-level projects ([03f4d0b](03f4d0b8bcba90b19302d6c6d2fbb92460b59957))
* Don't show child projects when the project is only a favorite ([0a17df8](0a17df87e950b8043578dbb7e9f12d5937802169))
* Don't try to convert a null date ([4ba02eb](4ba02ebbb6be4b96a42688b3ec8f29fe923aee0b))
* Don't try to map data from empty responses ([a118580](a11858070496614c492da321fe461b72c31afe5a))
* Don't try to map non-array data ([813d2b5](813d2b56a06cbd28a1bf0d01b64685a1b49188d0))
* Don't try to set a user language if none is saved ([68fd469](68fd4698ac443345dc7dbbf8cfebf76ec467b6ec))
* Don't try to set config from non-json responses ([7c1934a](7c1934aad0e5fcd0f785896d57efc34c6df935cd))
* Ensure all matched quick add magic parts are correctly removed from the task ([7b6a13d](7b6a13dd52dfa06e6093ae30adad1b86b66610e1))
* Ensure same protocol for configured api url (#3303) ([6c999ad](6c999ad14844b4f9ec74dc225895db6a12e4a781))
* Follow the happy path ([34182b8](34182b8bbb7a7e8eeb0ce698dc6da79785d05fc9))
* Force usage of @types for flexsearch instead of integrated types ([f60cebf](f60cebf42cb73d9fd2d9fea8de5bbeb96a724d47))
* Has-pseudo-class polyfill ([4703f9c](4703f9c4d5e3902d0fc389d447aa9a7da2e2dd4a))
* Ignore ts deprecations for now ([96e2c81](96e2c81b7ef2b7a0ff515dd01d9aeb28429cc0d5))
* Improve projectView storing and add migration ([842f204](842f204123afc3b9b4633b68de58cffb3af4f912))
* Improve the "pop" sound a bit ([3643ffe](3643ffe0d0357c89cb3517fafbb0c438188ac88d))
* Improve tooltip icon contrast ([a6cdf6c](a6cdf6c4bdceb1168f20e9d049c2e66f40c98aa1))
* Improve tooltip text ([2174608](21746088012f4fe0f750ed5e5cac916d506fb17b))
* Increase default auto-save timeout to 5 seconds ([f7ba3bd](f7ba3bd08fa9181180f99f4e5ebd5ec916fbcf19))
* Indentation ([e25273d](e25273df4899867ee146159d3d18125d387f8524))
* Lint ([292c904](292c90425ef96b99671702a0b28d87d660fa53dc))
* Lint ([4ff0c81](4ff0c81e373696b0505c2c080d558a20071562f3))
* Lint ([5d59392](5d593925666a09cbfda2f62577deb670033f93fb))
* Lint ([9ec29ca](9ec29cad300fe1c25cb355fb86e165ca920df511))
* Lint ([c294f9d](c294f9d28d3e793f8151265d5a16ed2fc53aea92))
* Lint ([c74612f](c74612f24adeb4aceafe9fc9b3264b1dfe84d128))
* Lint ([cd2b7fe](cd2b7fe185632253290838e405b8a2666b15ce24))
* Lint ([ed8de7e](ed8de7e3eb78f6723d5f675cca18c014b252ed64))
* List view: don't sort tasks after marking one "done" (#3285) ([6870db4](6870db4a72568f183134a6dd2d4af687dd7c839d))
* Load the correct language ([6593380](6593380013ff6043b846126ac67e6f96442a1c5b))
* Make check if projects are available work again ([5e65814](5e65814b8c5b37f3962856f2809f7cc85756da1e))
* Make computed side-effect free ([26bec05](26bec0517417afc52db93f0fdc4c48d47ed5c131))
* Make sure redirects to a saved view work as intended ([a64c0c1](a64c0c19e5a7da36ae4993fce443e4f37e3a4572))
* Make sure the unread notifications indicator is correctly positioned ([8b90b45](8b90b45739418f447b885fc9b37438e325f61b32))
* Make tests work again ([5685890](56858904938126dbfa8ade2e88e3ec6c4fff3a6f))
* Make type singular ([bc416f2](bc416f282f13d2b81782aba4b0d68b71f26c83e8))
* Make update available button use the correct text color all the time ([ae2b0f9](ae2b0f97c4bb50d4ff493af2af132f9740d16d49))
* Missing await ([391992e](391992effbb424a107ff060e7175884740a28c62))
* Missing variant prop for loading component ([2e9ade1](2e9ade11c3a3b6cb531d053f82a598a5ab851a93))
* Move parent project child id mutation to store ([26e3d42](26e3d42ed527afd6bf695ba3ad291e1c2b545bba))
* Move parent project handling out of useProject ([ba452ab](ba452ab88339b9ace987f1a18584a7950e00a776))
* Move the collapsible placeholder to the button ([1344026](1344026494fe47ac5604bff07b537a2765e840f6))
* Move types to dev dependencies ([739fe0c](739fe0caa13dc946e1801f290d8ab5f18cdc5faf))
* Only bind child projects data down ([3eca9f6](3eca9f6180e64f892e94d27eaa192cea780563a0))
* Only update daytime salutation when switching to home view ([c577626](c5776264c069000efbb62c64dfc2143d5fc4e0df))
* Passing readonly projects data to navigation ([d85be26](d85be26761240164b6bdcbe0601b46585b74fafa))
* Properly determine if there are projects ([a2cc9dd](a2cc9ddc8821a4b9b1ee1dd6109d1b3958a06ba6))
* Rebase re-add CustomTransition ([b93639e](b93639e14ecab06496086c3d2cc14f51d8f9f672))
* Recreate project instead of editing before ([175e31c](175e31ca629660d8d683b35b8e7c8052a62cd17d))
* Redundant ) ([6c2dc48](6c2dc483a20213f1f238e6224b9ecfb87faa2461))
* Remove getProjectById and replace all usages of it ([78158bc](78158bcba52d152a2ebf465242e25a55e6764470))
* Remove leftover suspense ([9d73ac6](9d73ac661fbf9315995c8a1f633708021591d2db))
* Remove leftovers of childIds ([bbaddb9](bbaddb9406910106b7d476a6550acff025e72655))
* Remove namespace routes ([10311b7](10311b79df36db44a8e96a446234c3c6d6aa6ec7))
* Remove namespace store reference ([ad2690b](ad2690b21cfc9ccc658737a726cc6b110089b635))
* Remove unnecessary fallback ([d414b65](d414b65e7d591f567067ce8085b9934207dc938a))
* Rename getParentProjects method to make it clear what it does ([39f699a](39f699a61ae91eb93c364137f76b595e7cad7561))
* Rename list to project for parsing subtasks via indentation ([fc8711d](fc8711d6d841d11847cd8567999373145ce3398d))
* Rename resolveRef ([f14e721](f14e721caf9434ac119f32c5e7f107bfbdd6746c))
* Return redirect ([7c964c2](7c964c29d487b5bcd2c125f81731e3b37374641a))
* Return updated project instead of the old one ([4ab5478](4ab547810c77e747e701ea865c13157d51aba461))
* Review findings ([5fb45af](5fb45afb12479eb135323299409bd91d8be24e39))
* Review findings ([85ffed4](85ffed4d9a26fc054fee51608bb83ccf2e3032f9))
* Review findings ([fb14eca](fb14eca6340ac4c761b8f61027662328bf55ade4))
* Route to create new project ([a5e710b](a5e710bfe594e06262b9ef46fa6b56ad637b8156))
* Set and use correct type for destructured props ([dbe1ad9](dbe1ad9353e165fd1e314cc72c7a4dece1c47d38))
* Set vue-ignore ([b6cd424](b6cd424aa30be3bd715c0b7555032fc80149ae7b))
* Show favorite on hover ([0be83db](0be83db40fa96478bfdb4a69e8a995d6debb6f52))
* Simplify sort ([85e882c](85e882cc5940067414004dabd01916f559fbd0ff))
* Sort in store ([46e8258](46e825820c465ebb9f8087e3afe6d74fad8d5159))
* SortBy type import ([d73b71a](d73b71a097755cdb075955a824c26bbaba222aaf))
* Spacing ([9162002](9162002e55d9ebfd0a6c8dbe28aab0c15f95b7e2))
* Style: "favorite" button being shown on projects without hovering ([ee4974a](ee4974a4948012b03adcedd956c0d907c57431c9))
* Switching to view type now ([060a573](060a573fe9006441131fd98c4618c5d294cf39b7))
* Tests ([69e94e5](69e94e58c451a5115c713696798f8fcbf8f787b3))
* Translation string ([f13db92](f13db9268a8862204522a8d68ad7b51edb9d91e1))
* Tsconfig as per https://github.com/vuejs/tsconfig#configuration-for-node-environments ([05b7063](05b70632c55ce34ee6471cc372334fc6e14c99a0))
* Tsconfig as per https://github.com/vuejs/tsconfig#configuration-for-node-environments ([ca9fe6f](ca9fe6ff215351c3f4c8de65a333f5cfd5876488))
* Undefined parent project when none was selected ([6cc11e6](6cc11e64ab392f8e8e69070000a748c04746e550))
* Undo further nesting of interactive items ([0acf447](0acf44778d0ab2a317bcfb3e89aa0292e2d5c2ed))
* Update logo change only every hour ([7126576](71265769cefb91be9a51fceff0a04095a9dd7e72))
* Use correct shortcut to open projects overview ([326b6ed](326b6eda6fce6554ea6e215c681466374657902a))
* Use menu tag everywhere ([0dd6f82](0dd6f82a0e198056724821c2bb56c6b9807ea451))
* Use onActivated ([a33fb72](a33fb72ef86112c6f29017bb951ff4e1ee611ed6))
* Use props destructuring everywhere ([3aa502e](3aa502e07d89314e885c252e1e3d4668fa64059b))
* Use strict comparison ([91e9eef](91e9eef5829d2a5ae27099fbd54029ed0ca46818))
* Use the color bubble as handle if the project has a color ([4857080](48570808e55e51751734ddaf4532ad651920d622))
* Use time constant ([a13c16c](a13c16ca03698a24860f8453cdb231c420d0077b))
* Wording ([985f998](985f998a821229d03c7d40d1a81f7fbe5121d585))


### Dependencies

* *(deps)* Install dependencies after rebase
* *(deps)* Pin dependency @tsconfig/node18 to 2.0.0
* *(deps)* Update all dev dependencies at once per day
* *(deps)* Update caniuse-and-related
* *(deps)* Update caniuse-and-related
* *(deps)* Update dependency @4tw/cypress-drag-drop to v2.2.4
* *(deps)* Update dependency @cypress/vite-dev-server to v5.0.5
* *(deps)* Update dependency @cypress/vue to v5.0.5
* *(deps)* Update dependency @faker-js/faker to v8
* *(deps)* Update dependency @faker-js/faker to v8.0.1
* *(deps)* Update dependency @faker-js/faker to v8.0.2
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v0.10.0
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v0.11.0
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v0.12.0
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v0.12.1
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v0.9.2
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v0.9.3
* *(deps)* Update dependency @kyvg/vue3-notification to v2.9.1
* *(deps)* Update dependency @rushstack/eslint-patch to v1.3.0
* *(deps)* Update dependency @rushstack/eslint-patch to v1.3.1
* *(deps)* Update dependency @rushstack/eslint-patch to v1.3.2
* *(deps)* Update dependency @tsconfig/node18 to v18
* *(deps)* Update dependency @tsconfig/node18 to v2.0.1
* *(deps)* Update dependency @types/codemirror to v5.60.8
* *(deps)* Update dependency @types/dompurify to v3
* *(deps)* Update dependency @types/dompurify to v3.0.1
* *(deps)* Update dependency @types/dompurify to v3.0.2
* *(deps)* Update dependency @types/marked to v4.3.0
* *(deps)* Update dependency @types/marked to v4.3.1
* *(deps)* Update dependency @types/marked to v5
* *(deps)* Update dependency @types/node to v18.15.1
* *(deps)* Update dependency @types/node to v18.15.10
* *(deps)* Update dependency @types/node to v18.15.11
* *(deps)* Update dependency @types/node to v18.15.12
* *(deps)* Update dependency @types/node to v18.15.13
* *(deps)* Update dependency @types/node to v18.15.2
* *(deps)* Update dependency @types/node to v18.15.3
* *(deps)* Update dependency @types/node to v18.15.5
* *(deps)* Update dependency @types/node to v18.15.6
* *(deps)* Update dependency @types/node to v18.15.7
* *(deps)* Update dependency @types/node to v18.15.8
* *(deps)* Update dependency @types/node to v18.15.9
* *(deps)* Update dependency @types/node to v18.16.0
* *(deps)* Update dependency @types/node to v18.16.1
* *(deps)* Update dependency @types/node to v18.16.10
* *(deps)* Update dependency @types/node to v18.16.11
* *(deps)* Update dependency @types/node to v18.16.14
* *(deps)* Update dependency @types/node to v18.16.16
* *(deps)* Update dependency @types/node to v18.16.17
* *(deps)* Update dependency @types/node to v18.16.18
* *(deps)* Update dependency @types/node to v18.16.19
* *(deps)* Update dependency @types/node to v18.16.2
* *(deps)* Update dependency @types/node to v18.16.3
* *(deps)* Update dependency @types/node to v18.16.4
* *(deps)* Update dependency @types/node to v18.16.5
* *(deps)* Update dependency @types/node to v18.16.6
* *(deps)* Update dependency @types/node to v18.16.7
* *(deps)* Update dependency @types/node to v18.16.8
* *(deps)* Update dependency @types/node to v18.16.9
* *(deps)* Update dependency @types/sortablejs to v1.15.1
* *(deps)* Update dependency @vitejs/plugin-legacy to v4.0.2
* *(deps)* Update dependency @vitejs/plugin-legacy to v4.0.3
* *(deps)* Update dependency @vitejs/plugin-legacy to v4.0.4
* *(deps)* Update dependency @vitejs/plugin-legacy to v4.0.5
* *(deps)* Update dependency @vitejs/plugin-vue to v4.1.0
* *(deps)* Update dependency @vitejs/plugin-vue to v4.2.0
* *(deps)* Update dependency @vitejs/plugin-vue to v4.2.1
* *(deps)* Update dependency @vitejs/plugin-vue to v4.2.2
* *(deps)* Update dependency @vitejs/plugin-vue to v4.2.3
* *(deps)* Update dependency @vue/eslint-config-typescript to v11.0.3
* *(deps)* Update dependency @vue/test-utils to v2.3.2
* *(deps)* Update dependency @vue/test-utils to v2.4.0
* *(deps)* Update dependency @vue/tsconfig to v0.3.2
* *(deps)* Update dependency @vue/tsconfig to v0.4.0
* *(deps)* Update dependency @vueuse/core to v10
* *(deps)* Update dependency @vueuse/core to v10.0.2
* *(deps)* Update dependency @vueuse/core to v10.1.0
* *(deps)* Update dependency @vueuse/core to v10.1.2
* *(deps)* Update dependency @vueuse/core to v10.2.0
* *(deps)* Update dependency @vueuse/core to v10.2.1
* *(deps)* Update dependency axios to v1.3.5
* *(deps)* Update dependency axios to v1.3.6
* *(deps)* Update dependency axios to v1.4.0
* *(deps)* Update dependency caniuse-lite to v1.0.30001465
* *(deps)* Update dependency caniuse-lite to v1.0.30001468
* *(deps)* Update dependency caniuse-lite to v1.0.30001470
* *(deps)* Update dependency caniuse-lite to v1.0.30001473
* *(deps)* Update dependency caniuse-lite to v1.0.30001477
* *(deps)* Update dependency caniuse-lite to v1.0.30001479
* *(deps)* Update dependency caniuse-lite to v1.0.30001481
* *(deps)* Update dependency caniuse-lite to v1.0.30001486
* *(deps)* Update dependency caniuse-lite to v1.0.30001487
* *(deps)* Update dependency caniuse-lite to v1.0.30001489
* *(deps)* Update dependency caniuse-lite to v1.0.30001500
* *(deps)* Update dependency caniuse-lite to v1.0.30001508
* *(deps)* Update dependency caniuse-lite to v1.0.30001511
* *(deps)* Update dependency codemirror to v5.65.13
* *(deps)* Update dependency css-has-pseudo to v6
* *(deps)* Update dependency csstype to v3.1.2
* *(deps)* Update dependency cypress to v12.10.0
* *(deps)* Update dependency cypress to v12.11.0
* *(deps)* Update dependency cypress to v12.12.0
* *(deps)* Update dependency cypress to v12.13.0
* *(deps)* Update dependency cypress to v12.14.0
* *(deps)* Update dependency cypress to v12.15.0
* *(deps)* Update dependency cypress to v12.16.0
* *(deps)* Update dependency cypress to v12.8.0
* *(deps)* Update dependency cypress to v12.8.1
* *(deps)* Update dependency cypress to v12.9.0
* *(deps)* Update dependency date-fns to v2.30.0
* *(deps)* Update dependency dayjs to v1.11.8
* *(deps)* Update dependency dayjs to v1.11.9
* *(deps)* Update dependency dompurify to v3.0.2
* *(deps)* Update dependency dompurify to v3.0.3
* *(deps)* Update dependency dompurify to v3.0.4
* *(deps)* Update dependency esbuild to v0.17.12
* *(deps)* Update dependency esbuild to v0.17.13
* *(deps)* Update dependency esbuild to v0.17.14
* *(deps)* Update dependency esbuild to v0.17.15
* *(deps)* Update dependency esbuild to v0.17.16
* *(deps)* Update dependency esbuild to v0.17.17
* *(deps)* Update dependency esbuild to v0.17.18
* *(deps)* Update dependency esbuild to v0.17.19
* *(deps)* Update dependency esbuild to v0.18.0
* *(deps)* Update dependency esbuild to v0.18.1
* *(deps)* Update dependency esbuild to v0.18.10
* *(deps)* Update dependency esbuild to v0.18.11
* *(deps)* Update dependency esbuild to v0.18.2
* *(deps)* Update dependency esbuild to v0.18.3
* *(deps)* Update dependency esbuild to v0.18.4
* *(deps)* Update dependency esbuild to v0.18.5
* *(deps)* Update dependency esbuild to v0.18.6
* *(deps)* Update dependency esbuild to v0.18.9
* *(deps)* Update dependency eslint to v8.37.0
* *(deps)* Update dependency eslint to v8.38.0
* *(deps)* Update dependency eslint to v8.39.0
* *(deps)* Update dependency eslint to v8.40.0
* *(deps)* Update dependency eslint to v8.41.0
* *(deps)* Update dependency eslint to v8.42.0
* *(deps)* Update dependency eslint to v8.43.0
* *(deps)* Update dependency eslint to v8.44.0
* *(deps)* Update dependency eslint-plugin-vue to v9.10.0
* *(deps)* Update dependency eslint-plugin-vue to v9.11.0
* *(deps)* Update dependency eslint-plugin-vue to v9.11.1
* *(deps)* Update dependency eslint-plugin-vue to v9.12.0
* *(deps)* Update dependency eslint-plugin-vue to v9.13.0
* *(deps)* Update dependency flexsearch to v0.7.31
* *(deps)* Update dependency floating-vue to v2.0.0-beta.21
* *(deps)* Update dependency floating-vue to v2.0.0-beta.22
* *(deps)* Update dependency floating-vue to v2.0.0-beta.24
* *(deps)* Update dependency happy-dom to v9
* *(deps)* Update dependency happy-dom to v9.1.9
* *(deps)* Update dependency happy-dom to v9.10.1
* *(deps)* Update dependency happy-dom to v9.10.9
* *(deps)* Update dependency happy-dom to v9.18.3
* *(deps)* Update dependency happy-dom to v9.20.1
* *(deps)* Update dependency happy-dom to v9.20.3
* *(deps)* Update dependency happy-dom to v9.7.1
* *(deps)* Update dependency happy-dom to v9.9.2
* *(deps)* Update dependency highlight.js to v11.8.0
* *(deps)* Update dependency histoire to v0.16.2
* *(deps)* Update dependency marked to v4.3.0
* *(deps)* Update dependency marked to v5
* *(deps)* Update dependency marked to v5.0.1
* *(deps)* Update dependency marked to v5.0.2
* *(deps)* Update dependency marked to v5.0.3
* *(deps)* Update dependency marked to v5.0.4
* *(deps)* Update dependency marked to v5.0.5
* *(deps)* Update dependency marked to v5.1.0
* *(deps)* Update dependency netlify-cli to v13.1.2
* *(deps)* Update dependency netlify-cli to v13.1.6
* *(deps)* Update dependency netlify-cli to v13.2.1
* *(deps)* Update dependency netlify-cli to v13.2.2
* *(deps)* Update dependency netlify-cli to v14
* *(deps)* Update dependency netlify-cli to v14.3.1
* *(deps)* Update dependency pinia to v2.0.34
* *(deps)* Update dependency pinia to v2.0.35
* *(deps)* Update dependency pinia to v2.0.36
* *(deps)* Update dependency pinia to v2.1.4
* *(deps)* Update dependency postcss to v8.4.22
* *(deps)* Update dependency postcss to v8.4.23
* *(deps)* Update dependency postcss to v8.4.24
* *(deps)* Update dependency postcss-preset-env to v8.1.0
* *(deps)* Update dependency postcss-preset-env to v8.2.0
* *(deps)* Update dependency postcss-preset-env to v8.3.0
* *(deps)* Update dependency postcss-preset-env to v8.3.1
* *(deps)* Update dependency postcss-preset-env to v8.3.2
* *(deps)* Update dependency postcss-preset-env to v8.4.1
* *(deps)* Update dependency postcss-preset-env to v8.4.2
* *(deps)* Update dependency postcss-preset-env to v8.5.0
* *(deps)* Update dependency postcss-preset-env to v8.5.1
* *(deps)* Update dependency rollup to v3.20.0
* *(deps)* Update dependency rollup to v3.20.1
* *(deps)* Update dependency rollup to v3.20.2
* *(deps)* Update dependency rollup to v3.20.3
* *(deps)* Update dependency rollup to v3.20.4
* *(deps)* Update dependency rollup to v3.20.5
* *(deps)* Update dependency rollup to v3.20.6
* *(deps)* Update dependency rollup to v3.20.7
* *(deps)* Update dependency rollup to v3.21.0
* *(deps)* Update dependency rollup to v3.21.1
* *(deps)* Update dependency rollup to v3.21.2
* *(deps)* Update dependency rollup to v3.21.3
* *(deps)* Update dependency rollup to v3.21.4
* *(deps)* Update dependency rollup to v3.21.5
* *(deps)* Update dependency rollup to v3.21.6
* *(deps)* Update dependency rollup to v3.21.7
* *(deps)* Update dependency rollup to v3.21.8
* *(deps)* Update dependency rollup to v3.22.0
* *(deps)* Update dependency rollup to v3.23.0
* *(deps)* Update dependency rollup to v3.23.1
* *(deps)* Update dependency rollup to v3.24.0
* *(deps)* Update dependency rollup to v3.24.1
* *(deps)* Update dependency rollup to v3.25.0
* *(deps)* Update dependency rollup to v3.25.1
* *(deps)* Update dependency rollup to v3.25.2
* *(deps)* Update dependency rollup to v3.25.3
* *(deps)* Update dependency rollup to v3.26.0
* *(deps)* Update dependency rollup-plugin-visualizer to v5.9.2
* *(deps)* Update dependency sass to v1.59.3
* *(deps)* Update dependency sass to v1.60.0
* *(deps)* Update dependency sass to v1.61.0
* *(deps)* Update dependency sass to v1.62.0
* *(deps)* Update dependency sass to v1.62.1
* *(deps)* Update dependency sass to v1.63.0
* *(deps)* Update dependency sass to v1.63.2
* *(deps)* Update dependency sass to v1.63.3
* *(deps)* Update dependency sass to v1.63.4
* *(deps)* Update dependency sass to v1.63.5
* *(deps)* Update dependency sass to v1.63.6
* *(deps)* Update dependency typescript to v5
* *(deps)* Update dependency typescript to v5.0.3
* *(deps)* Update dependency typescript to v5.0.4
* *(deps)* Update dependency typescript to v5.1.3
* *(deps)* Update dependency typescript to v5.1.5
* *(deps)* Update dependency typescript to v5.1.6
* *(deps)* Update dependency ufo to v1.1.2
* *(deps)* Update dependency vite to v4.2.0
* *(deps)* Update dependency vite to v4.2.1
* *(deps)* Update dependency vite to v4.2.2
* *(deps)* Update dependency vite to v4.3.0
* *(deps)* Update dependency vite to v4.3.1
* *(deps)* Update dependency vite to v4.3.2
* *(deps)* Update dependency vite to v4.3.3
* *(deps)* Update dependency vite to v4.3.4
* *(deps)* Update dependency vite to v4.3.5
* *(deps)* Update dependency vite to v4.3.6
* *(deps)* Update dependency vite to v4.3.7
* *(deps)* Update dependency vite to v4.3.8
* *(deps)* Update dependency vite to v4.3.9
* *(deps)* Update dependency vite-plugin-pwa to v0.14.5
* *(deps)* Update dependency vite-plugin-pwa to v0.14.6
* *(deps)* Update dependency vite-plugin-pwa to v0.14.7
* *(deps)* Update dependency vite-plugin-pwa to v0.15.0
* *(deps)* Update dependency vite-plugin-pwa to v0.15.1
* *(deps)* Update dependency vite-plugin-pwa to v0.15.2
* *(deps)* Update dependency vite-plugin-pwa to v0.16.1
* *(deps)* Update dependency vite-plugin-pwa to v0.16.3
* *(deps)* Update dependency vite-plugin-pwa to v0.16.4
* *(deps)* Update dependency vite-plugin-sentry to v1.3.0
* *(deps)* Update dependency vitest to v0.29.3
* *(deps)* Update dependency vitest to v0.29.4
* *(deps)* Update dependency vitest to v0.29.5
* *(deps)* Update dependency vitest to v0.29.7
* *(deps)* Update dependency vitest to v0.29.8
* *(deps)* Update dependency vitest to v0.30.0
* *(deps)* Update dependency vitest to v0.30.1
* *(deps)* Update dependency vitest to v0.31.0
* *(deps)* Update dependency vitest to v0.31.1
* *(deps)* Update dependency vitest to v0.31.2
* *(deps)* Update dependency vitest to v0.31.4
* *(deps)* Update dependency vitest to v0.32.0
* *(deps)* Update dependency vitest to v0.32.1
* *(deps)* Update dependency vitest to v0.32.2
* *(deps)* Update dependency vitest to v0.32.3
* *(deps)* Update dependency vue to v3.3.4
* *(deps)* Update dependency vue to v3.3.4
* *(deps)* Update dependency vue-flatpickr-component to v11.0.3
* *(deps)* Update dependency vue-router to v4.2.0
* *(deps)* Update dependency vue-router to v4.2.1
* *(deps)* Update dependency vue-router to v4.2.2
* *(deps)* Update dependency vue-router to v4.2.3
* *(deps)* Update dependency vue-router to v4.2.4
* *(deps)* Update dependency vue-tsc to v1.4.0
* *(deps)* Update dependency vue-tsc to v1.4.1
* *(deps)* Update dependency vue-tsc to v1.4.2
* *(deps)* Update dependency vue-tsc to v1.4.3
* *(deps)* Update dependency vue-tsc to v1.4.4
* *(deps)* Update dependency vue-tsc to v1.6.0
* *(deps)* Update dependency vue-tsc to v1.6.1
* *(deps)* Update dependency vue-tsc to v1.6.2
* *(deps)* Update dependency vue-tsc to v1.6.3
* *(deps)* Update dependency vue-tsc to v1.6.4
* *(deps)* Update dependency vue-tsc to v1.6.5
* *(deps)* Update dependency vue-tsc to v1.8.0
* *(deps)* Update dependency vue-tsc to v1.8.1
* *(deps)* Update dependency vue-tsc to v1.8.2
* *(deps)* Update dependency vue-tsc to v1.8.3
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update flake
* *(deps)* Update font awesome to v6.4.0
* *(deps)* Update histoire to v0.15.9
* *(deps)* Update histoire to v0.16.0
* *(deps)* Update histoire to v0.16.1
* *(deps)* Update lockfile
* *(deps)* Update node.js to v18.16.0
* *(deps)* Update node.js to v18.16.1
* *(deps)* Update node.js to v20 (#3411)
* *(deps)* Update pnpm to v7.29.3
* *(deps)* Update pnpm to v7.30.0
* *(deps)* Update pnpm to v7.30.1
* *(deps)* Update pnpm to v7.30.2
* *(deps)* Update pnpm to v7.30.3
* *(deps)* Update pnpm to v7.30.5
* *(deps)* Update pnpm to v7.31.0
* *(deps)* Update pnpm to v7.32.0
* *(deps)* Update pnpm to v8
* *(deps)* Update pnpm to v8.3.0
* *(deps)* Update pnpm to v8.3.1
* *(deps)* Update pnpm to v8.4.0
* *(deps)* Update pnpm to v8.5.0
* *(deps)* Update pnpm to v8.5.1
* *(deps)* Update pnpm to v8.6.0
* *(deps)* Update pnpm to v8.6.1
* *(deps)* Update pnpm to v8.6.2
* *(deps)* Update pnpm to v8.6.3
* *(deps)* Update pnpm to v8.6.4
* *(deps)* Update pnpm to v8.6.5
* *(deps)* Update pnpm to v8.6.6
* *(deps)* Update sentry-javascript monorepo to v7.43.0
* *(deps)* Update sentry-javascript monorepo to v7.44.0
* *(deps)* Update sentry-javascript monorepo to v7.44.1
* *(deps)* Update sentry-javascript monorepo to v7.44.2
* *(deps)* Update sentry-javascript monorepo to v7.45.0
* *(deps)* Update sentry-javascript monorepo to v7.46.0
* *(deps)* Update sentry-javascript monorepo to v7.47.0
* *(deps)* Update sentry-javascript monorepo to v7.48.0
* *(deps)* Update sentry-javascript monorepo to v7.49.0
* *(deps)* Update sentry-javascript monorepo to v7.50.0
* *(deps)* Update sentry-javascript monorepo to v7.51.0
* *(deps)* Update sentry-javascript monorepo to v7.51.2
* *(deps)* Update sentry-javascript monorepo to v7.52.0
* *(deps)* Update sentry-javascript monorepo to v7.52.1
* *(deps)* Update sentry-javascript monorepo to v7.53.0
* *(deps)* Update sentry-javascript monorepo to v7.53.1
* *(deps)* Update sentry-javascript monorepo to v7.54.0
* *(deps)* Update sentry-javascript monorepo to v7.55.0
* *(deps)* Update sentry-javascript monorepo to v7.55.2
* *(deps)* Update sentry-javascript monorepo to v7.56.0
* *(deps)* Update sentry-javascript monorepo to v7.57.0
* *(deps)* Update typescript-eslint monorepo to v5.55.0
* *(deps)* Update typescript-eslint monorepo to v5.56.0
* *(deps)* Update typescript-eslint monorepo to v5.57.0
* *(deps)* Update typescript-eslint monorepo to v5.57.1
* *(deps)* Update typescript-eslint monorepo to v5.58.0
* *(deps)* Update typescript-eslint monorepo to v5.59.0
* *(deps)* Update typescript-eslint monorepo to v5.59.1
* *(deps)* Update typescript-eslint monorepo to v5.59.11
* *(deps)* Update typescript-eslint monorepo to v5.59.2
* *(deps)* Update typescript-eslint monorepo to v5.59.5
* *(deps)* Update typescript-eslint monorepo to v5.59.6
* *(deps)* Update typescript-eslint monorepo to v5.59.7
* *(deps)* Update typescript-eslint monorepo to v5.59.8
* *(deps)* Update typescript-eslint monorepo to v5.59.9
* *(deps)* Update typescript-eslint monorepo to v5.60.0
* *(deps)* Update typescript-eslint monorepo to v5.60.1
* *(deps)* Update workbox monorepo to v6.6.0 (#3548)
* *(deps)* Update workbox monorepo to v6.6.1 (#3553)
* *(deps)* Update workbox monorepo to v7 (major) (#3556)

### Features

* *(assignees)* Show user avatar in search results
* *(datepicker)* Separate datepicker popup and datepicker logic in different components
* *(i18n)* Enable Danish translation
* *(i18n)* Enable Japanese translation
* *(i18n)* Enable Spanish translation
* *(i18n)* Use chinese name for chinese translation
* *(kanban)* Use total task count from the api instead of manually calculating it per bucket
* *(link share)* Add e2e tests for link share hash
* *(navigation)* Add hiding child projects
* *(navigation)* Allow dragging a project out from its parent project
* *(navigation)* Correctly show child projects
* *(navigation)* Make dragging a project to a parent work
* *(navigation)* Make dragging a project under another project work
* *(navigation)* Show favorite projects on top
* *(projects)* Allow setting a saved filter for tasks shown on the overview page
* *(projects)* Move hasProjects check to store
* *(quick add magic)* Allow fuzzy matching of assignees when the api results are unambiguous
* *(reminders)* Add confirm button
* *(reminders)* Add e2e tests for task reminders
* *(reminders)* Add more spacing
* *(reminders)* Add on the due / start / end date as a reminder preset
* *(reminders)* Add preset two hours before due / start / end date
* *(reminders)* Add proper time picker for relative dates
* *(reminders)* Highlight which preset or custom date is selected
* *(reminders)* Make adding new reminders less confusing
* *(reminders)* Make relative presets actually work
* *(reminders)* Move reminder settings to a popup
* *(reminders)* Only show relative reminders when there's a date to relate them to
* *(reminders)* Show resolved reminder time in a tooltip and properly bubble updated task down to the reminder component
* *(reminders)* Translate all reminder form strings
* *(sentry)* Only load sentry when it's enabled
* *(tests)* Add project tests derived from old namespace tests
* *(user)* Migrate color scheme settings to persistence in db
* *(user)* Migrate pop sound setting to store in api
* *(user)* Persist frontend settings in the api (#3594)* Rename files with list to project ([b9d3b5c](b9d3b5c75635577321acc1791219aed40c6c14a4))
* *(user)* Save quick add magic mode in api
* *(user)* Set default settings when loading persisted
* *(user)* Use user language from store after logging in
* Abstract BaseCheckbox ([8fc254d](8fc254d2db5738e5d370c9f346c8d0d1e31bb9d0))
* Add hotkeys for priority, delete and favorite on the `TaskDetailView` (#3400) ([e00c9bb](e00c9bb1afc8491039b5ffb50d4d8d9b38e6e086))
* Add message to add to home screen on mobile ([3c9083b](3c9083b90dd3e5f97109ba2a23d2f2f8cc7d6c7c))
* Add redirect for old list routes ([af523cf](af523cfcd71528c7e8d0b50874f4766f40f958d2))
* Add setting for infinite nesting ([cb218ec](cb218ec0c31a41ba41a713a3757f71ad550dd71c))
* Add transition to input icons ([abb5128](abb51284269d84de14d0a156c386c63dc596b9ab))
* Add vite-plugin sentry (#1991) ([5ca31d0](5ca31d00eeff28f4728a4d07b96d761a6f174207))
* Add vite-plugin sentry ([73947f0](73947f0ba4031cb0f9aff78f8a7e3316a36d59b4))
* Allow creating a new project directly as a child project from another one ([b341184](b34118485cc056146682cd4592c90e4662b307eb))
* Allow disabling icon changes ([efb3407](efb3407b8769a23f4352161d6db6267ce4b30eee))
* Allow hiding the quick add magic help tooltip with a button ([7fb85da](7fb85dacecdae597180553036243ab845d50ede5))
* Allow selecting a parent project when creating a project ([ce887c3](ce887c38f3a9e84c832bfbf62efa455df37a1a4f))
* Allow selecting a parent project when duplicating a project ([799c0be](799c0be8306cfc5150611153c59701e96d56893a))
* Allow selecting a parent project when editing a project ([ee8f80c](ee8f80cc70109a496959da167d14ffda4e2a6175))
* Allow to edit existing relative reminders ([5d38b83](5d38b8327fc323c571fced33442bdb923d6d3baa))
* Better vscode vitest integration ([314cbf4](314cbf471f8e9cff2a3fca6bbd969807401b5cda))
* Change the link share hash name ([2066056](20660564c16283c77029bc3c3125c6c3febde47e))
* Check link share auth from store instead ([c2ffe3a](c2ffe3a9dcfd1e067b8d92e1d69183c2a8acfa8f))
* Don't handle child projects and instead only save the ids ([760efa8](760efa854dcc83e74f96782339b79b8d27b853b2))
* Don't use child_projects property from api ([ebd9c47](ebd9c4702ed1c6920d47e5e42294e6d4fa3c73c0))
* Edit relative reminders (#3248) ([3f8e457](3f8e457d5250df0b3af34d8f3bb0c053b15a97be))
* Edit relative reminders ([14e2698](14e26988331ca72afae01b8264969458cdb4a509))
* Hide quick add magic help behind a tooltip (#3353) ([a988565](a988565227f57dfc728319d433532f71e61d6424))
* Highlight hint icon when hovering the input ([422d7fc](422d7fc693caf886a49d03ff48b56ae6ce825356))
* Improve datemathHelp.vue ([795b26e](795b26e1dde781e152ab03fc31fd95f9f106a452))
* Improve handling of an invalid api url ([24ad2f8](24ad2f892db0fce3458624c9dad8735130253fa0))
* Improve user assignments via quick add magic (#3348) ([d9f608e](d9f608e8b4be4da380a535edcce1782c6d21926d))
* Improve variable naming for ProjectCardGrid ([a4be973](a4be973e29e81db4e244427fc46a11b4c8c95f4c))
* Load all projects earlier than in the navigation and use the loading state of the store ([1d93661](1d936618faecb0ddcb10f7c900096a3705614dbd))
* Mark undone if task moved from isDoneBucket (#3291) ([30adad5](30adad5ae6568b5ef1125f206989d447fb999eee))
* Move namespaces list to projects list ([e1bdabc](e1bdabc8d670f7342f4f0777a30a961e3fd4601d))
* Move navigation item to component ([3db4e01](3db4e011d4b625cee940c58ee32d065b8c43f1bb))
* Move quick add magic to a popup behind an icon ([6989558](69895589636ee6369c9778f529bf1df953acb7b1))
* New image for the unauthenticated views ([bef25c4](bef25c49d535ff3940a0112a715f5b351e816468))
* Optimize print view for project views ([8e2c76a](8e2c76a33eec573afab0b754d0707f84e2cca962))
* Persist link share auth rule in url hash (#3336) ([da3eaf0](da3eaf0d357c24775ba8a4cf8f089e5042f73c00))
* Persist link share auth rule in url hash ([f68bb26](f68bb2625e5f619f365fdd421aeda2b8af879aab))
* Prepare for pnpm 8 (#3331) ([7d3b97d](7d3b97d422896e17ab9231c66e49da6c07967d7e))
* Rebuild main navigation so that it works recursively with projects ([06e8cdb](06e8cdb9d2907c846ca7c555b31571b5c1798433))
* Remove all namespace leftovers ([1bd17d6](1bd17d6e50034c159150095f1c51a966293a6726))
* Remove namespaces, make projects infinitely nestable (#3323) ([ac1d374](ac1d374191fca764a70d9851d9828a78ae27c075))
* Rename link share hash prefix ([b9f0635](b9f0635d9fcc764c7ee188c95ce59ac358f735cf))
* Rename list to project everywhere ([befa6f2](befa6f27bb607a57eb8ed49d0152b85cdab4cb95))
* Replace color dot with handle icon on hover ([a3e2cbe](a3e2cbeb27ad8b0d052df62c9f24f8dd3808ddda))
* Set the current language to the one saved by the user on login ([acb212a](acb212ab241e1ed873c943e9c5fa3bcfb2c83a91))
* Show all parent projects in project search ([6a8c656](6a8c656dbb0a4729035468aedc60fd06e80c17ed))
* Show all parent projects in task detail view ([63ba298](63ba2982c92d495de6c7e3526c3693dcfe0e3fba))
* Show avatar and full name in team overview ([b80f070](b80f07043104868d134761b34582b234d12274e1))
* Show initial list of users when opening the assignees view ([59c942a](59c942af735a40f68cfd01caadb22694113da8ae))
* Start adding relative reminder picker with more options ([9df6950](9df6950d1a4a361c075020319ce3037e19e0912d))
* Translate inbox project title ([f2ca2d8](f2ca2d850de5b4b3b3d90e3a5c41adebca2dc1a5))
* Type i18n improvements ([dea1789](dea1789a00981fb496f0c1f4c19a6f0749e4de70))
* Use new Reminders API instead of reminder_dates ([f747d5b](f747d5b2fcadb7459c389372dca4507b75cdd4fa))
* Wrap projects navigation in a <Suspense> so that we can use top level await ([2579c33](2579c33ee1d07234c3ad42d75f2c7a1f7bfdb149))


### Miscellaneous Tasks

* *(ci)* Remove netlify dependency (#3459)
* *(ci)* Sign drone config
* *(editor)* Disable deprecated marked options
* *(i18n)* Clarify translation string
* *(parseSubtasksViaIndentation)* Fix comment (#3259)
* *(reminders)* Remove reminderDates property
* *(sentry)* Always use the same version
* *(sentry)* Ignore missing commits
* *(sentry)* Only load sentry when enabled
* *(sentry)* Remove debug options
* *(sentry)* Remove sourcemaps after upload via plugin
* *(sentry)* Use correct chunks option
* *(task)* Move toggleFavorite to store
* *(task)* Use ref for task instead of reactive
* *(tests)* Enable experimental memory management for cypress tests
* *(user)* Cleanup* Update JSDoc example ([bfbfd6a](bfbfd6a4212d493912406c1c505b6c0a24f0f014))
* Add comment on overriding ([21ad830](21ad8301f28ba838c577acb72cb66ea00e176876))
* Add types for emit ([c567874](c56787443f6f9f6be0f8d8501dd4e6e7a768648a))
* Better function naming in password components ([a416d26](a416d26f7cfd163cadb0b6ded107b217ecad5d7c))
* Catch error when trying to play pop sound ([929d4f4](929d4f402342de309dd8e453252d22fcb9f362a6))
* Chore; extract code to reminder-period.vue ([0d6c0c8](0d6c0c8399c9fc73843bdbeb84ff19467edcaa90))
* Clarify users when can still be found even if they disabled it ([302ba2b](302ba2bec7f592f6b0b1fba84a5a1a9fd5f994de))
* Cleanup namespace leftovers ([2e33615](2e336150e086354b1623569aa98ab9c5be48c59a))
* Don't recalculate everything ([9c3259c](9c3259c660e8436f41b5494c9567319090c03bd6))
* Don't set the current project to null if it's undefined already ([e4d97e0](e4d97e05205e2c36143319ccf07ccac03f5de408))
* Don't show selection for parent project when no projects are available ([c30dcff](c30dcff45157e5b89b7bb6c2442271c15da33fc4))
* Don't wrap a computed in another computed ([afaf184](afaf1846ece65b8b2bbee971fafb31a535a4381b))
* Export favorite projects from store ([131022d](131022da427616765f8109ca8ac8f6bad1bdcbbb))
* Export not archived root projects ([b5d9afd](b5d9afd0f72aaf28b89f4877ce3ad2eabe6c3d7b))
* Export projects as array directly from projects store ([e4379f0](e4379f0a229b7b8572fddb029658713a0bbfca1d))
* Follow the happy path ([a33e2f6](a33e2f6c00f35f36aabb6b4d6e823396d29cdf3d))
* Format ([4ad9773](4ad9773022b5873fff09b7afade02c026ac5332f))
* Format ([638d187](638d187a24020d698327b0a0d04a5897672d3b79))
* Formatting ([b92d780](b92d780cda3ab7222c4a6ab7323d1dd3f679b514))
* Group return parameter ([5298706](52987060b11ac0418b6a88f1beabaee59165117d))
* Import const instead of redeclaring it ([61baf02](61baf02e26b292e3f02816a483eb7d92fb49d8ab))
* Improve prop type definition ([638f6be](638f6bea24980658d0f5fb3432d7b64c2ae06f75))
* Make fuzzy matching a parameter ([aeb73a3](aeb73a374f84f6b01d4be4cc784336a214a4cdfa))
* Move ProjectsNavigationWrapper back to navigation.vue ([65522a5](65522a57f1ceddfabeba235e17f8f81ee6bae47b))
* Move all options to component props ([db1c6d6](db1c6d6a41591c8ee5df2d2ee400aaaeda0d02bb))
* Move const ([0ce150a](0ce150af237985dda0cf44f24179ebae332e7585))
* Move duplicate project logic to composable ([b69a056](b69a05689be6e2c833c838cde052702600d245c7))
* Move loader class ([ac78e85](ac78e85e1726b6d7047db72ccbbaf29ac11d1696))
* Move loading styles to variant into the component ([76814a2](76814a2d3f68876934c5791bb4901fca5f95c00f))
* Move more logic to ProjectsNavigationItem.vue ([b567146](b567146d69f1c6a1eba6e37061bde7f627ff8654))
* Move positioning css ([7110c9a](7110c9a5ceb58e6e9675c0f91ddb34c9ab8f2cbc))
* Move styles to components ([25c3b7b](25c3b7bcbfb4ddc2163092ed7c1d5e4758967f1b))
* Move v-if ([12ebefd](12ebefd86a61ca5c82b104b4155a4989c8622713))
* Only apply padding where needed ([ddcd6a1](ddcd6a17dc659611910c2d4ed84fcff575e0ca3a))
* Re-add top menu spacing ([086f50d](086f50d4feed90aac0c458d3f53cbe59ae7402c8))
* Redirect to new project after creating from store ([6b824a4](6b824a49abe8854045c7670fcd6da50539c9fce5))
* Reduce nesting ([06a1ff6](06a1ff6f4bea4cc7447d528423de54f14583dca4))
* Refactor get parents project and move to projects store ([c32a198](c32a198a34edd3db7d6967010ce9dde401d1c864))
* Remove nesting ([a4c8fcc](a4c8fccb115f019840025659c7a8a4bac31eee04))
* Remove old comment ([4134fcb](4134fcbd752ab4cc7691907264b04cf64e11d012))
* Remove old todo ([4e21b46](4e21b463df9af5aec9a5b45c8331f5a9f9e8aeb9))
* Remove triggered notifications as it's not supported anywhere ([8a75790](8a75790453427287dc5a57ff3b59cd2b9cabd3f4))
* Remove type annotation for computed ([a3e289c](a3e289c06c992b24dcff21b1c4f8871676101d98))
* Remove unnecessary map ([336db56](336db56316dec7aeacf2174f5945764dc350769c))
* Remove unused class ([d4e4525](d4e452545afe94ed2e860cd14982462e080a4d49))
* Remove unused code ([652db56](652db56d42b39c05385ff7484fa43b0baa769759))
* Remove user margin from the component ([57c64bb](57c64bbf71342b4e9e2e9e3808412b5e0cf01006))
* Remove user margin from the component ([a1dd1d6](a1dd1d6664479e125e2f8ae87a9d2a57bf94fc9e))
* Remove wrapper div ([2c9693a](2c9693a83eeca832d49d519c5676ae30569628ca))
* Rename alias ([a803bc6](a803bc637e44893aa6921b70215a3206acdc5a91))
* Rename archived message key ([4dee3a9](4dee3a90e9a76cdd190eb28b3327bef1bcc34787))
* Rename flag ([6e09543](6e095436e9bfb6856c6aa469fc4cad93f239bad4))
* Rename getRedirectRoute ([59b05e9](59b05e9836946ed8b9dbb3926fc694641d8508a1))
* Rename prop ([2bb7ff1](2bb7ff1803d5a35bdd61a94e7a4d6fd03d5d1492))
* Replace section with a div ([9b10693](9b1069317283fc20c834eac981e0b2a500e32dba))
* Set project id from the outside ([6c9cbaa](6c9cbaadc821ab92e85b1f8e3fcb3fa85ea99670))
* Update nix flake ([f40035d](f40035dc7943e8199c553acfec838f21ea212c3e))
* Use <menu> instead of <ul> ([49fac7d](49fac7db1cbefce49712797869b956f31e8f541c))
* Use klona to clone project object ([55e9122](55e912221be4b4765cdb3a7bd0e3dc693478ac81))
* Use long variable name ([6f1baa3](6f1baa3219093147842efe10f92482364516c84c))
* Use long variable name ([a0d39e6](a0d39e6081f35e4ba6589b7840168b0c69b3210f))
* Use project id type ([a342ae6](a342ae67de1c884895ce3304cf6eb1757a38573a))
* Use startsWith for prefix matching ([10ac1ff](10ac1ff66a2bcd797f54c83dda13745fdf359f33))
* Use stores directly ([a7440ed](a7440ed296ec0e99c9dc81e43617b3b54fc518a7))
* [skip ci] Updated translations via Crowdin


## [0.20.5] - 2023-03-12

### Bug Fixes

* *(docker)* Add cap_net_bind to the nginx binary in the docker container
* *(docker)* Revert unprivileged user

### Dependencies

* *(deps)* Update dependency sass to v1.59.2
* *(deps)* Update dependency eslint to v8.36.0

## [0.20.4] - 2023-03-10

### Bug Fixes

* *(base)* Use Build Time Base Path
* *(docker)* Cross compilation with buildx
* *(docker)* Default api url
* *(docker)* Make sure the service worker and webmanifest are never cached
* *(filter)* Validate title before creating or editing a filter
* *(filter)* Don't allow marking a filter as favorite
* *(i18n)* Load language files before doing anything else (#3218)
* *(keyboard-shortcuts)* Use card prop
* *(list)* Make sure favorite lists are not duplicated in the menu when renaming them
* *(menu)* Don't show drag handle for not draggable menu items
* *(postcss-preset-env)* Client side polyfills (#3051)
* *(quick actions)* Don't throw an error message when selecting the last items with the arrow keys
* *(quick actions)* Hide edges of last entry on hover
* *(quick add magic)* Correctly parse "next {weekday}" on the beginning of the text
* *(quick-actions)* Nothing happening on team click (#3186)
* *(table view)* Correctly load sort order from local storage
* *(task)* Allow clicking on the whole task to open the task detail view
* *(tests)* Only look in src for tests
* Make sure global error handler handles unrejected promises correctly ([4576da0](4576da0dd394ee68801b1dc424c9550896d63737))
* Use Build Time Base Path (#2964) ([6572f75](6572f75e5d111f7f2dd06e8c2ad0e0d16091fca6))
* Always show update popup on top ([7cbf0ac](7cbf0acac503c508a44e0491ae51e6d5749dfa04))
* Button styles ([d40729c](d40729cbe70b760bcc64d56130a410b05ef9d3dc))
* Stop revealing elements on hover if hover is not supported (#3191) ([7b6f76d](7b6f76d1b4698d0d6c6889aaab3f1cdad80469f8))
* Sync sidebar transition with `<main>` (#3200) ([0f97ba6](0f97ba6ec904226ed91cd3ade8223e2959e9207a))
* Collapse menu on mobile when path changes ([1b06112](1b06112db4ba5ad4144b5868dd04e954be1d77f7))
* I18ze a string (#3210) ([b4dd23b](b4dd23b85d909f7e629e953f1d8543ccbf963a1c))


### Dependencies

* *(deps)* Update sentry-javascript monorepo to v7.33.0 (#3004)
* *(deps)* Update dependency axios to v1.2.4 (#3005)
* *(deps)* Update pnpm to v7.26.0 (#3002)
* *(deps)* Update dependency cypress to v12.4.0 (#3006)
* *(deps)* Update dependency @infectoone/vue-ganttastic to v2.1.4 (#3009)
* *(deps)* Update dependency vitest to v0.28.2 (#3008)
* *(deps)* Update dependency rollup to v3.11.0 (#3013)
* *(deps)* Update dependency @vitejs/plugin-legacy to v3.0.2 (#3012)
* *(deps)* Update dependency axios to v1.2.5
* *(deps)* Update sentry-javascript monorepo to v7.34.0
* *(deps)* Update pnpm to v7.26.1
* *(deps)* Update dependency @vue/test-utils to v2.2.8
* *(deps)* Update dependency vitest to v0.28.3 (#3019)
* *(deps)* Update dependency cypress to v12.4.1
* *(deps)* Update dependency rollup to v3.12.0
* *(deps)* Update dependency esbuild to v0.17.5
* *(deps)* Update dependency axios to v1.2.6
* *(deps)* Update dependency @vueuse/core to v9.12.0
* *(deps)* Update pnpm to v7.26.2
* *(deps)* Update dependency eslint to v8.33.0
* *(deps)* Update dependency netlify-cli to v12.10.0
* *(deps)* Update dependency happy-dom to v8.2.0
* *(deps)* Update dependency caniuse-lite to v1.0.30001449
* *(deps)* Update dependency typescript to v4.9.5
* *(deps)* Update typescript-eslint monorepo to v5.50.0
* *(deps)* Update dependency axios to v1.3.0 (#3036)
* *(deps)* Update dependency sass to v1.58.0
* *(deps)* Update dependency cypress to v12.5.0
* *(deps)* Update pnpm to v7.26.3
* *(deps)* Update dependency rollup to v3.12.1
* *(deps)* Update sentry-javascript monorepo to v7.35.0 (#3041)
* *(deps)* Update dependency pinia to v2.0.30 (#3042)
* *(deps)* Update dependency @vue/test-utils to v2.2.9
* *(deps)* Update dependency axios to v1.3.1
* *(deps)* Update dependency vue to v3.2.47
* *(deps)* Update dependency vite to v4.1.0
* *(deps)* Update dependency postcss-preset-env to v8 (#3000)
* *(deps)* Update dependency @vitejs/plugin-legacy to v4
* *(deps)* Update dependency @vitejs/plugin-legacy to v4.0.1
* *(deps)* Update sentry-javascript monorepo to v7.36.0
* *(deps)* Update dependency vite to v4.1.1
* *(deps)* Update dependency cypress to v12.5.1
* *(deps)* Update dependency @vue/test-utils to v2.2.10
* *(deps)* Update dependency vitest to v0.28.4
* *(deps)* Update dependency rollup to v3.13.0
* *(deps)* Update dependency axios to v1.3.2
* *(deps)* Update dependency rollup to v3.14.0
* *(deps)* Update dependency @types/node to v18.11.19
* *(deps)* Update dependency @histoire/plugin-screenshot to v0.13.0
* *(deps)* Update dependency histoire to v0.13.0
* *(deps)* Update caniuse-and-related
* *(deps)* Update dependency @histoire/plugin-vue to v0.13.0
* *(deps)* Update dependency happy-dom to v8.2.6
* *(deps)* Update typescript-eslint monorepo to v5.51.0
* *(deps)* Update dependency esbuild to v0.17.6
* *(deps)* Update dependency @cypress/vue to v5.0.4
* *(deps)* Update dependency @types/node to v18.13.0
* *(deps)* Update dependency vite-plugin-pwa to v0.14.2
* *(deps)* Update font awesome to v6.3.0
* *(deps)* Update pnpm to v7.27.0
* *(deps)* Update dependency @histoire/plugin-screenshot to v0.13.1
* *(deps)* Update dependency @histoire/plugin-vue to v0.13.1
* *(deps)* Update dependency vite-plugin-pwa to v0.14.3
* *(deps)* Update dependency histoire to v0.13.1
* *(deps)* Update dependency @histoire/plugin-screenshot to v0.13.2
* *(deps)* Update dependency @histoire/plugin-vue to v0.13.2
* *(deps)* Update dependency histoire to v0.13.2
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v0.8.2
* *(deps)* Update sentry-javascript monorepo to v7.37.0
* *(deps)* Update dependency esbuild to v0.17.7
* *(deps)* Update dependency rollup to v3.15.0
* *(deps)* Create a group for all histoire dependencies
* *(deps)* Update dependency @histoire/plugin-vue to v0.14.0
* *(deps)* Update dependency @histoire/plugin-screenshot to v0.14.0
* *(deps)* Update dependency @histoire/plugin-vue to v0.14.0
* *(deps)* Update dependency histoire to v0.14.0
* *(deps)* Update sentry-javascript monorepo to v7.37.1
* *(deps)* Update dependency histoire to v0.14.2
* *(deps)* Include histoire main package in histoire renovate group
* *(deps)* Histoire renovate group
* *(deps)* Update dependency eslint to v8.34.0
* *(deps)* Update histoire to v0.14.2
* *(deps)* Update dependency vite-plugin-pwa to v0.14.4
* *(deps)* Update dependency esbuild to v0.17.8
* *(deps)* Update dependency netlify-cli to v12.12.0
* *(deps)* Update dependency caniuse-lite to v1.0.30001451
* *(deps)* Update dependency vite-plugin-inject-preload to v1.3.0
* *(deps)* Update dependency vitest to v0.28.5
* *(deps)* Update sentry-javascript monorepo to v7.37.2
* *(deps)* Update dependency dompurify to v3 (#3107)
* *(deps)* Update typescript-eslint monorepo to v5.52.0
* *(deps)* Update dependency axios to v1.3.3
* *(deps)* Update dependency start-server-and-test to v1.15.4 (#3109)
* *(deps)* Update dependency sass to v1.58.1
* *(deps)* Update dependency vue-flatpickr-component to v11.0.2 (#3112)
* *(deps)* Update dependency @kyvg/vue3-notification to v2.9.0 (#3113)
* *(deps)* Update histoire to v0.15.1
* *(deps)* Update histoire to v0.15.3
* *(deps)* Update dependency vue-tsc to v1.1.0
* *(deps)* Pin node.js to 18.14.0
* *(deps)* Update dependency cypress to v12.6.0 (#3115)
* *(deps)* Update histoire to v0.15.4
* *(deps)* Update dependency vue-tsc to v1.1.2
* *(deps)* Update dependency sass to v1.58.2
* *(deps)* Update dependency ufo to v1.1.0
* *(deps)* Update node.js to v18.14.1
* *(deps)* Update dependency vite to v4.1.2
* *(deps)* Update sentry-javascript monorepo to v7.38.0
* *(deps)* Update dependency rollup to v3.16.0
* *(deps)* Update histoire to v0.15.7
* *(deps)* Update dependency blurhash to v2.0.5
* *(deps)* Update dependency @cypress/vite-dev-server to v5.0.3
* *(deps)* Update dependency @types/node to v18.14.0
* *(deps)* Update histoire to v0.15.8
* *(deps)* Update dependency @vueuse/core to v9.13.0
* *(deps)* Update dependency rollup to v3.17.0
* *(deps)* Update pnpm to v7.27.1
* *(deps)* Update dependency vue-tsc to v1.1.3
* *(deps)* Update dependency sass to v1.58.3
* *(deps)* Update dependency rollup to v3.17.1
* *(deps)* Update dependency esbuild to v0.17.9
* *(deps)* Update dependency vite to v4.1.3
* *(deps)* Update dependency @vue/test-utils to v2.3.0
* *(deps)* Update dependency caniuse-lite to v1.0.30001457
* *(deps)* Update dependency codemirror to v5.65.12
* *(deps)* Update dependency pinia to v2.0.31
* *(deps)* Update dependency vue-tsc to v1.1.4
* *(deps)* Update dependency rollup to v3.17.2
* *(deps)* Update dependency happy-dom to v8.6.0
* *(deps)* Update dependency netlify-cli to v12.13.2
* *(deps)* Update dependency esbuild to v0.17.10
* *(deps)* Update typescript-eslint monorepo to v5.53.0
* *(deps)* Update dependency vue-tsc to v1.1.5
* *(deps)* Update dependency pinia to v2.0.32
* *(deps)* Update node.js to v18.14.2
* *(deps)* Update dependency vite to v4.1.4
* *(deps)* Update dependency vue-tsc to v1.1.7
* *(deps)* Update dependency axios to v1.3.4
* *(deps)* Update dependency @types/node to v18.14.1
* *(deps)* Update dependency @cypress/vite-dev-server to v5.0.4
* *(deps)* Update dependency cypress to v12.7.0
* *(deps)* Update dependency vue-tsc to v1.2.0
* *(deps)* Update dependency vitest to v0.29.1
* *(deps)* Update pnpm to v7.28.0
* *(deps)* Update dependency eslint to v8.35.0
* *(deps)* Update dependency rollup to v3.17.3
* *(deps)* Update dependency netlify-cli to v13
* *(deps)* Update dependency happy-dom to v8.9.0
* *(deps)* Update dependency caniuse-lite to v1.0.30001458
* *(deps)* Update dependency start-server-and-test to v1.15.5
* *(deps)* Update dependency start-server-and-test to v2
* *(deps)* Update dependency @types/node to v18.14.2
* *(deps)* Update sentry-javascript monorepo to v7.39.0
* *(deps)* Update typescript-eslint monorepo to v5.54.0
* *(deps)* Update dependency ufo to v1.1.1
* *(deps)* Update dependency vitest to v0.29.2
* *(deps)* Update dependency rollup to v3.18.0
* *(deps)* Update dependency dompurify to v3.0.1
* *(deps)* Update sentry-javascript monorepo to v7.40.0
* *(deps)* Update dependency @types/node to v18.14.4
* *(deps)* Update dependency @types/node to v18.14.5
* *(deps)* Update dependency @types/node to v18.14.6
* *(deps)* Update dependency esbuild to v0.17.11
* *(deps)* Update dependency netlify-cli to v13.0.1
* *(deps)* Update dependency caniuse-lite to v1.0.30001460
* *(deps)* Update pnpm to v7.29.0
* *(deps)* Update sentry-javascript monorepo to v7.41.0
* *(deps)* Update typescript-eslint monorepo to v5.54.1
* *(deps)* Update dependency pinia to v2.0.33
* *(deps)* Update node.js to v18.15.0
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v0.9.0
* *(deps)* Update pnpm to v7.29.1
* *(deps)* Update dependency @vue/test-utils to v2.3.1
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v0.9.1
* *(deps)* Update sentry-javascript monorepo to v7.42.0
* *(deps)* Update dependency rollup to v3.19.0
* *(deps)* Update dependency vite-plugin-inject-preload to v1.3.1
* *(deps)* Update dependency @types/node to v18.15.0
* *(deps)* Update dependency autoprefixer to v10.4.14
* *(deps)* Update dependency rollup to v3.19.1

### Features

* *(config)* Support Setting Base Path in .env
* Use v-show for navigation buttons ([7ed1a37](7ed1a37de53cb8c15994e9524a52080170db5950))
* Unindent settings page (#2996) ([13a39be](13a39be3de4d0f7e0f6be9c20e0464e86b87c676))
* Small content auth improvements (#2998) ([2be7847](2be784766f54810f8969e48291ce9181f2854a5b))
* Move update from navigation to app ([3db5ea4](3db5ea45d768d10458eaab0f5ee9dad0df2996e4))
* Improve naming and styles ([eaeddda](eaeddda4e468c2040862d18c9b2d37a1c0ba099e))
* Use klona instead of lodash.clonedeep (#3073) ([7b96397](7b96397e3bfa43a393ca84439069290bc4c8a5c8))
* Refactor to composable ([c502f9b](c502f9b840ee2d65193aa4ef29c7f260b49db0d2))
* Header improvements ([e8db2c2](e8db2c2b458bcae592609d5a5bc3f1b333651b25))
* Persistent menuActive state with Local Storage (#3011) ([e3dd4ef](e3dd4ef78ac818add138d0323bf65abe8a4caa29))
* Fix calculation of token invalidation (#3077) ([d6b55c7](d6b55c757067413bbc34acd48af9fb553f36db8a))
* Use renovate js-app as preset (#3087) ([97c8970](97c8970dd60b2ba1e894ca0039524c8f6a5cd5df))
* Improve recommended vscode settings ([e0f0699](e0f06999beb0a9fb5da817323744307401e85e47))


### Miscellaneous Tasks

* *(refactor)* Improve `stores/config` types (#3190)
* *(services)* Add examples for some functions
* *(services)* Let `getAll`: always return `Model[]`
* Move class name to top ([c6ed925](c6ed9254247efeb43e0763e095b145d6ec1965e1))
* Simplify error handling for login and OpenId Auth ([e67088f](e67088fdb7bd3b24cea6ee37851ef45f1fb7bdad))
* Simplify getting the error text from an exception ([9adf1ab](9adf1aba895a02f416148ddf8b6925689d6e2687))
* Typo ([81a4f2d](81a4f2d9775716bc0056348664fc24185af040d4))
* Update funding links ([7cb0cd2](7cb0cd293d6d277172eccf2558a62427bc86dfe6))
* Update funding links ([b26ea45](b26ea45fe0d1d6f5f070ef42a5d68aa6db8e6b70))
* Remove minimist dependency (not used anywhere) ([f697640](f697640636466e8f035c7d31597ee589379fa017))
* Remove sponsor ([fa0e46a](fa0e46a3991ab423c9364b65439d9e8e5a28cb7b))
* Histoire add logo link ([af4a039](af4a039502b29e9e7e21cf30d44715c7af056c15))
* Improve `@/message` `action` type (#3209) ([0eb78e3](0eb78e32f994e7032725e38d564320a5a04cbf2a))
* Remove an unused duplicate key ([9db3aed](9db3aedde9566fb94717e1dd66a21abdbda6e84a))


### Other

* *(other)* Add Ipv6 support to nginx (#100)
* *(other)* Added ipv6 control script
* *(other)* Disable listening on IPv6 ports when IPv6 is not supported (#102)
* *(other)* Docker refactoring (#3018)
* *(other)* Persist menuActive state in Local Storage
* *(other)* Refactor to only used local storage value when on desktop viewport widths
* *(other)* Solve for resize()
* *(other)* [skip ci] Updated translations via Crowdin

## [0.20.3] - 2023-01-24

### Bug Fixes

* *(BaseButton)* Prop type
* *(ci)* Make sure the i18n sync cron job actually runs
* *(ci)* Sign drone config
* *(ci)* Sign drone config
* *(ci)* Tagging logic for release docker images
* *(ci)* Sign drone config
* *(cypress)* Use ts for updateUserSettings
* *(cypress)* Use env for API_URL (#2925)
* *(drone)* Use correct property value (#2920)
* *(drone)* Pnpm cache folder path (#2932)
* *(faker)* Remove mock types (#2921)
* *(i18n)* Incorrect translation string
* *(migration)* Actually pass migration oauth code from query param
* *(quick add magic)* Make sure assignees which don't exist are not removed from task title
* *(task)* Update task description when switching between related tasks
* *(task)* Don't show the list color on the task when only viewing the list (#2975)
* *(useOnline)* Only log if actually faking state (#2924)
* Close button hover for sidebar (#2981) ([9922fcb](9922fcba65c8dc2c46c4f085813c2fbc0d0a7df6))


### Dependencies

* *(deps)* Update dependency vite to v4.0.2 (#2861)
* *(deps)* Update dependency netlify-cli to v12.4.0 (#2862)
* *(deps)* Update typescript-eslint monorepo to v5.47.0 (#2864)
* *(deps)* Update dependency esbuild to v0.16.10 (#2865)
* *(deps)* Update dependency sass to v1.57.1 (#2866)
* *(deps)* Update dependency vue-tsc to v1.0.16 (#2867)
* *(deps)* Update dependency codemirror to v5.65.11
* *(deps)* Update dependency @vueuse/core to v9.8.0
* *(deps)* Update dependency vitest to v0.26.1
* *(deps)* Update dependency @vueuse/core to v9.8.1 (#2870)
* *(deps)* Update dependency @vueuse/core to v9.8.2
* *(deps)* Update sentry-javascript monorepo to v7.28.0
* *(deps)* Update dependency cypress to v12.2.0 (#2873)
* *(deps)* Update dependency vitest to v0.26.2 (#2874)
* *(deps)* Update dependency vite to v4.0.3 (#2876)
* *(deps)* Update pnpm to v7.19.0 (#2875)
* *(deps)* Update dependency rollup to v3.8.0 (#2877)
* *(deps)* Update sentry-javascript monorepo to v7.28.1 (#2878)
* *(deps)* Update dependency @vueuse/core to v9.9.0 (#2881)
* *(deps)* Update dependency rollup to v3.8.1 (#2879)
* *(deps)* Update dependency vite-svg-loader to v4 (#2882)
* *(deps)* Update dependency vue-tsc to v1.0.17 (#2883)
* *(deps)* Update dependency caniuse-lite to v1.0.30001441 (#2884)
* *(deps)* Update dependency netlify-cli to v12.5.0 (#2886)
* *(deps)* Update pnpm to v7.20.0 (#2887)
* *(deps)* Update dependency vue-tsc to v1.0.18 (#2888)
* *(deps)* Update dependency happy-dom to v8.1.1 (#2885)
* *(deps)* Update dependency @types/node to v18.11.18 (#2889)
* *(deps)* Update typescript-eslint monorepo to v5.47.1 (#2890)
* *(deps)* Update dependency esbuild to v0.16.11
* *(deps)* Update dependency esbuild to v0.16.12 (#2893)
* *(deps)* Update dependency rollup to v3.9.0 (#2894)
* *(deps)* Update dependency rollup-plugin-visualizer to v5.9.0 (#2896)
* *(deps)* Update dependency marked to v4.2.5 (#2880)
* *(deps)* Update pnpm to v7.21.0 (#2895)
* *(deps)* Update dependency eslint to v8.31.0
* *(deps)* Update dependency vue-tsc to v1.0.19
* *(deps)* Update dependency @types/codemirror to v5.60.6
* *(deps)* Update dependency rollup to v3.9.1
* *(deps)* Update dependency vitest to v0.26.3
* *(deps)* Update dependency vite-plugin-pwa to v0.14.1 (#2909)
* *(deps)* Update dependency esbuild to v0.16.13 (#2907)
* *(deps)* Update typescript-eslint monorepo to v5.48.0 (#2906)
* *(deps)* Update dependency vue-tsc to v1.0.20
* *(deps)* Update dependency cypress to v12.3.0
* *(deps)* Update dependency @vueuse/core to v9.10.0 (#2911)
* *(deps)* Update pnpm to v7.22.0 (#2910)
* *(deps)* Update dependency @vue/test-utils to v2.2.7 (#2914)
* *(deps)* Update dependency vite to v4.0.4 (#2908)
* *(deps)* Update sentry-javascript monorepo to v7.29.0 (#2915)
* *(deps)* Update dependency esbuild to v0.16.14
* *(deps)* Update dependency axios to v1
* *(deps)* Update dependency vue-tsc to v1.0.21
* *(deps)* Update dependency vue-tsc to v1.0.22
* *(deps)* Update dependency dompurify to v2.4.2
* *(deps)* Update dependency dompurify to v2.4.3 (#2931)
* *(deps)* Update dependency postcss to v8.4.21 (#2933)
* *(deps)* Update dependency esbuild to v0.16.15 (#2934)
* *(deps)* Update dependency vue-tsc to v1.0.24
* *(deps)* Update pnpm to v7.23.0 (#2940)
* *(deps)* Update dependency happy-dom to v8.1.3 (#2939)
* *(deps)* Update dependency esbuild to v0.16.16 (#2937)
* *(deps)* Update dependency caniuse-lite to v1.0.30001442 (#2938)
* *(deps)* Update dependency vitest to v0.27.0 (#2941)
* *(deps)* Update typescript-eslint monorepo to v5.48.1 (#2942)
* *(deps)* Update pnpm to v7.24.2 (#2944)
* *(deps)* Update sentry-javascript monorepo to v7.30.0 (#2945)
* *(deps)* Update pnpm to v7.24.3 (#2946)
* *(deps)* Update dependency vitest to v0.27.1 (#2947)
* *(deps)* Update dependency esbuild to v0.16.17 (#2948)
* *(deps)* Update dependency rollup to v3.10.0 (#2949)
* *(deps)* Update dependency eslint-plugin-vue to v9.9.0 (#2950)
* *(deps)* Update pnpm to v7.25.0 (#2951)
* *(deps)* Update dependency marked to v4.2.12 (#2952)
* *(deps)* Update dependency esbuild to v0.17.0 (#2953)
* *(deps)* Update dependency eslint to v8.32.0 (#2954)
* *(deps)* Update dependency vue-advanced-cropper to v2.8.8 (#2955)
* *(deps)* Update dependency pinia to v2.0.29 (#2956)
* *(deps)* Update dependency @kyvg/vue3-notification to v2.8.0 (#2957)
* *(deps)* Update dependency caniuse-lite to v1.0.30001445 (#2958)
* *(deps)* Update dependency happy-dom to v8.1.4 (#2959)
* *(deps)* Update dependency netlify-cli to v12.7.2 (#2960)
* *(deps)* Update sentry-javascript monorepo to v7.31.0
* *(deps)* Update dependency esbuild to v0.17.1 (#2963)
* *(deps)* Update typescript-eslint monorepo to v5.48.2 (#2962)
* *(deps)* Update dependency esbuild to v0.17.2 (#2965)
* *(deps)* Update dependency vitest to v0.27.2 (#2966)
* *(deps)* Update dependency @vueuse/core to v9.11.0 (#2967)
* *(deps)* Update sentry-javascript monorepo to v7.31.1 (#2973)
* *(deps)* Update dependency axios to v1.2.3 (#2974)
* *(deps)* Update dependency esbuild to v0.17.3 (#2976)
* *(deps)* Update pnpm to v7.25.1 (#2977)
* *(deps)* Update dependency @vueuse/core to v9.11.1
* *(deps)* Update dependency rollup to v3.10.1
* *(deps)* Update dependency vite-plugin-inject-preload to v1.2.0 (#2983)
* *(deps)* Update dependency vitest to v0.27.3 (#2984)
* *(deps)* Update dependency esbuild to v0.17.4 (#2985)
* *(deps)* Update dependency caniuse-lite to v1.0.30001447 (#2986)
* *(deps)* Update dependency happy-dom to v8.1.5 (#2987)
* *(deps)* Update dependency netlify-cli to v12.9.1 (#2988)
* *(deps)* Update sentry-javascript monorepo to v7.32.1 (#2991)
* *(deps)* Update dependency vitest to v0.28.1 (#2990)
* *(deps)* Update dependency @types/codemirror to v5.60.7 (#2993)
* *(deps)* Update typescript-eslint monorepo to v5.49.0 (#2994)
* *(deps)* Update dependency start-server-and-test to v1.15.3
* *(deps)* Update dependency @fortawesome/vue-fontawesome to v3.0.3 (#3003)

### Features

* *(cypress)* Remove getSettled
* *(cypress)* Use cy.session
* *(i18n)* Add Norwegian translation
* *(netlify)* Abstract createSlug helper function (#2923)
* *(postcss)* Mock plugin types (#2930)
* Enable ts for rollup-plugin-visualizer (#2897) ([09d1352](09d13520b060e47be18640865befde44f59332e3))
* Remove date-fns formatISO (#2899) ([1f25386](1f25386f54f376357722e1e589d3a8bd8288a033))
* Add-task usability improvements (#2767) ([4be53b0](4be53b098ca909194aefb464a93b6dae99f4b9ab))
* Remove formatISO from list-view-gantt.spec (#2922) ([a29131e](a29131e7d4be2c83c3e9046549924d1f7692c95e))
* Add histoire ([7be8e89](7be8e892e2480f17cb5de6a69d35287906151c0f))
* Add XButton story ([ccc85b9](ccc85b9a828488dc849758f1e89f3ba3f75967d1))
* Add card story ([35cfb2f](35cfb2f3ca42ac83a9b943fc59818c978ee95fcc))
* Add histoire (#2724) ([a4424e0](a4424e089cdfadb4ab3b753e6fdca818bbe82dc4))
* Add describe project better in package.json (#2971) ([14466bf](14466bf9b7b8a3fc455c0d601205abbaf8cba4f5))
* Add .env.local.example (#2972) ([e1b35ff](e1b35ff023679a7cb8448a06e9edeb8eccc2f727))
* Fix broken font preloading (#2980) ([4890149](489014944a1544846875910d7d5e17e3d71b7e2d))

### Miscellaneous Tasks

* *(config)* Remove unused URL_PREFIX const (#2926)
* *(package)* Use pnpm commands (#2919)
* *(tests)* Fix macos cypress and align with create vite (#2898)
* Improve migrate title (#2968) ([56fd25e](56fd25e888cae8343f64a4c14ac5a3a760bdc7be))
* Add has content="false" to gantt charts (#2969) ([903e9a9](903e9a9904c18ced59962fc03b4c36e5ac8cd688))
* Use es6 imports for deploy-preview-netlify (#2970) ([2a2c27a](2a2c27af9226f441ec80d9d4f560b55cd357126c))

### Other

* *(other)* [skip ci] Updated translations via Crowdin
* *(other)* Redirect to oidc provider if configured correctly (#2805)


## [0.20.2] - 2022-12-18

### Bug Fixes

* *(bug-report.yml)* List (#2845)
* *(quick add magic)* Don't create a new label multiple times if it is used in multiple tasks
* *(task)* Pass a list specified via quick add magic down to all subtasks created via indentation
* *(task)* Move task color bubble next to task index and done badge on mobile
* *(tasks)* Remove a task from its bucket when it is in the first kanban bucket
* *(tasks)* Missing space when showing parent tasks and list title
* *(tasks)* Translation for multiple related tasks now works
* Move createdUpdated styles to component (#2685) ([4c458a1](4c458a1ad0761920868e3863982d5175664b3e6e))
* Move heading styles to component (#2686) ([293402b](293402b6fdfc699661c7f287ff1759a9ce5bea17))
* Use scss for datemathHelp (#2690) ([06775cf](06775cf4c72cf81a125b91d49c8d81e8649af661))
* Reactive const assignment (#2692) ([4c4adfd](4c4adfdf4e79eff3e101d9f0bd68bc3e5bb76495))
* Remove vuex leftover from setModuleLoading (#2716) ([3aaacf4](3aaacf4533c761864d3081edb92c9380df43f8b1))
* Icon offset and color ([74ad98d](74ad98de680f8b56e42886cd1e33874bd05772fa))
* Only load buckets if listId set (#2741) ([7db79ff](7db79ff04e4ce87d62cae7f93b67570bbc5c13be))
* Add all json files in src (#2737) ([422e731](422e731fe0d44c2e3be603b549538a05a695b95c))
* Vite.config imports (#2843) ([318e8c8](318e8c83a68bcb2f7953553c036f677a97b01c21))

### Dependencies

* *(deps)* Update dependency rollup to v3.3.0 (#2689)
* *(deps)* Update dependency @types/dompurify to v2.4.0 (#2688)
* *(deps)* Update dependency @vue/test-utils to v2.2.2 (#2696)
* *(deps)* Update dependency caniuse-lite to v1.0.30001431
* *(deps)* Update dependency happy-dom to v7.7.0
* *(deps)* Update dependency netlify-cli to v12.1.1 (#2699)
* *(deps)* Update dependency postcss-preset-env to v7.8.3 (#2701)
* *(deps)* Update dependency vitest to v0.25.2 (#2702)
* *(deps)* Update pnpm to v7.16.0 (#2703)
* *(deps)* Update typescript-eslint monorepo to v5.43.0
* *(deps)* Update dependency ufo to v1
* *(deps)* Update dependency esbuild to v0.15.14 (#2706)
* *(deps)* Update dependency @vue/test-utils to v2.2.3 (#2707)
* *(deps)* Update dependency vite to v3.2.4
* *(deps)* Update dependency typescript to v4.9.3
* *(deps)* Update dependency cypress to v11.1.0
* *(deps)* Update font awesome to v6.2.1 (#2712)
* *(deps)* Update pnpm to v7.16.1 (#2717)
* *(deps)* Update dependency pinia to v2.0.24
* *(deps)* Update sentry-javascript monorepo to v7.20.0 (#2720)
* *(deps)* Update dependency eslint to v8.28.0
* *(deps)* Update dependency esbuild to v0.15.15
* *(deps)* Update dependency netlify-cli to v12.2.4
* *(deps)* Update dependency @vue/test-utils to v2.2.4
* *(deps)* Update pnpm to v7.17.0
* *(deps)* Update dependency marked to v4.2.3
* *(deps)* Update dependency codemirror to v5.65.10
* *(deps)* Update sentry-javascript monorepo to v7.20.1
* *(deps)* Update dependency pinia to v2.0.25
* *(deps)* Update dependency rollup to v3.4.0
* *(deps)* Update typescript-eslint monorepo to v5.44.0
* *(deps)* Update vueuse to v9.6.0 (#2742)
* *(deps)* Update dependency vitest to v0.25.3 (#2743)
* *(deps)* Update dependency cypress to v11.2.0
* *(deps)* Update sentry-javascript monorepo to v7.21.0
* *(deps)* Update dependency @4tw/cypress-drag-drop to v2.2.2
* *(deps)* Update sentry-javascript monorepo to v7.21.1 (#2747)
* *(deps)* Update dependency pinia to v2.0.26
* *(deps)* Update dependency @cypress/vue to v5.0.2
* *(deps)* Update dependency highlight.js to v11.7.0 (#2752)
* *(deps)* Update dependency eslint-plugin-vue to v9.8.0 (#2753)
* *(deps)* Update dependency @infectoone/vue-ganttastic to v2.1.3
* *(deps)* Update dependency rollup to v3.5.0 (#2756)
* *(deps)* Update pnpm to v7.17.1 (#2755)
* *(deps)* Update dependency esbuild to v0.15.16
* *(deps)* Update dependency pinia to v2.0.27 (#2757)
* *(deps)* Update dependency caniuse-lite to v1.0.30001434 (#2759)
* *(deps)* Update dependency netlify-cli to v12.2.7 (#2760)
* *(deps)* Update dependency @kyvg/vue3-notification to v2.7.0 (#2761)
* *(deps)* Update typescript-eslint monorepo to v5.45.0 (#2762)
* *(deps)* Update dependency ufo to v1.0.1 (#2763)
* *(deps)* Update dependency vue-tsc to v1.0.10 (#2764)
* *(deps)* Update sentry-javascript monorepo to v7.22.0 (#2765)
* *(deps)* Update dependency @types/node to v18.11.10 (#2768)
* *(deps)* Update dependency rollup to v3.5.1 (#2769)
* *(deps)* Update sentry-javascript monorepo to v7.23.0
* *(deps)* Update dependency @vue/test-utils to v2.2.5 (#2773)
* *(deps)* Update dependency eslint to v8.29.0 (#2774)
* *(deps)* Update dependency @cypress/vue to v5.0.3 (#2775)
* *(deps)* Update dependency vue-tsc to v1.0.11 (#2777)
* *(deps)* Update dependency @cypress/vite-dev-server to v5 (#2776)
* *(deps)* Update pnpm to v7.18.0 (#2778)
* *(deps)* Update dependency esbuild to v0.15.17 (#2779)
* *(deps)* Update dependency caniuse-lite to v1.0.30001436 (#2780)
* *(deps)* Update dependency @vue/test-utils to v2.2.6 (#2784)
* *(deps)* Update dependency esbuild to v0.15.18 (#2783)
* *(deps)* Update dependency netlify-cli to v12.2.8 (#2782)
* *(deps)* Update dependency happy-dom to v7.7.2 (#2781)
* *(deps)* Update dependency vite to v3.2.5 (#2785)
* *(deps)* Update dependency rollup to v3.6.0 (#2786)
* *(deps)* Update typescript-eslint monorepo to v5.45.1 (#2787)
* *(deps)* Update dependency vitest to v0.25.4 (#2788)
* *(deps)* Update dependency @types/node to v18.11.11 (#2789)
* *(deps)* Update pnpm to v7.18.1 (#2790)
* *(deps)* Update dependency dayjs to v1.11.7 (#2791)
* *(deps)* Update dependency cypress to v12 (#2792)
* *(deps)* Update dependency vitest to v0.25.5 (#2793)
* *(deps)* Update dependency marked to v4.2.4 (#2796)
* *(deps)* Update dependency esbuild to v0.16.1 (#2795)
* *(deps)* Update dependency cypress to v12.0.1 (#2794)
* *(deps)* Update sentry-javascript monorepo to v7.24.0 (#2797)
* *(deps)* Update sentry-javascript monorepo to v7.24.1 (#2798)
* *(deps)* Update sentry-javascript monorepo to v7.24.2 (#2799)
* *(deps)* Update dependency typescript to v4.9.4 (#2800)
* *(deps)* Update dependency rollup to v3.7.0 (#2801)
* *(deps)* Update dependency esbuild to v0.16.2 (#2802)
* *(deps)* Update typescript-eslint monorepo to v5.46.0 (#2803)
* *(deps)* Update dependency vitest to v0.25.6 (#2804)
* *(deps)* Update dependency @cypress/vite-dev-server to v5.0.1 (#2806)
* *(deps)* Update dependency esbuild to v0.16.3 (#2809)
* *(deps)* Update dependency sass to v1.56.2 (#2810)
* *(deps)* Update dependency @types/marked to v4.0.8 (#2812)
* *(deps)* Update dependency vue-tsc to v1.0.12 (#2811)
* *(deps)* Update dependency @types/node to v18.11.12 (#2808)
* *(deps)* Update dependency cypress to v12.0.2 (#2807)
* *(deps)* Update dependency @vitejs/plugin-vue to v4 (#2814)
* *(deps)* Update dependency @vitejs/plugin-legacy to v3 (#2813)
* *(deps)* Update dependency pinia to v2.0.28 (#2815)
* *(deps)* Update dependency @vitejs/plugin-legacy to v3.0.1 (#2818)
* *(deps)* Update dependency @cypress/vite-dev-server to v5.0.2 (#2819)
* *(deps)* Update dependency rollup to v3.7.1 (#2820)
* *(deps)* Update dependency rollup to v3.7.2 (#2822)
* *(deps)* Update dependency esbuild to v0.16.4 (#2821)
* *(deps)* Update dependency vitest to v0.25.7 (#2824)
* *(deps)* Update dependency @types/node to v18.11.13 (#2823)
* *(deps)* Update dependency happy-dom to v8 (#2831)
* *(deps)* Update dependency postcss to v8.4.20 (#2827)
* *(deps)* Update dependency caniuse-lite to v1.0.30001439 (#2828)
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v0.8.1 (#2826)
* *(deps)* Update dependency netlify-cli to v12.2.10 (#2829)
* *(deps)* Update dependency vite-plugin-pwa to v0.14.0 (#2833)
* *(deps)* Update dependency rollup to v3.7.3 (#2825)
* *(deps)* Update dependency vue-tsc to v1.0.13 (#2832)
* *(deps)* Update sentry-javascript monorepo to v7.25.0
* *(deps)* Update dependency vite to v4 (#2816)
* *(deps)* Update pnpm to v7.18.2 (#2834)
* *(deps)* Update typescript-eslint monorepo to v5.46.1 (#2837)
* *(deps)* Update dependency @4tw/cypress-drag-drop to v2.2.3 (#2836)
* *(deps)* Update dependency @types/node to v18.11.14 (#2839)
* *(deps)* Update dependency cypress to v12.1.0 (#2838)
* *(deps)* Update dependency rollup to v3.7.4 (#2840)
* *(deps)* Update dependency vitest to v0.25.8
* *(deps)* Update sentry-javascript monorepo to v7.26.0
* *(deps)* Update dependency esbuild to v0.16.5 (#2846)
* *(deps)* Update dependency @types/node to v18.11.15
* *(deps)* Update dependency esbuild to v0.16.6 (#2848)
* *(deps)* Update dependency esbuild to v0.16.7
* *(deps)* Update sentry-javascript monorepo to v7.27.0 (#2850)
* *(deps)* Update dependency @vueuse/core to v9.7.0 (#2851)
* *(deps)* Update dependency wait-on to v7 (#2852)
* *(deps)* Update dependency @types/node to v18.11.16 (#2853)
* *(deps)* Update dependency eslint to v8.30.0
* *(deps)* Update dependency rollup to v3.7.5 (#2857)
* *(deps)* Update dependency esbuild to v0.16.8 (#2854)
* *(deps)* Update dependency sass to v1.57.0 (#2856)
* *(deps)* Update dependency vue-tsc to v1.0.14 (#2860)
* *(deps)* Update dependency esbuild to v0.16.9 (#2859)
* *(deps)* Update dependency @types/node to v18.11.17 (#2858)

### Features

* *(ci)* Use docker buildx for multiarch builds* Filters script setup ([4bad685](4bad685f39388d59fdd8ff79a1766c55f75262c2))
* Move select filters to dedicated components ([bb58dba](bb58dba8e07d683c75637ec88a378e873711eb29))
* Add vite build target esnext (#2674) ([163d936](163d9366d3061c40b5db7f3aad5c2cea01948403))
* Filters script setup (#2671) ([4a550da](4a550da6a69a50126b9d4a555b6713687347c2d3))
* Reduce multiselect selector specificity (#2678) ([9f0f0b3](9f0f0b39f8eea399b7b03003afa5893d0b8016f8))
* Reduce contentAuth selector specificity (#2677) ([12a8f7e](12a8f7ebe9fc556a7b0bc6e2d74e81d424ccfcf8))
* Reduce ListWrapper selector specificity (#2679) ([599c1ba](599c1ba4b5b0861d89755addf016e8f797b49dfe))
* Reduce dropdown-item selector specificity (#2680) ([eb4c2a4](eb4c2a4b9df93ee35404cd7143cc88b3d44f9d59))
* Reduce attachments selector specificity (#2682) ([0f1f131](0f1f131f7a2a38ee57175edfd5ed1c932225af16))
* Reduce ready selector specificity (#2683) ([9d604f7](9d604f7a3bc057bbe27ac19e73ac59736154d9b7))
* Use img for logo so that it's not part of the main bundle (#2684) ([02de481](02de481297502ad4b0b2eb2fa3e06366cce6d630))
* Improve user component (#2687) ([708ef2d](708ef2d72efbdfe6261322937b0a8f76ee19b9e4))
* Reduce TaskDetailView selector specificity ([fba402f](fba402fcd056ee397ce54f97ed4fec98845c7933))
* Move transition in own component ([631a19f](631a19fa923dba2759603e6a8b224cb4d3e1a038))
* Feature/load-views-async (#2672)
* Use transition component everywhere ([8c44ed8](8c44ed83e6530f67cc923a5e6d1a26c14575884a))
* Move transition in component (#2694) ([77ff0aa](77ff0aa256fbf388210af09d88673475386b3553))
* Disable fullscreen for EasyMDE side-by-side mode (#2710) ([98b38af](98b38af43c3acc9822f167ebca295f5aecb4908d))
* Only automatically redirect to provider if the url contains ?redirectToProvider=true and it's the only one ([3891d5b](3891d5b87634c890265477680fafaa04ff06cc3e))
* Improve loadTask logic (#2715) ([8ef3092](8ef309243db4e37d306167455987572006858cad))
* Remove edit-task from list view (#2721) ([45ec162](45ec1623d525ed31a49b6be6d609802c341fad27))
* Move useAutoHeightTextarea to composable (#2723) ([33d4efe](33d4efecc45ef8da5360fb878b7d365d1901b56c))
* More horizontal space on mobile (#2722) ([b42e4cc](b42e4cca59e338278261bc3ec613eefedde6fcce))
* Change list-content style (#91) ([4b47478](4b47478440d0af1bf24c44ea614c0f62f20723f7))
* Grid for list cards ([42e9f30](42e9f306e84120ba51d9b527c7868148730bf892))
* Move avatar class to where it is used (#2725) ([da8df8b](da8df8b667fc57798c1de7d78c1a7f88b0419d38))
* Undent and order navigation css ([66be0e6](66be0e6ac4bcf48124b33267224187b56ac9320a))
* Outdent navigation logo styles ([ff9efe7](ff9efe7889256706ac86bb1face842cd2de6f935))
* Group navigation styles further ([4fc7b9c](4fc7b9c67e2088e82760005cd530ea97cf796a4c))
* Move link color location together ([d9984b2](d9984b28f7d01da0f9d8f0afd5b6f0edf35823c2))
* Use fetch instead of axios for deploy preview (#2719) ([93d95b0](93d95b0821f39719c4a28c144ebb583c2eac754e))
* Remove useRouteQuery (#2751) ([3ee0bc3](3ee0bc345d6cd65769789ec029c50e652d80e1ca))
* Use Intl.DateTimeFormat for gantt weekdays (#2766) ([3b95824](3b95824f5834d7de50210414c56b07889db895c7))
* Add @intlify/unplugin-vue-i18n (#2772) ([b44d11c](b44d11cfc04712b9f9ec9479ba3a77a26c453532))
* Use vite preview for serve:dist:dev (#2842) ([f6c6f52](f6c6f52abe71674fa5f3951cc0ba61798758bd03))
* Use variable fonts with subsetting (#2817) ([b6a89a0](b6a89a0cde3c769e38146b05c33ff4ca4e97bca2))

### Other

* *(other)* [skip ci] Updated translations via Crowdin

## [0.20.1] - 2022-11-11

### Bug Fixes

* *(auth)* Always redirect to external openid provider if only one is enabled
* *(ci)* Cache folder name
* *(gantt)* Don't try to load list NaN when opening a task from the gantt chart
* *(kanban)* Don't allow dragging a bucket if a task input is focused
* *(quick add magic)* Don't parse labels, assignees or lists as date expressions if they are called that
* *(table)* Sort tasks by index instead of id
* *(tasks)* Show any errors happening during task load* SetModuleLoading LoadingState type ([35f4bb1](35f4bb138554d300757420261d70d1a6bf6b9cc0))
* Better kanban updateBucket types ([964aba4](964aba4824418e431955881be284e35f412e873b))
* Disable props destructure error ([d6cb965](d6cb965ea7330f80f1e3c213442a049f63cba57e))
* Missing href ([5d601ca](5d601ca4b34cd7368ff6061659617fff2836cdbc))
* Multiselect modelValue prop type ([480aa88](480aa8813ec28e1228e02ba78dd3ee3037f4928a))
* Potential issue with refs in Avatar ([3c5bfcc](3c5bfcc6f3cece0f3bd6e4f862a187c17a2c4d6c))
* CoverImageAttachmentId ([e01df4d](e01df4d36996aa281ef73ee74f3ac5316a0b8a98))
* Don't show user deletion menu entry in user settings if the server disabled it ([09b76b7](09b76b7bd476b9de653e53de579f1c533d101d4d))
* Resolve issues with vue-easymde (#2629) ([eb59ca5](eb59ca5836ae8454885827bcf28a8476600bd122))
* Remove wrong loadTask params (#2635) ([f7728e5](f7728e538408d15fcbfcd9ce02cd235447dfa6f0))
* Remove duplicate store assignment (#2644) ([38cef79](38cef79f680ddf3612376a90c69198e01283a5a0))
* Flatpickr types (#2647) ([7fbb6e8](7fbb6e8f700157238f8924ce95424d79a34b7543))
* Sort task alphabetically ([612e592](612e592da799ee6a76d32c8ebc567aeadde3ee11))
* Too much recursion error when opening a task from the gantt chart ([d47791b](d47791b95793aabf1524544494621b237479c15d))
* Lint & formatting ([c2dd18e](c2dd18edaa8ac29446845a5028d1a04c1f39fc76))
* Gantt route sync ([7ec2b6c](7ec2b6c0d28a1ae1799b1ed7a781efbf4c4542d7))
* Gantt route sync (#2664) ([9450817](94508173dcfc75d606d490a536f80e10397fb69c))

### Dependencies

* *(deps)* Update dependency vite to v3.2.1
* *(deps)* Update dependency @vue/test-utils to v2.2.1 (#2591)
* *(deps)* Update pnpm to v7.14.1 (#2593)
* *(deps)* Update dependency vue-flatpickr-component to v11
* *(deps)* Update sentry-javascript monorepo to v7.17.3
* *(deps)* Update dependency eslint-plugin-vue to v9.7.0
* *(deps)* Update dependency caniuse-lite to v1.0.30001427
* *(deps)* Update dependency blurhash to v2.0.4
* *(deps)* Update dependency vitest to v0.24.4
* *(deps)* Update dependency @types/node to v18.11.8
* *(deps)* Update dependency vite to v3.2.2
* *(deps)* Update dependency @kyvg/vue3-notification to v2.5.0
* *(deps)* Update dependency @kyvg/vue3-notification to v2.5.1
* *(deps)* Update dependency @kyvg/vue3-notification to v2.6.0 (#2612)
* *(deps)* Update typescript-eslint monorepo to v5.42.0
* *(deps)* Update dependency rollup to v3.2.4 (#2614)
* *(deps)* Update dependency @kyvg/vue3-notification to v2.6.1 (#2615)
* *(deps)* Update dependency rollup to v3.2.5 (#2618)
* *(deps)* Update dependency @cypress/vite-dev-server to v3.4.0 (#2617)
* *(deps)* Update dependency marked to v4.2.0 (#2616)
* *(deps)* Update dependency @types/node to v18.11.9 (#2619)
* *(deps)* Update dependency vitest to v0.24.5 (#2621)
* *(deps)* Update dependency @cypress/vue to v4.2.2
* *(deps)* Update dependency marked to v4.2.1 (#2625)
* *(deps)* Update pnpm to v7.14.2
* *(deps)* Update dependency esbuild to v0.15.13 (#2627)
* *(deps)* Update sentry-javascript monorepo to v7.17.4 (#2628)
* *(deps)* Pin dependency @types/codemirror to 5.60.5
* *(deps)* Update dependency vite-plugin-pwa to v0.13.2 (#2632)
* *(deps)* Update dependency sass to v1.56.0 (#2633)
* *(deps)* Update dependency marked to v4.2.2 (#2636)
* *(deps)* Update dependency eslint to v8.27.0
* *(deps)* Update dependency caniuse-lite to v1.0.30001430 (#2639)
* *(deps)* Update dependency netlify-cli to v12.1.0 (#2640)
* *(deps)* Update dependency vite to v3.2.3
* *(deps)* Update dependency @vitejs/plugin-legacy to v2.3.1 (#2641)
* *(deps)* Update dependency vite-plugin-pwa to v0.13.3 (#2648)
* *(deps)* Update dependency @cypress/vite-dev-server to v4 (#2651)
* *(deps)* Update dependency vitest to v0.25.0 (#2650)
* *(deps)* Update dependency @cypress/vue to v5 (#2652)
* *(deps)* Update typescript-eslint monorepo to v5.42.1 (#2653)
* *(deps)* Update dependency @cypress/vue to v5.0.1 (#2655)
* *(deps)* Update sentry-javascript monorepo to v7.18.0
* *(deps)* Update dependency vitest to v0.25.1 (#2657)
* *(deps)* Update dependency @cypress/vite-dev-server to v4.0.1 (#2658)
* *(deps)* Update vueuse to v9.5.0 (#2660)
* *(deps)* Update dependency sass to v1.56.1 (#2661)
* *(deps)* Update dependency vue to v3.2.42
* *(deps)* Update dependency @fortawesome/vue-fontawesome to v3.0.2
* *(deps)* Update dependency vue to v3.2.43 (#2663)
* *(deps)* Update dependency vue to v3.2.44 (#2666)
* *(deps)* Update pnpm to v7.15.0 (#2667)
* *(deps)* Update dependency cypress to v11 (#2659)
* *(deps)* Update dependency dompurify to v2.4.1 (#2669)

### Features

* *(ci)* Use 'always' for pull
* *(ci)* Add kind everywhere
* *(ci)* Update cypress image
* *(ci)* Improve drone config (#2637)
* *(tests)* Add tests for gantt chart time range
* *(tests)* Add tests for gantt chart task detail open* Task store with composition api (#2610) ([839d331](839d331bf51f9a0e9742b9972dbd6a88fa38f1c3))
* Auth store with composition api (#2602) ([825ba10](825ba100f0c05e1ab98d401157c30aad8658afa6))
* Config store with composition api (#2604) ([15ef86d](15ef86d597ceb8731febf789f1b812a339273e40))
* Base store with composition api (#2601) ([b4f4fd4](b4f4fd45a4c98629de182033e808cf7b22a1fe4a))
* Attachments store with composition api (#2603) ([a50eca8](a50eca852fcb841166baa07a6cc405eeb70c6e9d))
* Namespaces store with composition api (#2607) ([0832184](08321842220798b478ffaef7e9e11c527cb5b3bd))
* Lists store with composition api (#2606) ([5ae8bac](5ae8bace820b05d3ad05f40ab51164ec2c35c068))
* Label store with composition api (#2605) ([1002579](1002579173bd4b89e157c78ac607abd7969d85bc))
* Type improvements ([599e28e](599e28e5e5d56e4ced338ec1c79fea7d4576b85a))
* Type global components and especially icon prop ([a2c1702](a2c1702eef64dd779c86940898bd49fc2c96233f))
* Rework BaseButton ([e8c6afc](e8c6afce7298267f2f77ece0a746218c2eb3f7b7))
* Rework XButton ([4cd0e90](4cd0e90feaab05a2275e92affda23dde7453013f))
* Rework dropdown-item ([02deb0b](02deb0beddbc9221bdcafd0d09cee383571dae55))
* Rework popup ([0b58973](0b58973d872d8d54c9a829a06c8535a7a7115613))
* SingleTaskInList script setup (#2463) ([44e6981](44e6981759261cdada6388384cbad96e5401b8a9))
* Add type info ([0182695](0182695cda1252a65df3f48fdc316e82cd7fadbd))
* Rename http-common to fetcher (#2620) ([096daad](096daad80a9c089e732116ce3b8aa4310a611368))
* Improved types (#2547) ([0ff0d8c](0ff0d8c5b89bd6a8b628ddbe6074f61797b6b9c1))
* MigrateService script setup (#2432) ([8b7b4d6](8b7b4d61a3b9dd01ab58b7e7dd30bf649b62fcf6))
* Sticky action buttons (#2622) ([f4bc2b9](f4bc2b94f0466a357361a69cfb3562e84d1ea439))
* Simplify editAssignees (#2646) ([d9a8382](d9a83820495f34ddbd776f70cabdc24bbb1c3f32))
* Remove comments from prioritySelect (#2645) ([6a93701](6a93701649d35622d13dda969aae4aedf145d4d0))
* ListKanban script setup (#2643) ([d85abbd](d85abbd77a8197e977fdbfec0ee309736cce05fa))
* Kanban store with composition api ([f0492d4](f0492d49ef5cd99d95085deec066cec85f4688b3))

### Miscellaneous Tasks

* *(ci)* Sign drone config* Remove comment ([1101fcb](1101fcb3fff1fce102a7418b1e2734a71cdf84e2))
* Improve multiselect hover types ([caa29c1](caa29c152d35b28658773b838de0a8909d0e509f))
* Remove unused processModel in services (#2624) ([7f00c7d](7f00c7dabd1e55ec0e9a86ca495f702a38ddb18d))
* Inline simple helper (#2631) ([e49f960](e49f960aea2ead5baca6965649821db6584cbac2))
* Move run.sh in scripts folder (#2649) ([5057b69](5057b69382ca65659b624206b381d8f1500bae82))

### Other

* *(other)* [skip ci] Updated translations via Crowdin

## [0.20.0] - 2022-10-28

### Bug Fixes

* *(filters)* Changing filter checkbox values not being emitted to parent components
* *(filters)* Make sure all checkboxes are aligned properly
* *(filters)* Page freezing when entering a date as a result of an endless loop
* *(gantt)* Only unmount chart if there aren't any loaded tasks yet
* *(gantt)* UseDayjsLanguageSync and move to separate file
* *(i18n)* Spelling typo
* *(i18n)* Rename "right" to permission so that it's clearer what it is used for
* *(labels)* Unset loading state after loading all labels
* *(lint)* Unnecessary catch clause
* *(list)* Automatically close task edit pane when switching between lists
* *(quick add magic)* Time parsing for certain conditions (#2367)
* *(sharing)* Correctly check if the user has admin rights when sharing
* *(subscription)* Don't remove every namespace but the one subscribing to
* *(subscription)* Make sure list subscription state is propagated everywhere for the current list
* *(task)* Make sure users can be assigned via quick add magic via their real name as well
* *(task)* Cancel loading state when creating a new task does not work
* *(task)* Cancel loading state when creating a new task does not work
* *(task)* New tasks with quick add magic not showing up in task list
* *(task)* Setting a priority was not properly saved
* *(task)* Setting progress was not properly saved
* *(task)* Setting a label would not show up on the kanban board after setting it
* *(task)* Stop loading when no list was specified while creating a task
* *(task)* Only show create list or import cta when there are no tasks
* *(task)* Marking checklist items as done
* *(task)* Focusing on assignee search field when activating it
* *(task)* Scroll the task field into view after activating it
* *(tasks)* Don't allow adding the same assignee multiple times
* *(teams)* Show an error message when no user is selected to add to a team
* *(tests)* Fake current time in gantt tests to make them more reliable
* *(tests)* Adjust gantt rows identifier* Authenticate per request (#2258) ([6e4a3ff](6e4a3ff1996f55d99896a0e8267c1915de09dd39))
* Add lodash.clonedeep types ([80eaf38](80eaf38090413b74524ddc5a7dfcc9a845a6ba26))
* Use correct model for generics ([3ba423e](3ba423ed238a5f8f445246793829c7645dfe42aa))
* Merge duplicate types ([106abfc](106abfc842ca0c916ef7574b0fe5c89940869ac2))
* CreateNewTask typing ([f9b5130](f9b51306c396ceb0d8fa0c4af3fea24d2b28b64b))
* Improve some types ([4a50e6a](4a50e6aae28d22c3d441f1fead4edce7d0e30ff1))
* Use definite assignment assertion operator ([96f5f00](96f5f00c073f71c71d85c351f86ad16a67db6992))
* Mark abstractModel and abstractService abstract ([d36577c](d36577c04e1eea00fb21a5fb774e7f6b1f667d54))
* Use IAbstract to extend model interface ([8be1f81](8be1f81848303d590adb890743dd688fbf5cdf1c))
* Use new assignData method for default data ([8416b1f](8416b1f44811ff477d81db20370ff68e899c7252))
* Don't push a select event when nothing was selected ([9616bad](9616badc33173483e0b5cc0c99655e0c9a4907f9))
* Don't try to set the bucket of a task when it was moved to a new list ([c06b781](c06b781837c66174be41f40c967fbfcbcc35495e))
* Mutation error in TaskDetailView ([b4cba6f](b4cba6f7d96334b46e5e2d6be5ac87432b01f0c0))
* DefaultListId ([878b5bf](878b5bf236f7d1ddc9825d8dca8415313b08fd94))
* Use typed useStore ([54de368](54de368642519fc900ce89e4ee38989555054a05))
* Don't encode attachment upload file blob as json ([d819b9b](d819b9b0ba08db24a77751061ae285fc11205c2c))
* Dragging a list on mobile Safari ([6bf5f6e](6bf5f6efd46c47293fb54b9e9a25d91d8c6bec0d))
* Introduce a ListView type to properly type all available list views ([23598dd](23598dd2ee649449f2176ae86acbc16ecbf01e6f))
* Use proper computed for available views list ([e67fc7f](e67fc7fb7e1678b1b691fee77d3237b222ad50c6))
* Only warn once if triggeredNotifications are not supported (#2344) ([f083f18](f083f181e2c8aa0af3ac1381303f61792d5975f5))
* Bucket title edit success message appearing twice ([4921788](49217889b50da73d0f4851c4ee21f0dec11c7958))
* Don't parse dates in urls ([92f24e5](92f24e59a794a25098f5fb50f2101d516829cd36))
* Vue-i18n global scope (#2366) ([602ab83](602ab8379e3fb11eb8b547d036921311f193fb12))
* Redirect to login when the jwt token expires ([91976e2](91976e23f989f39fb25d3341aa3f4b632ea66f35))
* Only try to save user settings when a user is authenticated ([2df2bd3](2df2bd38e2b9f86be7e7c5aab744f27cbf2644c3))
* Remove margin from the color bubble component itself ([4fce71f](4fce71f729878d47c3ec79d0c10fae8fbaabbd91))
* Test pnpm cache ([e5d04c9](e5d04c98dabc6b597ecc32dd01ab31c4dd9882d1))
* Remove console.log ([43e2d03](43e2d036d77731fcce18cbea1d82196b10016609))
* Explicitly install cypress ([62e227c](62e227c767a43578f4487e3dc244f4756e073f5d))
* Only pass date to flatpickr if it's a valid date ([ede5cdd](ede5cdd8cf5575bba96d3e7b6824a7ad7b414ea7))
* Loading state when creating a new task from list view ([aa64e98](aa64e9835c6b9ef2bb10ab8d2a1b4a695cb4321b))
* Make add task button 100% height ([3c9c5ef](3c9c5eff1258b6e04e3d0e9299110fa9b5c9757d))
* Lint ([2bf9771](2bf9771e2894acb7ad3e563b7b31442d91c49e1a))
* Color list titles so that they are visible on cards with a background ([62ed7c5](62ed7c5964f1252f09fe432c42aaf327da5a8c4f))
* Missed porting these getters and commits ([95ad245](95ad245b59b0c6398b0bca217572ca36f6ea5a54))
* Use https for api url (#2425) ([9f39365](9f3936544d5906f0031412139b53c286023c2405))
* Don't use corepack prepare at all ([a199fc7](a199fc7a8e7f621ee96b2079e9558987f1350493))
* Add types for node ([6a82807](6a828078a398ab920f0e29d0801b918ae092ef30))
* VueI18n global scope fallback warnings (#2437) ([e9cf562](e9cf562969e42cc3ce3ffba3ed093db7a2089395))
* Fix missed conversion to ref (#2445) ([94d6f38](94d6f38e89174f879be4e5b1897b52603b40a745))
* Don't emit a possible null task ([5f5ed41](5f5ed410df1a2fe73e821d7dee7ebd4c0b918069))
* Docker build ([5b60693](5b606936c3f7b0dc1232ad269f3666f8170c6e11))
* Update top header list title when saving a filter ([fd3c15d](fd3c15d0642a8d91260ba24eaae52e0ba62c2871))
* Type of config stores maxFileSize (#2455) ([78a6d38](78a6d38641c5e4e68f117e37ee36a4ca3b40a24b))
* Don't add class method to interface ([367ad1e](367ad1e5a5972ac6ff353275b31f309ebcf5cb4c))
* Attachment deletion (#2472) ([f1852f1](f1852f1f33401576ae5033f54613c96cd80e0f95))
* Add lodash.debounce types (#2487) ([00e0a23](00e0a23d48c19c440aea7857c8b162a0dfa34361))
* Initial modal scroll lock (#2489) ([eae7cc5](eae7cc5a6b506cbbbe694b831cba7c5d1febaf05))
* Unset cover image when the task does not have one ([054d70c](054d70cbe5344e39d0e5f277a7db2f26573e1efa))
* Lint ([43258ab](43258ab74e0733e91be3ade1f0b13dcf9342cc18))
* Lint ([84a1abf](84a1abf3477abbbee136979bd0bde08ae6c54ceb))
* Don't try to render auth routes when the user is not authenticated ([3af20b6](3af20b6220d8fcded9c8c2f0bdef21dc26d748f6))
* Lint ([f405b21](f405b2105bf4d1cfd4f6acf03210b37ac91eff5e))
* Make sure subscriptions are properly inherited between lists and namespaces ([a895bde](a895bde6612e7a2b22a84b6ca7c583bafc9ebc9e))
* Make sure subscription strings work consistently across languages ([172d353](172d353df7a86baa9c2759907c7f855679138cc0))
* Make sure subscriptions are properly inherited between namespaces and lists ([0a29197](0a29197715f22602faf353fb8fe850150aa710d1))
* Lint ([c6d6da3](c6d6da31712906f094a88dbfdb5e9b6db66c29e3))
* Move hourToDaytime to separate file in order to pass tests ([5afafb7](5afafb7c82837a3af58c7bdc18174a785691b885))
* Postcss-preset-env configuration (#2554) ([b80f82c](b80f82c4118bb372263130df80d15a2a79d2191e))
* Password reset ([7357530](73575302debbe095ce031e4871fb3797a801db18))
* Email confirmation ([e6f7ddc](e6f7ddc9ce90ddcb3b58b2c001320b6b2c3ac169))
* Lint ([643a5b6](643a5b6d7d00bfab4b338582c85217dffa7d9b22))
* Make sure services without a modelFactory override still return data ([8fdd3e7](8fdd3e785d3c55281b557827860d0532b94ac758))
* Make sure share modals don't have a create button ([ae27502](ae27502022469882656459b0a9e7e8a4b6972c58))
* Redirect with query parameters ([f61723d](f61723dac251c9d85102beae73c6a03df10bd4bf))
* Task detail view top spacing on mobile ([a695719](a6957191284a8da38e56b4ed3fe0a57b69d6e2b9))
* Make sure the filter button is always shown on the kanban board ([8023006](80230069c6f09ced484cd356b816df6b1dd296d6))
* Wait until everything is loaded before replacing the current view with the last or login view ([6083301](6083301d1f410ede5fe62127e484169d74ff6dc0))
* Show frontend version in about dialog ([5ddce38](5ddce387fe589c574adf0cce438732faf4ad9fd1))
* Building version into releases ([a0795db](a0795db0408b5fece13d8a74e9e243375883ca6f))
* Lint ([e13e477](e13e477682ef9fd647925f459d8d4527d3c55b9b))
* New task input styling ([c3cae78](c3cae78213b791c9e6fd8143ee59e3ca256c374a))
* Handle bar styling so they can actually be used ([10c6db3](10c6db3849e734d0508c8d435164a0f771175740))
* Make sure the date format is actually valid ([2c012e1](2c012e1a080bd9519384d65ee0653483aa52d1c3))
* Make tests work again with new selectors ([091beec](091beecc19cf5ff49fc252c4eeb98aa8a65ddb67))
* Use inherit for font family ([b7b4530](b7b4530a111d93e81fc6398dc3f7267cc6e255fb))
* Remove precision setting ([970a04d](970a04d9733f4cbdc35e5b772ce4a34fa71e6c4c))
* Fix imports ([d91bc50](d91bc5090a6cec38e655c944df7cf57ac16e4133))
* Use base store ([f5fd141](f5fd14124fa139f3e76f7a4915b2efc85de6c789))
* Correctly import all components ([31f2065](31f2065d2005b27ff8a0abbc4efaa7138cfe27c1))
* Update eslint env to 2022 ([0b194bb](0b194bb0cf326104c249c953194997a1f9a80dbf))
* Don't try to dynamically load dayjs locales ([b8e7b87](b8e7b87f96bdccf19066ce31d40cf40379014bbe))
* Disable dayjsLanguageSync function ([e1f49f2](e1f49f2ff15286ee8903c29dbe708cda90e5d70d))
* Scope ListGantt styles ([73eab6c](73eab6c5b5bfe0d72393ab378cce77ad5cbb59b6))
* Initial transformation of ganttBars ([407f5f2](407f5f2ef8c4759ea46f5fb74717bafb16f606c5))
* ParseBooleanProp ([8dea408](8dea4082bb0766297f74acef0352f8a6a6168d3c))
* Do not change language to the current one ([abc2649](abc26496cf0e20d0124af327d47e086b39e2bd23))
* Remove IE fallback ([b4f88bd](b4f88bd4a6ba50be1f972794c3e87b7a09f7c2ca))
* Improve return type ([0665538](066553838ad289d6c6c0a8b1c6ed0b84139ace54))
* Improve notifications (#2583) ([9ded3d0](9ded3d0cd69dd974ffea2531e3ca92438e420f29))
* Lint ([9894337](98943377b8344f1f5a8e38c23eff79d7678f51bc))
* Label multiselect styling on focus ([da2a7a2](da2a7a224e3c8015939e189692813bc215dbd72c))


### Dependencies

* *(deps)* Update sentry-javascript monorepo to v7.11.0 (#2274)
* *(deps)* Update sentry-javascript monorepo to v7.11.1 (#2275)
* *(deps)* Update dependency vitest to v0.22.1 (#2276)
* *(deps)* Update dependency sass to v1.54.8 (#2281)
* *(deps)* Update dependency caniuse-lite to v1.0.30001387 (#2285)
* *(deps)* Update dependency rollup to v2.79.0 (#2278)
* *(deps)* Update dependency marked to v4.1.0 (#2284)
* *(deps)* Update dependency netlify-cli to v11 (#2287)
* *(deps)* Update dependency vite to v3.0.9 (#2279)
* *(deps)* Update dependency date-fns to v2.29.2 (#2277)
* *(deps)* Update dependency esbuild to v0.15.6 (#2290)
* *(deps)* Update dependency vite-plugin-pwa to v0.12.4 (#2291)
* *(deps)* Update dependency rollup-plugin-visualizer to v5.8.0 (#2282)
* *(deps)* Update dependency easymde to v2.17.0 (#2283)
* *(deps)* Update dependency vue-tsc to v0.40.5 (#2292)
* *(deps)* Update dependency vue to v3.2.38 (#2293)
* *(deps)* Update dependency vue-router to v4.1.5 (#2294)
* *(deps)* Update vueuse to v9.1.1 (#2295)
* *(deps)* Update dependency @cypress/vue to v4.2.0 (#2296)
* *(deps)* Update dependency @faker-js/faker to v7.5.0 (#2297)
* *(deps)* Update dependency eslint to v8.23.0 (#2299)
* *(deps)* Update dependency cypress to v10.7.0 (#2298)
* *(deps)* Update dependency eslint-plugin-vue to v9.4.0 (#2300)
* *(deps)* Update sentry-javascript monorepo to v7.12.0 (#2307)
* *(deps)* Update dependency dompurify to v2.4.0 (#2306)
* *(deps)* Update typescript-eslint monorepo to v5.36.1 (#2304)
* *(deps)* Update dependency vite-svg-loader to v3.5.1 (#2302)
* *(deps)* Update dependency typescript to v4.8.2 (#2301)
* *(deps)* Update font awesome to v6.2.0 (#2303)
* *(deps)* Update dependency @kyvg/vue3-notification to v2.4.1 (#2305)
* *(deps)* Update sentry-javascript monorepo to v7.12.1 (#2308)
* *(deps)* Update dependency vite-plugin-pwa to v0.12.6 (#2309)
* *(deps)* Update dependency vue-tsc to v0.40.6 (#2310)
* *(deps)* Update dependency rollup-plugin-visualizer to v5.8.1 (#2311)
* *(deps)* Update dependency vitest to v0.23.0 (#2312)
* *(deps)* Update dependency esbuild to v0.15.7 (#2313)
* *(deps)* Update dependency caniuse-lite to v1.0.30001390 (#2314)
* *(deps)* Update dependency vue-tsc to v0.40.7 (#2315)
* *(deps)* Update dependency vitest to v0.23.1 (#2316)
* *(deps)* Update dependency @vitejs/plugin-legacy to v2.1.0 (#2317)
* *(deps)* Update dependency @vitejs/plugin-vue to v3.1.0 (#2318)
* *(deps)* Update dependency vite to v3.1.0 (#2319)
* *(deps)* Update vueuse to v9.2.0 (#2320)
* *(deps)* Update typescript-eslint monorepo to v5.36.2 (#2321)
* *(deps)* Update dependency vue-tsc to v0.40.9 (#2322)
* *(deps)* Pin dependency @types/lodash.clonedeep to 4.5.7 (#2323)
* *(deps)* Update dependency @vue/eslint-config-typescript to v11.0.1 (#2324)
* *(deps)* Update dependency vite-plugin-pwa to v0.12.7 (#2325)
* *(deps)* Update dependency vue-tsc to v0.40.10 (#2326)
* *(deps)* Update dependency postcss-preset-env to v7.8.1 (#2328)
* *(deps)* Update dependency vite-svg-loader to v3.6.0 (#2327)
* *(deps)* Update dependency vue-tsc to v0.40.11 (#2333)
* *(deps)* Update dependency sass to v1.54.9 (#2336)
* *(deps)* Update dependency vue-tsc to v0.40.13
* *(deps)* Update dependency vue to v3.2.39
* *(deps)* Update dependency typescript to v4.8.3 (#2341)
* *(deps)* Update dependency vitest to v0.23.2
* *(deps)* Update dependency autoprefixer to v10.4.9
* *(deps)* Update dependency caniuse-lite to v1.0.30001397
* *(deps)* Update dependency netlify-cli to v11.7.1
* *(deps)* Update dependency eslint to v8.23.1
* *(deps)* Update typescript-eslint monorepo to v5.37.0
* *(deps)* Update dependency blurhash to v2 (#2351)
* *(deps)* Update dependency date-fns to v2.29.3 (#2354)
* *(deps)* Update dependency autoprefixer to v10.4.10 (#2355)
* *(deps)* Update dependency cypress to v10.8.0 (#2359)
* *(deps)* Update dependency autoprefixer to v10.4.11 (#2363)
* *(deps)* Update dependency postcss-preset-env to v7.8.2
* *(deps)* Update dependency vite to v3.1.1 (#2365)
* *(deps)* Pin dependency @types/dompurify to 2.3.4
* *(deps)* Update sentry-javascript monorepo to v7.13.0
* *(deps)* Update dependency eslint-plugin-vue to v9.5.0 (#2371)
* *(deps)* Update dependency eslint-plugin-vue to v9.5.1 (#2373)
* *(deps)* Update dependency vite to v3.1.2
* *(deps)* Update dependency @types/sortablejs to v1.15.0
* *(deps)* Update dependency vitest to v0.23.4
* *(deps)* Update dependency esbuild to v0.15.8
* *(deps)* Update dependency vite-plugin-pwa to v0.12.8 (#2375)
* *(deps)* Update caniuse-and-related to v4.21.4 (#2379)
* *(deps)* Update dependency netlify-cli to v11.8.0 (#2380)
* *(deps)* Update dependency @vitejs/plugin-legacy to v2.2.0 (#2381)
* *(deps)* Update dependency vite to v3.1.3 (#2382)
* *(deps)* Update typescript-eslint monorepo to v5.38.0 (#2383)
* *(deps)* Update dependency vite-plugin-pwa to v0.13.0 (#2385)
* *(deps)* Update dependency easymde to v2.18.0 (#2386)
* *(deps)* Update dependency autoprefixer to v10.4.12
* *(deps)* Update dependency pinia to v2.0.22 (#2400)
* *(deps)* Update dependency @vue/eslint-config-typescript to v11.0.2
* *(deps)* Update dependency vite-plugin-pwa to v0.13.1
* *(deps)* Update dependency rollup to v2.79.1
* *(deps)* Update dependency codemirror to v5.65.9
* *(deps)* Update pnpm to v7.12.1
* *(deps)* Update dependency sass to v1.55.0
* *(deps)* Update dependency esbuild to v0.15.9
* *(deps)* Update pnpm to v7.12.2 (#2408)
* *(deps)* Update dependency caniuse-lite to v1.0.30001412 (#2421)
* *(deps)* Update dependency netlify-cli to v11.8.3 (#2422)
* *(deps)* Update dependency eslint to v8.24.0 (#2410)
* *(deps)* Update vueuse to v9.3.0 (#2423)
* *(deps)* Update dependency rollup-plugin-visualizer to v5.8.2 (#2420)
* *(deps)* Update typescript-eslint monorepo to v5.38.1 (#2426)
* *(deps)* Update dependency blurhash to v2.0.1
* *(deps)* Update dependency cypress to v10.9.0 (#2429)
* *(deps)* Update dependency @types/node to v16.11.62 (#2430)
* *(deps)* Update dependency typescript to v4.8.4
* *(deps)* Update dependency vue to v3.2.40
* *(deps)* Update dependency blurhash to v2.0.2
* *(deps)* Update sentry-javascript monorepo to v7.14.0 (#2440)
* *(deps)* Update dependency vite to v3.1.4 (#2439)
* *(deps)* Update dependency @vue/test-utils to v2.1.0
* *(deps)* Update dependency esbuild to v0.15.10
* *(deps)* Update dependency @cypress/vite-dev-server to v3.2.0 (#2448)
* *(deps)* Update dependency postcss to v8.4.17 (#2449)
* *(deps)* Update dependency marked to v4.1.1
* *(deps)* Update dependency @vitejs/plugin-vue to v3.1.2 (#2461)
* *(deps)* Update dependency @types/node to v16.11.63 (#2464)
* *(deps)* Update dependency caniuse-lite to v1.0.30001414 (#2465)
* *(deps)* Update pnpm to v7.13.0 (#2467)
* *(deps)* Update dependency netlify-cli to v12 (#2466)
* *(deps)* Update dependency vue-advanced-cropper to v2.8.5 (#2469)
* *(deps)* Update dependency blurhash to v2.0.3 (#2468)
* *(deps)* Update sentry-javascript monorepo to v7.14.1 (#2471)
* *(deps)* Update typescript-eslint monorepo to v5.39.0
* *(deps)* Update dependency @types/node to v16.11.64 (#2479)
* *(deps)* Update dependency eslint-plugin-vue to v9.6.0 (#2480)
* *(deps)* Update pnpm to v7.13.1
* *(deps)* Update dependency vue-advanced-cropper to v2.8.6 (#2483)
* *(deps)* Pin dependency @rushstack/eslint-patch to 1.2.0 (#2486)
* *(deps)* Pin dependency @types/lodash.debounce to 4.0.7 (#2488)
* *(deps)* Update dependency happy-dom to v7 (#2492)
* *(deps)* Update dependency vite to v3.1.5
* *(deps)* Update dependency happy-dom to v7.0.2
* *(deps)* Update sentry-javascript monorepo to v7.14.2
* *(deps)* Update pnpm to v7.13.2
* *(deps)* Update dependency vue-flatpickr-component to v9.0.8 (#2494)
* *(deps)* Update dependency vite to v3.1.6
* *(deps)* Update dependency happy-dom to v7.0.4 (#2499)
* *(deps)* Update dependency @cypress/vite-dev-server to v3.3.0 (#2501)
* *(deps)* Update dependency happy-dom to v7.0.6 (#2500)
* *(deps)* Update dependency happy-dom to v7.3.0 (#2502)
* *(deps)* Update dependency vitest to v0.24.0 (#2503)
* *(deps)* Update dependency vue-tsc to v1 (#2504)
* *(deps)* Update dependency happy-dom to v7.4.0 (#2505)
* *(deps)* Update dependency eslint to v8.25.0
* *(deps)* Update dependency vue-tsc to v1.0.1 (#2507)
* *(deps)* Update dependency pinia to v2.0.23 (#2509)
* *(deps)* Update dependency express to v4.18.2
* *(deps)* Update pnpm to v7.13.3 (#2511)
* *(deps)* Update dependency vue-tsc to v1.0.2 (#2510)
* *(deps)* Update dependency vue-tsc to v1.0.3 (#2512)
* *(deps)* Update dependency netlify-cli to v12.0.7 (#2514)
* *(deps)* Update dependency caniuse-lite to v1.0.30001418 (#2513)
* *(deps)* Update dependency vite to v3.1.7 (#2515)
* *(deps)* Update sentry-javascript monorepo to v7.15.0 (#2516)
* *(deps)* Update dependency vitest to v0.24.1 (#2517)
* *(deps)* Update pnpm to v7.13.4 (#2518)
* *(deps)* Update typescript-eslint monorepo to v5.40.0 (#2519)
* *(deps)* Update dependency @types/node to v16.11.65 (#2520)
* *(deps)* Update dependency minimist to v1.2.7 (#2521)
* *(deps)* Update dependency rollup to v3 (#2524)
* *(deps)* Update dependency @cypress/vite-dev-server to v3.3.1 (#2523)
* *(deps)* Update dependency cypress to v10.10.0 (#2525)
* *(deps)* Update dependency vue-tsc to v1.0.4 (#2526)
* *(deps)* Update dependency vue-tsc to v1.0.5 (#2527)
* *(deps)* Update dependency rollup to v3.1.0 (#2528)
* *(deps)* Update dependency @faker-js/faker to v7.6.0 (#2530)
* *(deps)* Update dependency vue-tsc to v1.0.6 (#2529)
* *(deps)* Update dependency postcss to v8.4.18 (#2532)
* *(deps)* Update dependency vue-tsc to v1.0.7 (#2533)
* *(deps)* Update dependency vite to v3.1.8 (#2534)
* *(deps)* Update dependency vue to v3.2.41 (#2538)
* *(deps)* Update dependency vitest to v0.24.3 (#2536)
* *(deps)* Update dependency @cypress/vue to v4.2.1 (#2535)
* *(deps)* Update dependency esbuild to v0.15.11 (#2539)
* *(deps)* Update dependency rollup to v3.2.0 (#2541)
* *(deps)* Update dependency vue-tsc to v1.0.8 (#2540)
* *(deps)* Update dependency rollup to v3.2.1 (#2545)
* *(deps)* Update dependency @types/node to v16.11.66 (#2544)
* *(deps)* Update dependency ufo to v0.8.6 (#2542)
* *(deps)* Update dependency rollup-plugin-visualizer to v5.8.3 (#2543)
* *(deps)* Update pnpm to v7.13.5
* *(deps)* Update dependency rollup to v3.2.2 (#2549)
* *(deps)* Update dependency netlify-cli to v12.0.9 (#2551)
* *(deps)* Update vueuse to v9.3.1 (#2552)
* *(deps)* Update dependency caniuse-lite to v1.0.30001420 (#2550)
* *(deps)* Update dependency happy-dom to v7.5.12 (#2553)
* *(deps)* Pin dependency @types/postcss-preset-env to 7.7.0 (#2555)
* *(deps)* Update dependency rollup to v3.2.3 (#2556)
* *(deps)* Update typescript-eslint monorepo to v5.40.1 (#2557)
* *(deps)* Update dependency @types/node to v16.11.68 (#2558)
* *(deps)* Update sentry-javascript monorepo to v7.16.0 (#2560)
* *(deps)* Update dependency esbuild to v0.15.12 (#2561)
* *(deps)* Update pnpm to v7.13.6 (#2562)
* *(deps)* Update dependency vue-flatpickr-component to v10 (#2563)
* *(deps)* Update dependency eslint to v8.26.0 (#2564)
* *(deps)* Update pnpm to v7.14.0 (#2565)
* *(deps)* Update dependency vue-tsc to v1.0.9 (#2566)
* *(deps)* Update dependency @types/node to v16.18.0 (#2567)
* *(deps)* Update dependency happy-dom to v7.6.0 (#2571)
* *(deps)* Update dependency @vue/test-utils to v2.2.0 (#2570)
* *(deps)* Update dependency caniuse-lite to v1.0.30001423 (#2568)
* *(deps)* Update dependency netlify-cli to v12.0.11 (#2569)
* *(deps)* Update dependency vue-router to v4.1.6 (#2572)
* *(deps)* Update typescript-eslint monorepo to v5.41.0 (#2573)
* *(deps)* Update dependency @types/node to v18 (#2574)
* *(deps)* Update vueuse to v9.4.0 (#2575)
* *(deps)* Update dependency cypress to v10.11.0 (#2576)
* *(deps)* Update dependency @types/node to v18.11.6
* *(deps)* Update dependency vite to v3.2.0 (#2580)
* *(deps)* Update dependency @types/node to v18.11.7 (#2581)
* *(deps)* Update dependency @vitejs/plugin-legacy to v2.3.0 (#2578)
* *(deps)* Update dependency @vitejs/plugin-vue to v3.2.0 (#2579)
* *(deps)* Update sentry-javascript monorepo to v7.17.0
* *(deps)* Update sentry-javascript monorepo to v7.17.1 (#2585)
* *(deps)* Update dependency autoprefixer to v10.4.13 (#2586)

### Features

* *(gantt)* Trying to load gantt-chart
* *(gantt)* Add task collection to useGanttFilter
* *(gantt)* Use time constants
* *(gantt)* Reset gantt filter
* *(gantt)* Disable useDayjsLanguageSync
* *(link shares)* Hide the logo if a query parameter was passed
* *(link shares)* Allows switching the initial view by passing a query parameter
* *(link shares)* Cleanup link share table
* *(link shares)* Allows switching the initial view by passing a query parameter (#2335)
* *(list)* Add info dialog to show list description (#2368)
* *(openid)* Show error message from query after being redirected from third party
* *(task)* Cover image for tasks (#2460)
* *(tests)* Add tests for task attachments* Settings background script setup (#2104) ([ff65580](ff655808b3cb562bd1c843ff70bf3641718ae61d))
* List settings edit script setup (#1988) ([f6437c8](f6437c81da73b7e3406c28b9bd7b201e376f15c3))
* Convert abstractService to ts ([74ad6e6](74ad6e65e88d6aa5702686dd0b6f55e2dc6b7b77))
* Add properties to models ([797de0c](797de0c5432face3887f4d77bcb7dd7ee2e7e0c1))
* Constants ([8fb0065](8fb00653e47c6f41a0e461c944b401d58b4a2351))
* Function attribute typing ([332acf0](332acf012c423d3201ec1811093226447cd065e8))
* Improve types ([c9e85cb](c9e85cb52b562cf9dcfac3ed54d8289e2b499992))
* Improve store and model typing ([3766b5e](3766b5e51ba9c40a6affa91ce5cc11519e2da5c3))
* Use lib ESNext setting for typescript ([79e7e4a](79e7e4a8aefe9f4d00bcbad76c4206c409384b61))
* Extend mode interface from class instead from interface ([a6b96f8](a6b96f857d949874ba75f657b887a7c997aa7c57))
* Improve store typing ([2444784](244478400ad8b8243ae2b29d741c03fa2b83601b))
* Add modelTypes ([7d4ba62](7d4ba6249e300b6711369476f5d6a84728668b0f))
* Convert services and models to ts (#1798) ([dbea1f7](dbea1f7a51f3cf5173b5f381944c4ef19ef97ec8))
* Add sponsor logo to readme (realm) ([e959043](e95904351fbd30776306225f3be55978d70ae42e))
* Show user display name when searching for assignees on a list ([65fd2f1](65fd2f14a067ea9d79b352af00f3c316be883fdf))
* Add keyboard shortcut to toggle task description edit (#2332) ([7f6f896](7f6f8963e7db236f3beb9e6a36fab4ba479b969b))
* Programmatically generate list of available views ([26d02d5](26d02d5593283c3ad2fb961348ba2f412cc9eaa8))
* Add fallback for useCopyToClipboard (#2343) ([7b398f7](7b398f73f604d6564a41c3ce5031883c677f02c7))
* Improve models ([1a11b43](1a11b43ca8d51bf998019fbc741e845b07d70157))
* Use v-model more consequent (#2356) ([db8b881](db8b8812af731fb6acbdd1aec173e37b84066eea))
* Make share link name italic ([224cea3](224cea33ced403f45c7d833ab576be44c89d199a))
* Move the url link to the bottom of the items ([6576b61](6576b6148ce1b02dbe6a335778592c4b72e275de))
* Color the task color button when the task has a color set ([51c806c](51c806c12b90aa124384497856590f5010b9ff49))
* Color the color button icon instead of the button itself ([bdf992c](bdf992c9bfe9de176a22f7b5a6fdae1bc5e5010f))
* Move the update available dialog always to the bottom ([a18c6ab](a18c6ab8d860a496905f58278315222992bacd07))
* Show the task color bubble everywhere ([2683fec](2683fec0a67f6afd16579bb44a6ceadc0edd565f))
* Color the task color button when the task has a color set (#2331) ([f70b1d2](f70b1d2902f91a88eaf33f1a9799489c20a6a143))
* Namespace settings archive script setup ([ad6b335](ad6b335d41e07e8ce2e74e4282d572ba4c04ea30))
* ListNamespaces script setup (#2389) ([ff5d1fc](ff5d1fc8c1961134ef3baec09be52b02c0b6898e))
* NewTeam script setup (#2388) ([e91b5fd](e91b5fde0216e15f739da22efbcaae3829e31ba1))
* Port label store to pinia | pinia 1/9 (#2391) ([d67e5e3](d67e5e386d7d1901694fe0004f580807754bcae1))
* Use pnpm ([d76b526](d76b526916d4aca279670d2690f7bb8e63e432a7))
* Move list store to pina (#2392) ([a38075f](a38075f376aa5cc2d8a06943cf8932366a0d4011))
* Task relatedTasks script setup ([943d5f7](943d5f79757b73f447c51641812e7766edeffe9e))
* Allow marking a related task done directly from the list ([ce0f58c](ce0f58c7833bbb37974709112cdedad88ae07cc8))
* DeleteNamespace script setup (#2387) ([0814890](0814890cac92b813b5b93bb42c7a40e2dc13cb94))
* Task relatedTasks script setup (#1939) ([d57e27b](d57e27b4a62aaa0f0a739f030515fff72a56f7fc))
* Use pnpm (#1789) ([f7ca064](f7ca064127863de4a4c1e3ae29d84d6bd5311cb9))
* Add hot reloading support ([1c58fcc](1c58fccd926586b2303ce41939a535b2044a78a9))
* Move namespaces store to stores ([9474240](9474240cb9159a0e1b42f82cb492cc267782ce4f))
* Port namespace store to pinia ([093ab76](093ab766d45247b3b1d12740dc6b24c6b48f21c4))
* Feat-attachments-script-setup (#2358) ([4dfcd8e](4dfcd8e70f54d2ed977d4b8de5fb8bf9469819aa))
* Convert namespaces store to pina (#2393) ([937fd36](937fd36f724f2b383fe51ae25a55ba90f58c8975))
* Move attachments store to stores ([c2ba1b2](c2ba1b2828439d3bd1e846a4bb9a4c456562c460))
* Port attachments store to pinia ([20e9420](20e94206388ab694248942996fdb67b7be87e76f))
* Move config to stores ([9e8c429](9e8c429864923215be5b110fdcb7c4a586c60f3d))
* Port config store to pinia ([a737fc5](a737fc5bc2affc87b209746ecf04c66e1f6077db))
* Filter-popup script setup (#2418) ([ba2605a](ba2605af1bb6f9ba7d3bd1b99ed862d510c6bb31))
* ListLabels script setup (#2416) ([89e428b](89e428b4d285f3465a40773fbda564c432fb371e))
* Possible fix for pnpm ci errors ([e8f0b56](e8f0b5665161e77bcc961ec0dc57c5b127b93a1f))
* NewLabel script setup (#2414) ([7f581cb](7f581cbe2780633fdfa03609824182fe93fe77e3))
* Possible fix for pnpm ci errors (#2413) ([bc83309](bc833091f2b919177ce75815b562818c93ea2884))
* Feat NewNamespace script setup (#2415) ([63f2e6b](63f2e6ba6f22502becf61aa89c729fa9d01cdc7b))
* ListList script setup (#2441) ([bbf4ef4](bbf4ef4697fc6338ad603e2491fe4aed61057cd8))
* Move auth to stores ([f30c964](f30c964c06987f87b615c3eec25197241175db96))
* Port auth store to pinia ([7b53e68](7b53e684aa405a7874f189dcb404c031dfed1388))
* Auth store type improvements ([176ad56](176ad565cc64e2212eedb1601c844e458d7e4bb6))
* Improve api-config (#2444) ([8f25f5d](8f25f5d353064f383e97bbc524ce6e00ba559d0f))
* Convert model methods to named functions ([8e3f54a](8e3f54ae42c21fdae62225892ad340877651df27))
* Migrate auth store to pina (#2398) ([9856fab](9856fab38f62f82a42d5cb3b69b232eb319b8050))
* Move tasks to stores ([1fdda07](1fdda07f650702b7e3943e0afc7532367ee20100))
* Port tasks store to pinia ([34ffd1d](34ffd1d5729341bdede217387a4a4c490d7d60d8))
* Move kanban to stores ([9f26ae1](9f26ae1ee6241b2ef529f01d3511380c9d7a4576))
* Port kanban store to pinia ([c35810f](c35810f28fc5aacefabad7526b0ac4e982d53cc7))
* Port tasks store to pina (#2409) ([8c394d8](8c394d8024a825b961e825543453d188c28fa370))
* Automatically create subtask relations based on indentation ([cc378b8](cc378b83fee2b326610cdda1997cc5236f947fbf))
* Automatically create subtask relations based on indentation (#2443) ([ec227a6](ec227a6872ababb612cb0b7e68ca0c20676117c1))
* Migrate kanban store to pina (#2411) ([d1d7cd5](d1d7cd535ed992fc0a8be8afaf13250ac9b61132))
* Move base store to stores ([df74f9d](df74f9d80cdd44315a29189ecb2f236482cb70f5))
* Port base store to pinia ([7f281fc](7f281fc5e98c5eb83f926100c7f79ee374c5a784))
* Rework loading state of stores ([1d7f857](1d7f857070651f676bbb5bd7e6d79c7fed56be5f))
* TaskDetail as script setup (#1792) ([2dc36c0](2dc36c032bad93654fbd64a68682685870972feb))
* Add github issue template ([9400637](940063784b3ec129e99fe18c4eb2b205ffb15163))
* Login script setup (#2417) ([63fb8a1](63fb8a1962f9ecd8c9a079e2770b4658c5559d84))
* Datepicker script setup (#2456) ([ff1968a](ff1968aa36254d788d0d80ba2d156ce66f4a9df8))
* Multiselect script setup (#2458) ([0620b8f](0620b8f0b308e358526bed0d82322ffb9c0627cf))
* ColorPicker script setup (#2457) ([b08dd58](b08dd58552edb763f007f355f5c0d36d6dccbd05))
* Migrate kanban card to script setup ([a5925ba](a5925baff03ac2809b7c601b45b93363b6188083))
* Migrate kanban card to script setup (#2459) ([3e21a8e](3e21a8ed6ee74d85628feedd8855c817af8de538))
* Add nix flake for dev shell ([12215c0](12215c043d45d2f2294e65671587a923997e6f6f))
* Fancycheckbox script setup (#2462) ([06c1a54](06c1a548867e37a74a8493bd44fef728e10c658b))
* Editor script setup ([db627ed](db627ed28af8432e6971ad08864d11e56d3512c6))
* Use floating-ui (#2482) ([f360ebf](f360ebfe9854aeae9cb426c67b1bb48aa74a9c08))
* Update eslint config ([4655e1c](4655e1ce34223337c953ebbe52f94ef811034e6b))
* Feature/update-eslint-config (#2484) ([6f2dedc](6f2dedcb488ec6a38182e85e702ec880263ecbd3))
* Move composables in separate files (#2485) ([c206fc6](c206fc6f3462be2e0ebc0bd16d96b3c0099fdda1))
* Add display of kanban card attachment image ([3d88fda](3d88fdaaddca15b98efa938f0b2813420d56ad84))
* Promote an attachment to task cover image ([877e425](877e4250554b31db2d57f44a7443c5d04c783e59))
* Add indicator if an attachment is task cover ([f01107f](f01107fd737e2205bf60498b3d2954a251c3d9d4))
* Show done tasks as strikethrough when searching for new tasks to relate ([74a9b9a](74a9b9ab1b31740fe84a7dddd91a04995c1eb58d))
* Allow users to leave a team they're in ([feeaca2](feeaca2c02fb233c35a81f786acd5cbdf5c5d21d))
* Add TickTick migrator support ([1af4f78](1af4f7811a63826c4aa4740a55f606757e22c7ae))
* Make salutation i18n static ([c20de51](c20de51a3c98792580c0a2f2751648582ac5ac0c))
* Get username from store getter ([c4d7f6f](c4d7f6fdfa18c221597b28198d5fa432b1e934dc))
* Use getter and helper in other components as well ([9de20b4](9de20b4c54d192a20f9135388de9fa13121ed322))
* Make salutation i18n static (#2546) ([29f6874](29f68747bbd7da50d37ae3238b6b19782ec8022b))
* Refactor password reset to use a single password field ([4ed665f](4ed665fbd9dc4db1ecb6afc1a75d1818c3518186))
* Rename useTaskList ([7ce8802](7ce880239ec3ce16313d93bfefa657c499bbfb29))
* Add basic implementation of ganttastic ([2b0df8c](2b0df8c2375ec5f9afe43207807e999bcc693d21))
* Allow passing props down to the gantt component ([49a2497](49a24977f96cff1e90e706321505ae43bf7efadf))
* Only load tasks which start in the currently selected range ([ed241d2](ed241d21bea91795a10cdc1af92561d435c9eedc))
* Dynamically set default date ([736e5a8](736e5a8bf55ccf7cbed23fd3af48122c459bcdc6))
* Dynamically set default date ([3b48ada](3b48adad675b0b20dc91a08f8ebbfe1dd1c3806b))
* Create new tasks ([ef46893](ef4689335b3e738b7e1338657e9dcd69c82fbcb9))
* Add open task detail when double clicking ([d2c4092](d2c40926ded479db92d0f3b77d2ece5842bcacbb))
* Scroll ([c8eac91](c8eac914d10a09453afb70d35c6d16faac9cd00c))
* Styling ([80c151c](80c151ca6c4a76a5f912505672eee471f77a3bba))
* Update task in gantt bar after dragging to make sure it changes its color ([ebd824b](ebd824bddf8d37a66d2dbf7f330b39c8849db9b2))
* Show done tasks strikethrough ([3eacc07](3eacc0754ff50fed2d5a50198480c5c8d697f6ce))
* Handle changing props ([29dcc02](29dcc02217dfe9d52b3cdd6166ca82cc8be1022e))
* Loading animation ([8c62a9e](8c62a9e198fb5b8221a13747e9510f5036ed3095))
* Create task when pressing the button ([0a9588e](0a9588e09730e83ddc61630012e62c0530a9997d))
* Increase the default date range ([5f7159e](5f7159ebc49e73bc4757c7cefa9a10ed14d65b46))
* Only use one watcher ([64fdae8](64fdae81ec8a1b807a1b1788a6954c8d7850dc36))
* Review changes ([f21a4e1](f21a4e1e9f558e999e1f6638847aeab4d73b9636))
* Update ganttastic version ([2f820e5](2f820e517f6dea384440a9574da4f82c02c86143))
* Improve types ([3b244df](3b244dfdbecf2f1feaa766b5c9e52c7e66dfe52a))
* Working route sync ([acdbf2f](acdbf2f8f5b8e28e923d7598696dadec373c7a67))
* Working gantt-chart ([eaf7778](eaf777864ac857275bc657bf39f1886460d307d2))
* Abstract to useGanttFilter / and useRouteFilter ([2c732eb](2c732eb0d55c9161b8d47cbc850421136994bff4))
* Simplify ListGantt styles ([c7dd20e](c7dd20ef57f037db0ac8bbdc583463ae98ffe9ac))
* Move useGanttTaskList in separate file ([7f4114b](7f4114b7032c24d9305c7c731ad1fef2f9390dcd))
* Remove gantt-chart wrapper ([aefda38](aefda38bdd8fa5f5b4f4d2c7486566f669dd6929))
* Use PascalCase for component name ([acb3ddc](acb3ddc73fd7a8240d42774c80c68b5a725c3734))
* Use ref for filters ([51dc123](51dc123d893517a30c2dbb26a68e877b493ec95e))
* Use plural for filters consequently ([6bf6357](6bf6357cbd281fa5b99b7aae9845fee90c758ae7))
* Move config preparation in separate function ([e74e6fc](e74e6fcc996cced93f040782ac278db6baea975e))
* Align with vue-flatpickr-component 10 ([874dc1e](874dc1e5fc9f76ad3d45f555b9d04585cd9a2704))
* Replace our home-grown gantt implementation with ganttastic (#2180) ([fd3e7e6](fd3e7e655dbbd59f9a94db0f18a3ef4876cec059))
* Improve useTaskList (#2582) ([d5258b7](d5258b73153a477a82c750482a6fd504c5823b7a))
* Unify savedFilter logic in service (#2491) ([9807858](9807858436e4b7d6de8dcb71b2a03a55ed8a7d52))
* Quick-actions script setup (#2478) ([386fd79](386fd79b4983b9d472d46219fc60c1a1a2cc1012))


### Miscellaneous Tasks

* *(ci)* Sign drone config
* *(ci)* Sign drone config
* *(gantt)* Wip daterange
* *(gantt)* Upgrade packages
* *(gantt)* Upgrade packages
* *(gantt)* Pnpm install after merge
* *(i18n)* Use global scope
* *(task)* Move cover image setter to store* Improve type imports ([af630d3](af630d3b8c1536c1a9a320172aaf19e000bb2517))
* Remove date mixins ([b0ee316](b0ee316a262ca71b9cfecbaaeccab7f9465ec09d))
* Remove global mixing ([4a247b2](4a247b2a7d6741bfec9fbdb387c9313d7b6381d1))
* Remove unnecessary defineComponent ([6f93d63](6f93d6343c1c518fec3591b83b999efcbccf9607))
* Better variable typing ([42e72d1](42e72d14a4a804aa38908cc2a9d6b4cb120c988a))
* Align docker cypress image version with drone ([2445f0e](2445f0eec8b130d8d71e5fc399a399c0d1cf6836))
* Minor fixes ([49f3b92](49f3b928cbc16031cf65fa3ed1cc908968e1083b))
* Automerge renovate dev dependency updates ([d822709](d822709991ee4dc52ee8aa56a03248c6e4a3a709))
* Rearrange non-dev dependencies ([b8d77a6](b8d77a617b0b205fbf8553d3ce060547c96f0f22))
* Remove &nbsp; ([d91d1fe](d91d1fecf1b34734ef8af21c3c34bdaaa6d53e09))
* Remove unused id ([5f678e2](5f678e2449529758cd6ade233c52a0c091889fd9))
* Set more expressive variable names for available views dropdowns ([7e7fa80](7e7fa807fd1c6a34c5236cca4fb20141ca9d0454))
* Improve types ([6d9c4a7](6d9c4a7aa083425e252b96729b57c16ab13fd295))
* Don't cache node_modules ([b542221](b542221dac6a14cd84aab446ceab0888bc98bb38))
* Don't use node alpine image ([6624db1](6624db1d49545524083d124698fa5b6e02bbfb0c))
* Use node alpine image ([dfb3561](dfb3561310bec49043a630136a2d51cc80184cc1))
* Optimise loading order (#2435) ([ca899d3](ca899d3b5172be6f39a60bdaffab58330225ecd9))
* Make const out of export download file name (#2436) ([878c6ea](878c6ea9e17527b3f199f4acf10588e910b5727c))
* Spread title ([3970d0f](3970d0fd315488427df0c4a37447eb52dca322b4))
* Use better variable names ([8ce242b](8ce242bb6595ef12442a6ba0fb37eb66c65dd71b))
* Break earlier if index === 0 ([d58f8b4](d58f8b4ba1d873abb0fc8dc4c2cec64a33b55ab8))
* Use jsDoc to explain param ([5bd7c77](5bd7c77b68f08ab4771f3d80d5191def9d634204))
* Small review adjustments ([af7f840](af7f8400e901c2f4d9c5c4cca7614af62892a75e))
* Remove unneeded this from PasswordReset.vue (#2473) ([c232170](c2321703a767395b77523d4551ea508396b7cae8))
* Remove IE edge fallback (#2477) ([3248dcd](3248dcd6636627548f2df869900a7943c0dde0ba))
* Add line-wrap ([eb80bfa](eb80bfa00de891ee12643d664e8610d1f3bc851f))
* Better wording for cover set button ([a773137](a7731370a0bcdd8a393036a617dd1953cd39f5df))
* Update happy-dom less frequently ([458df80](458df8044306642e5da813ff8341bed07f67f26a))
* Move helper function outside of composable ([aa2278a](aa2278a56411dc8045fa468b090755cf5d899d09))
* Use flatpickr range instead of two datepickers ([c289a6a](c289a6ae18fd5936b789270cc72408374a790edc))
* Use width property ([7a7a1c9](7a7a1c985e0feb8de62ddbdd54f36f2a09a9d765))
* Remove old component and dependencies ([6cb331e](6cb331ee0f26dffcbf700426da17acb6159aea3e))
* Use Loading component ([766b4c6](766b4c669ff52f6d6c888727e62142eaa90de54d))
* Use @/models ([d3925b8](d3925b8d80e16e25e9b82d057fb47ed9f41f61a0))
* Uppercase const ([98d0398](98d0398ca840d8d8077f850c8ca4e65784373b61))
* Don't set required if there's a default value ([ed5d3be](ed5d3be7cba7992eb18a3ed1844c085cf88b3bdd))
* Define types ([56a2573](56a25734d7557663e2ba43ba41f4922f0b10ed8b))
* Don't use for..in ([6975a2b](6975a2b286628294b8909bce3d43334cc383d987))
* Add types for template ref ([4be0977](4be097701449b74bbeb7218b539db65961539591))
* Don't use ref when not necessary ([fd9d0ad](fd9d0ad1553756414696315508bc2d8928f63d9d))
* Update lockfile ([957d8f0](957d8f05a5e9548138f8dce192513928deb02669))
* Better naming for input ([df02dd5](df02dd529181e9701ce586dba9025c83eeaf48d8))
* Clean up ([2acb70c](2acb70c56257202fe7d136b36ceaaa2fe122491e))
* Pnpm install after merge ([26e522c](26e522cf8c302f5d63b26134e5fa37bed5c808ef))
* Use vue-ganttastic release ([6c61907](6c619072b4863328c24588bb08a9543806942be1))
* Don't pass other params to ListGantt than route ([cf0eaf9](cf0eaf9ba1816b610ba1cbc9b4a6c661f00f61a5))
* Refactor parseTimeLabel to own function ([443e1a0](443e1a063dfff3cbb82a9f625e05bf7e2b606cbe))
* Add git-cliff to flake ([b817720](b817720907b0c4bb848e9624e3fdf71437ba0bde))


### Other

* *(other)* [skip ci] Updated translations via Crowdin


## [0.19.1] - 2022-08-17

### Bug Fixes

* *(dark mode)* Code background color
* *(dark mode)* Make a focused text only button actually readable
* *(lists)* Moving a list into another namespace on the first position* I18n scope ([5b8d142](5b8d142abba9559f6b259940d5f35ccb1c098496))
* Clear all localstorage when logging out ([51ffe93](51ffe930483bdd02118b512bb00a1ca50a5ce2e5))
* Search for assignees by username (#2264) ([c6e7390](c6e7390f137991a6d992ad62ddca46a07fd4bf4e))

### Dependencies

* *(deps)* Update dependency sass to v1.54.2 (#2219)
* *(deps)* Update vueuse to v9.1.0 (#2220)
* *(deps)* Update dependency sass to v1.54.3 (#2223)
* *(deps)* Update sentry-javascript monorepo to v7.9.0 (#2224)
* *(deps)* Update dependency vue-i18n to v9.2.1
* *(deps)* Update dependency vitest to v0.21.0
* *(deps)* Update dependency vue-i18n to v9.2.2 (#2228)
* *(deps)* Update dependency postcss to v8.4.16 (#2230)
* *(deps)* Update dependency vue-tsc to v0.39.5
* *(deps)* Update dependency caniuse-lite to v1.0.30001374 (#2231)
* *(deps)* Update dependency netlify-cli to v10.15.0 (#2232)
* *(deps)* Update dependency esbuild to v0.14.54 (#2233)
* *(deps)* Update typescript-eslint monorepo to v5.33.0 (#2235)
* *(deps)* Update dependency @faker-js/faker to v7.4.0 (#2234)
* *(deps)* Update dependency vite to v3.0.5 (#2237)
* *(deps)* Update dependency sass to v1.54.4 (#2238)
* *(deps)* Update dependency esbuild to v0.15.0 (#2239)
* *(deps)* Update dependency vue-tsc to v0.40.0 (#2241)
* *(deps)* Update dependency vitest to v0.21.1 (#2236)
* *(deps)* Update sentry-javascript monorepo to v7.10.0 (#2242)
* *(deps)* Update dependency rollup to v2.77.3 (#2245)
* *(deps)* Update dependency esbuild to v0.15.1 (#2244)
* *(deps)* Update dependency vue-tsc to v0.40.1 (#2243)
* *(deps)* Update dependency vite to v3.0.6 (#2252)
* *(deps)* Update dependency @vitejs/plugin-legacy to v2.0.1 (#2250)
* *(deps)* Update dependency @cypress/vue to v4.1.0 (#2249)
* *(deps)* Update dependency @vitejs/plugin-vue to v3.0.2 (#2251)
* *(deps)* Update dependency @cypress/vite-dev-server to v3.1.0 (#2248)
* *(deps)* Update dependency esbuild to v0.15.2 (#2255)
* *(deps)* Update dependency vite to v3.0.7 (#2254)
* *(deps)* Update dependency @vitejs/plugin-vue to v3.0.3 (#2253)
* *(deps)* Update dependency eslint to v8.22.0 (#2256)
* *(deps)* Update dependency rollup to v2.78.0 (#2257)
* *(deps)* Update dependency esbuild to v0.15.3
* *(deps)* Update dependency netlify-cli to v10.17.4 (#2262)
* *(deps)* Update dependency caniuse-lite to v1.0.30001376 (#2261)
* *(deps)* Update typescript-eslint monorepo to v5.33.1 (#2263)
* *(deps)* Update dependency vitest to v0.22.0 (#2265)
* *(deps)* Update dependency cypress to v10.5.0 (#2266)
* *(deps)* Update dependency @cypress/vite-dev-server to v3.1.1 (#2267)
* *(deps)* Update dependency postcss-preset-env to v7.8.0 (#2268)
* *(deps)* Update dependency vite to v3.0.8 (#2269)
* *(deps)* Update dependency esbuild to v0.15.4 (#2270)
* *(deps)* Update dependency cypress to v10.6.0 (#2271)
* *(deps)* Update dependency esbuild to v0.15.5 (#2272)

## [0.19.0] - 2022-08-03

### Bug Fixes

* *(ListList)* Use ButtonLink
* *(a11y)* Remove wrong aria-label
* *(button)* Min-height
* *(dark mode)* Dark mode adjustments (#1069)
* *(dark mode)* Disabled input colors
* *(dark mode)* Flatpickr colors
* *(docker)* Setting nginx run ports
* *(docker)* Properly replace api url
* *(editor)* Duplicate edit buttons for empty descriptions
* *(faker)* Imports
* *(gantt)* Use function to create default date
* *(gantt)* Correctly show month and year in gantt chart on safari
* *(kanban)* Transition animation for bucket footer when adding a new task
* *(kanban)* Make sure the buckets don't appear glued to the bottom
* *(kanban)* Background content scrolling when opening a task
* *(kanban)* Make sure the task position is calculated correctly
* *(kanban)* Error when moving a task to an empty bucket
* *(kanban)* Reset loading state after creating a task
* *(natural language parser)* Fix parsing short days
* *(natural language parser)* Parts of week days in other words
* *(password)* Watcher (#2097)
* *(quick-add-magic)* Use ButtonLink
* *(ready)* Remove class form fragment
* *(tests)* Wait until namespaces are loaded before checking if the history is present
* *(tests)* Add more waits for namespaces loaded
* *(tests)* Assert absence of last viewed headline more precisely
* *(tests)* Wait until lists are loaded
* *(tests)* Don't assert for h3 anymore
* *(tests)* Don't visit / directly but use navigation instead
* *(tests)* Make sure to create all lists before doing anything
* *(tests)* Make sure the namespace exists before trying to run the history tests
* *(tests)* Set correct user issuer for test users
* *(tests)* Remove old label task relations before adding a new one
* *(tests)* Correctly set task position in cypress test fixtures
* *(translations)* Typo
* *(user)* Settings wording
* *(vscode)* Example plugin name (#2076)* Remove attachment by id (#725) ([0376ef5](0376ef53e38a8b20137d710edb4ea0be4d0fb2d1))
* Use date-fns for gantt years (#734) ([077fe26](077fe264f009e9c60593daf04e48111e686cff79))
* Import bulma utilities global (#738) ([3ac25c9](3ac25c9f08d6a575c901c4164783f8ab75227d66))
* No drag delay when using mouse on touch device (#748) ([d88e299](d88e299358099c6ac5924d91bc660eb0fd80a3a5))
* Fix spelling in cypress README (#763) ([77352e7](77352e7a8c1eafdfa1a8c22152c7191bdaa5a61e))
* Prevent vue-shortkey use in elements with contenteditable (#775) ([17d11c6](17d11c6ce387ce15141a27574701aa3801f924b7))
* Computed in api-config (#777) ([3245752](3245752a8003026f86c50cdaeee99da5e902abb6))
* Quick add magic assignee prefix in explanation ([dedf6cb](dedf6cbf21f688ade4572e781299c6fc6f84de68))
* Lists disappearing when updating their namespace ([77f8b27](77f8b27dc67c220caad8984c7a11484645c8a588))
* Namespace collision of global error method with draggable error method ([ebeca48](ebeca48be42de1791235935365063bc86819eb2e))
* Breaking attribute coercion behavior ([697ea12](697ea12c8e032c3350256969455f22008a243e87))
* Remove unused function ([f762d8a](f762d8ad4d50c93c1f93dc901437ab2585db4ad6))
* Eslint settings (#787) ([feb34c8](feb34c8cc13f9220c0c27be7bf45ed90d543c4cb))
* Run tests with unstable api ([8b01dc6](8b01dc6b71ec1e7cbdee2beff5a85baee69ca899))
* Remove font preload of quicksand 300 (#794) ([166539c](166539c7e8ecbab80773fa76991ff068b267ac55))
* Date formatting for non-english languages ([a955488](a955488cdf7bbe54876b76548006979b0bd5eabe))
* Don't try to create a task with an empty title when creating multiple tasks at once ([4bd2c94](4bd2c94256156a358aee0c139522398e7652c77f))
* Don't enable editing when the user has no rights for it ([96ef25b](96ef25ba01a23c1f7b9812ce896d609b00bf8ff1))
* More spacing for last viewed tasks headline ([4163800](416380025ee426586b0533e6676ffa431fb9ee45))
* Quick add magic always disabled ([4a1b402](4a1b402e62c962e55c7d96011620e86f1889106e))
* Use dynamic imports instead of old async components for router views ([0c678b6](0c678b6e443e04800688a231f92d37f431c31fbb))
* New directive syntax ([3c89147](3c89147ee25f1afb84565a9280b333c0969dc806))
* Compiler warnings ([2b20f32](2b20f328cb8d61529d8909128e889ee05b64e80d))
* Directly set arrays, objects and delete directly ([db49b9b](db49b9b532b1da9773b262dcb85016d66722b6d9))
* Life cycle hook naming ([ecc3d3c](ecc3d3cf3f72a86a0b6da0c378cdefb59154e18e))
* Transition class names ([2ef2bb7](2ef2bb77008aacc3be2f84e2685a4f07c66fcd5a))
* Use vue3 v-model bindings ([51a740f](51a740f53c71323bfd7c10eb2d96291ebdd0f7fe))
* EmailPlaceholder translation ([8fc01f7](8fc01f774acf1da942239007bc5f952e6640b43c))
* Fix newList.vue ([aeabc42](aeabc42844541bfc068d6938eae2080a7c907ca8))
* Typo in translation string ([c3b6e13](c3b6e13009380838ddcc4767ed9b78a10bf3f8ea))
* Dropdown routes ([0cbffad](0cbffad49d7a7062f56ad2d7aee5f9acdf77dd9c))
* Vuex mutation error in edit list (#813) ([3f9917d](3f9917dfab7aaf238de5ea09e60a44c611474deb))
* Use correct translation key filter save success message (#823) ([a843cdd](a843cddbc9c98e20447724d63352e1d5c0cdeec9))
* Missing translation for error during link share auth ([cc22d8d](cc22d8d4e9e013ef6bc095c3b3722c2f479fd841))
* Wrong success message when adding and creating a label to a task ([22ef778](22ef7785fdf26d42f2d656ac1c6985cbf1073f99))
* Properly resolve relative date translations ([d583cb2](d583cb2094ed3450b2d5437391352058dd8e4b06))
* Translate months in gantt chart ([a558f5b](a558f5b35a2bd91773ce99ad6527531112e26eb8))
* Make task relation kinds translatable ([2a1004a](2a1004ac68064f0af84260bf73f7a6d53aba9806))
* Remove gzip compression of woff2 (#824) ([813982e](813982e833eb508df5f466c432a48ff593e460a3))
* Don't allow reordering tasks in filtered lists ([d284db6](d284db672ef90702bddc9de02cf10a46cd345928))
* Vue3 types ([59401bc](59401bc1da70ac030060a5d72bcf56b26a40b4eb))
* Unassign user success messgage (#831) ([36d4599](36d4599276acd44426fa93508a3f4d66218d0f73))
* Kanban drag task test ([4ae18ec](4ae18ec16278fd970f404d9ee357a393e336c1a2))
* Access namespace only if loaded ([e064c3b](e064c3bf96adc95ef7b4be1a0e5389484c1e1e5d))
* Give the dom some time to update for some tests to pass ([60ef07d](60ef07da0fa86bc8f4c9d4c539bc1dcafb71261e))
* Wait with redirect until route name is available ([eec02a5](eec02a55a4d54d41400d4b6fbad92661bbdd7ec7))
* Mutation errors by make a copy of the store settings ([3750b0f](3750b0f78b8d163ef0c75350ac5212ccf641a816))
* GetTaskById function ([9b2e9fc](9b2e9fc17f6f9bb25387dd6e5884d48feff6da22))
* Watch deep for multiselect modelValue changes ([0bf68ef](0bf68effb8cb67957715ff3f483bf1fefc466ef2))
* Watch deep for other arrays ([dfe401a](dfe401a9dc72f18b22fe045e967ad6d185d6410c))
* Use correct listId when deleting bucket ([d7ed5b8](d7ed5b8f1178e634b4776c567fccc772da21b319))
* AddTasksToBucket mutation ([7c3ece5](7c3ece58167f5a6dcc87558759cb6eb8a0fdb928))
* Use correct listId to load next tasks ([0b68a47](0b68a473efdb2ca39343be64e160fe489e4fcc1c))
* Remove broken getTaskIndices helper ([e0456cd](e0456cdfa1b076e52e4d59216c675908ca7078be))
* Add timeout to wait for move to finish ([fd77aaa](fd77aaa123d2bb725df01bc0f2a0d9e4e5742715))
* Watch deep in listSearch ([427f18d](427f18d59e3e0bcceeccabd04f2462d09b3376cd))
* Remove side effect from computed ([18c3148](18c31482dfd5d3f1e38060295b1fcd7e67abbf37))
* Use correct method for fakers uuid ([cc8b037](cc8b03778c755f02d68835a49ce3617c51114820))
* Disable service workers in cypress (#830) ([e6a935f](e6a935f49dca3301f0ef8cb138909f67a4fa72e1))
* Wrong async order ([50fa592](50fa592aada3f01de039bbdc636102c8f91d71e5))
* Editing a label works now ([69821fb](69821fb6635d9eb54c450e6bd4885090cce31dbb))
* Switch view height on devices with smaller font size ([b5b56a6](b5b56a6e4afab3f207ae7e31d5640412fed3c401))
* Task input height on devices with smaller font size ([c30c2e0](c30c2e00cb1a87bef2963ce2703500a9f0941bc4))
* Task input height after removing a line now works correctly ([3f96ce6](3f96ce6d60af580f9c0b745f54d2ba9a784f667d))
* User dropdown padding on mobile ([4fef047](4fef047d74ce284eb1af3cddd9f666f0b4397273))
* Wrong word in en base text string ([435535f](435535f8cd6f1dc540ce67f955225adecccafc4f))
* Add null check for parsedTask listId (#31) ([26568fe](26568fe5c61a47f1c5d1abc679d77cae13d6d393))
* Remove wrong active prop ([9c730d3](9c730d33811469abf4aa62ae7077844dd46d2303))
* Use componentData prop in draggable to set class ([80163ee](80163ee9923b86213e6bf81b3ca74639e6b2c58f))
* Pagination in vue 3 (#859) ([373a766](373a766f5c91c1289ab0f6947dc967f7d77646d5))
* Setting background to state mutation violation (#858) ([f05e811](f05e81190f9f0e17ce85642b1668a3d24af77092))
* Remove disabled prop for editor ([a6db1e7](a6db1e7391776c288609fe44357a4e9c11d2d257))
* Await namespace creation ([54d456e](54d456e886156631940d5e5ee21a7f029e356134))
* Call loadList just once ([7f5f44d](7f5f44d7f0326c487693b91a1df7ec9361f1f7a8))
* Use async action to change current list ([a60ad77](a60ad77bdcde3fffbaf343b8c5d8bd35e9633dac))
* Always sort tasks the same order in chrome and firefox ([df32893](df32893ce657bc57a6a5f9b850e23d7e814d76db))
* Kanban card elements spacing ([5766ae4](5766ae48d7676a4c921d26879a8989fbe8819a9c))
* New tasks were always created in the default list ([7e29dde](7e29dde7170b51b90d280ef3e02fae73e267c607))
* Label search in tasks not working ([529b3d2](529b3d2890c9f64862c13526bf5d255ff4576ae0))
* Task edit pane spacing ([e52c139](e52c139c9fde196ac1e6ea521bbddc509611229b))
* Creating a new task while specifying the list in quick add magic ([f884020](f884020c55a9761f2d3cb38d44e87db913fa475e))
* Await  getAuthUrl ([5636559](56365591cfdc8c70eba2cf5e56cf60f09dc08517))
* "TypeError: i18n.setLocaleMessage is not a function" when changing languages ([74d785d](74d785d60659f1525f62fad0a122e13886b3501d))
* Change the ui locale ([2fc96cb](2fc96cb6a7c4231e5513bd4b6832239d3468378f))
* Use lodash.debounce for searching unsplash background ([c107825](c1078255fc836f4609bda779cfac3f713b7b720b))
* Set the current list when opening a task ([1c8e26b](1c8e26bdc615a3477c139f3d8128d71c71027c43))
* Don't search for first letter images ([0cc7166](0cc71667677fb814c660289aaf842c9df71d7b89))
* Vuex store mutation error when moving a task with attributes on kanban ([9d48700](9d48700cd9ad5e0bb4394f32fcac94ba59f48d90))
* Lint ([2de94bc](2de94bc902573e8aa8767b96eecbef0ecf393ad4))
* Sort order by dueDate, then by id ([ae971b2](ae971b23bc3a6cd8565250d1b3a7b30b2f168a33))
* Loading labels after login ([1d46b85](1d46b851700a73cccb1c4d0b52b6657c3290b788))
* ATTR_ENUMERATED_COERCION spellcheck on kanban board ([36d5262](36d5262f1d3db57264ed84944bf4b71d3bcaeea1))
* Use correct prop for CreateEdit ([3f61c6b](3f61c6b21a30fb5cf7858ec2d0d4e0477e144910))
* Adding a list to favorites ([f4372ec](f4372ecd050932439e840d640de7292f80732737))
* Vuex store mutation violation when saving user settings ([4c24118](4c24118b4869e182eccf385e6ae798ba678fab6c))
* Lint ([1864359](18643597513d3e221bcfd67f4a99f280cc4dd9e6))
* ATTR_ENUMERATED_COERCION in a few places ([571b019](571b019c00bbbeb09a96bdcf9c6eab39199f7656))
* ATTR_ENUMERATED_COERCION errors with editor and contenteditable ([3ba9cd2](3ba9cd2d9989d46501b439c6fc1316d9f340cb90))
* ATTR_ENUMERATED_COERCION errors with contenteditable ([f795d2d](f795d2d0f31a28dd292ff9b237d6ce21ebd284aa))
* Remove nonexisting prop ([c7b4c25](c7b4c25caa49cdf2a149184cd9935e598d449237))
* Task attachment upload ([6d472bf](6d472bf5ca7c2b9621e3be5a8e9b2f5671c2e066))
* Update node in .nvmrc as well (#886) ([0fdfccc](0fdfcccee9b8185588dc62346e468c65ac57d3ea))
* Move .progress styles together as close as possible ([6ba974f](6ba974f9faf7912d796dc54de3b00e629149dc32))
* User dropdown-trigger background ([f496c9d](f496c9d678d6dc3a43df6f52e7de8f5eb19ee03f))
* Use :deep() selector instead of ::v-deep ([87d2b4f](87d2b4fed38e01aa31308ef299e94a17fce8b790))
* Label spacing ([2645edc](2645edc9e01c054ae4b780ea0f458a801314a505))
* Fix kanban height calculation with hack ([9e6afdb](9e6afdb7528e263e1814471baf2c81057d17025c))
* Use $shadow variable directly ([89cd8ea](89cd8eafc4ab08a67ac8ae532a945458f506a2ed))
* Logout error (#901) ([d4fe378](d4fe3781f73fb3c4a60ec16bc4d13c938440cd52))
* Use correct dash for english translation (#902) ([77fc5c0](77fc5c0c6f9e7c0edaecae9d62bd9329f11ab9c4))
* Create multiple tasks at once with multiline input now correctly uses the titles per line ([6394485](639448552405dad85bd1b80d81323f47d88da3db))
* Migration icons are not resolved properly (#864) ([e1a7fb4](e1a7fb4999c00f9c5cb9c63f2e508a03d0ccfb32))
* Check if notifications are available at all before checking if triggered are available ([8389587](8389587a60c6c31bc2f56fb0f9528feb4989bc9a))
* Showing deletion scheduled at for non-scheduled users ([54c5cab](54c5cabf48880f6be1b9bc0f941fd11342b0fc3b))
* Don't crash when an error does not contain a request ([85e85aa](85e85aa2bbc4b2b36ab0fbe225e0dd626f85320f))
* Don't try to check undefined relations ([a515b0c](a515b0c3a4e39b047cd0d9c16dfdf1ba21031af8))
* Loading tasks with infinite scroll in kanban buckets (#920) ([7aede35](7aede352f16687ff2746d1f905c3faab89926f1d))
* LoadTeam in EditTeam (#922) ([28a448a](28a448a1aabc2e228b359b38619105fa7e7cb682))
* Fix(style) restrict new task input size (#938) ([ee430b8](ee430b8687914b1bf6399177183f35fb0b28bf46))
* Navigation show and hide animation (#927) ([d3c303b](d3c303ba2aa3518691def7d19b07f0ca8bb8a045))
* Reloading in error component ([e831c3e](e831c3eb6540e88eeec40e2b93cb66a74e68eb06))
* Lint ([6e043e3](6e043e3b9e6addbad938252625d7f50cf3f372c6))
* Label spacing (#946) ([7e82aa8](7e82aa83e6b58954f1c41d9b31b75c513b17d47d))
* Adding or creating a label with quick add magic (#944) ([58986c4](58986c4a7a36154640e3128e4c8e0c1c3935c801))
* Typo in quick actions translation ([054f804](054f8044271a635ce1799a3532fa6d4f38f1c0f1))
* Fix ShowList margin in Home (#987) ([20e059c](20e059c921b6ac4d4571c1161121a2587f280481))
* Don't try to deploy a review env when not a pr ([be78fc1](be78fc177dadaaff0609384a2f4e1e46189d94f7))
* Preview deploy for PRs (#990) ([03eee06](03eee061ff96a4803a1fbe0d2ec409a5f3cc52f7))
* Show current host if configured api url is /api/v1 instead of "" (#994) ([31f3445](31f344503cf3cd347c7b2f6d9cffecee37f65917))
* Logo on ready screen ([1fa1644](1fa164453c0f2ea0ebd03b2decdc8f23a442d64d))
* Vikunja logo size when migrating ([0684806](0684806db0fa743be6d4a73dba0b4789c7f4eed5))
* CurrentPage of pagination component is undefined (#1002) ([6c6ccc6](6c6ccc647e59af6612050b0c47ccd3ffd129cbf9))
* Comment alignment (#1008) ([ed78a83](ed78a83ed9f3619fd0e5a8ae39e4241c676f0e4d))
* Api not found by default ([26213d5](26213d5e8c2dd66dc381b2dad8c54c789d515694))
* Kanban card spacing (#1005) ([ae5d3ec](ae5d3ecac5883bd26994fb034254d2e22570c9d3))
* Fix attribute coercion for contenteditable (#1025) ([b838e74](b838e7494dbdab6f2235b1dcfa9ac274ea112737))
* Fix #1046 logo overflow on login (#1050) ([44f8e3e](44f8e3ea9b1fbc95d3a8f18aca559b9039ecc342))
* Check for notification api (#1043) ([b029889](b029889f27bbbb0cb7376c462106b0eb0650a808))
* Deleting a namespace ([4ef54f1](4ef54f1bc24fc39c8c3c223af08ba9fabdfe349c))
* Remove mentioning of context (#1017) ([981babd](981babd691f465a6010e05be8194723d9623bbde))
* Edit task comment ([dc347ed](dc347ed8ba4fa0a0299981f46820662c0be2b976))
* Logo overflow on login (#1050) ([04c9441](04c94418d7889a1a2d912721a027cbe301f852f9))
* Upgrade cypress image (#1096) ([b7ad29f](b7ad29f05644b03f7918e7f99c34188857430584))
* Remove obsolete code (#1097) ([0c9dad9](0c9dad9891c87d7d19c1e93077338bc231a09221))
* Switching from a list with a background to settings would not remove the background ([734db07](734db0795c1467836d718d55f4931c4491b271db))
* UseColorScheme (#1117) ([baa8653](baa86530c89df37b94d4fc9997d1b83ae561a557))
* Cleanup some scss vars (#1118) ([769d94e](769d94e879f3698e33c07fcae1ed16f89f413890))
* Add import url suffix for vite svg loader (#1122) ([bc8b04f](bc8b04fc7a943f02aad904b487a927f6857ae519))
* Duplicate filter in gantt-component (#1121) ([e45bc83](e45bc831327086ff0838c83001e25dbdc72ed7cd))
* Unit test for "should recognize dates of the month in the past but next month" (#1131) ([20f0496](20f0496fa594784f5236a966d457f0ab9d8d30b5))
* Remove unused variable ([b96e89c](b96e89ca8c799b670b7bd7e8058ee85f74696da6))
* Home view (#1129) ([4137bab](4137bab7fc58663f69c2dca4d17db178cb79ab7a))
* Checklist update not working ([bba9a8e](bba9a8e0080d78c7ee6134d7940083a355a03295))
* Default sentry dsn in docker ([10fe38c](10fe38cef6048280f2766507292cb0d5124d2784))
* Unindent styles in pagination (#1172) ([cb9e1e8](cb9e1e891d432b5d00d471d828a60b654f1c9464))
* Spacing for deletion message ([a106511](a106511646d8c8c59fa2c58476e92ef3258f68b7))
* Use watcher to check for user query tokens ([807fb6a](807fb6a9fe404f65f0616aac78f707cd6e4ade5d))
* Saving default list (#1143) ([543dae2](543dae2f30c795f33f6846b9dfc3d87fa7d92719))
* Llama color (#1212) ([b3b7669](b3b766998347e3c916bcbf1f41983a9cb15e8fd6))
* Auth and move logic to router (#1201) ([063592c](063592ca3de68a3dd679e1458018f6e850f8a787))
* Move forgot password link next to password label ([f7eb160](f7eb1605092301921632d6df6996d5ee35e35661))
* Message spacing ([a1814ea](a1814ea29d74631f7befb5d51d85b2a16d1741a3))
* Disable login button ([9c04fb4](9c04fb4e40cbe1185a81c28696a546d638645c97))
* Add .vue suffix to fix typescript warning ([3eb0d58](3eb0d58f7935c593cd0adac59ed3ebbad69b13b8))
* Motd on mobile ([a4ec41e](a4ec41e9377bbe1b5e282aa8482bf0e94c0521ab))
* Remove unused var ([c46273c](c46273ca341cc12fcfe8440946cd97ee0117fdf0))
* Remove @ts-ignore ([27cd953](27cd9535bf7e59de20c0f6261c69525237427d60))
* PropType validation in message.vue ([9a3069c](9a3069c20d3011c61b9b4a54c11f6e153c444886))
* Lint ([9c5613a](9c5613ad98b742e008570d7f99e2424af64e335b))
* Disable broken stuff ([378f782](378f782d44feae63a19e19248f488def69e0e3c8))
* Pay attention to week start setting ([c24b8af](c24b8af00d4fcdb65e1421bce067e229340c5ddb))
* Date format ([729aa7d](729aa7d4cc4759521f1811b1464c0824dfd26699))
* Date range ([d6dd1fc](d6dd1fc0e31e07c9f1a8c97f1ef1fd7bed632a76))
* Checkboxes ([f691e96](f691e96e22b69ed487445295a3a78baa3eab0702))
* Loading spinner ([75cbc73](75cbc73b33ed76045ccf5fc3c34c41c04dc9e342))
* Z-index ([294e89b](294e89b6f749a092e993fe807004c0774f333a46))
* Lint ([0710cea](0710cea9e5b910f82350c42906cc243429a7e386))
* Test ([7dddfea](7dddfea79ea6539d195eaa2fabc808b6337dbb1d))
* Padding and centering of the kanban limit and dropdown ([8ae84ea](8ae84eaf42c1f7cf8cb26e555cfb77e70aabfef2))
* Blockquote styling in dark mode ([0befa58](0befa58908ae3f5a2467429d1ab0ff7fded0eb46))
* Re-add modal transitions ([16b0d03](16b0d0360159aed24cae41fabc4a88a37e9d9711))
* List loading ([5937f01](5937f01cc57d74f9bc69d58c05406537975189f5))
* List specs ([e78d47f](e78d47fdcf93052fdcf5d41abbe2bd63ca51e086))
* Task done label test ([da8cf13](da8cf13619e269a0fc02b3b733d9eeb0b5d9c860))
* Kanban tests ([58207db](58207db6c3a3133fd928fc77675dbe8778a01988))
* Sharing components ([700fce3](700fce3c2cf6fb6d7dea69771786ae5f4da9a1ad))
* Fix task remove label test ([f335826](f3358269e5346ec53652a7d9f3643f739ea21809))
* Closing modal ([e54d958](e54d95802bb961d7a9b507ca08fa952dde1e20cd))
* Check now just once ([6d62ca1](6d62ca1adaf4be5f94dac1a3840834e881b30aa3))
* Move local storage list view to router ([76f4cca](76f4cca5fe25c19be04493911f814b1e6a7c0f17))
* Don't set defined values for search and page ([e6e8a98](e6e8a9851446fb3523b4e31799a9d44c76f14d7a))
* Namespace new buttons on mobile (#1262) ([c618b7e](c618b7e0b6aaa20f7747a688867ee7111767ed71))
* Remove some of the typescript warnings ([49955eb](49955eb03a0c682092a08bc574aa32b807d92a29))
* Remove obsolet code (#1312) ([49a6569](49a6569db0c429caf656020cd4e248cb6ad71c92))
* Password validation field in test ([19a161f](19a161ff7858e702b708dd0798a079e39cc2bb39))
* Button size on task detail view ([4579dd3](4579dd3ce762b161a7d26b8fa0c5272c1279fd29))
* Don't reset active fields when saving ([68a76fa](68a76faacc004c3c9e3b4d0d91f98036b300d339))
* Make sure the app is fully ready before trying to redirect to the login page ([55826bb](55826bb8c9cdc2dffb409e51a3fc1eed967da99b))
* Editor cursor color ([0473c38](0473c385d6d8f7bd10c8c33e009f579ce308b5ba))
* Editor color in dark mode (#1338) ([76fe2ce](76fe2ceac6bfc6c74236acc940f3afbb84bd19b2))
* Don't recognize emails in quick add magic (#1335) ([ed88fb9](ed88fb91bcdefbc627b47bda3c35000948df8118))
* Flatpickr date not updating (#1336) ([6080e49](6080e49f26bb7ad764c8acf408649281c55defe5))
* Translation typo ([796a56d](796a56d5d8fcd3525405226c8f0edd950028f21f))
* Save user language when it wasn't saved previously ([c7ac81a](c7ac81a99f829f70e898ac97f057a9c1ed014348))
* Some typechecks ([26a94c7](26a94c7e8cd72811e27c1fdff672dc677d966664))
* Update available text color in dark mode ([b73165f](b73165fce4b6e286edb8f47a90f09b0383af1402))
* Keyboard shortcut message bottom margin ([cc3fcdf](cc3fcdf1c3557885ae230f0a1d37cce8e292eb78))
* Attachment meta data not aligned properly ([443a9c1](443a9c14b9c8158daf684063981caafc06815a55))
* Don't try to format invalid dates as ISO ([50c3bcd](50c3bcd793653415daa48c29a1900600a4786d5f))
* Check if a shortcut has an available function before trying to invoke it ([8233c8c](8233c8c9539869b27faf8a28272ab9629550fb51))
* Scrolling to heading if it wasn't available ([1818ed3](1818ed364899d1b1d46a6235804f4c580ac03eba))
* Vuex store manipulation warning when modifying task labels ([ff9e1b3](ff9e1b3fcad02fad290c1b663d33e640f85a8a8c))
* Label edit spacing ([6a6203f](6a6203f553972127dc136adbf2480087a8667780))
* Subscription prop validation ([ca938b8](ca938b8615af27c44e20e5172e4b3f346c540666))
* Lint ([0548649](054864925777d623a1b4156c10b700fa75d363e0))
* Show namespace count for long titles (#1057) ([375c3ad](375c3adfb1f4d1650087b560c8abb591d42e1855))
* Subscription prop validation linting ([c896ad5](c896ad58836a28e3123e7992f6112ab9d9f28f54))
* Use AsyncEditor again in comments and description ([5867f79](5867f79735d06ad8d4596c82dd59e469799124b0))
* Replace faker with community fork faker-js/faker (#1408) ([6db0559](6db0559b8193e1da090dc844e3084a1ca2dc7b8a))
* Vuex store mutation violation when archiving a namespace ([fdd2e7e](fdd2e7e53840924a79bfd43157954a3d0e9cfe4d))
* Subscription icon not rendered correctly ([b3697cb](b3697cb9bfc4b5971b1d68fdf1ffd07c8c0effd5))
* Don't try to parse date numbers with letters around them ([9319413](931941359b21972c85de492b1f98e7b2bcd600e1))
* Edge cases for dates where the next month had fewer days than the current one ([d913fa1](d913fa1745df56a083f4dd3dec8d4391740cc1c5))
* Ts errors in subscription ([24b7821](24b7821c5027b1f22cc23b7535c0c5ee992db571))
* Keyboard-shortcuts typing ([57965b1](57965b1ea3869f9735db13e5715560793eadde9b))
* CurrentList typing ([a9fb24a](a9fb24aa35239be38f255f7eae019519a496b345))
* Improve ListModel typing ([98b41a2](98b41a22c6824b069f4b1d7e7609e128ba13d240))
* Fix ts errors in various files ([de3c47d](de3c47dc69fa0c80f83004eab1f6bd455cb4ea9e))
* Use to.hash for returned element ([6894024](6894024ad43931cad49efbdf006b7c66d585e32c))
* Expose configureCompat types ([0bd235c](0bd235cea37678aba52eb1d74d43d23d1f266dc2))
* Mark broken test as skipped ([9995abf](9995abf64cc5c69496d55308435c143d00356c1e))
* Related task with the same namespace ([00ffe17](00ffe17eb838981bdca013a7731e9f8e2056f705))
* Related task within the same namespace ([20a9ad2](20a9ad2c9efea59a1752bae170744f500cba9092))
* Undefined prop subscription ([3e311e0](3e311e07cdd603d970b834fa5de6b8c926c474dd))
* Make isButton prop optional ([3d420c3](3d420c37708ae3568200ccf8214dd2d120a0af37))
* Don't try to load a language if there's none provided ([210a78b](210a78be86385b2d57a65563082e60bd11965217))
* Don't try to load a language if there's none provided ([ba20ac3](ba20ac3b89e11af897978c350f20b501fd028686))
* Custom date range with nothing specified ([16f48bc](16f48bcc2dbc081c5526040e789a1a9f07f1575b))
* Reset the flatpickr range when setting a date either manually or through a quick setting ([4d23fae](4d23fae9ad1d1238dfdecf9694adfa36313c6651))
* Now correctly showing the title of predefined ranges ([6c55411](6c55411f71b1790f7144624ba651df468ab37af8))
* Llama position ([a74fc47](a74fc473357617d49988794b211b98badf1975c7))
* Lint ([7135288](71352888007e69b92b9d00f38bc9fca0d77d6a2c))
* Sort tasks correctly by due date ([9e7c258](9e7c25834724d9df2c645a5dd9aafb2730747dfd))
* SetTitle import ([cbbcb7e](cbbcb7ef239c2f627c5e64ac169532ac01e25c4d))
* Correctly send filter values ([eeee1c8](eeee1c842ab2a0c4ec5d63208f36021c342ec177))
* Related tasks add button and task dates in read only view (#1268) ([581b2cb](581b2cb4ab211ee1added752eba6138dd9ca6b61))
* Lint ([aac777e](aac777e2864184d97fe4e26cd324a729927e8e8c))
* Styling ([a22792a](a22792a4b4d54ff02c6f6ed2efb668bc30fe161a))
* Don't reset flatpickr date ([4ac7d6b](4ac7d6b9df9b9135aec6140711ad44f74fc9e53b))
* Emit function name (#1511) ([10bcdc8](10bcdc880417469df34aa7193637a16c84cb9e78))
* Make logo change reactive (#1509) ([cf849da](cf849da104a103454e2f5270cb05201ce5795d3f))
* Mark query parameter as string ([badbae0](badbae0e9a02bb84c4ed8ddfd971f52cdcc58de0))
* Namespace archive success message ([8b90b8f](8b90b8f6a86a70c59daaea5320a40e54a612f2e1))
* Hack to fix wrong index position ([e2c81d8](e2c81d840f35c47167592bb7ca21648a09cef6e8))
* Use BaseButton in MenuButton and fix computed (#1532) ([d57c9af](d57c9af3329cad3ca7e5f12476eaa486f8adf5ee))
* Property spelling ([17dc276](17dc27697131cbdd7b633ab948903626d521219c))
* Replace slugify in deploy-preview-netlify with simple regex solution (#1543) ([28af46b](28af46bcd35cf90de4ce8f44a1a2588798b70627))
* Direct store manipulation in tasks (#1534) ([c419062](c419062e49b53edceda1ccb07ed226d4b92180ec))
* Lint ([622f08f](622f08fb1bb8faf1050111fc400e283384114f14))
* Popup not really hidden when hidden ([c7943ef](c7943ef8238bca266eabc80d4b9a41ce991d3e28))
* Modal not scrolling content when open ([da162d5](da162d5652ab5153742edf8130ecdca8817a9cc6))
* Api config domain name contains the current domain instead of the provided one (#1581) ([bdb53ec](bdb53ec8eef362e1bb805810556e30d38658a0b8))
* Don't try to sort tasks when none were returned ([8cdcfaf](8cdcfaf071544a17a9be9a0dcc11d51380873f40))
* Don't try to filter notifications if there are none ([731506f](731506fab756ea8739b378485f94810799a774d5))
* Don't try to validate nonexisting fields ([b83cec2](b83cec2f0ec3b3a42e8488d1c90de22a5bd95a50))
* Don't fire close event multiple times ([9a55482](9a554826819bda40a7620a1952c2a5f4218c7697))
* Removing a label from a task ([1256c37](1256c37b69b3b0c1f12ccb858f8e98fdaff431a0))
* Hide "title required" error after entering text ([45c0529](45c05296a6a95a16cbbf1404107cbd16b6b8cd34))
* Update page title when changing the task title ([7b62a08](7b62a0895d3f9dad1370d323914bd7720c8b7b7d))
* Undo task done from list view ([051dd98](051dd98ff7f530a4132bd97abb858e3b636e17dd))
* Missing app padding when opening the task detail modal ([6d0cbc5](6d0cbc51f6920c1e8d0e0ac0cc4c552f3648d040))
* Don't always show a scrollbar ([74ab197](74ab197dc61cca5e27f3951eaaa79182345e2211))
* Pop sound not saved and played when marking tasks done ([c06cc6a](c06cc6ad7aec3f91eb8fdbe9dc621a5b4577b6aa))
* Kanban board layout on mobile ([a23b4a9](a23b4a96ee25557ef8cc984cf8f716de47b72dd9))
* "invalid date" error when trying to set a date and none was set before ([b144802](b144802203dd23bd5024da50cee4f27bbfe1b5d7))
* Don't rotate kanban cards while dragging ([7f2189b](7f2189b45552da08c742bb9ef798b57533581d52))
* Keyboard shortcut text indicating what works where ([cf5460d](cf5460d2980fe23feb1efdd9f772403a626b32de))
* Aria-label for password field ([81993cc](81993cc2e68768d70418dd37f2337b98488428af))
* Modal close icon color in light mode on mobile ([63e04f8](63e04f874af268466b5b6aace8545c2130698ac9))
* Mobile menu backdrop ([d7b1d7d](d7b1d7da7f826f8dd25d21b3004c6a65eccb337d))
* Multiselect search results text color ([8f65031](8f650316dcaf4b481ffeb432e4509a744ac3d076))
* Related done tasks strikethrough ([87ac22b](87ac22b44829b29ca3b698ed7c5dc0d096ec745e))
* Load the list tasks only after the list itself was loaded (#1251) ([7f56a35](7f56a3537cced373e88e9320398e5a604023953d))
* Add task input layout on mobile (#1615) ([3639498](3639498b3f7d99270df285b689d924b37a4f50e5))
* Make sure a list background is set in store when adding one ([42c0fc6](42c0fc61854de6bb8fb3190ff5811057caa7f745))
* Setting the last viewed list after navigating away from it ([b7a976a](b7a976a9cf5329ec35f418c11d1997edc72d5b26))
* Lint ([a055a3e](a055a3ea52488287377920aba530eda9de15d3dc))
* Forgotten import ([4605061](46050611d86f4173048e4e7a2f7583326b62d0b4))
* Loading list views would sometimes not get loaded ([2e537f6](2e537f6d63690724fb83b31107ddf3e34f63edba))
* Indentation of nested checklist items ([ad8ca46](ad8ca462cb19a0e5d81e23ebf07c292381ea0219))
* Lint ([53787a6](53787a65dfccebeca938d84e3e6a30f47aa48304))
* Remove self and replace with this ([175b786](175b786ec6c807ef61aefc1153bb786b8e14f787))
* Service worker path ([fb2eb4c](fb2eb4c439580a34533a0bf0ee0adfb8f2d3b02d))
* Lint ([b65839d](b65839d0d76a13b95a38aa162c9b751ff22d7990))
* Type ([19b772f](19b772f8ee378417729c042d683ab6ae123b16e4))
* Create token ([898b22b](898b22b37794d8d42face1062a97c557c13f1141))
* CaldavToken model typehints ([58b0397](58b0397cec5b287d7a369228bb748eb45873c200))
* Menu on mobile devices ([010eca1](010eca1d0cb7f38474b5a26ca5e856d4b5a7aab4))
* Properly set list backgrounds when switching between lists ([b289754](b2897545e4dc26837f0f08a549cc195feac4f3c3))
* Reset all tasks before loading new ones ([480bfbc](480bfbceeff3ea682d26bb719fc17195172beaf3))
* Resetting the list when changing from a list view to a non-list view ([1eb19f8](1eb19f87645e60cae83dfe7b155718490d0ec37b))
* Rename caldavToken to ts (#1814) ([e3483b1](e3483b1a5a5658dd52a4c9534ca35ab9c0f533b4))
* Remove obsolete watchEffect (#1795) ([9c24380](9c2438026b91fb84b0d3e3f4f4c476b5769b827f))
* Uppercase types (#1810) ([080675b](080675b38f7ba87396a6872ca6304564dce27d30))
* Typos in translation files ([c962c8c](c962c8c3f411bb0bfedbeee2d0c70c0828940aea))
* Checklist summary design on home page (#1842) ([bf3e16c](bf3e16c6eee3c910331f77efaf4b34005ce3f7ec))
* Fix imports ([d325810](d325810e5570d247fa88fc7396eb6e7d07cd46a3))
* Update nvm node version (#1856) ([2083a52](2083a52a56b0f5214e2f89d6fe135bc5da8aa0c9))
* Subscription works correctly again ([89c81ae](89c81ae854bc6de83619e9f0aaf5774aeaa3ab97))
* Update notification spacing ([49946b2](49946b27662ed30ab0556c901c6bc91e4e6396f1))
* New task input focus ([24701a1](24701a17f5cc8f656bc3f0ede7aa3a5ff5cca888))
* Progress bar alignment in task list ([fbcf587](fbcf587e938f1d990f74669889e548875e0a537c))
* Date filters are now correctly converted ([87d4ced](87d4ceddb8a033f89f9764256f0d39e6da25fc3e))
* Actually deleting the list now works ([b40d6f7](b40d6f783c013c0d15bcfec656942947393be4fc))
* Remove user from team ([86efe9f](86efe9fd23978d9af2c7bbd1198c9d74b8bedda2))
* Dark mode for user and team settings ([ed85557](ed85557cf3031184a6bb9176c1371b2ac17723dc))
* List dropdown menu item hover background color ([8846b2f](8846b2f8625ec700fb0dd9e2a8e17cae7754d0f7))
* Favorite task list spacing in menu ([24aca5c](24aca5cfa687868dae6ff68367d4bb5231cae436))
* Spacing between username and notification ([ce3f285](ce3f285224595f49e02750d620676c7af80cd6eb))
* List hover background in dark mode ([2dba9e6](2dba9e6e571b3870ff5d39896aa8a971f2ed5403))
* Tooltip color in dark mode ([1a98305](1a983059697b26b45e23e6835aa9665cbc242718))
* Filter button alignments and backgrounds for link shares ([c2694dc](c2694dc08907d2e20eea90dd97d829e8dfbfb65b))
* List views not switchable on link share mobile ([21a8298](21a8298a968fe95432de0d13142c23b9e91d4baf))
* List title not set as page title after closing a task popup ([a38bd7e](a38bd7e971f3e1d4125d28dc99a702932b1f6958))
* Use a new notification service on every poll to make sure it uses a non-expired token ([3e7f598](3e7f598ee8858b201f7c4492e2fb3aa78ba782ea))
* Remove workarounds to properly overlay the top menu bar over everything else ([4b0d491](4b0d491359d2c5241742e86d83e3a1e6a6b2de41))
* Active color for editor buttons ([f1c9887](f1c98872437bc78057c78c420324db058ce93a8d))
* Lint ([1d9665f](1d9665fb8473b8483646187833d494c1e7688b88))
* Import in PasswordReset (#1923) ([4b6015d](4b6015da99ab7225774ca567b00cadc3d89ed343))
* Allow clicking on confirm for a date without requiring to click on another input field ([138b067](138b06752f8221c1ebae89e318ef36734662efbc))
* Direct state mutation when adding another reminder to a task ([44dc898](44dc8983c8474a617dec62fafa0722fabfb5e397))
* User menu not properly positioned on mobile ([90bb800](90bb80034693408bb950c10439963ec82aca87a1))
* Update banner spacing ([e3373d2](e3373d2e4e76a6576520986853d07dceccac7209))
* Navbar user dropdown spacing on mobile ([fee2fe7](fee2fe76ce1202bd95ffbf84d07f96937bbbfe88))
* Very long words overflowing in descriptions and comments ([9936d36](9936d3683ec69675fbf4169419667090581d5243))
* Throw error messages in dev mode (#1968) ([2359678](235967844a8be2e3fa12157eecac97967e407bee))
* Disabled attribute fallback (#1984) ([96fce73](96fce73192bb49bae50a2d34cb377c9779998396))
* Problem with newTaskInput ref (#1986) ([829eed0](829eed0b9f8c42096eb47a5f0656eef60e28a8ca))
* RepeatAfter initial modelValue ([72925fb](72925fb93837f580eec178a76cba4cb43e9b200b))
* Button prop type (#1966) ([f91424f](f91424f693bfe988fa6960c5d720a177f58d6c08))
* Watcher in listSearch (#1992) ([b4aa650](b4aa65018cfb891c2e6fe35a75060558b701b04a))
* Quick actions not properly styled ([e1e410b](e1e410b50b4f4aa98992cae446151be5c347923e))
* Replace vue.draggable.next with zhyswan-draggable ([1569042](156904247124ddd7890f77bccef0a7d87189a45e))
* New label text color in dark mode ([cadcaa9](cadcaa966f27eeb469d3a41b335a386718362a66))
* Properly reference task input textarea from parent component ([745d466](745d4660d80c6eb00d682f70a283d1cebd8cba94))
* Rely on api to properly sort tasks on home page (#1997) ([efed128](efed128f0325ae18342a0dd64913dc51b5ccec91))
* Sed replacement of SENTRY_DSN (#2036) ([d308d66](d308d665bdb11ac52726bccb6698b3e7ab989102))
* Top header still in foreground when menu is open ([a2c0696](a2c06967539ab2ae8534aaba942c07a8cca18dbd))
* Pride logo rounded corners on mobile ([9716517](9716517ffa8a60f1627131ca92b07c02aa18d03f))
* Use grey-100 instead of light so that it is properly set in dark mode ([d1f22c5](d1f22c5b43a8cf8a6ad395e7f3c443d14a41dfca))
* Show a proper error message when no list or default list was specified ([9bbc1bf](9bbc1bf9396354042d9a67581d0ed83edf9b3694))
* Don't try to load the namespace again when navigating away from the settings page ([aadf75c](aadf75c7bffc3ffc5367d7a126905b7a3f8b4a1c))
* Capitalize all priorities ([f2f5f90](f2f5f90adc655d0be196854165f1bba31650154b))
* Task default color should be set and evaluated properly ([37c3656](37c36560fb4f92567bb0117c0128f3ab3cd7de4c))
* Setting user settings in cypress tests ([9d0415e](9d0415e24c266e0f22cfa66301cf7d9868008802))
* Opening the list share dialog hangs everything ([978cb97](978cb9769ed1360112347739b730b270ccee1e32))
* Sharing lists and namespaces ([fab58a2](fab58a2e6d8fa722fab7c16256693e85ce86460c))
* Properly define focus expose for new task input field ([e0864fa](e0864fab3eb7807b52eb0b0102b5ca5ef42366d1))
* Archiving a list ([2b8a786](2b8a7868254db5fd739d6f0ec62e67ee802d8429))
* Fix import type ([d064f0a](d064f0acc099311dae63db86dbea0a1f6e247864))
* Fix linting ([5835848](58358481bcab645087135f32bf1c00adaf52cd3a))
* Re-enable some compilerOptions ([8f82dd2](8f82dd27835667654bdd869e4dbf0b070064948d))
* Cypress plugins import ([77466e3](77466e337353fdc1d7ee5eb7b21c55c752c4b6bd))
* Cypress plugins ([c6d214b](c6d214b9ebb469a0636c86c5b5b62dc335ac53c7))
* Button styling ([02f985d](02f985d8a3627f5536dfe0c88e2c97c0aaea8701))
* Add ButtonLink component ([12544c5](12544c52ca55e6ada6d0fc3b5c1fab4dfce1a9ef))
* Setting a label on a task fails if the kanban view is open in the background ([990639d](990639dd24918e3dca3163401aee70cb97fb5bdb))
* Make sure weekday parsing in quick add magic ignores the casing ([dff5d84](dff5d84ebbe8ac1f623bd985c47a9d5b45bf4037))
* Pass modal bindings to teleport target (#2109) ([6e54929](6e549291041c78ab81a3b5b6a738bdf64588841d))
* Datepicker button color and spacing for overdue dates ([ab7bf7d](ab7bf7d8f927e52ef38ccdbd43bc2d603b1bc416))
* Expose focus function for BaseButton ([cc07933](cc079336a8322f310de50e79337537024d01289b))
* Add a task relation with enter when only one search result is available ([e8705c6](e8705c66dde0f67c3ac570fc4155d17e462e6c9e))
* Task sorting in table ([4a8b7a7](4a8b7a726a06d2b3d40e7f7d1c0c650992963c3a))
* Task sorting by position in list view ([99a5afc](99a5afc817c65aa53dcafd5b5491ef9fb6202b3a))
* Make sure saved filter data is correctly populated when editing a filter ([a4c3939](a4c3939fb66a25e4e2b50098283378735f7585b2))
* Upgrade packages for vite 3.0 ([d96ea38](d96ea384dce1f282722e37161a1767a090def812))
* Datepicker confirm button overflow ([9fd2f4e](9fd2f4ea5caad1a307a6886379f029a83ad0aa6c))
* Use of sortable js with transition-group (#2160) ([0456f4a](0456f4a041300a2c076c808b5b844d0677ffaba0))
* Don't try to pass nonexistent props to filters ([6dc02c4](6dc02c45dd78485106b89537f9ca49328a4adbb7))
* Don't use transitions for elements where it is not possible ([c2d5370](c2d5370e4a88fc646dccd3c598c2953b6b40ca82))
* User avatar settings ([62bbffb](62bbffb17ef863a3a1575d6827e392cad3ee0e84))
* Quick actions arrow key navigation in dark mode ([f5bb697](f5bb6970322f825faf64841c97539c6a324ca8d4))
* Pagination on table view should not open the list view ([a4d3caf](a4d3cafdf121f9261e12b390072ff9acfd1157e1))
* Properly update state when duplicating a list ([e7de930](e7de930129c51ae5d68f915d8132543366aa5554))
* Don't allow marking a task as done in a read-only list ([175fb02](175fb02629f66887f4ede582e03fe520b1783b26))
* Lint ([8b0e88b](8b0e88b57435dfa0173910104c484986ce58b4e6))
* Vuex state mutation error when moving a kanban bucket ([9ddb55a](9ddb55a5efa8851427b71aff6f478e695bea1687))
* Logo spacing for link shares ([3becf87](3becf8738b6b6eeb040593c3d05005c8e50baa64))
* User menu dropdown ([8183fce](8183fce829c79837b53aad63b9d28ae6f6b4c30b))
* Don't allow negative repeat amounts ([71c8540](71c8540c74f8448a2fddb0791e28b22c76a6d4b6))
* Don't try to load lists after logging out ([4c560f1](4c560f1a031c21a3e735bdbad061b284a03b6618))
* General user settings empty when loading the settings page ([ff48178](ff48178051c4726093751bc3a2317e836ea8b99c))
* Transition error when deleting a task ([56147dc](56147dc9fbed5680a06de200dbd9111d92b5cf6f))
* Progress bar color in dark mode ([8b30726](8b3072672a795163acfe4b2b5065c4f59ca0dd1c))
* Default label color in dark mode ([31480ea](31480eae72cb936226aba3454f55a672d87059cb))
* Properly parse dates or null ([e82a83c](e82a83c8cf5e8721f80bb426c3dfdd9549e09a88))
* Don't replace the last edited task with the one currently editing ([ad7ed86](ad7ed86d36a9385149ea75eefa8b34f643050345))


### Dependencies

* *(deps)* Update dependency vite to v2.5.6 (#723)
* *(deps)* Update dependency marked to v3.0.3 (#726)
* *(deps)* Update dependency esbuild to v0.12.26 (#729)
* *(deps)* Update dependency sass to v1.39.2 (#733)
* *(deps)* Update workbox monorepo to v6.3.0 (#730)
* *(deps)* Update dependency typescript to v4.4.3 (#740)
* *(deps)* Update dependency esbuild to v0.12.28 (#744)
* *(deps)* Update dependency jest to v27.2.1 (#745)
* *(deps)* Update dependency vue-i18n to v8.25.1 (#747)
* *(deps)* Update typescript-eslint monorepo to v4.31.2 (#749)
* *(deps)* Update dependency marked to v3.0.4 (#753)
* *(deps)* Update dependency dompurify to v2.3.3 (#754)
* *(deps)* Update dependency @types/jest to v27.0.2 (#766)
* *(deps)* Update dependency eslint-plugin-vue to v7.18.0 (#761)
* *(deps)* Update dependency date-fns to v2.24.0 (#757)
* *(deps)* Update dependency vite to v2.5.10 (#746)
* *(deps)* Update dependency cypress to v8.4.1 (#750)
* *(deps)* Update dependency sass to v1.42.0 (#751)
* *(deps)* Update dependency browserslist to v4.17.1 (#770)
* *(deps)* Update dependency esbuild to v0.12.29 (#769)
* *(deps)* Update dependency autoprefixer to v10.3.5 (#771)
* *(deps)* Update dependency sass to v1.42.1 (#772)
* *(deps)* Update dependency vue-i18n to v8.26.0 (#779)
* *(deps)* Update dependency esbuild to v0.13.1 (#776)
* *(deps)* Update dependency vue-i18n to v8.26.1 (#784)
* *(deps)* Update dependency esbuild to v0.13.2 (#782)
* *(deps)* Pin dependency ufo to 0.7.9 (#780)
* *(deps)* Update dependency jest to v27.2.2 (#788)
* *(deps)* Update dependency autoprefixer to v10.3.6 (#792)
* *(deps)* Update typescript-eslint monorepo to v4.32.0 (#799)
* *(deps)* Update dependency cypress to v8.5.0 (#800)
* *(deps)* Update dependency jest to v27.2.3 (#801)
* *(deps)* Update dependency vue-i18n to v8.26.2 (#803)
* *(deps)* Update dependency esbuild to v0.13.3 (#802)
* *(deps)* Update dependency vite to v2.6.0 (#805)
* *(deps)* Update dependency jest to v27.2.4 (#806)
* *(deps)* Update dependency vite to v2.6.1 (#807)
* *(deps)* Update dependency vue-i18n to v8.26.3 (#810)
* *(deps)* Update dependency @fortawesome/vue-fontawesome to v2.0.4 (#835)
* *(deps)* Pin dependencies (#834)
* *(deps)* Update dependency @fortawesome/vue-fontawesome to v2.0.5 (#837)
* *(deps)* Update dependency @vitejs/plugin-legacy to v1.6.2 (#838)
* *(deps)* Update typescript-eslint monorepo to v5 (major) (#856)
* *(deps)* Update dependency date-fns to v2.25.0 (#853)
* *(deps)* Update dependency vite-plugin-vue2 to v1.9.0 (#851)
* *(deps)* Update dependency sass to v1.43.2 (#850)
* *(deps)* Update dependency cypress to v8.6.0 (#849)
* *(deps)* Update dependency vue-i18n to v8.26.5 (#847)
* *(deps)* Update dependency autoprefixer to v10.3.7 (#839)
* *(deps)* Update dependency ts-jest to v27.0.6 (#843)
* *(deps)* Update dependency eslint to v8 (#855)
* *(deps)* Update dependency @vue/eslint-config-typescript to v8 (#854)
* *(deps)* Update dependency vite to v2.6.7 (#845)
* *(deps)* Update dependency browserslist to v4.17.4 (#840)
* *(deps)* Update dependency typescript to v4.4.4 (#844)
* *(deps)* Update dependency esbuild to v0.13.7 (#841)
* *(deps)* Update dependency jest to v27.2.5 (#842)
* *(deps)* Update dependency marked to v3.0.7 (#846)
* *(deps)* Update dependency axios to v0.23.0 (#848)
* *(deps)* Update dependency ts-jest to v27.0.7 (#857)
* *(deps)* Update dependency esbuild to v0.13.8 (#861)
* *(deps)* Update dependency highlight.js to v11.3.0 (#863)
* *(deps)* Update dependency vuedraggable to v4.1.0 (#872)
* *(deps)* Update dependency highlight.js to v11.3.1 (#869)
* *(deps)* Update dependency jest to v27.3.0 (#866)
* *(deps)* Pin dependencies (#870)
* *(deps)* Update dependency vite to v2.6.9 (#873)
* *(deps)* Update dependency jest to v27.3.1 (#878)
* *(deps)* Update typescript-eslint monorepo to v5.1.0 (#877)
* *(deps)* Update dependency vite to v2.6.10 (#876)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.13 (#871)
* *(deps)* Update dependency @fortawesome/vue-fontawesome to v3.0.0-5 (#875)
* *(deps)* Update dependency eslint-plugin-vue to v7.20.0 (#881)
* *(deps)* Update dependency postcss to v8.3.10 (#882)
* *(deps)* Update node.js to v17 (#883)
* *(deps)* Update dependency postcss to v8.3.11 (#887)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.14 (#885)
* *(deps)* Update dependency sass to v1.43.3 (#888)
* *(deps)* Update dependency eslint to v8.1.0 (#890)
* *(deps)* Update dependency browserslist to v4.17.5 (#891)
* *(deps)* Update dependency esbuild to v0.13.9 (#892)
* *(deps)* Update dependency marked to v3.0.8 (#893)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.15 (#894)
* *(deps)* Update dependency vite to v2.6.11 (#896)
* *(deps)* Update dependency rollup to v2.58.3 (#895)
* *(deps)* Update dependency axios to v0.24.0 (#897)
* *(deps)* Update typescript-eslint monorepo to v5.2.0 (#898)
* *(deps)* Update dependency cypress to v8.7.0 (#900)
* *(deps)* Update dependency vite to v2.6.12 (#904)
* *(deps)* Pin dependencies (#905)
* *(deps)* Update dependency sass to v1.43.4 (#907)
* *(deps)* Update dependency @vitejs/plugin-vue to v1.9.4 (#908)
* *(deps)* Update dependency vite to v2.6.13 (#909)
* *(deps)* Update dependency esbuild to v0.13.10 (#910)
* *(deps)* Update dependency autoprefixer to v10.4.0 (#911)
* *(deps)* Update dependency @vue/eslint-config-typescript to v9 (#914)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.16 (#915)
* *(deps)* Update dependency esbuild to v0.13.11 (#916)
* *(deps)* Update dependency esbuild to v0.13.12 (#917)
* *(deps)* Update dependency rollup to v2.59.0 (#928)
* *(deps)* Update typescript-eslint monorepo to v5.3.0 (#932)
* *(deps)* Update vue monorepo to v3.2.21 (#934)
* *(deps)* Update dependency marked to v4 (#935)
* *(deps)* Update dependency browserslist to v4.17.6 (#936)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.17 (#937)
* *(deps)* Update sentry-javascript monorepo to v6.14.0 (#940)
* *(deps)* Update dependency @vue/eslint-config-typescript to v9.0.1 (#941)
* *(deps)* Update dependency eslint-plugin-vue to v8 (#913)
* *(deps)* Pin dependency vue-tsc to 0.28.10 (#955)
* *(deps)* Update sentry-javascript monorepo to v6.14.1 (#958)
* *(deps)* Update dependency eslint to v8.2.0 (#959)
* *(deps)* Update dependency vue-tsc to v0.29.0 (#960)
* *(deps)* Update dependency vue-tsc to v0.29.2 (#963)
* *(deps)* Update typescript-eslint monorepo to v5.3.1 (#962)
* *(deps)* Update dependency vite to v2.6.14 (#967)
* *(deps)* Update dependency esbuild to v0.13.13 (#964)
* *(deps)* Update dependency vue-tsc to v0.29.3 (#968)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.18 (#969)
* *(deps)* Pin dependencies (#974)
* *(deps)* Update dependency cypress to v9 (#975)
* *(deps)* Update dependency marked to v4.0.1 (#977)
* *(deps)* Update sentry-javascript monorepo to v6.14.2 (#979)
* *(deps)* Update dependency netlify-cli to v6.14.21 (#980)
* *(deps)* Update sentry-javascript monorepo to v6.14.3 (#982)
* *(deps)* Update dependency vue-tsc to v0.29.4 (#981)
* *(deps)* Update dependency rollup to v2.60.0 (#983)
* *(deps)* Update dependency marked to v4.0.3 (#988)
* *(deps)* Update dependency netlify-cli to v6.14.23 (#986)
* *(deps)* Pin dependency vite-svg-loader to 3.1.0 (#989)
* *(deps)* Pin dependency @github/hotkey to 1.6.0 (#995)
* *(deps)* Update dependency browserslist to v4.18.0 (#998)
* *(deps)* Update dependency vue-advanced-cropper to v2.7.0 (#999)
* *(deps)* Update dependency vite-plugin-pwa to v0.11.5 (#1000)
* *(deps)* Pin dependencies (#1003)
* *(deps)* Update dependency vue to v3.2.22 (#1006)
* *(deps)* Update dependency vue-tsc to v0.29.5 (#1007)
* *(deps)* Update dependency netlify-cli to v6.14.25 (#1009)
* *(deps)* Update dependency browserslist to v4.18.1 (#1010)
* *(deps)* Update typescript-eslint monorepo to v5.4.0 (#1011)
* *(deps)* Update dependency @vue/eslint-config-typescript to v9.1.0 (#1018)
* *(deps)* Update dependency esbuild to v0.13.14 (#1014)
* *(deps)* Update dependency @vue/compat to v3.2.22 (#1016)
* *(deps)* Update workbox monorepo to v6.4.1 (#1012)
* *(deps)* Update sentry-javascript monorepo to v6.15.0 (#1015)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.19
* *(deps)* Update dependency typescript to v4.5.2 (#1024)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.20
* *(deps)* Update dependency eslint-plugin-vue to v8.1.1 (#1026)
* *(deps)* Update dependency netlify-cli to v6.15.0 (#1028)
* *(deps)* Update dependency netlify-cli to v7 (#1029)
* *(deps)* Update dependency @types/jest to v27.0.3 (#1030)
* *(deps)* Update dependency vite-plugin-pwa to v0.11.6 (#1031)
* *(deps)* Update dependency date-fns to v2.26.0
* *(deps)* Update dependency postcss-preset-env to v7.0.1
* *(deps)* Update dependency marked to v4.0.4
* *(deps)* Update dependency netlify-cli to v7.0.1
* *(deps)* Update dependency netlify-cli to v7.0.2
* *(deps)* Update dependency eslint to v8.3.0
* *(deps)* Update dependency codemirror to v5.64.0
* *(deps)* Update dependency vue-tsc to v0.29.6
* *(deps)* Update dependency @vitejs/plugin-vue to v1.10.0
* *(deps)* Update dependency rollup to v2.60.1
* *(deps)* Update dependency esbuild to v0.13.15
* *(deps)* Update dependency slugify to v1.6.3
* *(deps)* Update dependency netlify-cli to v7.0.4
* *(deps)* Update dependency @vitejs/plugin-legacy to v1.6.3
* *(deps)* Update dependency @vueuse/core to v7 (#1066)
* *(deps)* Pin dependency bulma-css-variables to 0.9.33 (#1065)
* *(deps)* Update dependency netlify-cli to v7.1.0 (#1067)
* *(deps)* Update dependency vite-plugin-pwa to v0.11.7
* *(deps)* Update dependency @vueuse/core to v7.0.3 (#1071)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.21 (#1072)
* *(deps)* Update dependency @4tw/cypress-drag-drop to v2.1.0 (#1076)
* *(deps)* Update dependency netlify-cli to v8 (#1077)
* *(deps)* Update dependency @vueuse/core to v7.1.0 (#1078)
* *(deps)* Update dependency postcss to v8.4.0 (#1075)
* *(deps)* Pin dependency autoprefixer to 10.4.0 (#1080)
* *(deps)* Update dependency netlify-cli to v8.0.1 (#1081)
* *(deps)* Update dependency @vueuse/core to v7.1.1 (#1086)
* *(deps)* Update dependency marked to v4.0.5 (#1085)
* *(deps)* Update dependency postcss to v8.4.1 (#1083)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.22
* *(deps)* Update dependency sass to v1.43.5
* *(deps)* Update dependency netlify-cli to v8.0.2 (#1088)
* *(deps)* Update dependency netlify-cli to v8.0.3 (#1089)
* *(deps)* Update vue monorepo to v3.2.23 (#1090)
* *(deps)* Update dependency @vitejs/plugin-vue to v1.10.1 (#1091)
* *(deps)* Update dependency @vueuse/core to v7.1.2 (#1092)
* *(deps)* Update dependency postcss to v8.4.2 (#1093)
* *(deps)* Update dependency postcss to v8.4.3 (#1094)
* *(deps)* Update dependency esbuild to v0.14.0 (#1095)
* *(deps)* Update dependency postcss to v8.4.4 (#1100)
* *(deps)* Update dependency vite-plugin-pwa to v0.11.8 (#1102)
* *(deps)* Update dependency vue-tsc to v0.29.7 (#1106)
* *(deps)* Update dependency netlify-cli to v8.0.5 (#1108)
* *(deps)* Update dependency jest to v27.4.0 (#1107)
* *(deps)* Update dependency sass to v1.44.0 (#1110)
* *(deps)* Update dependency vue-tsc to v0.29.8 (#1111)
* *(deps)* Update dependency jest to v27.4.2 (#1115)
* *(deps)* Update dependency rollup to v2.60.2 (#1112)
* *(deps)* Update dependency esbuild to v0.14.1
* *(deps)* Update typescript-eslint monorepo to v5.5.0
* *(deps)* Update dependency date-fns to v2.27.0
* *(deps)* Update dependency netlify-cli to v8.0.6 (#1125)
* *(deps)* Update dependency vite-plugin-pwa to v0.11.9 (#1124)
* *(deps)* Update dependency marked to v4.0.6
* *(deps)* Update dependency netlify-cli to v8.0.13
* *(deps)* Update dependency netlify-cli to v8.0.14 (#1132)
* *(deps)* Update dependency jest to v27.4.3
* *(deps)* Update dependency netlify-cli to v8.0.15 (#1135)
* *(deps)* Update dependency eslint to v8.4.0 (#1136)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.23 (#1138)
* *(deps)* Update workbox monorepo to v6.4.2 (#1133)
* *(deps)* Update dependency esbuild to v0.14.2 (#1139)
* *(deps)* Update dependency vite-plugin-pwa to v0.11.10 (#1140)
* *(deps)* Update dependency ts-jest to v27.1.0 (#1141)
* *(deps)* Update dependency eslint-plugin-vue to v8.2.0 (#1145)
* *(deps)* Update vue monorepo to v3.2.24
* *(deps)* Update dependency eslint to v8.4.1 (#1149)
* *(deps)* Update typescript-eslint monorepo to v5.6.0 (#1148)
* *(deps)* Update dependency vite to v2.7.0 (#1151)
* *(deps)* Update dependency @vitejs/plugin-vue to v1.10.2 (#1150)
* *(deps)* Update dependency netlify-cli to v8.0.16 (#1147)
* *(deps)* Update dependency @vitejs/plugin-legacy to v1.6.4 (#1152)
* *(deps)* Update dependency dompurify to v2.3.4
* *(deps)* Update dependency vite to v2.7.1 (#1154)
* *(deps)* Update sentry-javascript monorepo to v6.16.0 (#1155)
* *(deps)* Update dependency ts-jest to v27.1.1
* *(deps)* Update dependency @vueuse/core to v7.2.1 (#1158)
* *(deps)* Update dependency @vueuse/core to v7.2.2
* *(deps)* Update dependency netlify-cli to v8.0.17
* *(deps)* Update dependency vite-svg-loader to v3.1.1
* *(deps)* Update dependency netlify-cli to v8.0.18
* *(deps)* Update dependency vite-plugin-pwa to v0.11.11
* *(deps)* Update dependency rollup to v2.61.0
* *(deps)* Update dependency jest to v27.4.4 (#1171)
* *(deps)* Update dependency typescript to v4.5.3 (#1169)
* *(deps)* Update dependency marked to v4.0.7 (#1170)
* *(deps)* Update dependency netlify-cli to v8.0.20 (#1168)
* *(deps)* Update dependency rollup to v2.61.1 (#1174)
* *(deps)* Update sentry-javascript monorepo to v6.16.1 (#1175)
* *(deps)* Update vue monorepo to v3.2.26 (#1179)
* *(deps)* Update dependency @vitejs/plugin-vue to v2 (#1180)
* *(deps)* Update dependency sass to v1.45.0 (#1177)
* *(deps)* Update dependency @vueuse/core to v7.3.0 (#1178)
* *(deps)* Update dependency cypress to v9
* *(deps)* Pin dependency @vueuse/router to 7.3.0 (#1182)
* *(deps)* Pin dependency caniuse-lite to 1.0.30001286 (#1185)
* *(deps)* Update dependency esbuild to v0.14.3 (#1187)
* *(deps)* Update dependency postcss to v8.4.5 (#1189)
* *(deps)* Update dependency vite to v2.7.2 (#1191)
* *(deps)* Update dependency netlify-cli to v8.1.1 (#1190)
* *(deps)* Update dependency typescript to v4.5.4 (#1194)
* *(deps)* Update dependency browserslist to v4.19.0 (#1195)
* *(deps)* Update dependency jest to v27.4.5 (#1193)
* *(deps)* Update typescript-eslint monorepo to v5.7.0 (#1192)
* *(deps)* Update dependency esbuild to v0.14.5 (#1200)
* *(deps)* Update dependency browserslist to v4.19.1 (#1198)
* *(deps)* Update dependency @vitejs/plugin-vue to v2.0.1 (#1196)
* *(deps)* Update dependency @github/hotkey to v1.6.1 (#1197)
* *(deps)* Update dependency netlify-cli to v8.1.4 (#1199)
* *(deps)* Update dependency vite-plugin-pwa to v0.11.12 (#1204)
* *(deps)* Update dependency postcss-preset-env to v7.0.2 (#1206)
* *(deps)* Update dependency caniuse-lite to v1.0.30001287 (#1205)
* *(deps)* Update dependency vite to v2.7.3 (#1207)
* *(deps)* Update dependency express to v4.17.2 (#1211)
* *(deps)* Update dependency netlify-cli to v8.1.8
* *(deps)* Update dependency ts-jest to v27.1.2
* *(deps)* Update dependency marked to v4.0.8 (#1217)
* *(deps)* Update dependency @vueuse/router to v7.4.0 (#1216)
* *(deps)* Update dependency caniuse-lite to v1.0.30001291 (#1214)
* *(deps)* Update dependency slugify to v1.6.4 (#1209)
* *(deps)* Update dependency @vueuse/core to v7.4.0
* *(deps)* Update dependency esbuild to v0.14.6 (#1218)
* *(deps)* Update dependency eslint to v8.5.0 (#1213)
* *(deps)* Update dependency codemirror to v5.65.0
* *(deps)* Update dependency vite to v2.7.4
* *(deps)* Update dependency netlify-cli to v8.1.9 (#1221)
* *(deps)* Update dependency netlify-cli to v8.2.0 (#1222)
* *(deps)* Update dependency netlify-cli to v8.2.1 (#1223)
* *(deps)* Update dependency netlify-cli to v8.2.3 (#1224)
* *(deps)* Update typescript-eslint monorepo to v5.8.0 (#1225)
* *(deps)* Update dependency netlify-cli to v8.2.4 (#1226)
* *(deps)* Update dependency sass to v1.45.1 (#1227)
* *(deps)* Update dependency netlify-cli to v8.3.0 (#1228)
* *(deps)* Update dependency netlify-cli to v8.4.1
* *(deps)* Update dependency vue-tsc to v0.30.0
* *(deps)* Update dependency vite to v2.7.6 (#1236)
* *(deps)* Update dependency netlify-cli to v8.4.2 (#1235)
* *(deps)* Update dependency caniuse-lite to v1.0.30001292 (#1234)
* *(deps)* Update dependency cypress to v9.2.0 (#1232)
* *(deps)* Update dependency postcss-preset-env to v7.1.0 (#1237)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.24 (#1238)
* *(deps)* Update dependency esbuild to v0.14.7
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.25 (#1240)
* *(deps)* Update dependency esbuild to v0.14.8 (#1242)
* *(deps)* Update dependency @vueuse/router to v7.4.1 (#1245)
* *(deps)* Update dependency @vueuse/core to v7.4.1 (#1244)
* *(deps)* Update dependency vite to v2.7.7 (#1247)
* *(deps)* Update dependency vue-tsc to v0.30.1 (#1248)
* *(deps)* Update dependency @vue/eslint-config-typescript to v10 (#1243)
* *(deps)* Update dependency rollup to v2.62.0 (#1246)
* *(deps)* Update typescript-eslint monorepo to v5.8.1 (#1253)
* *(deps)* Update dependency vite to v2.7.9 (#1254)
* *(deps)* Update dependency netlify-cli to v8.5.0 (#1255)
* *(deps)* Update dependency date-fns to v2.28.0 (#1256)
* *(deps)* Update dependency caniuse-lite to v1.0.30001294 (#1257)
* *(deps)* Update dependency esbuild to v0.14.9 (#1258)
* *(deps)* Update dependency autoprefixer to v10.4.1 (#1260)
* *(deps)* Update dependency netlify-cli to v8.6.0 (#1259)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.26 (#1263)
* *(deps)* Update dependency vite to v2.7.10 (#1265)
* *(deps)* Update dependency @vueuse/core to v7.4.3 (#1266)
* *(deps)* Update dependency @types/jest to v27.4.0
* *(deps)* Update dependency @vueuse/router to v7.4.3
* *(deps)* Update dependency @vueuse/router to v7.5.1 (#1273)
* *(deps)* Update dependency @vueuse/core to v7.5.1 (#1272)
* *(deps)* Update dependency sass to v1.45.2 (#1271)
* *(deps)* Update dependency esbuild to v0.14.10
* *(deps)* Update dependency caniuse-lite to v1.0.30001295
* *(deps)* Update dependency netlify-cli to v8.6.1
* *(deps)* Update dependency v-tooltip to v4.0.0-beta.5
* *(deps)* Update dependency postcss-preset-env to v7.2.0
* *(deps)* Update dependency slugify to v1.6.5
* *(deps)* Update dependency eslint to v8.6.0
* *(deps)* Update typescript-eslint monorepo to v5.9.0
* *(deps)* Update dependency v-tooltip to v4.0.0-beta.6
* *(deps)* Update dependency rollup to v2.63.0
* *(deps)* Update dependency vue-tsc to v0.30.2
* *(deps)* Update dependency caniuse-lite to v1.0.30001296
* *(deps)* Update dependency netlify-cli to v8.6.3
* *(deps)* Update dependency netlify-cli to v8.6.4
* *(deps)* Update dependency vitest to v0.0.131
* *(deps)* Pin dependency happy-dom to 2.25.1
* *(deps)* Update dependency @vueuse/router to v7.5.2
* *(deps)* Update dependency vitest to v0.0.132
* *(deps)* Update dependency @vueuse/core to v7.5.2
* *(deps)* Update dependency @vueuse/router to v7.5.3 (#1303)
* *(deps)* Update dependency vitest to v0.0.133
* *(deps)* Pin dependency @types/is-touch-device to 1.0.0 (#1308)
* *(deps)* Update dependency vue-advanced-cropper to v2.7.1
* *(deps)* Update dependency netlify-cli to v8.6.5
* *(deps)* Update dependency vitest to v0.0.134 (#1314)
* *(deps)* Update dependency sass to v1.46.0 (#1315)
* *(deps)* Update dependency netlify-cli to v8.6.6 (#1316)
* *(deps)* Update dependency v-tooltip to v4.0.0-beta.8 (#1317)
* *(deps)* Update dependency highlight.js to v11.4.0 (#1319)
* *(deps)* Update dependency netlify-cli to v8.6.8 (#1318)
* *(deps)* Update dependency netlify-cli to v8.6.9 (#1320)
* *(deps)* Update dependency marked to v4.0.9 (#1321)
* *(deps)* Update dependency v-tooltip to v4.0.0-beta.10 (#1324)
* *(deps)* Update dependency vitest to v0.0.135 (#1323)
* *(deps)* Update dependency netlify-cli to v8.6.12 (#1322)
* *(deps)* Update dependency vitest to v0.0.136 (#1325)
* *(deps)* Update dependency caniuse-lite to v1.0.30001297 (#1327)
* *(deps)* Update dependency v-tooltip to v4.0.0-beta.11 (#1326)
* *(deps)* Update dependency autoprefixer to v10.4.2 (#1329)
* *(deps)* Update dependency vitest to v0.0.139 (#1330)
* *(deps)* Update dependency netlify-cli to v8.6.15 (#1331)
* *(deps)* Update dependency v-tooltip to v4.0.0-beta.13 (#1332)
* *(deps)* Update dependency caniuse-lite to v1.0.30001298 (#1334)
* *(deps)* Update dependency sass to v1.47.0 (#1333)
* *(deps)* Update dependency esbuild to v0.14.11 (#1341)
* *(deps)* Update dependency netlify-cli to v8.6.16 (#1343)
* *(deps)* Update dependency v-tooltip to v4.0.0-beta.14 (#1344)
* *(deps)* Update dependency netlify-cli to v8.6.17 (#1345)
* *(deps)* Update dependency v-tooltip to v4.0.0-beta.15 (#1346)
* *(deps)* Update dependency vitest to v0.0.140 (#1348)
* *(deps)* Update typescript-eslint monorepo to v5.9.1 (#1347)
* *(deps)* Update dependency cypress to v9.2.1 (#1349)
* *(deps)* Update dependency netlify-cli to v8.6.18 (#1350)
* *(deps)* Update dependency vite-svg-loader to v3.1.2 (#1351)
* *(deps)* Update dependency netlify-cli to v8.6.19 (#1352)
* *(deps)* Update dependency vitest to v0.0.141 (#1355)
* *(deps)* Update dependency v-tooltip to v4.0.0-beta.16 (#1354)
* *(deps)* Update dependency netlify-cli to v8.6.21 (#1353)
* *(deps)* Update dependency easymde to v2.16.0 (#1356)
* *(deps)* Update dependency caniuse-lite to v1.0.30001299 (#1357)
* *(deps)* Update dependency postcss-preset-env to v7.2.2 (#1358)
* *(deps)* Update dependency eslint-plugin-vue to v8.3.0 (#1360)
* *(deps)* Update dependency netlify-cli to v8.6.22 (#1359)
* *(deps)* Update dependency v-tooltip to v4.0.0-beta.17 (#1362)
* *(deps)* Update dependency postcss-preset-env to v7.2.3 (#1361)
* *(deps)* Update dependency vite-plugin-pwa to v0.11.13 (#1364)
* *(deps)* Update dependency netlify-cli to v8.6.23 (#1363)
* *(deps)* Update dependency vitest to v0.0.142 (#1365)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.28
* *(deps)* Update dependency vitest to v0.1.12
* *(deps)* Update dependency sass to v1.48.0
* *(deps)* Update dependency happy-dom to v2.25.2
* *(deps)* Update dependency marked to v4.0.10
* *(deps)* Update dependency vite to v2.7.11
* *(deps)* Update dependency netlify-cli to v8.8.0 (#1372)
* *(deps)* Update dependency vite to v2.7.12 (#1373)
* *(deps)* Update dependency netlify-cli to v8.8.1 (#1374)
* *(deps)* Update dependency vitest to v0.1.13 (#1375)
* *(deps)* Update dependency netlify-cli to v8.8.2 (#1376)
* *(deps)* Update dependency rollup to v2.64.0 (#1377)
* *(deps)* Update dependency rollup-plugin-visualizer to v5.5.4 (#1381)
* *(deps)* Update dependency vitest to v0.1.16 (#1382)
* *(deps)* Update dependency easymde to v2.16.1
* *(deps)* Update dependency eslint to v8.7.0 (#1384)
* *(deps)* Update dependency vitest to v0.1.17 (#1385)
* *(deps)* Update dependency vue-tsc to v0.30.3 (#1386)
* *(deps)* Update vue monorepo to v3.2.27 (#1387)
* *(deps)* Update dependency vue-tsc to v0.30.4 (#1389)
* *(deps)* Update dependency vue-tsc to v0.30.5 (#1392)
* *(deps)* Update dependency caniuse-lite to v1.0.30001300 (#1391)
* *(deps)* Update dependency vitest to v0.1.18 (#1393)
* *(deps)* Update dependency vitest to v0.1.19
* *(deps)* Update dependency axios to v0.25.0 (#1399)
* *(deps)* Update dependency vitest to v0.1.20 (#1398)
* *(deps)* Update dependency happy-dom to v2.27.0 (#1397)
* *(deps)* Update typescript-eslint monorepo to v5.10.0 (#1396)
* *(deps)* Update dependency vitest to v0.1.21 (#1400)
* *(deps)* Update dependency vite to v2.7.13 (#1401)
* *(deps)* Update dependency cypress to v9.3.1 (#1402)
* *(deps)* Update dependency vue-tsc to v0.30.6 (#1404)
* *(deps)* Update dependency vitest to v0.1.23 (#1405)
* *(deps)* Update dependency sass to v1.49.0 (#1403)
* *(deps)* Update dependency happy-dom to v2.27.2 (#1406)
* *(deps)* Update dependency vitest to v0.1.24
* *(deps)* Update dependency codemirror to v5.65.1 (#1409)
* *(deps)* Update dependency typescript to v4.5.5 (#1410)
* *(deps)* Update dependency esbuild to v0.14.12 (#1413)
* *(deps)* Update dependency happy-dom to v2.28.0 (#1412)
* *(deps)* Update dependency caniuse-lite to v1.0.30001301 (#1414)
* *(deps)* Update dependency vitest to v0.1.25 (#1411)
* *(deps)* Update dependency rollup to v2.65.0 (#1415)
* *(deps)* Update dependency @vue/compat to v3.2.28 (#1416)
* *(deps)* Update dependency vue to v3.2.28 (#1417)
* *(deps)* Update dependency vitest to v0.1.26 (#1418)
* *(deps)* Update dependency @vueuse/router to v7.5.4 (#1420)
* *(deps)* Update dependency @vueuse/core to v7.5.4 (#1419)
* *(deps)* Update dependency @vitejs/plugin-vue to v2.1.0 (#1421)
* *(deps)* Update dependency vitest to v0.1.27 (#1422)
* *(deps)* Update dependency vue-tsc to v0.31.1 (#1423)
* *(deps)* Update dependency esbuild to v0.14.13 (#1426)
* *(deps)* Update dependency rollup to v2.66.0 (#1424)
* *(deps)* Update dependency vitest to v0.2.0 (#1427)
* *(deps)* Update dependency vue-advanced-cropper to v2.8.0 (#1425)
* *(deps)* Update dependency @vue/compat to v3.2.29 (#1428)
* *(deps)* Update dependency vue to v3.2.29 (#1429)
* *(deps)* Update dependency netlify-cli to v8.13.0 (#1431)
* *(deps)* Update sentry-javascript monorepo to v6.17.0 (#1432)
* *(deps)* Update dependency vitest to v0.2.1 (#1433)
* *(deps)* Update typescript-eslint monorepo to v5.10.1 (#1435)
* *(deps)* Update sentry-javascript monorepo to v6.17.1 (#1434)
* *(deps)* Update dependency happy-dom to v2.30.0 (#1437)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.29 (#1438)
* *(deps)* Update dependency rollup to v2.66.1 (#1439)
* *(deps)* Update dependency vitest to v0.2.2 (#1440)
* *(deps)* Update dependency vitest to v0.2.3 (#1441)
* *(deps)* Update dependency @faker-js/faker to v6.0.0-alpha.5 (#1436)
* *(deps)* Update dependency @vueuse/router to v7.5.5 (#1443)
* *(deps)* Update dependency @vueuse/core to v7.5.5 (#1442)
* *(deps)* Update sentry-javascript monorepo to v6.17.2 (#1444)
* *(deps)* Update dependency happy-dom to v2.30.1 (#1445)
* *(deps)* Update dependency esbuild to v0.14.14 (#1446)
* *(deps)* Update dependency caniuse-lite to v1.0.30001302 (#1447)
* *(deps)* Update dependency dompurify to v2.3.5 (#1448)
* *(deps)* Update dependency marked to v4.0.11 (#1449)
* *(deps)* Update dependency vitest to v0.2.4 (#1450)
* *(deps)* Update dependency eslint-plugin-vue to v8.4.0 (#1451)
* *(deps)* Update dependency marked to v4.0.12 (#1452)
* *(deps)* Update dependency caniuse-lite to v1.0.30001303 (#1453)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.30 (#1454)
* *(deps)* Update dependency vitest to v0.2.5 (#1456)
* *(deps)* Update sentry-javascript monorepo to v6.17.3 (#1457)
* *(deps)* Update dependency eslint to v8.8.0 (#1458)
* *(deps)* Update dependency caniuse-lite to v1.0.30001304 (#1459)
* *(deps)* Update dependency happy-dom to v2.31.0 (#1461)
* *(deps)* Update dependency netlify-cli to v8.15.0 (#1463)
* *(deps)* Update dependency postcss-preset-env to v7.3.0 (#1464)
* *(deps)* Update dependency happy-dom to v2.31.1 (#1465)
* *(deps)* Update dependency ufo to v0.7.10 (#1466)
* *(deps)* Update typescript-eslint monorepo to v5.10.2
* *(deps)* Update dependency cypress to v9.4.1
* *(deps)* Update dependency @github/hotkey to v2 (#1471)
* *(deps)* Update dependency esbuild to v0.14.16 (#1469)
* *(deps)* Update dependency sass to v1.49.4 (#1470)
* *(deps)* Update dependency postcss to v8.4.6
* *(deps)* Update dependency sass to v1.49.5
* *(deps)* Update dependency sass to v1.49.6 (#1474)
* *(deps)* Update dependency sass to v1.49.7 (#1475)
* *(deps)* Update dependency caniuse-lite to v1.0.30001305 (#1476)
* *(deps)* Update dependency esbuild to v0.14.17 (#1477)
* *(deps)* Update dependency rollup to v2.67.0 (#1478)
* *(deps)* Update sentry-javascript monorepo to v6.17.4 (#1479)
* *(deps)* Update dependency esbuild to v0.14.18 (#1480)
* *(deps)* Update dependency vitest to v0.2.6 (#1481)
* *(deps)* Update dependency caniuse-lite to v1.0.30001306 (#1482)
* *(deps)* Update dependency postcss-preset-env to v7.3.1 (#1483)
* *(deps)* Update dependency vitest to v0.2.7 (#1485)
* *(deps)* Update dependency caniuse-lite to v1.0.30001307 (#1484)
* *(deps)* Update dependency eslint-plugin-vue to v8.4.1 (#1486)
* *(deps)* Update dependency vue-tsc to v0.31.2 (#1488)
* *(deps)* Update dependency esbuild to v0.14.19 (#1490)
* *(deps)* Update dependency netlify-cli to v8.16.1 (#1492)
* *(deps)* Update dependency caniuse-lite to v1.0.30001309 (#1493)
* *(deps)* Update dependency rollup to v2.67.1 (#1494)
* *(deps)* Update dependency @vue/compat to v3.2.30 (#1495)
* *(deps)* Update dependency vue to v3.2.30 (#1496)
* *(deps)* Update typescript-eslint monorepo to v5.11.0 (#1502)
* *(deps)* Update sentry-javascript monorepo to v6.17.5 (#1501)
* *(deps)* Update dependency esbuild to v0.14.20 (#1500)
* *(deps)* Update dependency vitest to v0.2.8 (#1506)
* *(deps)* Update dependency @vueuse/router to v7.6.0
* *(deps)* Update dependency @vueuse/core to v7.6.0 (#1507)
* *(deps)* Update sentry-javascript monorepo to v6.17.6 (#1513)
* *(deps)* Update dependency caniuse-lite to v1.0.30001310 (#1514)
* *(deps)* Update dependency esbuild to v0.14.21 (#1515)
* *(deps)* Update dependency @vitejs/plugin-legacy to v1.7.0 (#1516)
* *(deps)* Update dependency @vitejs/plugin-vue to v2.2.0 (#1517)
* *(deps)* Update dependency vitest to v0.3.0 (#1519)
* *(deps)* Update dependency @vueuse/router to v7.6.1 (#1521)
* *(deps)* Update dependency rollup to v2.67.2 (#1525)
* *(deps)* Update dependency vitest to v0.3.2 (#1523)
* *(deps)* Update dependency caniuse-lite to v1.0.30001311 (#1524)
* *(deps)* Update dependency @faker-js/faker to v6.0.0-alpha.6 (#1526)
* *(deps)* Update dependency @fortawesome/fontawesome-svg-core to v1.3.0 (#1504)
* *(deps)* Update dependency wait-on to v6.0.1 (#1527)
* *(deps)* Update dependency eslint to v8.9.0 (#1530)
* *(deps)* Update sentry-javascript monorepo to v6.17.7 (#1528)
* *(deps)* Update dependency @vitejs/plugin-legacy to v1.7.1 (#1529)
* *(deps)* Update dependency vitest to v0.3.6
* *(deps)* Update dependency express to v4.17.3 (#1550)
* *(deps)* Update dependency @vueuse/router to v7.6.2 (#1555)
* *(deps)* Update dependency @vue/compat to v3.2.31 (#1553)
* *(deps)* Update dependency vue-tsc to v0.31.4 (#1552)
* *(deps)* Update dependency esbuild to v0.14.22 (#1549)
* *(deps)* Update dependency dompurify to v2.3.6
* *(deps)* Update dependency caniuse-lite to v1.0.30001312
* *(deps)* Update dependency @vueuse/core to v7.6.2
* *(deps)* Update dependency vue to v3.2.31
* *(deps)* Update sentry-javascript monorepo to v6.17.9
* *(deps)* Update dependency vue-advanced-cropper to v2.8.1
* *(deps)* Update dependency axios to v0.26.0
* *(deps)* Update dependency happy-dom to v2.34.0
* *(deps)* Update dependency cypress to v9.5.0
* *(deps)* Update dependency postcss-preset-env to v7.4.1
* *(deps)* Update dependency happy-dom to v2.36.0
* *(deps)* Update typescript-eslint monorepo to v5.12.0
* *(deps)* Update dependency happy-dom to v2.39.1
* *(deps)* Update dependency sass to v1.49.8
* *(deps)* Update dependency rollup to v2.67.3 (#1569)
* *(deps)* Update dependency vitest to v0.4.0 (#1568)
* *(deps)* Update dependency vitest to v0.4.1 (#1570)
* *(deps)* Update dependency vite to v2.8.3
* *(deps)* Update dependency browserslist to v4.19.2
* *(deps)* Update dependency sass to v1.49.8 (#1574)
* *(deps)* Update dependency rollup to v2.67.3
* *(deps)* Update dependency vite to v2.8.4 (#1575)
* *(deps)* Update dependency vitest to v0.4.1 (#1576)
* *(deps)* Update dependency @vitejs/plugin-vue to v2.2.2 (#1577)
* *(deps)* Update dependency esbuild to v0.14.23
* *(deps)* Update dependency browserslist to v4.19.3 (#1579)
* *(deps)* Update dependency vitest to v0.4.2 (#1580)
* *(deps)* Update dependency @faker-js/faker to v6.0.0-alpha.7 (#1587)
* *(deps)* Update dependency netlify-cli to v8.19.3 (#1589)
* *(deps)* Update dependency vitest to v0.4.3 (#1591)
* *(deps)* Update dependency vitest to v0.5.0 (#1592)
* *(deps)* Update dependency netlify-cli to v9 (#1590)
* *(deps)* Update dependency codemirror to v5.65.2 (#1593)
* *(deps)* Update typescript-eslint monorepo to v5.12.1 (#1595)
* *(deps)* Update dependency vitest to v0.5.1 (#1596)
* *(deps)* Update dependency rollup to v2.68.0 (#1597)
* *(deps)* Update dependency eslint-plugin-vue to v8.5.0 (#1598)
* *(deps)* Update dependency vitest to v0.5.3 (#1599)
* *(deps)* Update dependency happy-dom to v2.41.0 (#1600)
* *(deps)* Update dependency vitest to v0.5.4 (#1602)
* *(deps)* Update workbox monorepo to v6.5.0 (#1603)
* *(deps)* Update dependency vitest to v0.5.5 (#1604)
* *(deps)* Update sentry-javascript monorepo to v6.18.0 (#1605)
* *(deps)* Update dependency sass to v1.49.9 (#1606)
* *(deps)* Update dependency postcss to v8.4.7 (#1607)
* *(deps)* Update dependency vue-tsc to v0.32.0 (#1608)
* *(deps)* Update dependency rollup-plugin-visualizer to v5.6.0 (#1609)
* *(deps)* Update dependency ufo to v0.7.11 (#1610)
* *(deps)* Update dependency vitest to v0.5.7 (#1612)
* *(deps)* Update dependency eslint to v8.10.0 (#1611)
* *(deps)* Update dependency @vueuse/router to v7.7.0 (#1614)
* *(deps)* Update dependency @vueuse/core to v7.7.0 (#1613)
* *(deps)* Update dependency vitest to v0.5.8 (#1618)
* *(deps)* Update dependency netlify-cli to v9.8.3 (#1619)
* *(deps)* Update sentry-javascript monorepo to v6.18.1 (#1621)
* *(deps)* Update dependency vue-router to v4.0.13 (#1620)
* *(deps)* Update dependency vite to v2.8.5 (#1623)
* *(deps)* Update dependency @vitejs/plugin-vue to v2.2.4 (#1622)
* *(deps)* Update typescript-eslint monorepo to v5.13.0 (#1624)
* *(deps)* Update dependency typescript to v4.6.2 (#1626)
* *(deps)* Update dependency cypress to v9.5.1 (#1625)
* *(deps)* Update dependency vitest to v0.5.9 (#1627)
* *(deps)* Update dependency happy-dom to v2.43.0 (#1628)
* *(deps)* Update dependency vite to v2.8.6 (#1630)
* *(deps)* Update dependency rollup to v2.69.0 (#1631)
* *(deps)* Update dependency vue-tsc to v0.32.1 (#1629)
* *(deps)* Update dependency postcss-preset-env to v7.4.2 (#1633)
* *(deps)* Update dependency happy-dom to v2.43.1 (#1632)
* *(deps)* Update dependency esbuild to v0.14.24 (#1634)
* *(deps)* Update dependency caniuse-lite to v1.0.30001313 (#1636)
* *(deps)* Update dependency esbuild to v0.14.25 (#1637)
* *(deps)* Update workbox monorepo to v6.5.1 (#1635)
* *(deps)* Update dependency rollup to v2.69.1 (#1638)
* *(deps)* Update dependency happy-dom to v2.45.0 (#1640)
* *(deps)* Update dependency @vueuse/router to v7.7.1 (#1642)
* *(deps)* Update dependency @vueuse/core to v7.7.1 (#1641)
* *(deps)* Update dependency rollup to v2.69.2 (#1643)
* *(deps)* Update dependency flatpickr to v4.6.10 (#1644)
* *(deps)* Update dependency rollup to v2.70.0 (#1648)
* *(deps)* Update dependency browserslist to v4.20.0 (#1645)
* *(deps)* Update dependency netlify-cli to v9.12.3 (#1646)
* *(deps)* Update dependency postcss to v8.4.8 (#1647)
* *(deps)* Update dependency happy-dom to v2.45.1 (#1649)
* *(deps)* Update dependency vitest to v0.6.0 (#1651)
* *(deps)* Update dependency happy-dom to v2.46.0 (#1650)
* *(deps)* Update typescript-eslint monorepo to v5.14.0 (#1652)
* *(deps)* Update dependency @faker-js/faker to v6.0.0-beta.0 (#1653)
* *(deps)* Update dependency caniuse-lite to v1.0.30001314 (#1654)
* *(deps)* Update sentry-javascript monorepo to v6.18.2 (#1655)
* *(deps)* Update dependency axios to v0.26.1 (#1656)
* *(deps)* Update dependency caniuse-lite to v1.0.30001315 (#1657)
* *(deps)* Update dependency happy-dom to v2.46.3 (#1658)
* *(deps)* Update dependency flatpickr to v4.6.11 (#1659)
* *(deps)* Update dependency highlight.js to v11.5.0 (#1662)
* *(deps)* Update dependency eslint to v8.11.0 (#1661)
* *(deps)* Update dependency vue-tsc to v0.33.1 (#1665)
* *(deps)* Update dependency @vueuse/core to v8 (#1663)
* *(deps)* Update dependency vue-router to v4.0.14 (#1660)
* *(deps)* Update dependency @vueuse/router to v8 (#1664)
* *(deps)* Update dependency vitest to v0.6.1 (#1666)
* *(deps)* Update dependency rollup to v2.70.1 (#1671)
* *(deps)* Update dependency esbuild to v0.14.26 (#1670)
* *(deps)* Update dependency netlify-cli to v9.13.0 (#1667)
* *(deps)* Update dependency @vueuse/core to v8.0.1 (#1668)
* *(deps)* Update dependency @vueuse/router to v8.0.1 (#1669)
* *(deps)* Update dependency caniuse-lite to v1.0.30001316 (#1672)
* *(deps)* Update typescript-eslint monorepo to v5.15.0 (#1675)
* *(deps)* Update dependency happy-dom to v2.47.0 (#1673)
* *(deps)* Update dependency vue-tsc to v0.33.2 (#1674)
* *(deps)* Update dependency cypress to v9.5.2 (#1676)
* *(deps)* Update dependency caniuse-lite to v1.0.30001317 (#1679)
* *(deps)* Update dependency esbuild to v0.14.27 (#1678)
* *(deps)* Update font awesome to v6 (major) (#1505)
* *(deps)* Update dependency autoprefixer to v10.4.3 (#1682)
* *(deps)* Update dependency postcss to v8.4.11 (#1684)
* *(deps)* Update dependency ufo to v0.8.0 (#1685)
* *(deps)* Update dependency browserslist to v4.20.2 (#1683)
* *(deps)* Update dependency @faker-js/faker to v6.0.0 (#1681)
* *(deps)* Update dependency autoprefixer to v10.4.4 (#1686)
* *(deps)* Update dependency happy-dom to v2.49.0 (#1680)
* *(deps)* Update dependency postcss to v8.4.12 (#1687)
* *(deps)* Update dependency ufo to v0.8.1 (#1689)
* *(deps)* Update dependency vitest to v0.6.3 (#1688)
* *(deps)* Update dependency @vueuse/core to v8.1.1 (#1690)
* *(deps)* Update dependency vitest to v0.7.0 (#1692)
* *(deps)* Update dependency @vueuse/router to v8.1.1 (#1691)
* *(deps)* Update dependency @types/flexsearch to v0.7.3 (#1677)
* *(deps)* Update dependency vitest to v0.7.4 (#1693)
* *(deps)* Update dependency caniuse-lite to v1.0.30001319 (#1695)
* *(deps)* Update dependency vitest to v0.7.6 (#1698)
* *(deps)* Update dependency @vueuse/router to v8.1.2 (#1697)
* *(deps)* Update yarn to v1.22.18 (#1694)
* *(deps)* Update dependency @vueuse/core to v8.1.2 (#1696)
* *(deps)* Update dependency postcss-preset-env to v7.4.3 (#1699)
* *(deps)* Update dependency vue-tsc to v0.33.5 (#1701)
* *(deps)* Update dependency netlify-cli to v9.13.3 (#1700)
* *(deps)* Update dependency happy-dom to v2.49.1 (#1703)
* *(deps)* Update dependency vitest to v0.7.7 (#1702)
* *(deps)* Update dependency happy-dom to v2.49.2 (#1704)
* *(deps)* Update sentry-javascript monorepo to v6.19.0 (#1705)
* *(deps)* Update dependency vue-tsc to v0.33.6 (#1706)
* *(deps)* Update typescript-eslint monorepo to v5.16.0 (#1707)
* *(deps)* Update sentry-javascript monorepo to v6.19.1 (#1708)
* *(deps)* Update font awesome to v6.1.1 (#1710)
* *(deps)* Update dependency happy-dom to v2.50.0 (#1711)
* *(deps)* Update dependency vue-tsc to v0.33.7 (#1712)
* *(deps)* Update dependency vitest to v0.7.8 (#1713)
* *(deps)* Update dependency vitest to v0.7.10 (#1714)
* *(deps)* Update sentry-javascript monorepo to v6.19.2 (#1715)
* *(deps)* Update dependency caniuse-lite to v1.0.30001320 (#1716)
* *(deps)* Update dependency vue-tsc to v0.33.9 (#1719)
* *(deps)* Update dependency typescript to v4.6.3 (#1717)
* *(deps)* Update dependency vitest to v0.7.11 (#1718)
* *(deps)* Update dependency @vueuse/core to v8.2.0 (#1720)
* *(deps)* Update dependency esbuild to v0.14.28 (#1723)
* *(deps)* Update dependency @vueuse/router to v8.2.0 (#1721)
* *(deps)* Update dependency eslint to v8.12.0 (#1722)
* *(deps)* Update dependency vitest to v0.7.12 (#1724)
* *(deps)* Update workbox monorepo to v6.5.2 (#1725)
* *(deps)* Update dependency netlify-cli to v9.13.5 (#1726)
* *(deps)* Update typescript-eslint monorepo to v5.17.0 (#1727)
* *(deps)* Update dependency cypress to v9.5.3 (#1729)
* *(deps)* Update dependency @faker-js/faker to v6.1.1 (#1728)
* *(deps)* Update dependency happy-dom to v2.51.0 (#1733)
* *(deps)* Update dependency vitest to v0.8.0 (#1731)
* *(deps)* Update dependency caniuse-lite to v1.0.30001322 (#1730)
* *(deps)* Update sentry-javascript monorepo to v6.19.3 (#1735)
* *(deps)* Update dependency esbuild to v0.14.29 (#1736)
* *(deps)* Update dependency vite to v2.9.0 (#1742)
* *(deps)* Update dependency happy-dom to v2.52.0 (#1741)
* *(deps)* Update dependency vitest to v0.8.1 (#1740)
* *(deps)* Update dependency @vitejs/plugin-legacy to v1.8.0 (#1738)
* *(deps)* Update dependency @vitejs/plugin-vue to v2.3.0 (#1739)
* *(deps)* Update dependency caniuse-lite to v1.0.30001323 (#1748)
* *(deps)* Update dependency @vueuse/core to v8.2.2 (#1744)
* *(deps)* Update dependency sass to v1.49.10 (#1747)
* *(deps)* Update dependency happy-dom to v2.53.0 (#1749)
* *(deps)* Update dependency @vitejs/plugin-vue to v2.3.1 (#1746)
* *(deps)* Update dependency vite-svg-loader to v3.2.0 (#1743)
* *(deps)* Update dependency @vueuse/router to v8.2.2 (#1745)
* *(deps)* Update dependency vite to v2.9.1 (#1750)
* *(deps)* Update dependency ufo to v0.8.3 (#1754)
* *(deps)* Update dependency happy-dom to v2.54.0 (#1753)
* *(deps)* Update dependency @vueuse/core to v8.2.3 (#1751)
* *(deps)* Update dependency @vueuse/router to v8.2.3 (#1752)
* *(deps)* Update dependency happy-dom to v2.55.0 (#1755)
* *(deps)* Update dependency vitest to v0.8.2 (#1756)
* *(deps)* Update dependency esbuild to v0.14.30 (#1758)
* *(deps)* Update dependency sass to v1.49.11 (#1757)
* *(deps)* Update dependency caniuse-lite to v1.0.30001324 (#1759)
* *(deps)* Pin dependencies (#1760)
* *(deps)* Update dependency blurhash to v1.1.5 (#1761)
* *(deps)* Update dependency vitest to v0.8.3 (#1762)
* *(deps)* Update dependency vitest to v0.8.4 (#1763)
* *(deps)* Update dependency @vueuse/core to v8.2.4 (#1764)
* *(deps)* Update dependency @vueuse/router to v8.2.4 (#1765)
* *(deps)* Update dependency netlify-cli to v9.16.1 (#1766)
* *(deps)* Update dependency esbuild to v0.14.31 (#1767)
* *(deps)* Update dependency caniuse-lite to v1.0.30001325 (#1768)
* *(deps)* Update dependency @faker-js/faker to v6.1.2 (#1770)
* *(deps)* Update typescript-eslint monorepo to v5.18.0 (#1771)
* *(deps)* Update sentry-javascript monorepo to v6.19.4 (#1772)
* *(deps)* Upgrade minimist to 1.2.6
* *(deps)* Update dependency esbuild to v0.14.32 (#1773)
* *(deps)* Update dependency eslint-plugin-vue to v8.6.0 (#1774)
* *(deps)* Update dependency @vueuse/core to v8.2.5 (#1775)
* *(deps)* Update sentry-javascript monorepo to v6.19.5 (#1780)
* *(deps)* Update dependency esbuild to v0.14.34 (#1779)
* *(deps)* Update dependency sass to v1.50.0 (#1778)
* *(deps)* Update sentry-javascript monorepo to v6.19.6 (#1781)
* *(deps)* Update dependency @vueuse/router to v8.2.5 (#1776)
* *(deps)* Update dependency caniuse-lite to v1.0.30001327 (#1783)
* *(deps)* Update dependency marked to v4.0.13 (#1782)
* *(deps)* Update dependency eslint to v8.13.0 (#1784)
* *(deps)* Update dependency vue-tsc to v0.34.0
* *(deps)* Update dependency vue-tsc to v0.34.1
* *(deps)* Update dependency vue-tsc to v0.34.2 (#1801)
* *(deps)* Update dependency vue-tsc to v0.34.4
* *(deps)* Update dependency vue-tsc to v0.34.5
* *(deps)* Update dependency highlight.js to v11.5.1
* *(deps)* Update dependency marked to v4.0.14
* *(deps)* Update dependency netlify-cli to v9.16.5
* *(deps)* Update typescript-eslint monorepo to v5.19.0
* *(deps)* Update dependency cypress to v9.5.4
* *(deps)* Update dependency vue-flatpickr-component to v9.0.6
* *(deps)* Update dependency @vitejs/plugin-legacy to v1.8.1
* *(deps)* Update dependency vue to v3.2.32
* *(deps)* Update dependency vue-tsc to v0.34.6
* *(deps)* Update dependency caniuse-lite to v1.0.30001331
* *(deps)* Update dependency esbuild to v0.14.36
* *(deps)* Update dependency vite to v2.9.3
* *(deps)* Update dependency vite to v2.9.4
* *(deps)* Update dependency rollup to v2.70.2
* *(deps)* Update dependency vite to v2.9.5
* *(deps)* Update dependency @vueuse/router to v8.2.6
* *(deps)* Update dependency caniuse-lite to v1.0.30001332
* *(deps)* Update dependency vue to v3.2.33
* *(deps)* Update workbox monorepo to v6.5.3 (#1820)
* *(deps)* Update dependency codemirror to v5.65.3 (#1841)
* *(deps)* Update typescript-eslint monorepo to v5.20.0 (#1840)
* *(deps)* Update dependency vite-plugin-pwa to v0.12.0 (#1839)
* *(deps)* Update dependency vue-tsc to v0.34.7 (#1838)
* *(deps)* Update dependency sass to v1.50.1 (#1837)
* *(deps)* Update dependency @vueuse/core to v8.2.6 (#1828)
* *(deps)* Update dependency flatpickr to v4.6.13 (#1826)
* *(deps)* Update dependency @vueuse/router to v8.3.0 (#1844)
* *(deps)* Update dependency @vueuse/core to v8.3.0 (#1843)
* *(deps)* Update dependency vue-tsc to v0.34.8 (#1847)
* *(deps)* Update dependency esbuild to v0.14.37 (#1846)
* *(deps)* Update node.js to v18 (#1845)
* *(deps)* Update dependency vue-tsc to v0.34.9 (#1848)
* *(deps)* Update dependency @faker-js/faker to v6.2.0 (#1851)
* *(deps)* Update dependency @vueuse/router to v8.3.1 (#1850)
* *(deps)* Update dependency esbuild to v0.14.38 (#1852)
* *(deps)* Update dependency @vueuse/core to v8.3.1 (#1849)
* *(deps)* Update dependency eslint-plugin-vue to v8.7.0 (#1853)
* *(deps)* Update dependency eslint-plugin-vue to v8.7.1 (#1854)
* *(deps)* Update dependency vitest to v0.9.4
* *(deps)* Update dependency vue-tsc to v0.34.10
* *(deps)* Update dependency autoprefixer to v10.4.5 (#1858)
* *(deps)* Update dependency vite-svg-loader to v3.3.0 (#1859)
* *(deps)* Update dependency cypress to v9.6.0 (#1866)
* *(deps)* Update typescript-eslint monorepo to v5.21.0 (#1867)
* *(deps)* Update dependency eslint to v8.14.0 (#1855)
* *(deps)* Update dependency netlify-cli to v10 (#1862)
* *(deps)* Update dependency vitest to v0.10.0 (#1864)
* *(deps)* Update dependency express to v4.18.0 (#1868)
* *(deps)* Update dependency sass to v1.51.0 (#1869)
* *(deps)* Update dependency browserslist to v4.20.3 (#1860)
* *(deps)* Update dependency happy-dom to v3 (#1870)
* *(deps)* Update sentry-javascript monorepo to v6.19.7 (#1871)
* *(deps)* Update dependency postcss-preset-env to v7.4.4 (#1872)
* *(deps)* Update dependency vite to v2.9.6 (#1873)
* *(deps)* Update dependency happy-dom to v3.1.0 (#1874)
* *(deps)* Update dependency axios to v0.27.2 (#1865)
* *(deps)* Bump ejs from 3.1.6 to 3.1.7 (#49)
* *(deps)* Update dependency caniuse-lite to v1.0.30001334 (#1875)
* *(deps)* Update dependency typescript to v4.6.4 (#1876)
* *(deps)* Update dependency vue-tsc to v0.34.11 (#1877)
* *(deps)* Update dependency express to v4.18.1 (#1878)
* *(deps)* Update dependency netlify-cli to v10.1.0 (#1882)
* *(deps)* Update dependency autoprefixer to v10.4.6 (#1881)
* *(deps)* Update dependency rollup to v2.71.1 (#1880)
* *(deps)* Update dependency postcss to v8.4.13 (#1879)
* *(deps)* Update dependency caniuse-lite to v1.0.30001335 (#1883)
* *(deps)* Update dependency marked to v4.0.15 (#1884)
* *(deps)* Update dependency @vitejs/plugin-legacy to v1.8.2 (#1885)
* *(deps)* Update dependency vite to v2.9.7 (#1886)
* *(deps)* Update dependency @faker-js/faker to v6.3.0 (#1887)
* *(deps)* Update dependency autoprefixer to v10.4.7 (#1888)
* *(deps)* Update dependency vitest to v0.10.1 (#1889)
* *(deps)* Update typescript-eslint monorepo to v5.22.0 (#1890)
* *(deps)* Update dependency @faker-js/faker to v6.3.1 (#1891)
* *(deps)* Update dependency postcss-preset-env to v7.5.0 (#1892)
* *(deps)* Update dependency vitest to v0.10.2 (#1893)
* *(deps)* Update dependency @vueuse/core to v8.4.0 (#1895)
* *(deps)* Update dependency @vueuse/router to v8.4.0 (#1896)
* *(deps)* Update dependency vue-router to v4.0.15 (#1897)
* *(deps)* Update dependency @vueuse/core to v8.4.1 (#1898)
* *(deps)* Update dependency @vueuse/router to v8.4.1 (#1899)
* *(deps)* Update dependency @vitejs/plugin-vue to v2.3.2 (#1900)
* *(deps)* Update dependency vite to v2.9.8 (#1901)
* *(deps)* Update dependency rollup to v2.72.0 (#1902)
* *(deps)* Update dependency caniuse-lite to v1.0.30001336 (#1903)
* *(deps)* Update dependency @vueuse/router to v8.4.2 (#1907)
* *(deps)* Update dependency vitest to v0.10.4 (#1906)
* *(deps)* Update dependency happy-dom to v3.1.1 (#1904)
* *(deps)* Update dependency @vueuse/core to v8.4.2 (#1905)
* *(deps)* Update dependency caniuse-lite to v1.0.30001337 (#1908)
* *(deps)* Update dependency caniuse-lite to v1.0.30001338 (#1909)
* *(deps)* Update dependency vitest to v0.10.5 (#1910)
* *(deps)* Update dependency ufo to v0.8.4 (#1911)
* *(deps)* Update dependency eslint to v8.15.0 (#1912)
* *(deps)* Update dependency rollup to v2.72.1 (#1913)
* *(deps)* Update dependency @types/sortablejs to v1.13.0 (#1915)
* *(deps)* Update dependency netlify-cli to v10.3.0 (#1916)
* *(deps)* Update typescript-eslint monorepo to v5.23.0 (#1918)
* *(deps)* Update dependency cypress to v9.6.1 (#1917)
* *(deps)* Update dependency vue-tsc to v0.34.12 (#1920)
* *(deps)* Update dependency happy-dom to v3.2.0 (#1921)
* *(deps)* Update dependency rollup to v2.73.0 (#1946)
* *(deps)* Update dependency vue-tsc to v0.34.13 (#1945)
* *(deps)* Update dependency esbuild to v0.14.39 (#1944)
* *(deps)* Update dependency dompurify to v2.3.8 (#1943)
* *(deps)* Update dependency vite to v2.9.9 (#1942)
* *(deps)* Update dependency @vitejs/plugin-vue to v2.3.3 (#1941)
* *(deps)* Update dependency vue-tsc to v0.34.15 (#1948)
* *(deps)* Update dependency happy-dom to v3.2.1 (#1949)
* *(deps)* Update vueuse to v8.5.0 (#1953)
* *(deps)* Update dependency caniuse-lite to v1.0.30001341 (#1951)
* *(deps)* Update dependency netlify-cli to v10.3.1 (#1952)
* *(deps)* Update dependency happy-dom to v3.2.2 (#1954)
* *(deps)* Update typescript-eslint monorepo to v5.24.0 (#1955)
* *(deps)* Update dependency postcss to v8.4.14 (#1959)
* *(deps)* Update typescript-eslint monorepo to v5.25.0 (#1957)
* *(deps)* Update dependency marked to v4.0.16 (#1956)
* *(deps)* Update dependency eslint-plugin-vue to v9 (#1958)
* *(deps)* Update dependency vue to v3.2.34 (#1960)
* *(deps)* Update dependency happy-dom to v4
* *(deps)* Update dependency postcss-preset-env to v7.6.0
* *(deps)* Update dependency rollup to v2.74.1
* *(deps)* Update dependency sass to v1.52.0 (#1965)
* *(deps)* Update dependency esbuild to v0.14.42 (#1998)
* *(deps)* Update dependency sass to v1.52.1 (#1999)
* *(deps)* Update dependency vue to v3.2.36 (#2001)
* *(deps)* Update dependency eslint-plugin-vue to v9.1.0 (#2014)
* *(deps)* Update dependency happy-dom to v4.1.0 (#2004)
* *(deps)* Update dependency postcss-preset-env to v7.7.0 (#2005)
* *(deps)* Update vueuse to v8.6.0 (#2010)
* *(deps)* Update dependency typescript to v4.7.2 (#2007)
* *(deps)* Update dependency vue-tsc to v0.35.2 (#2008)
* *(deps)* Update typescript-eslint monorepo to v5.27.0 (#2009)
* *(deps)* Update dependency vitest to v0.13.1 (#1914)
* *(deps)* Update dependency happy-dom to v5 (#2012)
* *(deps)* Update dependency eslint to v8.16.0 (#2003)
* *(deps)* Update dependency rollup to v2.75.5 (#2006)
* *(deps)* Update dependency codemirror to v5.65.5
* *(deps)* Update dependency vue-tsc to v0.36.0 (#2016)
* *(deps)* Update dependency sass to v1.52.2 (#2017)
* *(deps)* Update dependency postcss-preset-env to v7.7.1 (#2018)
* *(deps)* Update dependency eslint to v8.17.0 (#2020)
* *(deps)* Update dependency browserslist to v4.20.4 (#2029)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.36 (#2025)
* *(deps)* Update dependency vitest to v0.14.1 (#2022)
* *(deps)* Update dependency vue to v3.2.37 (#2026)
* *(deps)* Update typescript-eslint monorepo to v5.27.1 (#2028)
* *(deps)* Update dependency vite to v2.9.10 (#2027)
* *(deps)* Update sentry-javascript monorepo to v7 (major) (#2013)
* *(deps)* Update dependency rollup to v2.75.6 (#2030)
* *(deps)* Update dependency vue-tsc to v0.37.3 (#2021)
* *(deps)* Update dependency typescript to v4.7.3 (#2019)
* *(deps)* Update dependency esbuild to v0.14.43 (#2033)
* *(deps)* Update yarn to v1.22.19 (#2032)
* *(deps)* Update dependency @fortawesome/vue-fontawesome to v3.0.0 (#2031)
* *(deps)* Update dependency cypress to v10 (#2015)
* *(deps)* Update dependency codemirror to v6 (#2035)
* *(deps)* Update sentry-javascript monorepo to v7.1.1 (#2034)
* *(deps)* Update dependency happy-dom to v5.2.0 (#2037)
* *(deps)* Update dependency vue-router to v4.0.16 (#2039)
* *(deps)* Update dependency vitest to v0.14.2 (#2041)
* *(deps)* Update dependency sass to v1.52.3 (#2038)
* *(deps)* Update dependency eslint-plugin-vue to v9.1.1 (#2043)
* *(deps)* Update dependency cypress to v10.1.0 (#2042)
* *(deps)* Update dependency vite to v2.9.12 (#2040)
* *(deps)* Update dependency caniuse-lite to v1.0.30001352 (#2045)
* *(deps)* Update dependency vue-tsc to v0.37.5 (#2044)
* *(deps)* Update dependency marked to v4.0.17 (#2046)
* *(deps)* Update dependency @vue/eslint-config-typescript to v11 (#2047)
* *(deps)* Update dependency vue-tsc to v0.37.7 (#2048)
* *(deps)* Update dependency happy-dom to v5.3.1 (#2052)
* *(deps)* Update dependency vue-tsc to v0.37.8 (#2051)
* *(deps)* Update typescript-eslint monorepo to v5.28.0 (#2049)
* *(deps)* Update dependency vitest to v0.15.0 (#2053)
* *(deps)* Update dependency vitest to v0.15.1 (#2054)
* *(deps)* Update dependency @4tw/cypress-drag-drop to v2.2.0 (#2058)
* *(deps)* Update dependency vue-tsc to v0.37.9 (#2057)
* *(deps)* Update dependency vue-advanced-cropper to v2.8.2 (#2056)
* *(deps)* Update dependency esbuild to v0.14.44 (#2055)
* *(deps)* Update dependency vite-svg-loader to v3.4.0 (#2059)
* *(deps)* Update vueuse to v8.7.3 (#2060)
* *(deps)* Update dependency esbuild to v0.14.45 (#2061)
* *(deps)* Update dependency typescript to v4.7.4 (#2064)
* *(deps)* Update dependency @fortawesome/vue-fontawesome to v3.0.1 (#2063)
* *(deps)* Update sentry-javascript monorepo to v7.2.0 (#2062)
* *(deps)* Update dependency eslint to v8.18.0 (#2065)
* *(deps)* Update vueuse to v8.7.4 (#2066)
* *(deps)* Update dependency vue-tsc to v0.38.0 (#2067)
* *(deps)* Update dependency esbuild to v0.14.46 (#2068)
* *(deps)* Update dependency vue-tsc to v0.38.1 (#2069)
* *(deps)* Update dependency rollup to v2.75.7 (#2071)
* *(deps)* Update dependency caniuse-lite to v1.0.30001357 (#2070)
* *(deps)* Update dependency vitest to v0.15.2
* *(deps)* Update typescript-eslint monorepo to v5.29.0
* *(deps)* Update dependency esbuild to v0.14.47
* *(deps)* Update vueuse to v8.7.5
* *(deps)* Update dependency @faker-js/faker to v7
* *(deps)* Update dependency sass to v1.53.0
* *(deps)* Update dependency postcss-preset-env to v7.7.2 (#2079)
* *(deps)* Update typescript-eslint monorepo to v5.30.0 (#2088)
* *(deps)* Update dependency cypress to v10.3.0 (#2087)
* *(deps)* Update dependency vite to v2.9.13 (#2086)
* *(deps)* Update dependency vue-tsc to v0.38.2 (#2084)
* *(deps)* Update dependency happy-dom to v5.3.4 (#2083)
* *(deps)* Update sentry-javascript monorepo to v7.3.1 (#2081)
* *(deps)* Update dependency vue-advanced-cropper to v2.8.3 (#2080)
* *(deps)* Update dependency esbuild to v0.14.48 (#2089)
* *(deps)* Update dependency vite-plugin-pwa to v0.12.1 (#2090)
* *(deps)* Update dependency vitest to v0.16.0 (#2082)
* *(deps)* Update dependency @4tw/cypress-drag-drop to v2.2.1 (#2085)
* *(deps)* Update dependency happy-dom to v5.4.0 (#2092)
* *(deps)* Update dependency vite-plugin-pwa to v0.12.2 (#2091)
* *(deps)* Update dependency eslint to v8.19.0 (#2096)
* *(deps)* Update typescript-eslint monorepo to v5.30.3 (#2095)
* *(deps)* Update sentry-javascript monorepo to v7.4.1 (#2094)
* *(deps)* Update dependency happy-dom to v6
* *(deps)* Update typescript-eslint monorepo to v5.30.4
* *(deps)* Update dependency vitest to v0.17.0
* *(deps)* Update caniuse-and-related (#2100)
* *(deps)* Update dependency vue-router to v4.1.0 (#2101)
* *(deps)* Update sentry-javascript monorepo to v7.5.0 (#2102)
* *(deps)* Update dependency netlify-cli to v10.9.0 (#2024)
* *(deps)* Update dependency @cypress/vue to v3.1.2 (#2122)
* *(deps)* Update dependency dompurify to v2.3.9 (#2131)
* *(deps)* Update dependency @kyvg/vue3-notification to v2.3.5 (#2130)
* *(deps)* Update typescript-eslint monorepo to v5.30.6 (#2129)
* *(deps)* Update dependency vue-tsc to v0.38.5 (#2128)
* *(deps)* Update dependency vite-plugin-pwa to v0.12.3 (#2127)
* *(deps)* Update dependency happy-dom to v6.0.3 (#2125)
* *(deps)* Update dependency esbuild to v0.14.49 (#2124)
* *(deps)* Update dependency @vue/test-utils to v2.0.2 (#2123)
* *(deps)* Update dependency @cypress/vite-dev-server to v2.2.3 (#2121)
* *(deps)* Update dependency vite to v2.9.14 (#2126)
* *(deps)* Update dependency marked to v4.0.18 (#2133)
* *(deps)* Update dependency ufo to v0.8.5 (#2134)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.37 (#2135)
* *(deps)* Update dependency eslint-plugin-vue to v9.2.0 (#2137)
* *(deps)* Update dependency rollup to v2.76.0 (#2138)
* *(deps)* Update dependency vitest to v0.18.0 (#2139)
* *(deps)* Update dependency highlight.js to v11.6.0 (#2140)
* *(deps)* Update dependency vue-router to v4.1.2 (#2136)
* *(deps)* Update dependency rollup-plugin-visualizer to v5.7.0 (#2141)
* *(deps)* Update vueuse to v8.9.2 (#2143)
* *(deps)* Update sentry-javascript monorepo to v7.6.0 (#2142)
* *(deps)* Update vueuse to v8.9.3 (#2148)
* *(deps)* Update dependency vitest to v0.18.1
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.38
* *(deps)* Update dependency rollup-plugin-visualizer to v5.7.1
* *(deps)* Update sentry-javascript monorepo to v7.7.0
* *(deps)* Update dependency vue-tsc to v0.38.7
* *(deps)* Update dependency rollup to v2.77.0
* *(deps)* Update dependency happy-dom to v6.0.4 (#2164)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.39 (#2163)
* *(deps)* Update vueuse to v8.9.4 (#2161)
* *(deps)* Update dependency eslint to v8.20.0 (#2159)
* *(deps)* Update dependency vite to v3
* *(deps)* Update dependency vite to v3 (#2149)
* *(deps)* Update dependency @vitejs/plugin-vue to v3.0.1 (#2147)
* *(deps)* Update typescript-eslint monorepo to v5.30.7 (#2168)
* *(deps)* Update dependency vite to v3.0.2 (#2166)
* *(deps)* Update dependency dompurify to v2.3.10 (#2167)
* *(deps)* Update dependency vue-i18n to v9.2.0-beta.40 (#2172)
* *(deps)* Update dependency cypress to v10.3.1 (#2175)
* *(deps)* Update dependency vue-tsc to v0.38.9 (#2162)
* *(deps)* Update dependency @github/hotkey to v2.0.1 (#2182)
* *(deps)* Update dependency vitest to v0.20.2
* *(deps)* Update dependency vitest to v0.20.2 (#2188)
* *(deps)* Update font awesome to v6.1.2 (#2198)
* *(deps)* Update dependency vite to v3.0.4 (#2193)
* *(deps)* Update dependency eslint-plugin-vue to v9.3.0 (#2192)
* *(deps)* Update dependency esbuild to v0.14.51 (#2191)
* *(deps)* Update dependency netlify-cli to v10.13.0 (#2190)
* *(deps)* Update caniuse-and-related (#2189)
* *(deps)* Update dependency sass to v1.54.0 (#2186)
* *(deps)* Update dependency date-fns to v2.29.1 (#2185)
* *(deps)* Update dependency autoprefixer to v10.4.8 (#2202)
* *(deps)* Update dependency rollup to v2.77.2 (#2203)
* *(deps)* Update dependency vue-tsc to v0.39.4 (#2187)
* *(deps)* Update dependency @kyvg/vue3-notification to v2.3.6 (#2205)
* *(deps)* Update typescript-eslint monorepo to v5.31.0 (#2207)
* *(deps)* Update dependency vue-router to v4.1.3 (#2206)
* *(deps)* Update vueuse to v9 (major) (#2209)
* *(deps)* Update sentry-javascript monorepo to v7.8.0 (#2208)
* *(deps)* Update dependency vue-i18n to v9.2.0 (#2210)
* *(deps)* Update dependency eslint to v8.21.0 (#2211)
* *(deps)* Update sentry-javascript monorepo to v7.8.1 (#2212)
* *(deps)* Update typescript-eslint monorepo to v5.32.0 (#2213)
* *(deps)* Update workbox monorepo to v6.5.4 (#2204)
* *(deps)* Update dependency vitest to v0.20.3 (#2215)
* *(deps)* Update dependency cypress to v10.4.0 (#2216)
* *(deps)* Update dependency sass to v1.54.1 (#2218)
* *(deps)* Update dependency esbuild to v0.14.53 (#2217)* Chore(deps): update node.js to v17 (#883) ([2004478](2004478c8860c1d0a6d325833a70ce1afb481d94))


### Documentation

* Add fixmes ([858e7d6](858e7d60a66e55650c44d1660040e039ded196d4))


### Features

* *(BaseButton)* Add target _blank for links by default
* *(a11y)* Use <time> tag for dates everywhere
* *(a11y)* Use better markup tags everywhere
* *(a11y)* Add aria-expanded
* *(a11y)* Honor prefer-reduced-motion
* *(a11y)* Make sure the contrast for the primary color works with dark and light themes
* *(ci)* Remove vue3 branch trigger
* *(ci)* Update translations only on cron schedule
* *(docker)* Show what api url the container is using on startup
* *(kanban)* Show loading indicators when handling tasks
* *(natural language)* Make natural language prefixes configurable (#795)
* *(quick actions)* Select the item when only one result is available
* *(shares)* Filter out users or teams a list is already shared with
* *(tests)* Replace cypress-file-upload with .selectFile() (#1460)
* *(tests)* Change cypress default viewport size* Use computed for api domain (#722) ([432c6ba](432c6babf2c50d63c3d796c1bd69c1614568c9fc))
* Import bulma utilities global (#718) ([0ed3cf2](0ed3cf25538d6f9546db0adf9b02bce9fd867329))
* Simplify heading blur logic (#727) ([dae441a](dae441a37312c633fdf1304c6f6c0fdc84f1b0d3))
* Use store getters to check auth (#731) ([0295113](0295113f5064f43bbd3a66794b9f723d41bb0dc0))
* Feat simplify taskList mixin (#728) ([50c1a2e](50c1a2e4d59aeedc4a9f362210a0ed6cf5da1c4e))
* Close modals with esc key (#741) ([728dfc5](728dfc52e5fc9b34a3529f06a46d45ea6580cd75))
* Move fontawesome icons import to dedicated file (#742) ([9122a18](9122a184d6bfaa48b25cc3cfedb91d1eb3198021))
* Move pagination to dedicated component (#760) ([7719ef1](7719ef1bef149d537648544612aa002c95a352d6))
* Reduce file size by removing by removing fonts (#759) ([6b1bf27](6b1bf27bf87e403913490c8d8cba2b72170f1852))
* Add variant hint-modal to modal component (#764) ([4f2378f](4f2378ff023b08213adfd16a08608dd962600c51))
* Feature/remove-attachment-upload-mixin (#724) ([41331c8](41331c8a867da5b33edab19be505111ec7824c35))
* Update to new slot syntax ([4454e6c](4454e6cf2234f5de69fc52972994cb2a1d763b51))
* Formatting ([0da7a46](0da7a46612cb0b8cb227dc3e7f7356d367d2750a))
* Move from life cycle to data or watcher ([f51371b](f51371bbe0c3daf4de7bc9f987c9094bada2adcf))
* Move unique functions from taskList to List ([fe27a43](fe27a432c76e3fb1262819c39e0df28f936bd18c))
* Define node version in .nvmrc file (#789) ([c551bf5](c551bf5836af5da610e5342fa86aac51f39a9b7c))
* Add types for vite (#790) ([e997854](e9978548d8cded485061fd4d191beee7d014cecb))
* Don't try to load task comments if they are disabled ([e918b82](e918b82cfa7dbf5d645a5da6542a89b19d5bba46))
* Add example configuration for vscode with volor (#791) ([7579222](7579222bb049a71c114e44e42f815037a681537f))
* Allow quickly creating multiple tasks at once with multiline input (#796) ([442e6b1](442e6b12e00a71efd876db191d1c74fe2cd9c975))
* Make checklists work with '-' instead of '*' ([e17116d](e17116dac15aa720de2181710da1ede54a0869ae))
* Don't show bullet points of checklists ([2691a84](2691a846108c28f407eae0d1ed792ef4ea76091d))
* Task checklist improvements (#797) ([96523f1](96523f1fbf5f6f9eee8b9ea481bcf24265e41292))
* Native color picker instead of verte ([4ee7a8b](4ee7a8bac6a55b09e0a800815411c5d75a3ddc56))
* Add vue3 in compat mode ([78a5096](78a5096e0d5769dc168f676388eb09955e6deb62))
* Use new async component definition ([421ff9a](421ff9a1886b1f7e35cedbc1567e61a373d8b0dd))
* Use vue-fontawesome for vue3 ([b75c79f](b75c79fd5e8e0d1410287a47ba98c7b1bfe3fa69))
* Use vue-flatpickr-component for vue3 ([b31da0c](b31da0cefe55933ba64fa939163dd9ec08e218f7))
* Use vue-router 4 for vue3 ([7251821](72518212dac590f024ba3dd2d19d21adf0d35784))
* Use vuex 4 for vue3 ([3d6aca3](3d6aca3510a28d588be092c9a42a4851fcf93d4a))
* Use vue-i18n 9 for vue3 ([7c3c294](7c3c2945f82af77672b0d2472dfab766379070f6))
* Upgrade to packages for vue 3 ([e779681](e779681905daa23d963da9e99829a66125a07d59))
* Forked vue-easymde ([a745966](a74596698448746de633b87fca0dc915decd181a))
* Remove createTask mixin ([672d63f](672d63fbed88c9130669b0f1582eae30286d0989))
* Always use index for buckets ([8d88b37](8d88b3792dfd73343e3eda4196ca199ffdbf24f3))
* Simplify filter-popup ([15640e9](15640e98ec0495c0a743da0d64dbeaaa1adae79c))
* Watch taskId instead of whole route ([6b35810](6b358107b62f26367d697d87a7bc7eea318c5751))
* Fix some Kanban errors with vue3 ([7bb1b1b](7bb1b1b769be50a9f5ce3252864fded665b63bbc))
* Watch taskId instead of whole route (#812) ([64abb1c](64abb1ce379c3400357cf0bbb7595cf5682f2e11))
* Compress media files (#818) ([b59b5de](b59b5def57e93f9ad68b768dedffe388926ba3b4))
* Show checklist summary on kanban cards ([99fb78d](99fb78dbd4201f8bf522fa96a74e95c0752ca0b2))
* Hide priority label for done tasks (#825) ([7e1a7f3](7e1a7f3f19f91f5c41f7c3be8c834417ca55ef60))
* Remove lodash dependency (#743) ([faa2daa](faa2daa87627abfe679b3714650efa9db1308466))
* Add legacy build ([17d7910](17d791027ccf0cc3a08a54f3a677168c4960569c))
* Improve kanban implementation ([d66ad12](d66ad12f5c7a9d82e912fbd598745fc0e890d664))
* Create randomId helper and use everywhere ([35c861b](35c861b711a0dd64a9ee77da03bb26010ce62c36))
* Also use createRandomID helper in editor ([18d7ca0](18d7ca0820713c6db5bcfddeef6bd8f2ff34d65d))
* Some vue3 package improvements ([d1b61a1](d1b61a1489df3592574493ca1fe55a92ab3276e8))
* Show up to 4 recent lists on the overview page ([97dd55d](97dd55d9464e3cb24bf1b057932bf64d4a295928))
* Redirect the user to the last page they were on before logging in after login ([9a2f95e](9a2f95ecc60241a9c6157acc384ad3a664bee57d))
* Review changes ([fa54e1f](fa54e1f1de284c31c7a089870046cb660f2b0c40))
* Don't rethrow same error and handle errors globally ([3b940cb](3b940cb56c2e29f83b5b329dea45f6059be7ca48))
* Use async / await where it makes sense ([bb94c1b](bb94c1ba3afbed2b9acf7d8595e3c7aaf620f608))
* Use computed for namespace title ([12a3c23](12a3c238b8ff3ec56d170698fbef07e23821d1be))
* Don't update the bucket after updating task position ([cc32ca2](cc32ca244c89f7baf1679343f08216f4224fd3ae))
* Feature/vue3-make-workbox-debug-configurable (#862) ([dd0e04b](dd0e04b10621ee2a08c479f4fe35aa190839ba9b))
* Keep errorMessage local (#865) ([0a1d008](0a1d0084e21d19acca0afa779aea3c8886242e47))
* Throw errors (#867) ([a70b922](a70b92253a53407a26f5673af8f9edb5b6be80a3))
* Rework style imports ([1f5283d](1f5283d5488050022dd626b4490def30cb57b2c2))
* Moved some card.scss styles to the card.vue component ([a33758e](a33758e37ed787585ed116c6abc2d591602cf1d0))
* Merge fancycheckbox.scss with component styles ([b9eba00](b9eba0060323669406272ad87331605c994e48fd))
* Merge multiselect.scss with component styles ([b304712](b304712b1e83eee209f9eef564cbbfb930621a28))
* Move scrollbar styles and add variables locally ([6195637](61956370018868917e04160d1422f1acff41bd83))
* Merge api-config.scss styles with component ([2650814](265081417d807c4bcce0ea432c28ec039f826ef1))
* Merge attachment styles with component ([08f84bf](08f84bf7e3e161d2cd6115560555fb35a13166f7))
* Merge color-picker.scss with component styles ([be35c73](be35c73f6ea584b0cc563f542f8dea105624a59d))
* Merge comments.scss styles with component ([46ebd45](46ebd45a74971c54a3944374329c6786aa16d101))
* Merge datepicker.scss styles with component ([3cb68c9](3cb68c945f5fa01e7df7bc0e8cf0dba792c3efc1))
* Merge gantt.scss with component styles ([ba1942e](ba1942e7570d1556c7fec3d3ecec9b326e51eb99))
* Merge kanban.scss styles with component ([9ca8857](9ca8857d890cd7c323d0101e51a1fdce82b7db75))
* Merge keyboard-shortcuts.scss styles with component ([f74cf51](f74cf516d2c1719dc0c3c078665642a94579a2c9))
* Merge legal.scss with component ([4223d23](4223d23ce5b1382012a1581c174e17d69bc65879))
* Merge list-backgrounds.scss with component ([4d15f7a](4d15f7ae987defed343efd13fe411a4870a317e1))
* Divide most list.scss styles into components ([87f7a51](87f7a515a6709a128f0d219983f6ce7f5cb2e1bd))
* Merge migrator.scss with component styles ([0eb8766](0eb87663e011a2c3c0eacf36f74b8835fb74a482))
* Moved most of namespaces.scss styles to the ListNamespaces.vue component ([0f7caaf](0f7caafd54d5c09b1e283cb510f16d5f883cadde))
* Merge notifications.vue with component styles ([a5a716e](a5a716e09ba0390c6c95b37fcb2d5fade4f7fe58))
* Merge quick-actions.scss with component styles ([0aff057](0aff057f7b70fd93dffc0282effbf48e8be1e337))
* Merge reminders.scss with component styles ([3701867](370186726a6cabddea70fee82f6736ef721cc094))
* Merge switch-view.scss with component styles ([55bed2e](55bed2e5e8d1529fa0e9527b84561dd6ce753789))
* Merge table-view.scss with component styles ([f7c7ea5](f7c7ea50eb6d4302bdfb3913369b2fd12c64c1b3))
* Moved most task.scss styles to the TaskDetailView.vue component ([c9e23cd](c9e23cdd29c2ba753cc56f7ce22039d1c2165f2c))
* Merge taskRelations.scss with component styles ([e0fd5f8](e0fd5f8fe0e8579d8c969feb770df91ab60ee232))
* Divide most tasks.scss styles into components ([14dd49e](14dd49e4b0ce5120c00a50c0ef2109211e59cc75))
* Merge teams.scss with component styles ([4d2c27e](4d2c27e74e957b9c2efc76dc2156d5d42731952d))
* Moved some background.scss styles to the contentLinkShare.vue component ([2aecf32](2aecf3245b70055a2051b720533dc9cdd1678a71))
* Divide most content.scss styles into components ([3e45678](3e456782dfbff6d1a64b05f7dbffabc5fdb02bc3))
* Moved some link-share.scss styles to the contentLinkShare and linkSharing components ([57d5afa](57d5afa530da75938a3f36d615027c52cc8b2fbf))
* Divide most navigation.scss styles into components ([7824ddc](7824ddc13fc86b7b574cbcd2d6b01207a4c9f3e8))
* Merge offline.scss with component ([986130a](986130a0ac7d7c7ff783af52a237758dbf5f837e))
* Merge update-notification.scss styles with the update.vue component ([7ca355d](7ca355db66fd914905b6ac815bfced9222703c0b))
* Add FIXME comments ([4f8cce0](4f8cce0f4597aa884445a9ff8499baf2a9760de7))
* Move some form.scss styles to button.vue ([19a4b17](19a4b17004a71902aba27b419227cc56a720c534))
* Add Done component ([c6b24dd](c6b24dd8f146ae6f9e0e10aafa55fd921e188e2a))
* Add close task popup link (#880) ([877b243](877b243c6980916c468588e23931d39e2db65b6e))
* Add vietnamese and italian languages ([48224e2](48224e28b8fbc8f485697346ab220955698e8851))
* Improve icons (#903) ([3bd9b02](3bd9b02768c5a64959f77d77e51b1a5783a39d48))
* Add sentry (#879) ([1774fdc](1774fdc604d823863603ee48b71f109f03a68d21))
* Move user settings to multiple components (#889) ([5040a76](5040a76781a01d92a21872bc1b69abaecaee09ce))
* Add czech language ([ab1f504](ab1f5047a1fde85b0b53cfa19d7abf2316700d25))
* Allow openid users to export their data without a password (#918) ([5b406b0](5b406b0172118317c0c89823e93fabce57848698))
* Add releases to sentry (#919) ([1873c74](1873c747761573d26107a79c26a7e476968a75d2))
* Disable password settings for users authenticated with third party auth (#921) ([ecb5be4](ecb5be4b1757661f352993e464b3736dfd5178c2))
* Show indicator on a repeating task (#925) ([d8d4803](d8d4803e2d907e634d82e425325f601fb46ff268))
* Use script setup for ShowTasksinRange.vue (#931) ([108e7af](108e7af57847c4c96ac53f50d6610b4311d53820))
* Add vue-tsc (#949) ([e23f3c2](e23f3c2570ce002818f364a3b6b016fd755e5391))
* Allow selecting multiple labels at once (#945) ([9b7882d](9b7882de7a911627b258b773962a163822a9f48a))
* #947 remove reset color button if no color set (#957) ([8f43619](8f43619f7365dbb5b08d1e191acc3c7692105fef))
* Add .editorconfig for scss and css files (#970) ([1cef4f6](1cef4f6e0b9c5191ccc39707f1b576537a887818))
* Add preview deploys with netlify (#972) ([e49fd16](e49fd16a3acd5cbfd430fbf54575e43a05f7adc3))
* Properly return 404 when the file does not exist (#966) ([052cd36](052cd36085c8be09c7d5d24f180f35f2e2817e6c))
* Wrap edit-task with card (#948) ([8e6e52b](8e6e52bf02c66a8159223b2c7281ec726bc04e1f))
* Add vite-svg-loader and add Logo component (#971) ([30cc89f](30cc89fe25dbea5385847cec03961a8d1c39a276))
* Remove ssl generation from docker image ([73651ef](73651ef964af57e717d6449fb7d5937028539c58))
* Add button to clear active filters (#924) ([31f0c38](31f0c384ac3a45b242464b2a8bd8424ef8cbdfc7))
* Defer everything until the api config is loaded (#926) ([0a2d5ef](0a2d5ef8200379f2fba401c7a024409c68c6e840))
* Search in quick actions (#943) ([0fe4338](0fe433891ad5c3f8e5ae99f5e310bea34865e6c4))
* Show namespace of related tasks if they are different than the current one (#923) ([db605e0](db605e0d219605f86c66dab70186a7d053f44c3e))
* Add v-shortcut directive for keyboard shortcuts (#942) ([feea191](feea191ecf68fb22e466466a516254513051fd7e))
* Use script setup for filter views (#951) ([e63fd58](e63fd587c81fd8fc2596fd097a1bf613480233f2))
* Re-style the keyboard shortcuts menu (#996) ([fcadbc3](fcadbc352b5fccd282f3e083479f3a8ff5fc5c13))
* Use flexsearch for all local searches (#997) ([507a73e](507a73e74c2551e9e3d9829a148a884d7d6203b3))
* Feature/use-setup-api-for-user-and-about-pages (#929) ([d0d4096](d0d4096f8b4b80959c164c57bf2288dea3e6e82d))
* Directly open general settings when opening user settings and none selected (#1001) ([665cc84](665cc841745fc0c8dc4c00149468aa85b8c2bfc5))
* Add postcss-preset-env (#1022) ([2656c74](2656c74f374696d11f7158130fd5bc5e346437bb))
* Always use latest browserlist (#1021) ([ed6dc94](ed6dc948738239421b6bcd0882019f1f7730fa1c))
* Improve namespace explanation (#1040) ([ae36c04](ae36c041a7453bc0b1d840dd95bb992f40a77933))
* Use popper.js v2 vue3 version of v-tooltip (#1038) ([91580f9](91580f97a1cde14267190267c5518f2a51e033e3))
* Reduce import size by only importing used modules (#1023) ([b688f35](b688f3544642243346c2ca110c678aef471a1dfd))
* Add packageManager field to package.json (#1099) ([59e915c](59e915cc10c490aa1a92b842792902fa6f99c15c))
* Add message component (#1082) ([f8d009a](f8d009a6aafddf7492cd1361f8c9ef2b6c1503cb))
* Convert home view to script setup and ts (#1119) ([716de2c](716de2c99c3170a67c61b98de74694c29a369f2b))
* Harden textarea auto height algorithm (#985) ([84284a6](84284a62117654961160bf555a3f0ec6dde88f72))
* Convert simple components to script setup and use typescript (#1120) ([ac630ac](ac630ac775bb4b222ff1b0dd01f20ece57c522e4))
* Recurring for quick add magic (#1105) ([8b8e413](8b8e413af0f3cffc5f720437e5a4473c401c46eb))
* Add support to set the marble avatar in user settings (#1156) ([1a119f9](1a119f97c584e425195485fd57b5b1eacacba694))
* Use script setup and ts in app auth components ([c3c4d2a](c3c4d2a0a57d523db6acc76288c76615d7936a20))
* Restyle unauthenticated screens (#1103) ([32353e3](32353e3b76d7aa6c95536f9b876de50ca3a666ab))
* Build openid redirect url dynamically ([ccaed02](ccaed029f27386a8c2505744101c4dde41bd4d76))
* Redirect to calculated url everywhere ([b7aa789](b7aa7891e988ba2231ca9c3e58569c2a190622e5))
* Improve input validation for register form ([05e054f](05e054f501be1a73d63788468a995f092d279e43))
* Replace password comparison with password toggle ([aa12bff](aa12bffcbc09dfb9070a92232400c0d103d45c51))
* Change wording ([1d916e7](1d916e7e03add8f8a4f8e809eed7600080bd3579))
* Improve error handling of login fields ([66d5e85](66d5e851e823e9667380e759e689fecfbe88e1ec))
* Add tooltip and aria-label ([fda0b81](fda0b81d9c653a1a93803a98582ebe9b10f3f433))
* Add extra prop for message center text ([1fc1c20](1fc1c20c87217dbc067cfea2a429e923ef2ee8b6))
* Change links to login / register pages ([5558d91](5558d91f4470481c2b4c80d7b28d094bcafa4c53))
* Feat/alphabetical-sort (#1162) ([7ebca9a](7ebca9afc5afd879faf1c6d1dc4aaa945bf775ae))
* Improve playPop helper (#1229) ([943e554](943e554a586eab7d0b52a0aa536984ee5cd59fbb))
* Move password to separate component ([0322daf](0322daf4d459f9108cb71902381f1dd1a2b06c57))
* Add new component for a datepicker with range ([8115563](8115563d674abae796c3d1084a5af89654b652c2))
* Make active class work ([3d1c1e4](3d1c1e41c7bc71fcef4a6253b429a16a84252a6c))
* Make the custom button actually do stuff ([12317c5](12317c56b3bed4f7d5e38362ad7c4352c1323997))
* Disable time ([a5b23a7](a5b23a704866a144842e976a46dc0163b261398c))
* Add more date ranges and make sure they actually make sense ([8f8d25e](8f8d25ece18939a1bcfd9c5237d8658567ed5203))
* Move date filter to popup and improve styling ([932f177](932f1774ecb2e0f6e59e32e997027c5e9c90368d))
* Save and restore the user language on the server (#1181) ([4a7d2d8](4a7d2d8414238b38c4eeec6d9a928b6bfb8dbaf0))
* Replace jest with vitest ([8114012](8114012997376480ee3d0788b1cbecd24e623648))
* Move the calculation of the current salutation to a different function ([de77393](de7739390513fa69255a4461fee7ecafca71aa01))
* Return full translation key ([27534a9](27534a98e916a1e6813b5940d4507bc1e22a209b))
* Use useNow to provide auto updates ([d2577f1](d2577f1df6d47363db2770b6c8a0b95ef6ea5ea8))
* Convert to composable useDateTimeSalutation ([cb37fd7](cb37fd773d9163a772c6ad1f84e6641334cf75f3))
* Create BaseButton component (#1123) ([cdbd1c2](cdbd1c2ac47d2c74585175f9fae2fc940347fb81))
* Implement modals with vue router 4 ([5a0c0ef](5a0c0eff9f0bb822a597164c2b87da7480ce4498))
* Make taskList a composable ([281c922](281c922de1ea931fbbfa4b7db0a5c97aa4498f0d))
* Unify modal view ([c70211a](c70211ad32e51659c542d5a8f6333a5cd701decd))
* Mount list views as route-views ([7eed062](7eed0628d0bd9846950ff025dc647bd7e6dc6523))
* Save current list view just once ([29a9335](29a93358446dbf293f3c08e2f2857bec4e5fc9d7))
* Review changes ([2db820d](2db820d926fbfd00ba5ff0c68fec243020db2620))
* Provide listId prop via router ([5916a44](5916a44724ca237daf13e6ac396f27451bfb5887))
* Run vue-tsc in ci (#1295) ([9b85817](9b85817ddba50a5641a78fa0cce14610b63ebfe8))
* Changed green "Done" button to read "Mark task done" (#1340) ([044f2b9](044f2b927dba715024d2946bba2a1a7e6433a68d))
* Move lists between namespaces (#1430) ([c98ab42](c98ab42e7560eef9df04280dd3904c39272c3628))
* Make subscription a BaseButton ([187e62a](187e62a7ec504257af03eee732eb63b6ba6a7d5c))
* Improve Sort component ([8937b42](8937b423219bc450775b752ccec7e9957f2690c0))
* Use es2022 for @typescript-eslint/parser ([a325e4b](a325e4b721ddee156ae6c7ca9dca414b36f3fdaa))
* Add cypress dashboard record (#1462) ([c21f236](c21f2362498f6e01c3cf6a37e5e5bedd7871adde))
* Don't open task detail in modal for list and table view ([de626ea](de626eab31e092ad9364876996575071260e705d))
* Merge TaskDetailViewModal with modal ([6827390](6827390b77ae6e186e7b0163651c19ca9a247d2f))
* Implement modals with vue router 4 (#816) ([a57676b](a57676bf546dc866010c1f33db97c427fa6b44c7))
* Add slot for trigger button in <datepicker-with-range> component ([c41397f](c41397f5dbcbfae3b8bc718e093c87e2dfaa8dcc))
* Move logic of ShowTasksInRange component to ShowTasks ([43e8335](43e83350bd3e98960fbbfa695c956893661f148c))
* Use object and loop to set date options ([32bdf16](32bdf168920c09cb2fbb6008cb4c10040094967a))
* Move everything to fancy date math ranges ([6667df5](6667df5f1fa525ca46c008cbed11206604f597fc))
* Make sure showTasks can handle dynamic dates ([dabe87a](dabe87af4b2a4c3cc198e8b805e45e93ffae6b11))
* Add two inputs to toggle flatpickr ([8d5bfbe](8d5bfbe828f55688c430200742b42d1af7c274fb))
* Make sure date ranges work with date picker and vice-versa ([1e46849](1e46849c784907fd3eff8f49b98027f26b529584))
* Add explanation of how date math works ([e7fa1d3](e7fa1d3383d19daca68c30ee431059f3d47f2589))
* Add more pre-defined ranges ([0ae8a0e](0ae8a0e6ef8103731ead828e2e487e01d6a529f1))
* Add prop to maybe show selected date ([3a12be5](3a12be505d9c28f987eb81b9436d4bbb9cb3eaaf))
* Add date range filter to task filters ([7aa2cfc](7aa2cfc8d4bb482642b97b1388190706fcac13e6))
* Add remember me style login (#1339) ([3d3ccf6](3d3ccf629a19b4425ffc0e1a97fc90f8cfc4b1a4))
* Add authenticated http factory to create an axios instance with bearer header ([59da668](59da6686d08071db7011bc928dc50c5c3a78553b))
* Add setting for time zone to user settings ([a812793](a812793eadb83d430bc5ae70d4542d23cfeaac88))
* Add timezone setting (#1379) ([2ea3499](2ea3499bf748936574edfe9c3573e23ac758c57c))
* Reduce dependency on router and move everything to route props instead ([84f177c](84f177c80e516a066363356b4df783eb5606105a))
* Add more default attributes to the rel attribute in link mode (#1491) ([2a4bf25](2a4bf25d20b308d0607672baa7ea3aff89d437a2))
* Simplify config mutation (#1498) ([1e0607c](1e0607cb86b010603bb76947b28c01120e743930))
* Add Polish, Dutch and Portuguese translations ([80664b6](80664b6182a939ed07aa891646c0e6764acd1009))
* Increase task drop area size for bucket list ([69654b8](69654b823ea24c6fc1a4a8d33ac9eeccfbf2b53b))
* Restore styling / fix styling issues ([45e1ae6](45e1ae66d69eaff6d7ff3294211e0542009b20df))
* Increase task drop area size for bucket list (#1512) ([cb395f3](cb395f3f69e5a364c9ecc5769da4d64bff80e9cb))
* Enable strictNullChecks ts setting (#1538) ([72d6701](72d67014040a5d97ffe9be314efe779db788879b))
* Make profile picture clickable (#1531) ([eac07d3](eac07d31692dc573284d13c7d93738d1723dcf13))
* Convert api-config to script setup and ts (#1535) ([b84fe4c](b84fe4c88ba244865d09a8e9f5e51a1fce20cb7d))
* Change port to 4173 ([98cb14a](98cb14a86c2918f1a087a6a180bf37e14edb0620))
* Rename percent done to progress (#1542) ([8ea9d75](8ea9d7541f07985d9cceed995a1872fa11b33cc8))
* Use AuthenticatedHTTPFactory for refreshToken (#1546) ([8df73c9](8df73c973bfebc1ec47b52a211ff35381ece51b8))
* Change preview api url (#1584) ([9f5e68a](9f5e68a125e90b67dcfabba53a738bc71ecbcaa2))
* Rotate task cards slightly while moving them between buckets ([17ba56f](17ba56f12d69b9f5a0fa9a123d3a3865367bd6ca))
* Add a few new keyboard shortcuts ([f4b0e68](f4b0e683229a667f730e1c0fae7d509dc978bbcc))
* Prevent scrolling the rest of the page when a modal is open ([574ecff](574ecff12db59a8450b8aaa5632e95b45999bcb3))
* Use vueuse to lock scrolling ([f9b7e2f](f9b7e2fd7657c6386be84ce093e0a5768f7690b5))
* Add date math for filters (#1342) ([9b09fad](9b09fadbd0aa0ceab2d7dd6636dd650f7e71c2b6))
* Directly create a new task from relations when none was selected ([dfed1f4](dfed1f438a168606617d2ecef55cd83ff724201f))
* Use blurHash when loading list backgrounds (#1188) ([4cff3eb](4cff3ebee138289c915d8e70e89e009f3be28fea))
* Rename js files to ts ([15b6713](15b67136fe2419ee7a132e863297b91ca0117f0a))
* Add lang ts to script block ([a3329f1](a3329f1b421bbfdde1b82c7feea41997ed21b304))
* Use defineComponent wrapper ([ba9f693](ba9f69344a53485e744dd05a4b733e6a8dfb8939))
* Convert some helpers to typescript ([b5f867c](b5f867cc66af6c70c875bc1abf4dbc545916063b))
* Convert navigation to script setup and ts ([658ca4c](658ca4c95593d8c9b294aa4e613fcafe6dcf801e))
* Add TSDoc definition to some models ([16d8c22](16d8c2224bd76f8a82c9454e85672770b4cd7703))
* Convert create-edit to script setup and ts ([0e14e30](0e14e3053d92ee8ca046a96badaf358557d221d5))
* Manage tokens ([8e5a318](8e5a318d4c1fe749867174657ec43cc1f886cf7a))
* Flatten and reorder after all ([50575ff](50575ffd687ceaf067076bf3b531861362c8c967))
* Remove duplicate rel attribute ([b1159f3](b1159f331f9ef704dc96f5ea27af5b2deff2fa41))
* Manage caldav tokens (#1307) ([0b31cce](0b31cce56778b5bb54b71962e2bbfa0ea06f2fb6))
* Nginx improvements (#1545) ([52fdc26](52fdc2614bf703c17c550df75832830440043965))
* Improve password component (#1802) ([ed8eb84](ed8eb846179f06bcb7a00938ed0bb62e2d3451e0))
* Add scroll snapping to kanban view ([8473bd6](8473bd6a8b7956481bad7758c700d7f4997541a3))
* Use BaseButton in PoweredByLink.vue (#1825) ([f7e4c58](f7e4c5819c7046a6379a25bcdf80f098aa7aa8e5))
* Improve dropdown (#1788) ([e0023b1](e0023b14e8d0818d5170c5f542a76b60eb157f13))
* Remove copy-to-clipboard (#1797) ([17a42dc](17a42dc2e7c78a84e1696c1b41229f11cec5884b))
* Show the number of tasks we're about to remove when deleting a list ([62adf17](62adf171ecc66f930fb83f871dc345d49b3d85cb))
* Simplify namespace search (#1835) ([8578225](8578225982d59ad187469b1362aaf529005fdb47))
* Move filter popup to a modal ([0007c30](0007c306726622eaccf9ba6f63434dbfb5aaabaf))
* EditLabels script setup (#1940) ([9a4e011](9a4e0117b2e945c6561b1fc3a7360338fe90c9a1))
* User deletion script setup (#1936) ([7682685](76826855e4668b3727b980fcf9262e61b9fdcd84))
* User Avatar script setup (#1935) ([fe698a6](fe698a6f84364f1017ac244644bbb3aef92443ae))
* Task reminders script setup (#1934) ([0a89e8d](0a89e8dc6bfe0eee138ff44eb3c818138dbbff37))
* User PasswordUpdate script setup (#1933) ([3ecd1d8](3ecd1d8db67492ed7228e47b17ec9b60586e1132))
* EmailUpdate script setup (#1932) ([6538a35](6538a3591eeb48748617f6895f0c629cdb4cb2bc))
* EditAssignees script setup (#1931) ([72e43b7](72e43b7bbf36f4d6e99f31ec8ddfe80bbf66ef57))
* Comments script setup (#1930) ([9a42713](9a42713b044627a2d6cafb6f65bd688cbbc45384))
* RepeatAfter script setup (#1928) ([6737bb3](6737bb37b48c799b171de08097fc6aa88aa461a4))
* Feat quick-add-magic script setup (#1926) ([1bf3786](1bf378608e513d3e797a22bd5abbf4cdf5f3dede))
* PrioritySelect script setup (#1925) ([99d1c40](99d1c40cfd38c82869448578e7da6e94cf3e888e))
* Checklist-summary script setup (#1924) ([49a73a1](49a73a154ba77a043b01ce15fce5d11734386f9e))
* PercentDoneSelect script setup (#1922) ([8d785cb](8d785cbf291a67600e7f73c904c8c6e156ff6c44))
* Add success message after deleting a comment ([246d679](246d6794d8346688d6dae69fb93a9b3dca8b7fd5))
* User DataExport script setup ([d11fae1](d11fae1c38cd03f9767996fe2155248e817e515b))
* User General script setup (#1938) ([2c270d0](2c270d063ed766886eda9b57fea6ba3df3e0293f))
* UserTeam script setup (#1976) ([0e41b78](0e41b787129ef7826e4e94f7eaddd02134703b54))
* Make user settings links config driven (#1990) ([6bab108](6bab1088c7dde74ed6604f22ab8d08ad366d3be6))
* NewList script setup (#1989) ([5291fc1](5291fc1192b360ff2350e7ce0468fcfa48967231))
* Remove bulma styles ([c6ee8a0](c6ee8a04e2a6f1232cd389ede87adf4f8bfc0e5e))
* TOTP script setup ([c1e4eba](c1e4eba7f550426e120834740f643c383c9db491))
* Migrate script setup ([27f7541](27f7541b25cc05fabe197ea826072849037ca627))
* Archive list script setup ([93b2482](93b2482d4c1ba37a8474f76597aea2dac139a13a))
* Edit-task script setup ([cdf359d](cdf359da002e0300c669627f7958ced9a5239b13))
* ListTeams script setup ([17b77c2](17b77c25c1f4d284584004fbad8f5b1eded1e63d))
* Improve colorIsDark helper ([297d283](297d283090058744f2ec9b9ddd2de5fae0ce4565))
* Description script setup (#1927) ([c7f8ae2](c7f8ae256b2082783dbe30112cfbf853b401cd39))
* Vue-easymde script setup (#1983) ([e6af477](e6af4772fbbdc51a190dbf1772b708016bb37f31))
* Defer-task script setup (#1929) ([1d869a0](1d869a0497bdd6d8a4ce69350f829c2d470f2d76))
* LinkSharing script setup (#1977) ([ae4c73b](ae4c73b6eb4ee2a4a4f86ad4d5348f4db45288d1))
* Remove vue3 compat mode ([53dc7d1](53dc7d12f7cab58e900aba842f99fc4236680b9e))
* Feature/fix-vue-i18n-9.2.31 (#1994) ([5ef939a](5ef939a230f7e2adcc62fbffd56aae7b74aec565))
* Add alt+r shortcut to bring up reminder input on task detail view ([72c123f](72c123f3f9b56b4f6dc3e216aa0ce0b3b465b567))
* OpenIdAuth script setup ([d996e39](d996e39a86bb9f3aa0bdf5673176fb8ac2321bb3))
* Add print styles ([6fc87e1](6fc87e1515c509cabd71e9de7464f21140539554))
* Add option to configure overdue reminders email time ([31c49ae](31c49aed4be58f22692ff7e0214dba92af9e9504))
* Only allow editing of a user's own comments ([a3192c3](a3192c30e9fc1681e66cf2ec7859e916c48ff69c))
* Ask for confirmation before deleting a label (#1996) ([e468595](e468595ce420251e6db792743432f997a8980dde))
* Enable quick add magic by default ([24f3477](24f3477d4b7b1d305193ec84065e7ce5aba8ca8e))
* Enable kanban scroll snap only for mobile devices ([8eed0be](8eed0be0720ce903673a2b7616ab061043a1b0c1))
* Add inputmode=generic to totp fields ([580b012](580b0129934e054ca2b6ee449d04e8d4513fc658))
* Move eslint config to external file to support comments ([513a51f](513a51fb73b2a8b87dca6e1a6a95f522c8b645c6))
* Improve ts setup ([c6aac15](c6aac15d2419b9c9ec6bb7371337697a5d54ac7c))
* Setup cypress ([7fe9f17](7fe9f17e43c1d18380dd92da1c88212050a9dc90))
* Use inline-block for BaseButton ([9e1ec72](9e1ec72739b08ff6fbc79634c543b37b74686362))
* Use BaseButton where easily possible ([3b9bc5b](3b9bc5b2f86f203eda408afa3f0bbccdddddde33))
* Select a value when there is one exact match in multiselect ([6973d76](6973d76e1790808df087a29e2bd6f1dcf8d2211b))
* Allow marking a task done from a filter ([579cff6](579cff647d0cc7dfd830784559c41a21b143614f))
* Allow for easy reset of a repeating amount ([9cebf53](9cebf5305a4658fe2d7b4d19e2c96c5b8541b29f))
* Add issue template ([4666087](4666087aa988b33a8b5029598f4ee4b1c1fd22e9))
* Add more testcases for parsing weekdays ([518417c](518417c0de61573d1373cd3eb2ebd1c2d72fbe2b))


### Miscellaneous Tasks

* *(ci)* Temporarily disable cache
* *(ci)* Use latest version of s3 plugin
* *(ci)* Make sure you cannot tamper the deploy script in a PR
* *(quick add magic)* Clarify the use of spaces for lists and labels
* *(tests)* Remove test result upload to s3 since we now have cypress dashboard* Define default label background color once (#713) ([87c70ce](87c70cec0e91f03f9bb58a247ef34be849c11a45))
* Create progress dots dynamically (#715) ([96ef926](96ef926ddecaa4dcfcfd12600bfe0af1f75395b0))
* Make method event independent (#719) ([d0e46e5](d0e46e59e84ab85239a5878cf22f4ff556ffcdda))
* Define default filters and params at one location (#721) ([b5df941](b5df941e39835429d94312cdcfedfb42892b44d3))
* Move constants in folder (#732) ([07a6a31](07a6a31f47dbd9b209663fd325df7c275a8b9e39))
* Remove obsolete css vendor prefixes (#739) ([47ad115](47ad1157380a497dcd17ed740f031add67e0de4d))
* Some small changes in the cypress README.md (#793) ([8cd4bbc](8cd4bbccf6fdd9db2895c8509da7f853e50e8a1e))
* Change cypress settings to run tests in cypress without needing to modify the config ([d13f3b9](d13f3b9b19a2872474260021383b0610b155a4f2))
* Some editor improvements ([117980a](117980a8fc68d25c995397479e9ecd0bf382d4ab))
* Remove unneeded babel packages and add peerDependencies (#828) ([3c5c3ca](3c5c3cad107d8de727acf63943988b2df610b029))
* Add vue3 branch as drone branch trigger ([43b2236](43b22360a513596d37f903ead51ee46871462fa5))
* Remove unneeded var ([6fee114](6fee11461066ceffbd08d65ffca2ee3802249a1d))
* Make functions of linkSharing less dependent on component state ([1964c13](1964c1352cfce581b097bbe8c37be2abd4844b39))
* Remove console.log ([a3a3ef8](a3a3ef850c48f1f971376b68bc312dd782de445f))
* Upgrade vue3 packages ([6f51921](6f51921588655653ba28966a8695ea9a86a83bf7))
* Don't resolve when returning from promise & improve list store module ([a776e1d](a776e1d2f30fc4889430d8e8d8dbe376a76e28f8))
* Simplify MENU_ACTIVE mutation ([1d43d1b](1d43d1bd652d027bd7fc219c27ace1e891a5ae02))
* Cleanup ([c329c37](c329c37c7b10549caff1b2b56b73e385427cd696))
* Remove vue3 from the drone branch trigger ([eb7b1bf](eb7b1bf4328710bf887d3e6973f859561d9f8e1c))
* Re-add vue3 branch ([1fc857d](1fc857d9a2bdaecf458897ed074dc9c1855f974c))
* Remove unused method ([c1a981c](c1a981c60bba5dda6bfdecea8109f51dfcf1f2c4))
* TRANSITION_GROUP_ROOT silence transition-group warning ([852b864](852b864ee6608a3c4051020ebdcfd416eafbcc24))
* Remove obsolete _all.scss variables ([a0ca6bb](a0ca6bb8fb62629d5df5fd0705650fa6771c04f5))
* Remove unneeded styles from tasks.scss ([4a61262](4a6126287a40d9540129e1ee3232465edcf58738))
* Small CSS format changes ([32a0106](32a0106819f49974380087c3182ac9cc5f2be8e7))
* Don't spread arguments (#933) ([d1ff800](d1ff800b415254922f0e25d0d65bd7265b57c89c))
* Remove setting loading state in register component (#939) ([b34213c](b34213c30188ffa27b7fcdd48cc07c8b84d3ac96))
* Remove weblate ping script ([a47d106](a47d1069268d6ffda25e0168bcd0a535b67eee6e))
* Remove some unused notification styles (#953) ([b7207c6](b7207c6eaf2a5ca069e40f70e20fecce6a14f5fb))
* Use a class to set the logo size (#1004) ([bb64452](bb64452382297165ca4160a11c5c4eb919ab1260))
* 0.18.2 release preparations ([9b24387](9b243873c52b55efaed95f442856acbbeb3fa7ac))
* Explicitly add caniuse-lite to dependencies ([8440869](8440869bcd7bc8da033532f3209f0a13aa510c42))
* Directly use redirectToProvider function ([36fb250](36fb250d1f5c159b3c375602e3dbf986c977441c))
* Simplify focus directive ([f944c35](f944c35e99c6f7ede46c037abc9b7e24c379b84d))
* Move password field toggle to scss file ([8397608](8397608fefe3905b9f4d4c95108883ee8c3aec1d))
* Cleanup and reorganize the date selection ([7408c37](7408c37dec3c2809a3a58054dd43242409566729))
* Use ts ([b274a79](b274a796d42bf9769b8e994893e14993417020db))
* Cleanup old stuff ([e93be0d](e93be0d04c0ca347ee57fe8774a3e4119ab07456))
* Move task sorting to computed ([0d6ef8f](0d6ef8f18afcbbcfa9afc72452eb00c1d3231273))
* Make showNulls and showOverdue computed ([d825960](d825960836e7557ed877939420f1c6b1338546db))
* Move datepicker popup to real popup component ([950fdce](950fdce111332e09e10fcea7be9521ce59469587))
* Make select date button actually a button ([1648bcd](1648bcdb70e2038df43da56e5bdeb550fa15edc1))
* Chore(addTask) improve order (#1297) ([e28f0f5](e28f0f5be439ce2cfaed7e527268189f80a6646d))
* Update netlify-cli only weekly ([9446550](9446550ce990613316eb718072e96ed9e6d4485c))
* Remove console.log ([959b53b](959b53b3a670d57fbfe4b049b38b3925f1eb1b56))
* Ignore wrong second argument argument for cause ([6ff621a](6ff621ada174bdf0ce52272714e08283ef366230))
* Rename function ([dfa3025](dfa30258aa3ee725c94af584e22d39c046cd5d10))
* Remove vikunjaReady from store ([24a1544](24a154422d8d0112e64eef5da70bf92cf0c44abf))
* Remove unrequired type ([8d13b97](8d13b979ec299efadd82274c01bd474e5668974a))
* Use v-else ([4e8a030](4e8a03066ebab60a907db1592f64a2860d87d9c0))
* Remove unused style ([ccd8602](ccd8602bfde6ff0a7636ee3501d3b53a5777ea21))
* Completely move logic of ShowTasksInRange component to ShowTasks and remove it ([ecf679d](ecf679d8e191de6a35152b9d8ec0fd6cb31f3cb6))
* Convert ShowTasks component to script setup and ts ([bcd34ef](bcd34efe91f50c35959bcc5187e49601613e9c99))
* Cleanup ([6d6f2b4](6d6f2b4e33e8e00eaf567d75f6915eca997791af))
* Refactor trigger to slot ([c5d598c](c5d598cac466e527aa9161901fb256933d1c7ed1))
* Use more BaseButtons ([18f7adf](18f7adf4204edcfcdfd21b49cff81d58b3c3b494))
* Watch values instead of listening to changes ([2041362](204136266f0c5b856fe8ac02b0d2819e1b861e84))
* Move date math explanation to separate component ([eefe6bd](eefe6bd413514f9281919271bc3b746aaf6918d5))
* Change import order and useStore ([f435ca9](f435ca99f4522c105fd2902ff91ce9c65be874d4))
* Rename date ranges export ([60be8b4](60be8b428e2fe39ab7f47bf81834f96999c359a7))
* Change return ([356b291](356b291a57ccd313a9a370664afb2c78edac33bc))
* Fix nesting and positioning ([a78ca6f](a78ca6fad368c1bcb5f1c1cd825213a3265c679e))
* Use a primary button to select range in upcoming ([436c041](436c0416d78b3d79e967339e4c7a26a3f7b9f96e))
* Simplify nesting ([4268eee](4268eee1f2dc1c21bbf34b4ed4b1d39dd2ad1ac0))
* Cleanup unnecessary css ([1e4ef96](1e4ef9615010685f00fad48f756fcef5d4067a22))
* Programmatically convert filter values to snake_case ([204e94a](204e94aa740236476254c818229bf5e47027a0e1))
* Move styling to the correct component ([77bf347](77bf34715591d9f25535b5ac9bd030ce9df656a5))
* Use BaseButton ([b1ec5b5](b1ec5b58ee3254bd513b1e6e3527b5a4d0844c18))
* Rename el ([7cd89b7](7cd89b7bf1268be9b29523c83e163e94d52f490e))
* Align wording in task detail view ([60f58af](60f58af41aa6ed8afb8f821d1bf5fc25d99c2d0c))
* Remove rel for help docs ([a6480cd](a6480cdb751a902e0af49300e78f9428a0fc7551))
* Rename i18n key for datemath help ([4195953](41959536967c736fbad936188a48f44c8f062fac))
* Remove abstractions ([18f5f8d](18f5f8da7d21414f7e797e1d04c3d648df283dd2))
* Hack the planet ([74766ce](74766ce1d0707609d0835ea0457f528a6d0e6e93))
* Return key directly ([564f669](564f669ed41190b574f884532ac14f7fdcda3202))
* Return the title directly ([95d8cdf](95d8cdffe4cee8f090d9516f0e7f1e5ece8cf124))
* Remove showAll prop and make it a computed instead ([4ce9ac9](4ce9ac9c669254fe32d66116bb1e4e5a5cbba167))
* Move converting params to service ([db47c1f](db47c1f10c65192d539410587b89aa952c856b35))
* Move to script setup ([75f09ec](75f09ec5dbd173bc8d20d1d0af798fbeaba3df78))
* Put action buttons right ([7bdefd9](7bdefd9a3e8ea878c0ce2f24ec7a9c3fbc0e549c))
* Clarify token is required for non-local users ([6b899be](6b899be202783ae5d9ebf6ce4913bfbac4019a0c))
* Use ts for caldav component ([cb06746](cb067461aa360dff90f144ca3267af45598d3aa4))
* Use findIndex to remove caldav token ([0299ed3](0299ed32f3366f6a1093329646105f2ae5bfa4d2))
* Make server functions async ([f042651](f0426519868c72f15e31ac24ea87899cb4a09aae))
* Extract getting all tokens into a composable ([043bf62](043bf62ef38555cca91f45cc9794cd071da15cd0))
* Check for no results ([af6385b](af6385bc606fa1a2b1a7af3396c90cbd86314a9d))
* Use function statements everywhere ([ca330fe](ca330fe63b3bf328ae4c1ce36513beee66657db2))
* Move success message after state changes ([da4f5a0](da4f5a0f758cf4dcaedb5407fcf58b31e1dab58d))
* Fix CalDAV casing ([cd245e4](cd245e467c2bee0fa7fb7efd9a670625fb91526b))
* Return new model instead of modifying the existing ([d865af5](d865af58a8206ecd8f1f2ed8edd767c0de37dd5a))
* Use h5 ([460a4db](460a4dbdbe2848295a22ab409cff7d57e628bbee))
* Rename to useTokens ([b9fa081](b9fa08116d5c1f728a783c25b466c6386d9489b7))
* Directly use newToken.value ([343be4d](343be4d5d6048a89d42b33433aeff7d458a680de))
* Use .then instead of await ([041f888](041f8884923904ccbf1d9897f12738d8a1f25ef4))
* Use BaseButton ([eb7667e](eb7667e27edc47893d6d661265498d304ffde998))
* Fix type ([ba1a1fc](ba1a1fc0413e0cd2776d0034948bc7a455b9f662))
* Simple Login view improvements (#1791) ([b9637e1](b9637e1bb6d543b7fdb97c783c9d96d8d55049dc))
* Fix spelling (#1786) ([656c020](656c020125e1ac6c0d2b0a6bdf8871329f630224))
* Add some types (#1790) ([53c669b](53c669b108b4eaac0fe624df42643350edf52d04))
* Move Modal to misc folder (#1834) ([f19221c](f19221cb1035424522966d299bf3b522718fee9e))
* Improve error handling in dev build ([1eaca64](1eaca64e2aa058b507a8cca54add568d7118ad9f))
* Replace the same i18n string with a single entry ([8257586](8257586c9077bc889ea3a0838cdb2480c4d2a1c1))
* Convert edit team to script setup ([cbecea6](cbecea62ae44bad07149655850185a0e81402005))
* Change dependency update frequencies ([ae93bbd](ae93bbd781976aeea321fc8754e92a26351f617d))
* Refactor notifications component to use ts and setup ([315da42](315da424ec42a6f2b75d43bc5fd89e8ef85939e8))
* Convert update available component to ts and script setup ([b2c2118](b2c2118c58a00fcc57cc16f959873be2078362d9))
* Update browserslist at most weekly and group it ([c7fb8fc](c7fb8fc7f2cbcf87d89d4071434c553b1d9b8a1e))
* Migrate namespace edit component to script setup ([0997c38](0997c3868da2b41f7cfda7c5039ab6f8ac14dda9))
* Remove unused import ([4070d64](4070d644041b685790a17cd4909289d1636cfff5))
* Rename js files to ts ([321850e](321850ec208167658069a1f99375c9c9bdc089e1))
* Update lockfile ([5aa6cce](5aa6cce185952e90cce757763442af508749a532))
* Use the <dropdown> and <dropdown-item> components everywhere ([cdb63b5](cdb63b578def76edf6a62e83e7ecb2374646f144))
* Add git-cliff config ([bafef06](bafef06e908b6c4e053482424ff5692667cb9f1a))


### Other

* *(other)* "feat: always use latest browserlist (#1021)"
* *(other)* Allow specifying listen ports (#27)
* *(other)* Enhance link share tooltip (#808)
* *(other)* Fix download export user data title
* *(other)* Merge branch 'main' into feature/vue3-implementation-improvements
* *(other)* Merge branch 'main' into vue3
* *(other)* Migrate to bulma-css-variables and introduce dark mode (#954)
* *(other)* Some dropdown.vue improvements
* *(other)* Try to cache list views
* *(other)* [skip ci] Updated translations via Crowdin


## [0.18.1] - 2021-09-08

### Bug Fixes
* Kanban-card mutatation violation (#712) ([4fc8858](4fc8858c64e9acf9072136c9bca256ec46249fdf))
* Call to /null from background image (#714) ([c9631c1](c9631c1e7126d70fac335a1c86b4e37ad889ae98))


### Features
* Make it possible to fake online state via dev env (#720) ([c409532](c4095327adec74a099e129403772e5e86f1359f8))


### Other

* *(other)* Update dependency axios to v0.21.4 (#705)

Reviewed-on: https://kolaente.dev/vikunja/frontend/pulls/705
Co-authored-by: renovate <renovatebot@kolaente.de>
Co-committed-by: renovate <renovatebot@kolaente.de>

* *(other)* Update typescript-eslint monorepo to v4.31.0 (#706)

Reviewed-on: https://kolaente.dev/vikunja/frontend/pulls/706
Co-authored-by: renovate <renovatebot@kolaente.de>
Co-committed-by: renovate <renovatebot@kolaente.de>

* *(other)* Fix translation badge

* *(other)* Update dependency vite-plugin-vue2 to v1.8.2 (#707)

Reviewed-on: https://kolaente.dev/vikunja/frontend/pulls/707
Co-authored-by: renovate <renovatebot@kolaente.de>
Co-committed-by: renovate <renovatebot@kolaente.de>

* *(other)* Fix rearranging tasks in a kanban bucket when its limit was reached

* *(other)* Update dependency vite to v2.5.4 (#708)

Reviewed-on: https://kolaente.dev/vikunja/frontend/pulls/708
Co-authored-by: renovate <renovatebot@kolaente.de>
Co-committed-by: renovate <renovatebot@kolaente.de>

* *(other)* Update dependency vite to v2.5.5 (#709)

Reviewed-on: https://kolaente.dev/vikunja/frontend/pulls/709
Co-authored-by: renovate <renovatebot@kolaente.de>
Co-committed-by: renovate <renovatebot@kolaente.de>

* *(other)* Update dependency jest to v27.1.1 (#716)

Reviewed-on: https://kolaente.dev/vikunja/frontend/pulls/716
Co-authored-by: renovate <renovatebot@kolaente.de>
Co-committed-by: renovate <renovatebot@kolaente.de>

* *(other)* Update dependency @4tw/cypress-drag-drop to v2 (#711)

Reviewed-on: https://kolaente.dev/vikunja/frontend/pulls/711
Co-authored-by: renovate <renovatebot@kolaente.de>
Co-committed-by: renovate <renovatebot@kolaente.de>

* *(other)* Fix data export download progress

* *(other)* Fix missing translation when creating a new task on the kanban board

* *(other)* Fix sort order for table view

* *(other)* Fix task attributes overridden when saving the task title with enter

* *(other)* 0.18.1 release preparations

## [0.18.2] - 2021-11-23

### Fixed

* fix(docker): properly replace api url
* fix: edit saved filter title

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
* PWA improvements (#622)
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
* Fix attachments being added multiple times
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
* Fix notification parsing for team member added
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
* Don't hide all lists of namespaces when losing network connectivity
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
* Make sure attachments are only added once to the list after uploading + Make sure the attachment list shows up every
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
* Move focus directive to separate file
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
* Add telegram release notification (#98)
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
* Fix team management (#121)
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
* Task assignees (#21)

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
