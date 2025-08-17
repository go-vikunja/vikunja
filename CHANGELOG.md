# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

All releases can be found on https://code.vikunja.io/vikunja/releases.

## [1.0.0-rc1] - 2025-08-17

### Bug Fixes

* *(desktop)* Correctly parse release rename flag
* *(desktop)* Do not try to publish release artifacts on tag

## [1.0.0-rc0] - 2025-08-17

### Breaking Changes

* **BREAKING**: *config*: read all env variables into config store explicitly
* **BREAKING**: disable 368 releases
* **BREAKING**: config for auth providers now use a map instead of an array
* **BREAKING**: remove echo log options - unify with general http logging
* **BREAKING**: rename right to permission (#1277)
* **BREAKING**: store sqlite file relative to rootpath (#934)
* **BREAKING**: remove deprecated config settings

### Bug Fixes

* *(a11y)* Add inline task labels
* *(a11y)* Add keyboard shortcuts button label
* *(a11y)* Add labels menu items
* *(a11y)* Add labels to comment interactions
* *(a11y)* Add labels to editor buttons
* *(a11y)* Add labels to logo
* *(a11y)* Add labels to project description label
* *(a11y)* Add labels to reaction button
* *(a11y)* Add notification icon label
* *(a11y)* Add task input label
* *(a11y)* Hide unfocusable buttons
* *(api)* Allow api tokens to retrieve the user who created the token
* *(api)* Return 404 response when using a token and the route does not exist
* *(attachments)* Add .jpeg to previewable image (#2770)
* *(attachments)* Always show dropzone
* *(attachments)* Check permissions when accessing all attachments
* *(attachments)* Left align attachment title
* *(attachments)* Return error message when attachment upload is not multipart form request
* *(auth)* Check for dot in username during validation
* *(auth)* Check for existence of field before casting
* *(auth)* Convert to int when failed password value is not int
* *(auth)* Do not allow commas in usernames
* *(auth)* Do not load login an openid auth components async
* *(auth)* Don't try to find openid provider when none are configured
* *(auth)* Fail when link share token is not parsable
* *(auth)* Fix regex in JWT decoding causing login issues for Chinese/Japanese user names (#809)
* *(auth)* Hide two factor authentication when using non-local user
* *(auth)* Jwt token decoding
* *(auth)* Load oidc provider before trying to use it
* *(auth)* Make sure routes from the "other" group work as intended
* *(auth)* Make text always white on image
* *(auth)* Move read all notifications to notification group
* *(auth)* Only use query parameters instead of local storage for password reset token (#770)
* *(auth)* Redirect to logout url after logging out
* *(auth)* Restrict max password length to 72 bytes
* *(auth)* Return correct error when trying to do a user action as a link share
* *(auth)* Return ldap as auth provider name when using it
* *(auth)* Return proper error when a jwt claim contains wrong data
* *(auth)* Set default data to empty when initializing password reset
* *(auth)* Size of password reset request button
* *(avatar)* Fallback to username when no name is set
* *(avatar)* Use keyvalue store to cache gravatar instead of map
* *(avatars)* Always return correct mime type for cached avatar
* *(background)* Validate unsupported formats and show error message (#1123)
* *(build)* Always use git tag for version number
* *(build)* Do not use our own goproxy
* *(build)* Prepare xgo once during release
* *(build)* Set correct frontend version in Docker builds (#1114)
* *(build)* Update corepack before using it
* *(caldav)* Do not crash with error 400 when fetching the list of all projects
* *(caldav)* Fetch saved filter
* *(caldav)* Make CalDAV REPORT request properly respond with VTODO objects (#1116)
* *(caldav)* Make sure colors are correctly saved and returned
* *(caldav)* Reject invalid project id with error 400
* *(caldav)* Return other status codes than 500 when projects are not found (#3065)
* *(caldav)* Use subpath for caldav url in frontend
* *(checkbox)* Use sibling css selector instead of has
* *(ci)* Correctly pass build args
* *(ci)* Correctly update translations from crowdin
* *(ci)* Crowdin download
* *(ci)* Do not build linuxx 368 docker images
* *(ci)* Do not build mage again in release
* *(ci)* Do not use pnpm during node js setup
* *(ci)* Embed release version in frontend (#999)
* *(ci)* Format desktop build script
* *(ci)* Improve cypress parallelization
* *(ci)* Login
* *(ci)* Move describe output to own job
* *(ci)* Move replacing version in desktop release to action
* *(ci)* Pin node setup to v4
* *(ci)* Push swagger docs changes via ssh
* *(ci)* Replace unstable version in filename
* *(ci)* Reuse frontend built in test
* *(ci)* Setup go when testing so that go compile steps are cached
* *(ci)* Smoke test database connection
* *(ci)* Sync translations from crowdin
* *(ci)* Syntax error
* *(ci)* Update before installing dependencies
* *(ci)* Update rate limit for unauthenticated routes
* *(ci)* Use correct host for migration tests
* *(ci)* Use correct version string for golangci-lint
* *(ci)* Use correct xgo cache path
* *(ci)* Use deploy key to push crowdin changes
* *(ci)* Use latest version for docker buildx setup
* *(ci)* Wait for database (#604)
* *(cmd)* Report error when the connection to the mail server failed
* *(colors)* Truncate longer hex color values
* *(comment)* Add validation check for the max comment length
* *(config)* Do not attempt to parse config values from env when they contain an invalid data type
* *(config)* Log fatal when timezone parse fails (#1077)
* *(config)* Set value when env variable contains string value
* *(cypress)* Assert visibility to make the test less flakey (#1138)
* *(date)* Do not format time values using dayjs for use in date pickers
* *(datepicker)* Set correct date ranges (#1021)
* *(datepickr)* Reset styles
* *(db)* Refactor filtering with subqueries (#701)
* *(desktop)* Use app.use to serve frontend files
* *(desktop)* Use pnpm in ci
* *(dev)* Enable go hardening workaround
* *(dev)* Ensure frontend assets before compile in magefile (#992)
* *(dev)* Ignore utils lint case
* *(dev)* Zed frontend task
* *(docker)* Use pnpm install directly
* *(docs)* Clarify team member username instead of id
* *(docs)* Improve swagger output by setting swaggertype and enums (#705)
* *(docs)* Remove invalid comma
* *(editor)* Add rounded edges to code highlight
* *(editor)* Make bubbling changes from outside work
* *(editor)* Make pasting a file work again
* *(editor)* Pasted link capturing trailing text (#912)
* *(editor)* Prevent links from extending after space (#1059)
* *(editor)* Restore the current value, not the one from a previous task
* *(editor)* Upload image via toolbar button
* *(events)* Do not crash filter event handler when triggered by a link share user
* *(events)* Report async errors via Sentry
* *(export)* Update only current user export file id
* *(favorites)* Do not return subtasks on favorites page
* *(files)* Configure the files path in files init instead of globally
* *(files)* Only use service rootpath for files when the files path is not absolute
* *(files)* Use absolute path everywhere
* *(filter)* Correctly create task positions during filter creation
* *(filter)* Correctly replace quoted values
* *(filter)* Correctly use user time zone when filtering for date fields
* *(filter)* Do not override filter include nulls query
* *(filter)* Do not replace labels keyword when the value is 'label'
* *(filter)* Do not show tasks in filter results when they are filtered out by labels
* *(filter)* Do not try to set timezone when doer does not exist
* *(filter)* Don't treat ' ' as label value
* *(filter)* Don't treat word value in filter value as operator
* *(filter)* Highlight multiple labels multiple times
* *(filter)* Make sure tasks are in a correct bucket and position when they are part of a date filter
* *(filter)* Use correct syntax for not in query in typesense
* *(filter)* Validate fields before using them
* *(filters)* Change assertion based on the environment
* *(filters)* Clarify usage of reminders in filters
* *(filters)* Correctly replace the same filter input part when it occurs multiple times
* *(filters)* Correctly transform and populate saved filter when creating and editing
* *(filters)* Do not crash when a filter is invalid
* *(filters)* Do not crash when paginating bucket with empty filter
* *(filters)* Do not replace filter or project values when the id value resolves to undefined
* *(filters)* Explicitly search in json when using postgres
* *(filters)* Ignore invalid task fields when recomputing task positions
* *(filters)* Immediately propagate changes
* *(filters)* Increase year value when using mysql and year < 1
* *(filters)* Make sure year is always at least 1
* *(filters)* Prevent position and bucket ID overriding position of existing tasks
* *(filters)* Return more details when the provided filter time zone is invalid
* *(filters)* Use correct filter string instead of object
* *(filters)* Validate filter expression when creating or updating filter
* *(frontend)* Mark only clicked task item (#891)
* *(gantt)* Reload the gantt chart when switching between projects
* *(home)* Explicitly use filter for tasks on home page when one is set
* *(i18n)* Add hr-HR to dayjs import languages
* *(i18n)* Add missing translations
* *(i18n)* Add translation for favorite project description
* *(i18n)* Adjust task strings
* *(i18n)* Capitalize Bulgarian label (#2784)
* *(i18n)* Change casing of Ukrainian language in selector
* *(i18n)* Make notification settings link translatable
* *(i18n)* Pass language to notification mail function
* *(i18n)* Remove duplicate api strings
* *(i18n)* Return proper error when language is empty
* *(i18n)* Translate all parts of reminder notifications
* *(i18n)* Translations json
* *(i18n)* Use actually set language for dates
* *(i18n)* Use correct Norwegian dialect for dayjs locales
* *(i18n)* Use only one function to get translations
* *(i18n)* Use same casing for all dayjs languages
* *(kanban)* Always make cover image full width
* *(kanban)* Correctly check bucket limit for saved filters or view filters
* *(kanban)* Correctly paginate filtered kanban buckets
* *(kanban)* Correctly set default bucket id when duplicating project
* *(kanban)* Disable create button when bucket limit is reached
* *(kanban)* Do not allow creating tasks in full bucket
* *(kanban)* Do not allow creating tasks in full bucket in frontend
* *(kanban)* Do not close task input after creating tasks
* *(kanban)* Do not mark first bucked as done bucket in filter bucket mode
* *(kanban)* Do not set bucket when it is null
* *(kanban)* Do not set filter by default
* *(kanban)* Increase dates when moving a task into the done bucket
* *(kanban)* Load full task when moving task between buckets
* *(kanban)* Make bucket query fixed per-view (#1007)
* *(kanban)* Make kanban full width on mobile
* *(kanban)* Make loading tasks for a bucket work
* *(kanban)* Make task creation loading spinner actually visible
* *(kanban)* Mark tasks done when creating them in the done bucket
* *(kanban)* Only stop adding tasks when a limit is set
* *(kanban)* Save updated position to store
* *(kanban)* Set new bucket id on task after moving it
* *(kanban)* Use full updated kanban bucket when moving task
* *(label)* Ignore existing ID during creation
* *(labels)* Correctly fall back to variable colors when no label color is set (#1124)
* *(labels)* Only show each label once
* *(labels)* Remove input interactivity when label edit is disabled
* *(labels)* Test error assertion
* *(labels)* Trigger task updated for bulk label task update
* *(labels)* Trigger task.updated event when removing a label from a task
* *(ldap)* Crop avatar when syncing
* *(ldap)* Return meaningful error when providing wrong credentials
* *(ldap)* Update user name and email during login
* *(link share)* Use selected view when opening link share
* *(mage)* Actually pass the cli parameter to the function
* *(mage)* Do not check files in hidden directories
* *(mail)* Do not fail testmail command when the connection could not be closed.
* *(migration)* Add more debug logging
* *(migration)* Cast to text
* *(migration)* Check if the provided file is a valid csv before importing
* *(migration)* Check if uploaded csv is empty
* *(migration)* Detect header lines in csv file when importing from TickTick (#937)
* *(migration)* Do not crash when relating a task to itself
* *(migration)* Do not fail when an attachment is too large
* *(migration)* Ensure project background gets exported and imported
* *(migration)* Fetch members when they do not exist
* *(migration)* Handle file errors in frontend
* *(migration)* Make sure tasks are associated to the correct view and bucket for data imported from Vikunja dump
* *(migration)* Reset buckets before creating related tasks so that they are actually created (#1015)
* *(migration)* Return proper error when uploaded file is not a zip file
* *(modal)* Do not prevent scrolling on mobile
* *(modal)* Make scrolling on iOS Safari work
* *(modal)* Make sure modal and its content scrolls properly on mobile
* *(modal)* Make sure multiple modals are stacked on top of each other
* *(multiselect)* Make selectPlaceholder optional
* *(notifications)* Handle user mentioned notification
* *(notifications)* Only add project subscription as task subscription when the user is not already subscribed to the task
* *(notifications)* Test assertion
* *(openid)* Check different provider types
* *(openid)* Lint
* *(openid)* Log error when config is still using array value
* *(openid)* Manually fetch providers
* *(password)* Validate password before sending request to api
* *(positions)* Directly look in the database to fetch tasks when recalculating their position
* *(project)* Add position in test fixtures
* *(project)* Correctly handle invalid project id error
* *(project)* Correctly migrate old project view filters
* *(project)* Correctly set done bucket after duplicating project
* *(project)* Make order stable in duplicate test
* *(project)* Only show create task cta when the user has permission to write to the project
* *(project)* Permission query on mysql
* *(project)* Reset id before creating
* *(project)* Restructure project drag handle
* *(project)* Set correct right after duplicating
* *(project)* Show description in title attribute without html
* *(project)* Transfer ownership after deleting a user
* *(projects)* (un-)archive child projects when archiving parent (#775)
* *(projects)* Adjust test assumptions
* *(projects)* Check with the current user if they have access to the project
* *(projects)* Correctly calculate the number of tasks and projects to delete
* *(projects)* Correctly check inherited permissions
* *(projects)* Description not visible on mobile
* *(projects)* Do not hide 6th project on project overview
* *(projects)* Do not try to fetch project permissions when no projects exist
* *(projects)* Only add conditions to query when they are non-empty
* *(projects)* Remove unnecessary join
* *(projects)* Return 0 if no parent project exists
* *(projects)* Return list of projects when accessing as link share
* *(projects)* Trigger only single mutation
* *(quick actions)* Add close button on mobile
* *(quick actions)* Always allow creating a new project or task, regardless of context
* *(quick actions)* Do not space between elements on mobile
* *(quick actions)* Quote label when it contains spaces (#1013)
* *(quick actions)* Show saved filters in search results
* *(quick actions)* Use default project when creating a new task via quick add magic without specifying a project
* *(release)* Use openrc for alpine (#1016)
* *(reminders)* Notify subscribed users as well
* *(restore)* Make sure all json columns are properly restored
* *(restore)* Restore encoded float values properly
* *(rtl)* Content list spacing
* *(rtl)* Don't convert logical properties to absolute
* *(rtl)* Icon button
* *(rtl)* Keyboard shortcuts trigger position
* *(rtl)* Make header work
* *(rtl)* Modal positioning
* *(rtl)* Task add input layout
* *(rtl)* User dropdown spacing
* *(saved filters)* Check permissions when accessing tasks of a filter
* *(service worker)* Use correct workbox version
* *(settings)* Move time zone selection to dropdown
* *(settings)* Properly align checkboxes
* *(settings)* Space input and label evenly
* *(settings)* Use correct test assertion (#1217)
* *(subscription)* Always return task subscription when subscribed to task and project
* *(subscriptions)* Cleanup and simplify fetching subscribers for tasks and projects logic
* *(subscriptions)* Correctly inherit subscriptions
* *(subscriptions)* Do not panic when a task does not have a subscription
* *(subscriptions)* Ignore task subscription when the user is subscribed to the project
* *(table)* Make sorting for two-word properties work
* *(task)* Add task to filter view after it was updated
* *(task)* Add tasks table prefix for sort order (#1003)
* *(task)* Align task title on mobile popup
* *(task)* Ambiguous description search (#1032)
* *(task)* Cleanup old task positions and task buckets when adding an updated or created task to filter
* *(task)* Correctly validate all task fields
* *(task)* Cyclomatic complexity
* *(task)* Do not allow moving a task to the project the task already belongs to
* *(task)* Do not allow saving an empty description
* *(task)* Do not show close button when the task was not opened via modal
* *(task)* Do not update all project_view ids
* *(task)* Dragging and dropping on mobile
* *(task)* Improve task delete modal on mobile
* *(task)* Make print styles work when printing task detail view from kanban
* *(task)* Make sure task comment url only contains one slash
* *(task)* Mark related task as done from the task detail view
* *(task)* Move task into new kanban bucket when moving between projects
* *(task)* Multiple overlapping defer due date popups
* *(task)* Open focused task when pressing enter
* *(task)* Open related task in popup when the other task was opened in a popup
* *(task)* Paginate task comments
* *(task)* Set current project after moving a task
* *(task)* Set done at date when moving a task to the done bucket
* *(task)* Show new due date immediately after deferring in list view
* *(task)* Specify task index when creating multiple tasks at once
* *(tasks)* Add new task only once to list when added
* *(tasks)* Also delete corresponding task positions (#2840)
* *(tasks)* Ambiguous done column in task sorting (#1011)
* *(tasks)* Creating subtasks with quick add magic should show up once
* *(tasks)* Default reminder to current date
* *(tasks)* Do not show import hint when using a filter as home tasks and already imported
* *(tasks)* Hide add button text on tablet
* *(tasks)* Prefix created and updated columns when sorting by them
* *(tasks)* Subtasks missing in list view (#1000)
* *(team)* Do not allow leaving exernal teams
* *(test)* Cypress test selector
* *(test)* Formatting
* *(test)* Set language in test
* *(test)* Use a date different from today as due date
* *(test)* Use correct assertion
* *(test)* Use correct selector for modal header
* *(test)* Use fixed date in due date test
* *(test)* Wait for project to be loaded
* *(test)* Wait for redirect
* *(tests)* Faker usage
* *(types)* Readd DOM.Iterable types
* *(typesense)* Add new tasks to typesense properly
* *(typesense)* Fetch task comments without permission check
* *(typesense)* Force position to always be float instead of auto-inferring
* *(typesense)* Index tasks one by one
* *(typesense)* Make sure task positions are recreated properly when updating them
* *(typesense)* Only fail silently when a project was not found during indexing
* *(typesense)* Use emplace instead of upsert to update documents
* *(typesense)* Use typesense bulk insert, log all errors
* *(typesense)* Use upsert instead of emplace when updating tasks in typesense
* *(typing)* Ensure HTMLElement refs (#918)
* *(typo)* Simpl -> Simple -> GetProjectsMapSimpleByTaskIDs (#2906)
* *(user)* Do not allow changing name in settings when the user originates from an external auth provider
* *(user)* Do not create user with existing id
* *(user)* Ensure deletion tokens can only be used by the user who created them
* *(user)* Persist status on email updates (#1084)
* *(user)* Show medium priority by default
* *(user)* Use correct link for user deletion
* *(users)* Refresh initials avatar refresh after name change (#1047)
* *(view)* Add unique index for task buckets (#1005)
* *(view)* Correctly get paginated task results
* *(view)* Correctly resolve bucket filter when paginating
* *(view)* Correctly resolve label for filtered views or buckets
* *(view)* Do not crash when saving a view
* *(views)* Add migration for filtered kanban buckets
* *(views)* Delete task buckets and task positions as well when deleting a view
* *(views)* Do not create task bucket and task position entries when duplicating a project
* *(views)* Enable search in bucket filters
* *(views)* Make searching in view filters work
* *(webhook)* Do not fail to send webhook when loading the project fails
* *config*: read all env variables into config store explicitly
* Add \n between scoped and unscoped commits in git cliff config ([9b85f3b](9b85f3bd0c8ce1d6123970b84386006b3c425de7))
* Add canRemove prop ([9d985f7](9d985f7e962b44ec0f8c8fe41e9c3653a198fe81))
* Add close button to keyboard shortcut button ([adaafaf](adaafafe9000b0acb1063dd8de5b026340af12f8))
* Add greater unicode range to font subset ([aea2f70](aea2f708d34ef0b0a87caefaeb842a1c0bbf1b83))
* Add migration for non-unique task buckets ([159961b](159961b5e0461d784eb91b73fe4ab91649d7f410))
* Add missing error messages to translations ([398d0c7](398d0c7ab500fa5027c013da9201d6b640173f38))
* Add newline at end of line (#827) ([bb9dc03](bb9dc03351acbc763d25dfb3d241c8a88c98cb98))
* Add test:e2e-nix command to make running cypress in devenv work ([2395c22](2395c22ad8baae19792d208b00bd6caa8f7a2b69))
* Adjust benchmark so that it only checks the task fetching ([e7f5142](e7f5142e3d890b4c63d5e48d896f9f713e824f5b))
* Adjust test for project view ([3ada7b6](3ada7b657b2721b40272ae63c159f0fb28de7da7))
* Allow setting task/project color to black ([12eb913](12eb91365ac1a0f80f371bee4eb44866c7dbda2a))
* Always add public url to allowed cors origins ([842e7f5](842e7f524b37218a4a24b05da094daf3bcb2dcb5))
* Ambiguous title column in task search (#1012) ([f9b31ab](f9b31ab4bf0aff0daa30235528129911eb8e4f77))
* Check if all required keys are available when parsing openid configuration ([4512045](4512045cbfccbf872dfa8ab322aa62a86ccf5ca0))
* Close modals by pressing escape (#932) ([5823688](58236884ddf4531ed3ce7f441c665d2f6e9030f8))
* Comment typo and misspellings ([d48d88d](d48d88d4423f1d19eb85c2f58f2c9baeef641bd4))
* Convert all css properties to logical ([16f7fa0](16f7fa087a7fa80dbb92a02c329facc9ef8a8288))
* Correct Team.createdBy type (#965) ([a2ac4bd](a2ac4bdc7f0deaa0df0c46fac6c065e87f376779))
* Correct comments ([342bbd6](342bbd6192284059c18f45a8229d78a926efc638))
* Correct license header references (#882) ([296577a](296577a875bbe9ebae4fe4fd699bd62d0673b86f))
* Correct trailing comma in tsconfig (#970) ([06d5791](06d5791568a28835244616d43dc8fd7f60f8752f))
* Correct unknown subscription entity typo (#883) ([62f0487](62f048767cce0c55f40efa706db133a6861e4a10))
* Correctly cache unsplash background ([a5591c1](a5591c16032f5746a7d84a383a6d6d114565561c))
* Correctly return cached intitals avatar ([c3fd659](c3fd659851e66160eeaf435afd70bbf27deb944c))
* Correctly return cached provider ([566657c](566657c54a4f37a61ff96b9c03a8924d239e2425))
* Correctly set default props in UserTeam ([3f4bf59](3f4bf592836f4744ec6cc4c22e3aeeb7566b870e))
* Create missing indexes on postgres ([a70c472](a70c472aa3ebe4a50fa1c241f848e0a3d2b270af))
* Cypress selector (#753) ([4937127](4937127898403e38e28e825f95ce25dd71294224))
* Demo banner positioning ([63732a3](63732a37c010fdb67c0cf001fd3e0df76e0393ef))
* Deprecated import in useTitle ([c7e708c](c7e708cf7d0f485fafc2cdca6bf7884f20bda72c))
* Disable 368 releases ([6fe22ab](6fe22aba395fa3ee7dde241014addf3d51c062d2))
* Do not fail to load projects without views via link share ([9eb5c62](9eb5c62b0122e4ab0f8e5d12480b248712ebf64e))
* Do not prefix tasks all the time ([3ad5797](3ad57973077a7e2143d3c0c820b0aae3c48a07dc))
* Do not try to reindex tasks into typesense when it is not set up ([8ab3873](8ab387396d2337fdc1ea54851dbcfbb0721c910b))
* Don't panic when using api token when not correctly put into context (#1119) ([42534cd](42534cdd79c02ff7d73ff023c3b681ac810c5730))
* Download badge in readme ([4238a5f](4238a5f6a3324ff26b86c893c09dcc38c7089c10))
* Emit for DatepickerWithValues ([fb91e73](fb91e73a3c9d0324b1c9fd6559a47190de5e8b58))
* Error message check on mysql ([29107e9](29107e9865a4a2aeaf3818b09c6a5639704117db))
* Error reporting ([9219f70](9219f7032ef41b6832c9a8f5e7d717fa53cc7be2))
* Fix "null" in project views (#1158)
* Fix(tasks): do not return subtasks multiple times when they are related to multiple tasks
* Gantt reset button ([4532cdf](4532cdfa006d86f25ffaacbfe8833bacdb5b172c))
* Generate config in ci ([8776465](8776465fa20b3d9892f1914637a8ee577e351bcb))
* Git ignore all dist folders ([8bada3e](8bada3e96710f30aff55b336b4bf79b3ad4b4e89))
* Global component types ([b5cb984](b5cb98498a40cbd166ee7a5ec0ce240a06b46af5))
* Guard invalid user lookups (#1034) ([a8025a9](a8025a9e364871619c331a533dfb563f9eb051e3))
* Guard null access in composables (#951) ([78fa574](78fa5742c34490bb1f045a244b36ef87879081ec))
* Hide empty sidebar navs ([7736440](773644050f68338d69459f3505fbfa02b48c97ea))
* Hide icon if description is missing ([abb4126](abb4126bce69ef55cf1cd139f745f48a5780f283))
* I18n missing translation key ([cbbc4c0](cbbc4c0372e54a2a381d02957aa4f53f3ed432a3))
* Improve ldap sanitization (#1155) ([ad0cf7a](ad0cf7a13cad1d7e9b9922651c42c6d407aa2370))
* Improve markdown paste detection (#939) ([a9714f6](a9714f6a4ae6274cb6dc88a1c38ddbdf35de6460))
* Include diacritics characters in generated font files ([3b8258d](3b8258d57ed011a00bb057ed9394f13418246da6))
* LinkSharing race condition (#2932) ([6a2a8c1](6a2a8c106bcf6eb12c52f41fc5772d10350f8cc4))
* Lint ([16c9d2f](16c9d2f6f97a35572262209e75695d52fd126cd8))
* Lint ([2394d1f](2394d1f8e0c999e66d58463134b7a3fdd98b535c))
* Lint ([302424b](302424b0477fca797d933130e0bf27094af8a71e))
* Lint ([32bdef8](32bdef841f9ff2875fb0da385d1f1a64b3ddb3b3))
* Lint ([61333c9](61333c9b7f7bd8865b521c5b2cf03f2387577dd4))
* Lint ([703a88e](703a88e99f536a696d56eeac8af9928b09f8859f))
* Lint ([acdb45a](acdb45a92c15bb2858088eeef1f3ae1a110f5918))
* Lint ([bf5cafc](bf5cafc03f661bc534f47adfff4b92f06ca2ab57))
* Lint issues ([703c641](703c641aeb586e8158ad619827d432574ec28fcf))
* Log correct response status ([388af80](388af80ece203a3f5d2cdcbb5ca73fa9fb219270))
* Lowlight imports for v3 ([a61e2d0](a61e2d064d18c99c298d3a84b9de39fd24ce3638))
* Make search in saved filter work ([624907a](624907ad6a8a4aff771051eef428dcfd317c8e22))
* Make sure task text items are flex ([0c5c385](0c5c385a862d6154a441214fdaf166e1da704d14))
* Make user data export download return 404 for nonexistent files  (#1227) ([7762d77](7762d7746e2a1815e9c7c30fd4525e970265ecbb))
* Markdown spelling ([d79a601](d79a60183d7b2e3692314557b30192d2e66105e4))
* Method name ([47538ca](47538ca8103ede8f7b686cfa61b3e312c76eff3a))
* MySQL constraint violations returning HTTP 500 instead of 400 for task bucket duplicates (#1154) ([9712dbe](9712dbe2ab53ad0021f7d9373c032ab1f11b32ab))
* Normalize frontend version string (#1289) ([7371f77](7371f7729edd9f2b7b07d4c84c01af758ff828e4))
* Only pass filter string from view to popup ([85823b3](85823b3e9747d4ab1c60b8b3b76220a1d8030e4f))
* Open keyboard shortcuts by pressing ? (#928) ([3e7c009](3e7c0099c88212f48285ec6a5ff75323f35854b9))
* Panic on restoring with numeric position fields (#1089) ([ecc95e9](ecc95e9139aec70a6f483457b1d6cf46a2998b26))
* Param type (#895) ([508cdf0](508cdf027cdc1c3e65f4442d1337bfd900f642a9))
* Partial fix to allow list tasks in ios reminders app (#2717) ([84dbc5f](84dbc5fd8467418be6390f4fe9eee9abdc50bf45))
* Pass pointer when fetching provider ([2b497e6](2b497e62650c925db98f6d27791989c265575ae7))
* Pin xgo to 1.22.x ([de1eac5](de1eac5d368b17b22cc649476e866f0ecbf94c92))
* Postcss-easing-gradient types ([6d3a30c](6d3a30c799ac8db3ee0e3de83560ebd39b354517))
* Prevent login screen flash when already authenticated (#933) ([eac6c98](eac6c98bf15435250d819cbedcefe2c1476e2702))
* Reactive ancestor projects ([90951d4](90951d4003d1ae46f4913785aae3a2d8a4867edb))
* Release bucket name ([af11a65](af11a6527f3813f1049f6bed2640f6e7eade50f7))
* Remove @types/lodash.clonedeep ([98c10ac](98c10acb50e13755ec53c900a95d570e7b1acc8d))
* Remove babel vscode extension ([e3793bc](e3793bcfbd70510418247a5c1872f81b2cca51f5))
* Remove console log ([413798b](413798b3211ee53e0e1c70a8b616979906c70635))
* Remove date-fns (#3039) ([021d71b](021d71b90e62237b1faf6d8eac51f0429cf8586f))
* Remove defineProps ([f3e77eb](f3e77eb1f076489617fc8d1bc9e56418a61ad4e1))
* Remove dompurify stub types ([13d52c7](13d52c721df576c7a507f71967644887a37e4f52))
* Remove fmt output in token check ([efff695](efff6955c5ec267cd3ce3e7821285b8e364b785f))
* Remove postcss-easings type ([c52b7ef](c52b7efc970e88a2bb8009da90118e5fc3c669a3))
* Remove second notification on undo task update (#1060) ([4090b13](4090b1377237367f064c8580e51f7171ff631409))
* Remove unused import ([3e46457](3e46457c03d1496141ceb66aa505979e9fa4e8f0))
* Reset id before creating ([c252c8f](c252c8f0cd1c09a74b564e67a855cca4cd436585))
* ResetEmptyTitleError (#2889) ([07df606](07df606c68e486d430ee835625e9a2946734dc6a))
* Return correct mimetype for openapi docs.json ([44b3e46](44b3e463255765ef85f35385aa4b42069b3af9d8))
* Return meaningful error message when selecting an invalid timezone ([65df9e5](65df9e5ef9c25619dae96f039b7b78f7a703f193))
* Show 404 on task detail page when the task does not exist (#1014) ([0f3da11](0f3da11bc43d7cc3432d6ac25bb295c141d909a0))
* Show close button on mobile popups ([ce57d85](ce57d85f04eca3ee1f797deff37d730ff18ed6fc))
* Specify cols when upgrading ([942c2e4](942c2e4af6eb6040e287d34d92061a40c3dbffa3))
* Start server when listening on socket ([b85befb](b85befb86abd6dc5bfb09d9ebaee8a736d1688e6))
* Strip label syntax with parentheses from task title (#1300) ([fc55563](fc55563dddb160a30ebbe69e9058a2f357483060))
* Style issues ([63319e1](63319e19adb5f82675d2c4d46747a95055eb8a9b))
* Subscription should only be visible for the user who subscribed  (#1183) ([e108374](e10837476ae2d7d8653eba7d6de22fd98302eea1))
* Swagger docs ([9aa197b](9aa197b196fdbbb5787494dd87c490ce23010866))
* Switch to wine electron builder ([2693419](26934199594f33e24f62462d9beeb8dde3c4d09b))
* TOTP account lock notification typo (#858) ([8632bd2](8632bd2063daecad205e63766d43791c3fc4e464))
* Task overdue at the same time as the notification ([8d8406d](8d8406df0582e7347517a8ea251abcde40c55b3d))
* TaskCollection types (#754) ([a62ac80](a62ac800c4a52aeb0407a0d2a4dceb7ffbd75247))
* Test selector ([c5b82fc](c5b82fc591c9056758eafeeae7f23c9652e8fb7f))
* Textarea autosize for LanguageTool ([7ef1e0a](7ef1e0a3e564ae245efc56b0b2764d2561d2929a))
* TipTap reactive prop destructuring ([30daf08](30daf08b54c9f2612a0c33bd06c65a41d5bd7f83))
* Typing reactive in ProjectSearch ([9814ff9](9814ff966776717281eb819a5b02a87ca2f99ed8))
* Upgrade xgo ([f826fb9](f826fb9a91007d22082af3421cf8b2876c3bb036))
* Upgrade xgo docker image everywhere ([d6194b8](d6194b8f1003e1e85708db7b0f9502a875292b80))
* Upload avatar caching ([45e7f6e](45e7f6e316d4167910de70a10010fdaf3d935679))
* Use @tsconfig/node22 ([64ffba2](64ffba28130aa4a66a7fa65c520bc584e14e38d8))
* Use assertions which are more specific ([7985a65](7985a6500a70ec498005fd7a9f999257496902d5))
* Use modern-compiler for sass files as well ([452cc66](452cc66b329149d2f3b5daf6b6ee07788c8109e5))
* Vite config linting ([a24c64d](a24c64da8ffa5f7de5ec1f05615895c0da4dc668))
* Vue/no-boolean-default NoAuthWrapper ([460d6ac](460d6ac8a48f89fbc6e66a0c6ca977c7fcc8ed8d))
* Workbox outDir for build:test ([c118e78](c118e788b8d180413923eb2e631dc8cb2d9cf4b9))

### Dependencies

* *(deps)* Update module github.com/coreos/go-oidc/v3 to v3.11.0
* *(deps)* Update flake
* *(deps)* Update go toolchain to 1.22.5
* *(deps)* Update dependency node to v20.16.0
* *(deps)* Update golangci
* *(deps)* Update tiptap to 2.6.6
* *(deps)* Update github.com/wneessen/go-mail to v0.4.4
* *(deps)* Update dependency dayjs to v1.11.13
* *(deps)* Update module github.com/typesense/typesense-go to v2
* *(deps)* Update module golang.org/x/term to v0.24.0
* *(deps)* Update module golang.org/x/text to v0.18.0
* *(deps)* Update dependency vue-i18n to v9.14.0
* *(deps)* Update dependency pinia to v2.2.2
* *(deps)* Update dependency @sentry/vue to v8.28.0
* *(deps)* Update dependency @kyvg/vue3-notification to v3.3.0
* *(deps)* Update module github.com/threedotslabs/watermill to v1.3.7
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.23
* *(deps)* Update module github.com/gabriel-vasile/mimetype to v1.4.5
* *(deps)* Update dependency sortablejs to v1.15.3
* *(deps)* Update dependency axios to v1.7.7
* *(deps)* Update dependency tailwindcss to v3.4.10
* *(deps)* Update vueuse to v11
* *(deps)* Update module golang.org/x/oauth2 to v0.23.0
* *(deps)* Update module golang.org/x/crypto to v0.27.0
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.6.1
* *(deps)* Update module dario.cat/mergo to v1.0.1
* *(deps)* Update dependency vue to v3.5.3
* *(deps)* Update module github.com/prometheus/client_golang to v1.20.3
* *(deps)* Update module golang.org/x/image to v0.20.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency go to v1.23.1
* *(deps)* Update dependency vue-router to v4.4.3
* *(deps)* Update dev-dependencies
* *(deps)* Update pnpm to v9.10.0
* *(deps)* Update dependency @sentry/vue to v8.29.0
* *(deps)* Update dependency vue to v3.5.4
* *(deps)* Update dependency express to v4.20.0
* *(deps)* Update module github.com/getsentry/sentry-go to v0.29.0
* *(deps)* Update dependency @sentry/vue to v8.30.0
* *(deps)* Update dependency vue-i18n to v10
* *(deps)* Update dependency tailwindcss to v3.4.11
* *(deps)* Update dependency vue-i18n to v10.0.1
* *(deps)* Update dependency vue-router to v4.4.4
* *(deps)* Update dependency express to v4.21.0
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v5
* *(deps)* Update dependency vue to v3.5.5
* *(deps)* Update dependency vue-router to v4.4.5
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency date-fns to v4
* *(deps)* Update dev-dependencies
* *(deps)* Update vueuse to v11.1.0
* *(deps)* Update module github.com/prometheus/client_golang to v1.20.4
* *(deps)* Update tiptap to v2.7.0
* *(deps)* Update dev-dependencies
* *(deps)* Update tiptap to v2.7.1
* *(deps)* Update dependency tailwindcss to v3.4.12
* *(deps)* Update dependency vue to v3.5.6
* *(deps)* Update tiptap to v2.7.2
* *(deps)* Update module github.com/typesense/typesense-go to v2
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency vuemoji-picker to v0.3.1
* *(deps)* Update pnpm to v9.11.0
* *(deps)* Update desktop lockfile
* *(deps)* Update dependency vue to v3.5.7
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency vue-i18n to v10.0.2
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency vue to v3.5.8
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v5.1.0
* *(deps)* Update dependency vue-i18n to v10.0.3
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v5.2.0
* *(deps)* Update dependency @sentry/vue to v8.31.0
* *(deps)* Update dependency tailwindcss to v3.4.13
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency @sentry/vue to v8.32.0
* *(deps)* Update tiptap to v2.7.3
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency vue to v3.5.9
* *(deps)* Update dependency dompurify to v3.1.7
* *(deps)* Update tiptap to v2.7.4
* *(deps)* Update dependency vue to v3.5.10
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency pinia to v2.2.3
* *(deps)* Update tiptap to v2.8.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency pinia to v2.2.4
* *(deps)* Update dependency go to v1.23.2
* *(deps)* Update pnpm to v9.12.0
* *(deps)* Update dependency @sentry/vue to v8.33.0
* *(deps)* Update dependency rollup to v4.24.0
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.24
* *(deps)* Update module golang.org/x/text to v0.19.0
* *(deps)* Update module golang.org/x/image to v0.21.0
* *(deps)* Update module golang.org/x/sys to v0.26.0
* *(deps)* Update dependency vue to v3.5.11
* *(deps)* Update dependency @sentry/vue to v8.33.1
* *(deps)* Update module golang.org/x/term to v0.25.0
* *(deps)* Update dependency vue-i18n to v10.0.4
* *(deps)* Update module golang.org/x/crypto to v0.28.0
* *(deps)* Update dependency caniuse-lite to v1.0.30001667
* *(deps)* Update pnpm to v9.12.1
* *(deps)* Update dependency @kyvg/vue3-notification to v3.4.0
* *(deps)* Update dependency @types/node to v20.16.11
* *(deps)* Update dependency express to v4.21.1
* *(deps)* Update dependency typescript to v5.6.3
* *(deps)* Update dependency @sentry/vue to v8.34.0
* *(deps)* Update dependency vue to v3.5.12
* *(deps)* Update module github.com/yuin/goldmark to v1.7.5
* *(deps)* Update module github.com/yuin/goldmark to v1.7.6
* *(deps)* Update module github.com/getsentry/sentry-go to v0.29.1
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.6.2
* *(deps)* Update module github.com/prometheus/client_golang to v1.20.5
* *(deps)* Update dependency tailwindcss to v3.4.14
* *(deps)* Update module github.com/gabriel-vasile/mimetype to v1.4.6
* *(deps)* Update module github.com/yuin/goldmark to v1.7.7
* *(deps)* Update module github.com/yuin/goldmark to v1.7.8
* *(deps)* Update pnpm to v9.12.2
* *(deps)* Update module github.com/wneessen/go-mail to v0.5.1
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.7.0
* *(deps)* Update dependency @sentry/vue to v8.35.0
* *(deps)* Update tiptap to v2.9.0
* *(deps)* Update tiptap to v2.9.1
* *(deps)* Update dependency caniuse-lite to v1.0.30001672
* *(deps)* Update dev-dependencies
* *(deps)* Update pnpm to v9.12.3
* *(deps)* Update dependency @kyvg/vue3-notification to v3.4.1
* *(deps)* Update dev-dependencies
* *(deps)* Update module github.com/swaggo/swag to v1.16.4
* *(deps)* Update dev-dependencies
* *(deps)* Update devenv
* *(deps)* Update module github.com/threedotslabs/watermill to v1.4.0
* *(deps)* Update dependency workbox-precaching to v7.3.0
* *(deps)* Update dependency node to v22
* *(deps)* Update dependency pinia to v2.2.5
* *(deps)* Update dev-dependencies
* *(deps)* Set workbox version to 7.3.0
* *(deps)* Update node.js to v22.11.0
* *(deps)* Update vueuse to v11.2.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency @sentry/vue to v8.36.0
* *(deps)* Update dev-dependencies
* *(deps)* Update module github.com/threedotslabs/watermill to v1.4.1
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency pinia to v2.2.6
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency @sentry/vue to v8.37.1
* *(deps)* Update dependency go to v1.23.3
* *(deps)* Update dev-dependencies
* *(deps)* Update module github.com/wneessen/go-mail to v0.5.2
* *(deps)* Update module golang.org/x/sync to v0.9.0
* *(deps)* Update dev-dependencies
* *(deps)* Update module golang.org/x/image to v0.22.0
* *(deps)* Update module golang.org/x/sys to v0.27.0
* *(deps)* Update module golang.org/x/crypto to v0.29.0
* *(deps)* Update module golang.org/x/oauth2 to v0.24.0
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v5.3.0
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v5.3.1
* *(deps)* Update dependency rollup to v4.25.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency dompurify to v3.2.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency @sentry/vue to v8.38.0
* *(deps)* Update pnpm to v9.13.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency tailwindcss to v3.4.15
* *(deps)* Update pnpm to v9.13.1
* *(deps)* Update pnpm to v9.13.2
* *(deps)* Update dependency sass-embedded to v1.81.0
* *(deps)* Update dependency vue to v3.5.13
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v6
* *(deps)* Update vulnerable dependencies of dependencies
* *(deps)* Update font awesome to v6.7.0
* *(deps)* Update pnpm to v9.14.1
* *(deps)* Update dependency @sentry/vue to v8.39.0
* *(deps)* Update module github.com/gabriel-vasile/mimetype to v1.4.7
* *(deps)* Update tiptap to v2.10.0
* *(deps)* Update dependency dompurify to v3.2.1
* *(deps)* Update pnpm to v9.14.2
* *(deps)* Update font awesome to v6.7.1
* *(deps)* Update vueuse to v11.3.0
* *(deps)* Update tiptap to v2.10.2
* *(deps)* Update dependency @sentry/vue to v8.40.0
* *(deps)* Update module github.com/stretchr/testify to v1.10.0
* *(deps)* Update dependency sortablejs to v1.15.4
* *(deps)* Update dependency vue-router to v4.5.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency @types/node to v22.9.4
* *(deps)* Update dependency axios to v1.7.8
* *(deps)* Update tiptap to v2.10.3
* *(deps)* Update dev-dependencies to v6
* *(deps)* Update dev-dependencies
* *(deps)* Update vueuse to v12
* *(deps)* Update dependency pinia to v2.2.7
* *(deps)* Update dependency vue-i18n to v10.0.5
* *(deps)* Update dependency sortablejs to v1.15.5
* *(deps)* Update dependency sortablejs to v1.15.6
* *(deps)* Update pnpm to v9.14.3
* *(deps)* Update dependency pinia to v2.2.8
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency @sentry/vue to v8.41.0
* *(deps)* Update dependency dompurify to v3.2.2
* *(deps)* Update pnpm to v9.14.4
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency @sentry/vue to v8.42.0
* *(deps)* Update dependency tailwindcss to v3.4.16
* *(deps)* Update dependency axios to v1.7.9
* *(deps)* Update dependency pinia to v2.3.0
* *(deps)* Update module golang.org/x/sync to v0.10.0
* *(deps)* Update dependency sass-embedded to v1.82.0
* *(deps)* Update dependency go to v1.23.4
* *(deps)* Update module github.com/getsentry/sentry-go to v0.30.0
* *(deps)* Update module golang.org/x/sys to v0.28.0
* *(deps)* Update module golang.org/x/text to v0.21.0
* *(deps)* Update module golang.org/x/image to v0.23.0
* *(deps)* Update module github.com/labstack/echo-jwt/v4 to v4.2.1
* *(deps)* Update dependency cypress to v13.16.1
* *(deps)* Update module github.com/labstack/echo-jwt/v4 to v4.3.0
* *(deps)* Update dependency vuemoji-picker to v0.3.2
* *(deps)* Update dependency express to v4.21.2
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v6.0.1
* *(deps)* Update dev-dependencies
* *(deps)* Update pnpm to v9.15.0
* *(deps)* Update dependency rollup to v4.28.1
* *(deps)* Update dependency dompurify to v3.2.3
* *(deps)* Update dev-dependencies to v8.18.0
* *(deps)* Update dependency @sentry/vue to v8.43.0
* *(deps)* Update module github.com/labstack/echo/v4 to v4.13.1
* *(deps)* Update dev-dependencies
* *(deps)* Update module github.com/labstack/echo/v4 to v4.13.2
* *(deps)* Update dependency sass-embedded to v1.83.0
* *(deps)* Update dependency @sentry/vue to v8.45.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency @sentry/vue to v8.45.1
* *(deps)* Update font awesome to v6.7.2
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency @sentry/vue to v8.46.0
* *(deps)* Update dependency tailwindcss to v3.4.17
* *(deps)* Update dependency cypress to v13.17.0
* *(deps)* Update dependency @sentry/vue to v8.47.0
* *(deps)* Update module github.com/labstack/echo/v4 to v4.13.3
* *(deps)* Update tiptap to v2.10.4
* *(deps)* Update pnpm to v9.15.1
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency node to v22.12.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v6.0.2
* *(deps)* Update vueuse to v12.1.0
* *(deps)* Update vueuse to v12.2.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency happy-dom to v16
* *(deps)* Update dependency vue-i18n to v11
* *(deps)* Update dev-dependencies
* *(deps)* Update pnpm to v9.15.2
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v6.0.3
* *(deps)* Update vueuse to v12.3.0
* *(deps)* Update tiptap to v2.11.0
* *(deps)* Update dev-dependencies
* *(deps)* Update module golang.org/x/sys to v0.29.0
* *(deps)* Update module github.com/coreos/go-oidc/v3 to v3.12.0
* *(deps)* Update module golang.org/x/oauth2 to v0.25.0
* *(deps)* Update module github.com/gabriel-vasile/mimetype to v1.4.8
* *(deps)* Update pnpm to v9.15.3
* *(deps)* Update dev-dependencies
* *(deps)* Update module golang.org/x/crypto to v0.32.0
* *(deps)* Update dependency @sentry/vue to v8.48.0
* *(deps)* Update dependency node to v22.13.0
* *(deps)* Update dev-dependencies
* *(deps)* Update node.js to v22.13.0
* *(deps)* Update module github.com/getsentry/sentry-go to v0.31.1
* *(deps)* Update module mvdan.cc/xurls/v2 to v2.6.0
* *(deps)* Update module github.com/threedotslabs/watermill to v1.4.2
* *(deps)* Update dependency @cypress/vite-dev-server to v6
* *(deps)* Update dev-dependencies
* *(deps)* Update module github.com/spf13/afero to v1.12.0
* *(deps)* Update dependency @sentry/tracing to v7.120.3
* *(deps)* Update vueuse to v12.4.0
* *(deps)* Update dependency wait-on to v8.0.2
* *(deps)* Update module github.com/wneessen/go-mail to v0.6.0
* *(deps)* Update tiptap to v2.11.1
* *(deps)* Update tiptap to v2.11.2
* *(deps)* Update dependency eslint to v9.18.0
* *(deps)* Update pnpm to v9.15.4
* *(deps)* Update dev-dependencies
* *(deps)* Update module github.com/threedotslabs/watermill to v1.4.3
* *(deps)* Update goreleaser/nfpm docker tag to v2.41.2
* *(deps)* Pin dependency @tiptap/starter-kit to 2.11.2
* *(deps)* Update dependency electron to v34
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency @sentry/vue to v8.49.0
* *(deps)* Update dependency @sentry/vue to v8.50.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency lowlight to v3
* *(deps)* Update module github.com/threedotslabs/watermill to v1.4.4
* *(deps)* Update dependency go to v1.23.5
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency vitest to v3.0.2
* *(deps)* Update dependency caniuse-lite to v1.0.30001695
* *(deps)* Update module github.com/wneessen/go-mail to v0.6.1
* *(deps)* Update dependency rollup to v4.31.0
* *(deps)* Update dependency pinia to v2.3.1
* *(deps)* Update dev-dependencies
* *(deps)* Update ws, vulnerable dependencies of dependencies
* *(deps)* Update dependency node to v22.13.1
* *(deps)* Update dev-dependencies
* *(deps)* Update vueuse to v12.5.0
* *(deps)* Update tiptap to v2.11.3
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency @sentry/vue to v8.51.0
* *(deps)* Update node.js to v22.13.1
* *(deps)* Update dev-dependencies
* *(deps)* Upgrade cypress docker image in ci
* *(deps)* Upgrade vitest to 3.0.5
* *(deps)* Bump esbuild from 0.24.2 to 0.25.0
* *(deps)* Update dompurify to 3.2.4
* *(deps)* Update oauth2 and go-jose
* *(deps)* Update devenv
* *(deps)* Update golangci-lint
* *(deps)* Bump vue-i18n from 11.0.1 to 11.1.2
* *(deps)* Bump @babel/helpers to 7.26.10
* *(deps)* Bump axios to 1.8.2
* *(deps)* Bump github.com/redis/go-redis/v9 from 9.7.0 to 9.7.3
* *(deps)* Bump github.com/golang-jwt/jwt/v5 from 5.2.1 to 5.2.2
* *(deps)* Update go.sum
* *(deps)* Update golang.org/x/net to 0.38.0
* *(deps)* Update pnpm to v9.15.9 (#436)
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v6.0.5 (#439)
* *(deps)* Update dependency axios to v1.8.4 (#440)
* *(deps)* Update module github.com/labstack/echo-jwt/v4 to v4.3.1 (#441)
* *(deps)* Update module github.com/wneessen/go-mail to v0.6.2 (#443)
* *(deps)* Update dev-dependencies (#447)
* *(deps)* Update node.js to v22.14.0 (#449)
* *(deps)* Update module github.com/go-sql-driver/mysql to v1.9.1 (#452)
* *(deps)* Update module github.com/spf13/afero to v1.14.0 (#454)
* *(deps)* Update module golang.org/x/image to v0.25.0 (#457)
* *(deps)* Update vueuse to v12.8.2 (#459)
* *(deps)* Update module github.com/spf13/cobra to v1.9.1 (#455)
* *(deps)* Update module github.com/coreos/go-oidc/v3 to v3.13.0 (#451)
* *(deps)* Update module github.com/spf13/viper to v1.20.1 (#456)
* *(deps)* Update module github.com/threedotslabs/watermill to v1.4.6 (#442)
* *(deps)* Update tiptap to v2.11.7 (#444)
* *(deps)* Update dev-dependencies (#460)
* *(deps)* Pin dependencies (#435)
* *(deps)* Update dependency go to v1.24.1 (#446)
* *(deps)* Update module github.com/prometheus/client_golang to v1.21.1 (#453)
* *(deps)* Update dependency marked to v15.0.7 (#464)
* *(deps)* Update vueuse to v13 (major) (#463)
* *(deps)* Update dependency express to v5 (#466)
* *(deps)* Update dependency pinia to v3 (#467)
* *(deps)* Update dependency vite to v6.2.4 [security] (#469)
* *(deps)* Update dev-dependencies (#472)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.25 (#598)
* *(deps)* Update module github.com/typesense/typesense-go/v2 to v3 (#599)
* *(deps)* Update ghcr.io/techknowlogick/xgo:go-1.23.x docker digest to 770db5a (#602)
* *(deps)* Update golangci/golangci-lint-action action to v7 (#462)
* *(deps)* Update module github.com/arran4/golang-ical to v0.3.2 (#605)
* *(deps)* Update postgres docker tag to v17 (#607)
* *(deps)* Update dependency go to v1.24.2 (#603)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.26 (#611)
* *(deps)* Update module github.com/ganigeorgiev/fexpr to v0.5.0 (#615)
* *(deps)* Pin useblacksmith/setup-go action to 647ac64
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.27
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency dompurify to v3.2.5
* *(deps)* Update module github.com/coreos/go-oidc/v3 to v3.14.1
* *(deps)* Update dev-dependencies
* *(deps)* Pin actions/setup-node action to cdca736
* *(deps)* Update dev-dependencies
* *(deps)* Pin kolaente/s3-action action to f154255
* *(deps)* Update module golang.org/x/sync to v0.13.0
* *(deps)* Update module golang.org/x/oauth2 to v0.29.0
* *(deps)* Update module golang.org/x/sys to v0.32.0
* *(deps)* Update kolaente/s3-action action to v1.2.1
* *(deps)* Update dependency vue-i18n to v11.1.3
* *(deps)* Update histoire to 1.0.0-alpha-2
* *(deps)* Update dev-dependencies
* *(deps)* Update module golang.org/x/text to v0.24.0
* *(deps)* Update module golang.org/x/image to v0.26.0
* *(deps)* Update postgres:17 docker digest to fe3f571
* *(deps)* Update module golang.org/x/term to v0.31.0
* *(deps)* Update module golang.org/x/crypto to v0.37.0
* *(deps)* Update dependency marked to v15.0.8
* *(deps)* Update dependency ufo to v1.6.1
* *(deps)* Update module github.com/go-sql-driver/mysql to v1.9.2
* *(deps)* Update vueuse to v13.1.0
* *(deps)* Update module github.com/prometheus/client_golang to v1.22.0
* *(deps)* Update dependency cypress to v14.3.0 (#650)
* *(deps)* Update pnpm to v10.8.0 (#649)
* *(deps)* Update mariadb:11 docker digest to 81e8930 (#651)
* *(deps)* Update dependency pinia to v3.0.2 (#653)
* *(deps)* Update dependency electron to v35.1.5 (#654)
* *(deps)* Update dependency @sentry/vue to v9 (#461)
* *(deps)* Update pnpm to v10 (#636)
* *(deps)* Update module github.com/getsentry/sentry-go to v0.32.0 (#658)
* *(deps)* Update module github.com/getsentry/sentry-go/echo to v0.32.0 (#657)
* *(deps)* Update dev-dependencies (#659)
* *(deps)* Update dependency @types/node to v22.14.1 (#665)
* *(deps)* Update dependency rollup to v4.40.0 (#666)
* *(deps)* Update actions/setup-node digest to 49933ea (#668)
* *(deps)* Update dependency @faker-js/faker to v9.7.0 (#667)
* *(deps)* Update pnpm to v10.8.1 (#669)
* *(deps)* Update dev-dependencies to v8.30.1 (#672)
* *(deps)* Update module github.com/yuin/goldmark to v1.7.9 (#671)
* *(deps)* Update module github.com/go-ldap/ldap/v3 to v3.4.11 (#670)
* *(deps)* Update module github.com/yuin/goldmark to v1.7.10 (#674)
* *(deps)* Update dev-dependencies (#676)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.28 (#675)
* *(deps)* Update dependency @sentry/vue to v9.13.0 (#450)
* *(deps)* Update ghcr.io/techknowlogick/xgo:go-1.23.x docker digest to 46a3479 (#679)
* *(deps)* Update dev-dependencies (#680)
* *(deps)* Update module github.com/gabriel-vasile/mimetype to v1.4.9 (#681)
* *(deps)* Update dev-dependencies (#683)
* *(deps)* Update dev-dependencies (#684)
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v6.0.8 (#685)
* *(deps)* Update pnpm to v10.9.0 (#686)
* *(deps)* Update dependency marked to v15.0.9 (#688)
* *(deps)* Update dev-dependencies (#689)
* *(deps)* Update dev-dependencies (#690)
* *(deps)* Update node.js to v22.15.0 (#692)
* *(deps)* Update node.js to v22.15.0 (#693)
* *(deps)* Update dependency @sentry/vue to v9.14.0 (#694)
* *(deps)* Update dev-dependencies (#695)
* *(deps)* Update dependency marked to v15.0.10 (#696)
* *(deps)* Update docker/build-push-action digest to 14487ce (#698)
* *(deps)* Update dev-dependencies (#703)
* *(deps)* Update dependency marked to v15.0.11 (#702)
* *(deps)* Update dependency axios to v1.9.0 (#700)
* *(deps)* Update actions/download-artifact digest to d3f86a1 (#699)
* *(deps)* Update dependency @types/node to v22.15.2 (#707)
* *(deps)* Update dependency vue-router to v4.5.1 (#706)
* *(deps)* Update pnpm to v10.10.0 (#709)
* *(deps)* Update module github.com/yuin/goldmark to v1.7.11 (#708)
* *(deps)* Update dev-dependencies (#713)
* *(deps)* Update dependency vite to v6.3.4 [security] (#722)
* *(deps)* Update dev-dependencies (#723)
* *(deps)* Update dependency electron to v36 (#718)
* *(deps)* Update dev-dependencies (#725)
* *(deps)* Update golangci/golangci-lint-action digest to 9fae48a (#727)
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.8.0 (#720)
* *(deps)* Update tiptap to v2.12.0 (#717)
* *(deps)* Update module github.com/huandu/go-clone/generic to v1.7.3 (#715)
* *(deps)* Update mariadb:11 docker digest to 11706a6 (#730)
* *(deps)* Update dependency @sentry/vue to v9.15.0 (#716)
* *(deps)* Update ghcr.io/techknowlogick/xgo:go-1.23.x docker digest to d45f463 (#729)
* *(deps)* Update dependency @sentry/vue to v9.16.0 (#732)
* *(deps)* Update dependency go to v1.24.3 (#731)
* *(deps)* Update module golang.org/x/crypto to v0.38.0 (#735)
* *(deps)* Update dependency @sentry/vue to v9.16.1 (#734)
* *(deps)* Update module golang.org/x/oauth2 to v0.30.0 (#737)
* *(deps)* Update module golang.org/x/image to v0.27.0 (#736)
* *(deps)* Update module dario.cat/mergo to v1.0.2 (#739)
* *(deps)* Pin useblacksmith/build-push-action action to 5646913 (#733)
* *(deps)* Update dev-dependencies (#741)
* *(deps)* Update actions/setup-go digest to d35c59a (#742)
* *(deps)* Update docker/dockerfile:1 docker digest to 9857836 (#745)
* *(deps)* Update dependency @sentry/vue to v9.17.0 (#746)
* *(deps)* Update dependency @types/node to v22.15.17 (#747)
* *(deps)* Update dependency @vitejs/plugin-vue to v5.2.4 (#749)
* *(deps)* Update dev-dependencies (#751)
* *(deps)* Update module github.com/go-testfixtures/testfixtures/v3 to v3.15.0 (#752)
* *(deps)* Update dev-dependencies (#759)
* *(deps)* Update dependency @sentry/vue to v9.18.0 (#760)
* *(deps)* Update postgres:17 docker digest to 8648313 (#712)
* *(deps)* Pin useblacksmith/build-push-action action to 5646913 (#762)
* *(deps)* Update pnpm to v10.11.0 (#764)
* *(deps)* Update vueuse to v13.2.0 (#766)
* *(deps)* Update dependency @sentry/vue to v9.19.0 (#767)
* *(deps)* Update useblacksmith/build-push-action digest to 5501e3f (#768)
* *(deps)* Pin useblacksmith/cache action to c5fe29e (#763)
* *(deps)* Update useblacksmith/build-push-action digest to f0d8aee (#769)
* *(deps)* Update node.js to v22.15.1 (#771)
* *(deps)* Update dev-dependencies (#772)
* *(deps)* Update dependency vue to v3.5.14 (#773)
* *(deps)* Update module github.com/getsentry/sentry-go/echo to v0.33.0 (#779)
* *(deps)* Update node.js to v22.15.1 (#783)
* *(deps)* Update dev-dependencies (#785)
* *(deps)* Update go-testfixtures/testfixtures to latest main
* *(deps)* Pin cypress/browsers docker tag to 05d30b9 (#789)
* *(deps)* Update module github.com/pquerna/otp to v1.5.0 (#790)
* *(deps)* Update node.js to 152270c (#784)
* *(deps)* Update golangci/golangci-lint-action action to v8 (#738)
* *(deps)* Update dependency eslint to v9.27.0 (#793)
* *(deps)* Update module github.com/yuin/goldmark to v1.7.12 (#795)
* *(deps)* Update useblacksmith/build-push-action digest to e09a088 (#792)
* *(deps)* Update dev-dependencies (#796)
* *(deps)* Update dependency @sentry/vue to v9.20.0 (#798)
* *(deps)* Update dependency dompurify to v3.2.6 (#799)
* *(deps)* Update dev-dependencies (#800)
* *(deps)* Update dependency marked to v15.0.12 (#801)
* *(deps)* Update dependency @sentry/vue to v9.21.0 (#802)
* *(deps)* Update dependency @sentry/vue to v9.22.0 (#805)
* *(deps)* Update cypress/browsers:latest docker digest to 753c6dd (#804)
* *(deps)* Update dev-dependencies (#806)
* *(deps)* Update node.js to v22.16.0 (#810)
* *(deps)* Update node.js to v22.16.0 (#811)
* *(deps)* Update postgres:17 docker digest to 2718f68 (#814)
* *(deps)* Update node.js to 9f3ae04 (#812)
* *(deps)* Update postgres:17 docker digest to bbdcc04 (#815)
* *(deps)* Update module github.com/labstack/echo/v4 to v4.13.4 (#820)
* *(deps)* Update postgres:17 docker digest to ea51edb (#822)
* *(deps)* Update postgres:17 docker digest to 6efd0df (#823)
* *(deps)* Update module github.com/olekukonko/tablewriter to v1 (#750)
* *(deps)* Update cypress/browsers:latest docker digest to ceabc12 (#824)
* *(deps)* Update dev-dependencies (#825)
* *(deps)* Update dependency vue-i18n to v11.1.4 (#826)
* *(deps)* Update dependency rollup-plugin-visualizer to v6 (#828)
* *(deps)* Update dependency vue to v3.5.15 (#829)
* *(deps)* Update github.com/go-testfixtures/testfixtures/v3 to v3.16.0
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.9.0 (#831)
* *(deps)* Update dependency vue-i18n to v11.1.5 (#830)
* *(deps)* Update vueuse to v13.3.0 (#832)
* *(deps)* Update module github.com/olekukonko/tablewriter to v1.0.7 (#835)
* *(deps)* Update dependency @sentry/vue to v9.23.0
* *(deps)* Update dependency vue to v3.5.16
* *(deps)* Update dev-dependencies
* *(deps)* Update node.js to 41e4389
* *(deps)* Update ghcr.io/techknowlogick/xgo:go-1.23.x docker digest to 229c595
* *(deps)* Update dependency @sentry/vue to v9.24.0
* *(deps)* Update cypress/browsers:latest docker digest to d74c0a3
* *(deps)* Update pnpm to v10.11.1
* *(deps)* Update mariadb:11 docker digest to fcc7fcd
* *(deps)* Update dependency @sentry/vue to v9.25.0
* *(deps)* Update dependency @sentry/vue to v9.25.1
* *(deps)* Update cypress/browsers:latest docker digest to 201bee8
* *(deps)* Update dependency @sentry/vue to v9.26.0
* *(deps)* Update dependency pinia to v3.0.3
* *(deps)* Update tiptap to v2.13.0
* *(deps)* Update module golang.org/x/sync to v0.15.0 (#874)
* *(deps)* Update module golang.org/x/crypto to v0.39.0 (#890)
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.10.0 (#889)
* *(deps)* Update dependency go to v1.24.4 (#886)
* *(deps)* Update dependency @sentry/vue to v9.27.0 (#888)
* *(deps)* Update ghcr.io/techknowlogick/xgo:go-1.23.x docker digest to a56d0c3 (#885)
* *(deps)* Update crowdin/github-action digest to f214c87 (#876)
* *(deps)* Update postgres:17 docker digest to 30a7233 (#901)
* *(deps)* Update pnpm to v10.12.1 (#902)
* *(deps)* Update tiptap to v2.14.0 (#906)
* *(deps)* Update module golang.org/x/image to v0.28.0 (#905)
* *(deps)* Update devenv
* *(deps)* Update dev-dependencies (#840)
* *(deps)* Update dependency @sentry/vue to v9.28.0 (#910)
* *(deps)* Update mariadb:11 docker digest to 1e66902 (#914)
* *(deps)* Update postgres:17 docker digest to b562fd5 (#915)
* *(deps)* Update dependency @sentry/vue to v9.28.1 (#922)
* *(deps)* Update postgres:17 docker digest to cb51e9f (#921)
* *(deps)* Do not update services as often
* *(deps)* Update useblacksmith/build-push-action digest to 574eb0e (#924)
* *(deps)* Update postgres:17 docker digest to 6cf6142 (#923)
* *(deps)* Update dev-dependencies (#925)
* *(deps)* Update dependency happy-dom to v18 (#926)
* *(deps)* Update dependency @sentry/vue to v9.29.0 (#936)
* *(deps)* Update module github.com/go-sql-driver/mysql to v1.9.3 (#935)
* *(deps)* Update dev-dependencies (#941)
* *(deps)* Update dependency axios to v1.10.0 (#942)
* *(deps)* Update dependency vite-plugin-vue-devtools to v7.7.7 (#946)
* *(deps)* Update docker/setup-buildx-action digest to 18ce135 (#949)
* *(deps)* Update brace-expansion to 2.0.2 and 1.1.12
* *(deps)* Update brace-expansion to 2.0.2 and 1.1.12
* *(deps)* Update dev-dependencies (#967)
* *(deps)* Update dependency vue-i18n to v11.1.6 (#968)
* *(deps)* Update dependency @sentry/vue to v9.30.0 (#972)
* *(deps)* Update cypress/browsers:latest docker digest to b290f97 (#976)
* *(deps)* Update dev-dependencies (#977)
* *(deps)* Update docker/setup-buildx-action digest to e468171 (#978)
* *(deps)* Update dependency vue to v3.5.17 (#979)
* *(deps)* Update dependency electron to v36.5.0 (#980)
* *(deps)* Update vueuse to v13.4.0 (#982)
* *(deps)* Update dependency rollup to v4.44.0 (#986)
* *(deps)* Update tiptap to v2.14.1 (#985)
* *(deps)* Update tiptap to v2.22.0 (#987)
* *(deps)* Update dependency caniuse-lite to v1.0.30001724 (#989)
* *(deps)* Update tiptap to v2.22.3 (#990)
* *(deps)* Bump github.com/go-chi/chi/v5 from 5.1.0 to 5.2.2 (#988)
* *(deps)* Update dependency vue-i18n to v11.1.7 (#991)
* *(deps)* Update module github.com/getsentry/sentry-go to v0.34.0 (#996)
* *(deps)* Update dependency @sentry/vue to v9.31.0 (#995)
* *(deps)* Update pnpm to v10.12.2 (#994)
* *(deps)* Update module github.com/getsentry/sentry-go/echo to v0.34.0 (#997)
* *(deps)* Update dev-dependencies to v8.35.0 (#998)
* *(deps)* Update pnpm to v10.12.3 (#1006)
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.11.0 (#1008)
* *(deps)* Update dependency @types/node to v22.15.33 (#1017)
* *(deps)* Update node.js to v22.17.0 (#1018)
* *(deps)* Update crowdin/github-action digest to 297234b (#1020)
* *(deps)* Update dependency @sentry/vue to v9.32.0 (#1022)
* *(deps)* Update node.js to v22.17.0 (#1027)
* *(deps)* Update node.js to 9f3f2c6 (#1028)
* *(deps)* Update dev-dependencies (#1030)
* *(deps)* Update node.js to 5340cbf (#1029)
* *(deps)* Update pnpm to v10.12.4 (#1033)
* *(deps)* Update dependency postcss-preset-env to v10.2.4 (#1035)
* *(deps)* Update tiptap to v2.23.0 (#1036)
* *(deps)* Update dependency marked to v16 (#1037)
* *(deps)* Update dependency @sentry/vue to v9.33.0 (#1039)
* *(deps)* Bump github.com/go-viper/mapstructure/v2 from 2.2.1 to 2.3.0 (#1043)
* *(deps)* Update dev-dependencies (major) (#1045)
* *(deps)* Update dev-dependencies (#1044)
* *(deps)* Update postgres:17 docker digest to 45518b2 (#1052)
* *(deps)* Update dev-dependencies (#1050)
* *(deps)* Update tiptap to v2.23.1 (#1051)
* *(deps)* Update pnpm to v10.12.4 (#1042)
* *(deps)* Update dependency @sentry/vue to v9.34.0 (#1062)
* *(deps)* Update postgres:17 docker digest to 3962158 (#1061)
* *(deps)* Update cypress/browsers:latest docker digest to 95587c1 (#1026)
* *(deps)* Update dev-dependencies (#1064)
* *(deps)* Update mariadb:11 docker digest to 1e4ec03 (#1066)
* *(deps)* Update dependency vue-tsc to v3 (#1065)
* *(deps)* Update vueuse to v13.5.0 (#1067)
* *(deps)* Update dependency vue-i18n to v11.1.8 (#1070)
* *(deps)* Update tiptap to v2.24.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency vue-i18n to v11.1.9
* *(deps)* Update module github.com/threedotslabs/watermill to v1.4.7
* *(deps)* Update tiptap to v2.24.1 (#1087)
* *(deps)* Update dependency vite to v7.0.2 (#1088)
* *(deps)* Update node.js to 10962e8 (#1096)
* *(deps)* Update module github.com/getsentry/sentry-go/echo to v0.34.1 (#1095)
* *(deps)* Update tiptap to v2.25.0 (#1091)
* *(deps)* Update module github.com/olekukonko/tablewriter to v1.0.8 (#1090)
* *(deps)* Update dependency @sentry/vue to v9.35.0 (#1092)
* *(deps)* Update dev-dependencies (#1093)
* *(deps)* Update dependency go to v1.24.5 (#1099)
* *(deps)* Update ghcr.io/techknowlogick/xgo:go-1.23.x docker digest to 55a8e62 (#1098)
* *(deps)* Update pnpm to v10.13.1 (#1100)
* *(deps)* Update dependency @sentry/vue to v9.36.0 (#1101)
* *(deps)* Update crowdin/github-action digest to 9fd07c1 (#1108)
* *(deps)* Update dependency @sentry/vue to v9.38.0 (#1109)
* *(deps)* Update module github.com/jaswdr/faker/v2 to v2.6.0 (#1110)
* *(deps)* Update module golang.org/x/text to v0.27.0 (#1107)
* *(deps)* Update module golang.org/x/image to v0.29.0 (#1105)
* *(deps)* Update module golang.org/x/sync to v0.16.0 (#1103)
* *(deps)* Update useblacksmith/cache digest to 71c7c91 (#1102)
* *(deps)* Update module golang.org/x/term to v0.33.0 (#1106)
* *(deps)* Update module golang.org/x/crypto to v0.40.0 (#1111)
* *(deps)* Update tiptap to v2.26.1 (#1112)
* *(deps)* Update dev-dependencies (#1115)
* *(deps)* Update module github.com/go-testfixtures/testfixtures/v3 to v3.17.0 (#1120)
* *(deps)* Update module github.com/golang-jwt/jwt/v5 to v5.2.3 (#1125)
* *(deps)* Update dependency @sentry/vue to v9.39.0 (#1126)
* *(deps)* Update cypress/browsers:latest docker digest to d5ae1a6 (#1127)
* *(deps)* Update node.js to 9db789c (#1129)
* *(deps)* Update node.js to fc3e945 (#1130)
* *(deps)* Update node.js to v22.17.1 (#1131)
* *(deps)* Update mariadb:11 docker digest to acf55e2 (#1128)
* *(deps)* Update dependency vue-i18n to v11.1.10 (#1132)
* *(deps)* Update dev-dependencies (#1136)
* *(deps)* Update module github.com/swaggo/swag to v1.16.5 (#1137)
* *(deps)* Update node.js to v22.17.1 (#1135)
* *(deps)* Update mariadb:11 docker digest to 2bcbaec (#1134)
* *(deps)* Update cypress/browsers:latest docker digest to 30192f4 (#1133)
* *(deps)* Update dependency @sentry/vue to v9.40.0 (#1141)
* *(deps)* Update dependency marked to v16.1.0 (#1143)
* *(deps)* Update dependency vue-tsc to v3.0.2 (#1144)
* *(deps)* Update dependency marked to v16.1.1 (#1145)
* *(deps)* Pin paradedb/paradedb docker tag to 9627c2a (#1147)
* *(deps)* Update dev-dependencies (#1148)
* *(deps)* Update dependency esbuild to v0.25.8 (#1151)
* *(deps)* Update module xorm.io/xorm to v1.3.10 (#1156)
* *(deps)* Update form-data to 4.0.4
* *(deps)* Update form-data to 4.0.4 in desktop
* *(deps)* Update @eslint/plugin-kit to 0.3.4
* *(deps)* Update postgres:17 docker digest to 378ef4a (#1157)
* *(deps)* Update postgres:17 docker digest to 4d89c90 (#1159)
* *(deps)* Update font awesome to v7 (major) (#1160)
* *(deps)* Update dependency vue to v3.5.18 (#1163)
* *(deps)* Update dev-dependencies to v8.38.0 (#1164)
* *(deps)* Update module github.com/yuin/goldmark to v1.7.13 (#1161)
* *(deps)* Update dependency axios to v1.11.0 (#1167)
* *(deps)* Update dependency vue-i18n to v11.1.11 (#1166)
* *(deps)* Pin actions/github-script action to 60a0d83 (#1168)
* *(deps)* Update dev-dependencies (#1169)
* *(deps)* Update golangci lint to 2.2.2
* *(deps)* Update paradedb/paradedb:latest-pg17 docker digest to c19d4ec (#1171)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.29 (#1172)
* *(deps)* Update dependency @fortawesome/vue-fontawesome to v3.1.0 (#1162)
* *(deps)* Update dependency @sentry/vue to v9.41.0 (#1173)
* *(deps)* Update dev-dependencies (#1174)
* *(deps)* Update dependency @sentry/vue to v9.42.0 (#1176)
* *(deps)* Update paradedb/paradedb:latest-pg17 docker digest to e641c93 (#1179)
* *(deps)* Update dev-dependencies (#1180)
* *(deps)* Update dependency rollup to v4.45.3 (#1182)
* *(deps)* Update dependency vite-plugin-vue-devtools to v8 (#1185)
* *(deps)* Update dev-dependencies (#1184)
* *(deps)* Update vueuse to v13.6.0 (#1187)
* *(deps)* Update dependency @sentry/vue to v9.42.1 (#1188)
* *(deps)* Update module github.com/olekukonko/tablewriter to v1.0.9 (#1189)
* *(deps)* Update module github.com/jaswdr/faker/v2 to v2.6.1 (#1191)
* *(deps)* Pin ghcr.io/go-vikunja/dex-testing docker tag to 7440cd3 (#1190)
* *(deps)* Update module github.com/swaggo/swag to v1.16.6 (#1194)
* *(deps)* Update sentry-javascript monorepo (#1195)
* *(deps)* Update linkifyjs to 4.3.2
* *(deps)* Update dev-dependencies (#1196)
* *(deps)* Update module github.com/coreos/go-oidc/v3 to v3.15.0 (#1197)
* *(deps)* Update ghcr.io/go-vikunja/dex-testing:main docker digest to 7a582d4 (#1199)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.30 (#1202)
* *(deps)* Update module github.com/golang-jwt/jwt/v5 to v5.3.0 (#1203)
* *(deps)* Update dev-dependencies to v1.0.0-alpha.3 (#1204)
* *(deps)* Update ghcr.io/go-vikunja/dex-testing:main docker digest to cdcb86c (#1205)
* *(deps)* Update module github.com/prometheus/client_golang to v1.23.0 (#1206)
* *(deps)* Update pnpm to v10.14.0 (#1207)
* *(deps)* Update dependency @sentry/vue to v9.44.0 (#1208)
* *(deps)* Update node.js to v22.18.0 (#1210)
* *(deps)* Update dev-dependencies (#1212)
* *(deps)* Update module github.com/getsentry/sentry-go to v0.35.0 (#1211)
* *(deps)* Update ghcr.io/go-vikunja/dex-testing:main docker digest to bc9b660 (#1214)
* *(deps)* Update module github.com/getsentry/sentry-go/echo to v0.35.0 (#1213)
* *(deps)* Update dependency @sentry/vue to v10 (#1209)
* *(deps)* Update crowdin/github-action digest to 590c05e (#1215)
* *(deps)* Update docker/metadata-action digest to c1e5197 (#1216)
* *(deps)* Update ghcr.io/go-vikunja/dex-testing:main docker digest to 9cdd39c (#1219)
* *(deps)* Update cypress/browsers:latest docker digest to a0f4875 (#1218)
* *(deps)* Update dependency vue-tsc to v3.0.5 (#1220)
* *(deps)* Update paradedb/paradedb:latest-pg17 docker digest to 99008f7 (#1228)
* *(deps)* Update dependency marked to v16.1.2 (#1229)
* *(deps)* Update node.js to v22.18.0 (#1231)
* *(deps)* Update dependency @sentry/vue to v10.1.0 (#1232)
* *(deps)* Update docker/login-action digest to 184bdaa (#1234)
* *(deps)* Update node.js to 1b2479d (#1236)
* *(deps)* Update dev-dependencies (#1237)
* *(deps)* Update ghcr.io/go-vikunja/dex-testing:main docker digest to 76ba260 (#1238)
* *(deps)* Update docker/dockerfile:1 docker digest to 3838752 (#1239)
* *(deps)* Update tiptap to v3 (#1241)
* *(deps)* Pin dependency @floating-ui/dom to 1.7.3 (#1243)
* *(deps)* Pin dependencies (#1242)
* *(deps)* Update paradedb/paradedb:latest-pg17 docker digest to df0a755 (#1244)
* *(deps)* Update module github.com/jaswdr/faker/v2 to v2.7.0 (#1245)
* *(deps)* Update dependency sass-embedded to v1.90.0 (#1246)
* *(deps)* Update actions/download-artifact action to v5 (#1247)
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.12.0 (#1240)
* *(deps)* Update dependency @sentry/vue to v10.2.0 (#1249)
* *(deps)* Update ghcr.io/techknowlogick/xgo:go-1.23.x docker digest to 37bfe9d (#1250)
* *(deps)* Update dependency go to v1.24.6 (#1251)
* *(deps)* Update dev-dependencies (#1252)
* *(deps)* Update module github.com/jaswdr/faker/v2 to v2.8.0 (#1253)
* *(deps)* Update actions/cache digest to 0400d5f (#1255)
* *(deps)* Update module golang.org/x/sys to v0.35.0 (#1256)
* *(deps)* Update module golang.org/x/term to v0.34.0 (#1257)
* *(deps)* Update module golang.org/x/text to v0.28.0 (#1258)
* *(deps)* Update cypress/browsers:latest docker digest to 2c4e104 (#1259)
* *(deps)* Update module golang.org/x/crypto to v0.41.0 (#1260)
* *(deps)* Update module golang.org/x/image to v0.30.0 (#1262)
* *(deps)* Update dev-dependencies (#1261)
* *(deps)* Update dependency @sentry/vue to v10.3.0 (#1263)
* *(deps)* Update dev-dependencies (#1264)
* *(deps)* Update dependency @cypress/vite-dev-server to v7 (#1265)
* *(deps)* Update dependency browserslist to v4.25.2 (#1266)
* *(deps)* Update tiptap to v3.1.0 (#1267)
* *(deps)* Update actions/checkout digest to 08eba0b (#1268)
* *(deps)* Update dependency @sentry/vue to v10.4.0 (#1269)
* *(deps)* Update actions/checkout action to v5 (#1271)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.31 (#1270)
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.12.1 (#1272)
* *(deps)* Update ghcr.io/go-vikunja/dex-testing:main docker digest to 7de4780 (#1273)
* *(deps)* Update tmp to 0.2.5
* *(deps)* Pin softprops/action-gh-release action to 72f2c25 (#1274)
* *(deps)* Update mariadb:11 docker digest to 5d71ac3 (#1275)
* *(deps)* Update mariadb:11 docker digest to 272084c (#1278)
* *(deps)* Update postgres:17 docker digest to 0d5b8e3 (#1279)
* *(deps)* Update dependency @sentry/vue to v10.5.0 (#1280)
* *(deps)* Update postgres:17 docker digest to aef2e62 (#1283)
* *(deps)* Update module github.com/getsentry/sentry-go to v0.35.1 (#1286)
* *(deps)* Update postgres:17 docker digest to 0f9d52b (#1288)
* *(deps)* Update module github.com/getsentry/sentry-go/echo to v0.35.1 (#1287)
* *(deps)* Update dev-dependencies (#1290)
* *(deps)* Update golangci-lint to 2.4.0 (#1291)
* *(deps)* Update tiptap to v3.2.0 (#1292)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.32 (#1294)
* *(deps)* Update postgres:17 docker digest to 19ad070 (#1293)
* *(deps)* Update dependency go to v1.25.0 (#1276)
* *(deps)* Update mariadb docker tag to v12 (#1281)
* *(deps)* Update dev-dependencies (#1295)
* *(deps)* Update postgres:17 docker digest to 29e0bb0 (#1301)
* *(deps-dev)* Bump vite from 6.0.11 to 6.0.12
* *(dev)* Add required deps to run font scripts in devenv
* *(deps)*: update goreleaser/nfpm docker tag to v2.40.0 (#2647)
* *deps)*: update dependency flexsearch to v0.7.43 (#2095)
* *(deps)*: update golangci/golangci-lint docker tag to v1.61.0 (#2678)
* *(deps)*: update dev-dependencies (#2754)
* *(deps)*: update node.js to v22 (#2783)
* *(deps)*: update goreleaser/nfpm docker tag to v2.41.0 (#2776)
* *(deps)*: update module github.com/go-testfixtures/testfixtures/v3 to v3.14.0 (#2615)
* *(deps)*: bump golang.org/x/net from 0.33.0 to 0.36.0 ([7a9fed4](7a9fed4581417784081aaa6eb079c80942345c17))

### Documentation

* *(api)* Use correct return type for the /user endpoint
* *(desktop)* Adjust dev instructions
* *(desktop)* Fix release version instruction
* *(filters)* Fix typos with filter query docs
* *(frontend)* Fix dialog typo
* *(frontend)* Fix env file example name
* *(plugins)* Add new config variables to docs
* *(web)* Fix typos
* Add AGENTS.md file with instructions for AI coding agents ([3813a3b](3813a3b35ce3254cdd60fc659b488fa35690ed7b))
* Add link for crus coding agent instructions ([813bdb5](813bdb58ff4224114bff184f8ee7af6f71c50f54))
* Adjust frontend readme ([77d1616](77d1616feacc89d7af768f6dba9bad3bbc835794))
* Clarify Todoist redirect url ([bf2d56c](bf2d56c9d42d6beb023ac6ea4cc10c0db7e58a6c))
* Clarify enabled providers ([003db05](003db05b66c086b42001e27269de3ac4e9593144))
* Clarify licensing (#894) ([bcfe6d2](bcfe6d241d3ad5c38cf7361839d7013a58f421e0))
* Clarify migrators ([b9cdc9f](b9cdc9fbe516493ea976b935eea9877f7c59adfe))
* Clarify return value of /tasks ([debdcd4](debdcd4dd3fa6a9da847562a11e73e7510f2671d))
* Clarify typesense url protocol ([d046599](d0465991928c49c219e7c7d9e97582e5e7e3d61b))
* Correctly document providers in config ([c5a97ef](c5a97ef0a353e7f9a2da7521e36e31108274f99b))
* Delete caldav token uses DELETE instead of GET ([9efdde8](9efdde8f1a325666ae70cdab87d2cb1fccbf5227))
* Fix comments in web package ([ef80fa7](ef80fa77b4f775f6c0d5bcec4e5cbb07105b2ff4))
* Fix typo (#1122) ([5ee3077](5ee3077f5d00a67cea4cd11c220b56bb0a063679))
* Fix typo in contributing section ([07e96e9](07e96e94b2fddad0a7d3326b1a7d468c9d65d368))
* Fix typos ([cab303f](cab303fe64399cb06f160f0c49872f06682d1816))
* Fix typos in models services guide ([75e43ab](75e43ab3f2aedd7ebe936d4cdf16b290dc755993))
* Format config json ([dea123d](dea123dbeafe4cd2a425b1c667067983815c7a84))
* Intro for migrators ([5643027](56430274543f7f8e4f41821b13f60885ebae7f1f))
* Mention AGPL-3.0-or-later in CLI help (#893) ([4146e91](4146e91616a33bafa03a74d9967be88551f35398))
* Update code agent instructions ([8864d04](8864d043906f316dd1c87d8ca0c00d4996972d11))

### Features

* *(auth)* Add ForceUserInfo option to OpenID provider (#797)
* *(auth)* Add ldap config variables to docs
* *(auth)* Allow automatic sso login from param (#3071)
* *(auth)* Allow passing custom settings links to user account via openid claims
* *(auth)* Authenticate users via ldap
* *(auth)* Ldap group sync
* *(auth)* Make ldap user filter configurable
* *(auth)* Make sure local auth and ldap can both work when configured at the same time
* *(auth)* Refactor group sync
* *(auth)* Rename oidc_id to external_id
* *(auth)* Require auth to fetch avatars (#930)
* *(auth)* Show login form when only ldap is enabled
* *(auth)* Sso fallback mapping (#3068)
* *(auth)* Sync avatar from OpenID providers (#821)
* *(auth)* Test ldap in ci
* *(auth)* Use config variable to check if we should verify tls
* *(auth)* Verify ldap config before trying to connect
* *(build)* Add RELEASE_VERSION argument to be able to pass release version via env
* *(caldav)* Return proper caldav intervals instead of FREQ=SECONDLY (#1230)
* *(ci)* Add build docker image gh action
* *(ci)* Add desktop release to GitHub actions
* *(ci)* Add tests for paradedb
* *(ci)* Build desktop app natively
* *(ci)* Build frontend before test
* *(ci)* Comment on closed issues when closed by commit or PR
* *(ci)* Create draft release when tagging
* *(ci)* Disable postgres durability features when testing
* *(ci)* Migrate some workflows to Blacksmith (#617)
* *(ci)* Pass cypress ci build id (#848)
* *(ci)* Publish desktop releases with GitHub actions only
* *(ci)* Run cypress tests in parallel
* *(ci)* Store desktop release files as artifact
* *(ci)* Use blacksmith docker action
* *(ci)* Use blacksmith docker in favor of github docker build cache
* *(ci)* Use docker image when testing with cypress parallel
* *(ci)* Use slightly smaller runner to build binaries
* *(ci)* Use xgo cache between binary builds
* *(cli)* Add cli command to delete orphan task positions
* *(cmd)* Allow to specify custom filename for dump command (#2775)
* *(config)* Only read file sub-keys from files
* *(dev)* Add devcontainers configuration
* *(dev)* Add frontend and api to launch config
* *(dev)* Add mailpit to devenv services
* *(dev)* Adjust devcontainer definition
* *(dev)* Allow passing struct name to dev:make-migration mage command (#931)
* *(dev)* Use proxy server in dev mode (#3069)
* *(docker)* Add opencontainer labels
* *(dump)* Add flag to allow specifying dump path
* *(editor)* Add message to list of allowed link protocols (#1284)
* *(editor)* Support custom protocol for links
* *(event)* Simplify dispatching task updated event from only a task id
* *(filter)* Allow dragging tasks in saved filter
* *(filter)* Automatically focus filter input when opening filter popup
* *(filter)* Rebuild filter input component
* *(filter)* Use new filter input component in view edit
* *(filters)* Add "not in" operator for filters
* *(filters)* Allow filtering by created and updated task fields
* *(filters)* Show when the current view has a filter as well and both will be used
* *(frontend)* Balance heading text (#952)
* *(frontend)* Improve generics in multiselect dropdown (#920)
* *(frontend)* Improve route filter generics (#953)
* *(gantt)* Rebuild the gantt chart (#1001)
* *(home)* Correctly check if tasks exist before showing import hint
* *(i18n)* Add Bulgarian for language selection
* *(i18n)* Add Finnish translation for selection
* *(i18n)* Add Hebrew translation for selection
* *(i18n)* Add Lithuanian translation for selection
* *(i18n)* Add Turkish as language for selection
* *(i18n)* Add params replacement to translation functions
* *(i18n)* Add pluralization function for translation strings
* *(i18n)* Add translations for migration notifications
* *(i18n)* Automatically set language during registration
* *(i18n)* Make overdue reminder translations more explicit (#761)
* *(i18n)* Use plural translations in humanize duration
* *(kanban)* Add debug option to show task position on card
* *(kanban)* Create To-Do, Doing, Done buckets when creating a new kanban view
* *(kanban)* Show project on kanban card if it's not the same as the current one
* *(labels)* Show priority labels based on minimum priority setting (#3075)
* *(labels)* Sort labels alphabetically
* *(ldap)* Add tests
* *(ldap)* Also look for username only when checking group membership
* *(ldap)* Do not allow changing user avatar when synced from ldap
* *(ldap)* Make group sync configurable
* *(ldap)* Make member id attribute configurable
* *(ldap)* Sync avatar from ldap
* *(link share)* Add feature test for link share avatar access (#944)
* *(list)* Add j/k keyboard navigation (#1040)
* *(navigation)* Use focus-visible for nav items
* *(notifications)* Include link to settings in notifications
* *(plugins)* Add rudimentary plugin system
* *(plugins)* Allow plugins to register routes
* *(positions)* Add more debug logs
* *(projects)* Add support for ParadeDB when searching for project
* *(projects)* Optionally return max right when querying all projects
* *(rtl)* Basic rtl layout for rtl languages
* *(rtl)* Mirror task description icon
* *(settings)* Restructure general settings view
* *(settings)* Show extra settings links on user settings page
* *(settings)* Show save button only when something was changed
* *(sharing)* Add config so that users only find members of their teams
* *(table view)* Require ctrl-click to sort by multiple task properties (#950)
* *(task)* Always insert new tasks at the top
* *(task)* Autosave description on route leave (#1140)
* *(task)* Cancel editing task title with escape (#2730)
* *(task)* Expand reactions via parameter
* *(task)* Use focus-visible for task focus styles
* *(tasks)* Add parameter to expand comments on a task
* *(tasks)* Add support for ParadeDB when searching tasks
* *(tasks)* Fetch comments with the task
* *(test)* Add benchmark for task search (#963)
* *(test)* Add e2e tests for openid (#1178)
* *(ts)* Improve module declarations (#957)
* *(user)* Add avatar cache flushing (#1041)
* *(user)* Use name for initals avatar, not username
* *(webhooks)* Expand buckets in webhooks
* *config*: store sqlite file relative to rootpath (#934)
* Add /token/test route ([ca98b7d](ca98b7da73e641db023af03b943acdad8d7b822b))
* Add Korean translation for selection ([6ee6b2f](6ee6b2ffeec4b38d5eff827990e80c39e5c8fd9a))
* Add display setting for dates (#1192) ([b444cf8](b444cf8d4310d833bd00afee3814174496251b9c))
* Add email filter to user list command (#973) ([34a5196](34a5196a05ee2b16019a032fe7535a9f6a8f8c1e))
* Add expand property to read one task ([333e35e](333e35e648fb49ed74002a4f6ce50293c20fdebd))
* Add generic types to multiselect (#2618) ([b7fc293](b7fc29327a5bfe8c69b99e10e8f91b0afed67235))
* Add install-e2e-binaries input flag for frontend setup composite action (#878) ([7d954cd](7d954cd5bf451c8cfcba6da99252fc30774e3a52))
* Add keyboard shortcut to open project from task detail (#940) ([9bab44d](9bab44d5ded326abc25009621cbe1169fcd456a3))
* Add keyvalue.Remember function ([c7a9838](c7a98386c2b18b77264c0f76cdca3dde717e50f7))
* Add logical utils ([d290f2e](d290f2e99c74962263e19b71faa844c7339366e1))
* Add logo change toggle setting (#1031) ([bddba86](bddba8646db5077f9771e89d86bc5f8960564b94))
* Add missing peer dependency ([ad28703](ad2870335d4fe183e08840622c480fb0251e6211))
* Add prefix key support to keyvalue store (#1038) ([ae92822](ae92822ee0a5159a1049fb6a24b20d2860765e2e))
* Add tailwind with prefix (#2513) ([bbfd527](bbfd5270db78f800a6940e73e36b525b4724c582))
* Add team member with enter (#711) ([a21430d](a21430d88a1388d624af2ef67ebe19e07e3473fa))
* Add tonight as a quick add date option (#2866) ([8dfb34c](8dfb34c863dce9f908038f0ac3667eb7b8a284ea))
* Add utm tag to powered by link ([204dccf](204dccf08b435fa677f411fda76db59bc33eb368))
* Add vite-plugin-vue-devtools ([e8a07fc](e8a07fc8e0a4c93790ada72cd7e1bcfd785bba02))
* Align caching and node version (#608) ([a245405](a2454057aef1816c8bac9cad807a000c418a8dda))
* Allow setting schema for connection in postgres (#2777) ([28d5cd7](28d5cd7b287c75508f01af35a1fa361dfa67f6c4))
* Allow to write date via Quick Add Magic in DMY format with dots (#744) ([fbec58a](fbec58a50bf91cf1d80731173be470dc840c9c2b))
* Always add project to webhook payload ([70e1fda](70e1fdae91ef9b28c15a4d932d237c8e992179ec))
* Arm 'vue/no-setup-props-reactivity-loss' rule ([522f1cb](522f1cb5962482773cacfaa3a2b23c01d326d944))
* Auto tls ([daa7ad0](daa7ad053c35a97933ca79aee007c388538bab5d))
* Cache docker (#758) ([3e540cf](3e540cff5f943ecd6c284af2ed11e4e8d75b7d69))
* Check configured default timezone on startup (#903) ([f070268](f070268c30738639e6ddbfdbc03ffe1434c05c20))
* Ci env do not track (#859) ([91f8df3](91f8df34fc923578bfafba556cec39ac52500ff8))
* Composite action for frontend setup ([1deb674](1deb674da156b359b525ada895d1381b66dbeb15))
* Config for auth providers now use a map instead of an array ([05349dd](05349ddb5c9d62a59ab1b997984d0b97f24df247))
* Consistent sorting ([ed5d983](ed5d983d18ba2e88e62dfea60b46d67132c32248))
* Convert pasted markdown to html so that it is correctly rendered (#3041) ([f52a321](f52a321acf19b8925a5285abf09ae3ed51ea4ca8))
* Do not load notifications while in the background ([c35c70e](c35c70e71ff378b6b877d0422ca66277f6e3f59a))
* Docker layer cache (#808) ([865a764](865a7640da809378abc6a02190d6d43e29d1e183))
* Don't log all headers when debug log is enabled ([ea42fef](ea42fef2dad3cb88f54b2fdc8be7d984a5e2d254))
* Downscaled image previews for task attachments (#2541) ([75ce261](75ce261f740051f6feb30dc200066f945e88ce95))
* Enable cors by default for desktop app (#904) ([433b8b9](433b8b9115e81aac364ea7261ccfdfaa5652c948))
* Expand buckets ([7f6cb1e](7f6cb1e06e846aa41fd810ea2d370a3b3e1e374e))
* Explicit pnpm ci args (#755) ([ac244d3](ac244d3915197168ea8c7ca7e9688971e2a7cf91))
* Generate yml config from json ([3c70bd6](3c70bd630dfebeda77183982173ac548ea6b77fc))
* HasAttachments as store computed ([1f55e3f](1f55e3f866e89577c599f7eec26e66933909f4d2))
* Improve ProjectSettingsDelete ([e57f04e](e57f04ec23e9ff8aa9877d2ea7d571c2a44790b0))
* Improve ProjectSettingsViews ([e8be657](e8be657d9775a592987b684250809eb3d170a6f1))
* Improve clean-translations script (#964) ([9fcede5](9fcede5729d3b1e4c99d48a41a1eb87b69f347d0))
* Improve docker layers (#803) ([75db483](75db48348a2d4c8354c9b9003c619b2312f31487))
* Improve label store ([2a6ba7e](2a6ba7e7f0162bb050575908e7ca80f4b0292398))
* Improve priority visibility ([d35454c](d35454c099f1f4aceb513634b7c531272fa8d550))
* Improve project edit form ([9c115b7](9c115b7f5c6d779124090e300fef27c8ad453714))
* Improve projects store ([1e523a1](1e523a1a39810ffd578f15fbefe034c70ed5a066))
* Inline dynamic component definitions in routes (#2812) ([a899933](a8999336f7c20c21075a8c309203ffde101cf09f))
* Install stylelint ([504e201](504e201da215156c2ff9885e98c43431b9670568))
* Introduce shared health check logic (#1073) ([4d36771](4d367713628023903d2f900758d1c3c28ab8e6e8))
* Load any config value from file ([c7914bc](c7914bc2452b8d7e3cc343d4d0413926fb86d3de))
* Load project in project view ([4c972e1](4c972e1bc480d484a22cc2eea09b20e965a92ce7))
* Log request headers when debug logs are enabled ([9fc6cdd](9fc6cdd07632a95d7136e4f82dbac10809dd8e98))
* Make time reactive (#2627) ([cb8fd09](cb8fd098248613195c881b305908bcb8dabbe738))
* Make used bcrypt rounds configurable ([a88124c](a88124cfce29e187d5b44049c82a95b95aedcfaa))
* Move loading logic from ready to base store ([a6644d9](a6644d9c89664d97882c231a34363fed1869e8f5))
* Move to slog for logging ([ca83ad1](ca83ad1f984fe5177e76f1822caed4789a7ed66b))
* Move useProjectBackground to composables ([44c659a](44c659aa34e80a16af38e232894e32f42ef2270f))
* Only build sourcemaps for sentry ([cf6836f](cf6836f8570f2d13c12d764d1940a64d661f582a))
* Permalinks for task comments (#2442) ([5f9d0fe](5f9d0fe763dcbee3273e0f90b84aa35dca091f9d))
* Preferably award admin access to project users with write access on user deletion (#2772) ([1f76a8b](1f76a8bb641024bd0c5722976ec7fd0f6dafce49))
* Reactive flatpicker language (#2628) ([79071a1](79071a19096650348ce580c787fb6db12d31ef3b))
* Remove 'frontend-dependencies' step ([1c2cdf9](1c2cdf9240556b1acb7dfc8f54b32ec03de19936))
* Remove @vitejs/plugin-legacy (#2921) ([cff602c](cff602c2462dc088ebf388da5034e3045f3176e0))
* Remove cypress install ([2aeff57](2aeff579fb1b9b8e8da419af0cc67cdfb86c790d))
* Remove cypress install v2 (#664) ([59e2d7e](59e2d7e650440359179ecbe7dd9df58bdcfea4b3))
* Remove dedicated test build and preview script (#661) ([43c80a6](43c80a6b9971a4a90c85af55cb0b289c98304611))
* Remove echo log options - unify with general http logging ([62200f6](62200f6e0ffbc598750a1acbafa1bc529782b811))
* Remove flexsearch stub types (#756) ([ed3b7e1](ed3b7e14986d7ed6417f3413939df0efe7c67d20))
* Remove postcss-easings ([194a323](194a3239afccb002906c1597509f98a1652415ff))
* Rename right to permission (#1277) ([a81a3ee](a81a3ee0e5a371773ba2fc12c058f4eb46b3dbcb))
* Renovate – use best-practices ([3e0e981](3e0e981da1e0d5f8d73b4f16b8b020540dc6c13b))
* Replace absolute left position with inset-inline-start ([a25a4a0](a25a4a00c9e36e2f630e39b14eaf86ff4272ec4d))
* Replace border-bottom with logical properties ([5cd256c](5cd256c4856cf72d539a4f26453a5f0e58ed3383))
* Replace border-bottom with logical properties ([cdd4e46](cdd4e46daa028ee616086d2508a3d3e09e9fbdbb))
* Replace border-left with logical properties ([55180b6](55180b60a11bb704f20ccd493aabfcd1e2d75732))
* Replace border-top with logical properties ([21943b6](21943b61ebb5cdaff0d60f8040a6e591d7fe318b))
* Replace border-top with logical properties ([dd199c4](dd199c4ddeecc70ec46b26b427daa604d90420ad))
* Replace bottom with logical properties ([a992488](a9924881c2c7f2c644bbfc89b9d2519101e87fae))
* Replace right with logical properties ([0f5e001](0f5e0019aef45d5a8e4f61dde6708cb2187f1bdf))
* Replace top with logical properties ([0159ddc](0159ddc31366482866cb402086e948996fc34826))
* Show frontend version separately if different (#974) ([16e1449](16e14490a60daa5c964d70e9823e330a93ecb472))
* Show user export status in settings (#1200) ([4042f66](4042f66efa05dc41cc55bb7ad03917e39a06685a))
* Simplify ProjectView ([144571e](144571e4485b6d5c071c2ed2ab79017039844f93))
* Switch from nix flakes to devenv ([ad3c5fc](ad3c5fcee5648d942512a31faf3cab4f15272d85))
* Translate notifications ([e11a302](e11a3026b94fd502f61b1e984d096f0a838ec1b0))
* Unify component name ([1f56b36](1f56b3615cef23e755454e8ce9647c7db4c40e13))
* Update no auth image ([a5e71ea](a5e71ea6ce2d5235789051ead2e4785f3661e23e))
* Upgrade to pnpm 10 (#616) ([b7c9770](b7c977003a8788f37bff5d524b1872d485d99489))
* Use GitHub actions for build and release ([34d6023](34d60232481aab2aa749603406178fd90d1c73b5))
* Use TipTap starter-kit ([140765a](140765ad20f658f8a26bc79c1d2d375e30c5d085))
* Use hetzner object storage for releases ([0472aca](0472acac980d31c37863d34a8eed988ee1dc6523))
* Use implicit naming for project title ([d6772a3](d6772a3d59d4e4f5e063ef663640ef9023895298))
* Use keyvalue.Remember where it makes sense ([fcdcdcf](fcdcdcf46a2d59ad635c4dda9797889822321c56))
* Use pnpm dedupe (#757) ([80f23c6](80f23c6a4c4c5fdddae2d5194174153eadc54bbb))
* Use position sticky for demo bar ([49fa32a](49fa32aad6f1b677cb5703dab78791274a90a31a))
* Use radio button for configMode change ([b0b8262](b0b8262aac77e9d4dc8c2be478fcc5a4fa58802a))
* Use sass-embedded ([e8bf5e3](e8bf5e33f77ad1e471d312f2903569701cc6b22e))
* Use withDefaults for AssigneeList ([811a933](811a933cd380207ea8abf23757fd1fe620dd118d))
* Use withDefaults for Reminders ([6990be7](6990be705c6d47047147dea2930196246f1e2885))
* Use withDefaults in Description (#2453) ([e9a932e](e9a932e0f04de9a9ecda22410a1d7507c6d09d83))
* Validate expand api parameter ([bc0c0b1](bc0c0b103fe183d948816f722efec606452ca401))
* WithDefaults for Flatpickr ([289bb73](289bb73e9ec7e63b04cb2e903c1296c15b7b7639))
* WithDefaults for RelatedTasks ([70e027a](70e027a84e50e4586afc75cf7222097692aa282f))

### Miscellaneous Tasks


* *(attachments)* Refactor building image preview
* *(auth)* Refactor creating users in openid and ldap
* *(auth)* Refactor registration enabled setting in /info
* *(auth)* Rename error
* *(auth)* Rename external team id find methods
* *(avatar)* Decouple upload from web handler
* *(caldav)* Refactor fetching projects
* *(ci)* Debug crowdin
* *(ci)* Debug crowdin
* *(ci)* Debug crowdin
* *(ci)* Remove drone config
* *(ci)* Rename frontend-build step for better naming consistency
* *(ci)* Sign drone config
* *(ci)* Update s3-action
* *(ci)* Use correct git email for automated commits
* *(ci)* Use latest s3 action
* *(ci)* Use main of s3 action
* *(config)* Append .file to config values when reading
* *(config)* Migrate config renovate.json (#437)
* *(db)* Simplify MultiFieldSearch
* *(desktop)* Remove unused connect-history-api-fallback (#948)
* *(dev)* Add sass-embedded patchelf script alias to devenv
* *(dev)* Add test:all mage command
* *(dev)* Add zed tasks
* *(dev)* Insert final newline
* *(dev)* Replace old fonts when updating
* *(dev)* Update devenv
* *(dev)* Update gitignore for AI tools
* *(dev)* Use latest devenv docker container for devcontainer
* *(dev)* Use unstable devenv image for devcontainer (#993)
* *(devenv)* Do not install cypress on darwin
* *(docker)* Add more files and folders to dockerignore
* *(docker)* Use new env format
* *(docs)* Clarify usage of related model creation
* *(errors)* Always add internal error to echo error
* *(files)* Use absolute file path to retrieve and save files
* *(filter)* Move FilterInputDocs component
* *(filter)* Remove old filter input component
* *(frontend)* Enforce tab indentation
* *(frontend)* Migrate vue component options (#917)
* *(i18n)* Improve overdue task emails translation
* *(i18n)* Update translations via Crowdin
* *(logging)* Simplify log template string
* *(magefile)* Use tx.Sync instead of Sync2
* *(openid)* Add more debug logging when retrieving token
* *(openid)* Move openid team struct to openid package
* *(openid)* Use general external team sync
* *(plugins)* Ignore plugins dev folder
* *(project)* Do not use fmt.Sprintf directly
* *(projects)* Only pass users to checks
* *(renovate)* Update github actions only once a month
* *(subscription)* Return subscription entity type using json Marshaler
* *(tasks)* Add more details to error message
* *(tasks)* Add more details to error message
* *(tasks)* Move drag options to direct attributes instead of v-bind
* *(test)* Cleanup and improve e2e tests
* *(typesense)* Add more debug logging
* *(user)* Refactor invalidating upload avatar cache
* *(utils)* Remove deprecated MakeRandomString function
* *(web)* Always set internal error
* *(web)* Directly use new db session
* *(web)* Move web handler package to Vikunja
* *(web)* Remove redundant use of fmt.Sprintf
* *(web)* Remove unused echo context
* *(web)* Use config directly
* *(web)* Use errors.As instead of type assertion
* *(web)* Use logger directly
* *(web)* Use web auth factory directly
* *(webhook)* Refactor reloading event data
* 0.24.2 release preperations ([b031c97](b031c9772fbbe4c3d7b25f556bbb6776e734df57))
* 0.24.3 release preperation ([7329029](732902919ba531174e8479ce03a80cd9e7d99420))
* 0.24.4 release preperation ([ca048d0](ca048d07f9a1cc66215b771e0cc0e2c848c84e93))
* 0.24.5 release preperation ([3c13d3b](3c13d3b63536d0735da450a61cb09c6b527bc457))
* Add Bug type to bug issue template ([9cef2c4](9cef2c4c97098bf187bd5eb11abb710993d2cbb5))
* Add dateFrom and dateTo props with undefined defaults in ShowTasks component ([0a1a67f](0a1a67f2486d203c0e24ddee05b355efe8b3121b))
* Add debug logging around provider failure ([bbd3567](bbd3567e43a18c05730fbe260c4bd5cb8f9ce72b))
* Add go and direnv to recommended vscode extensions ([89f78cd](89f78cd36970fb38df4e6e6ebffcc62d7c63968a))
* Add missing eof newlines (#969) ([5b9d4fc](5b9d4fcc720fe65548ab6c3f7c974c12b28d13af))
* Add more debug logging when returning error ([4ea3c01](4ea3c01b5f4903db62b7e32c9a5bef68a8b12ed8))
* Adjust comment about bucketless tasks (#1004) ([57dfdc5](57dfdc5168f53a094c24505966303a6d0c60ea9a))
* Allow v-html in FilterInput ([1830695](18306956ae829afa6db54a11f5db02f522029d4b))
* Cleanup ([e1893ff](e1893ff5732275986871bf6d9a53415157182ce3))
* Cleanup unused helper ([22579df](22579dffae4f7adf9b7a5f5405b9e74da387b5c2))
* Correctly set default for prop in ReminderDetail ([1be0c96](1be0c96b65ae594c498ecaf6bea2469a970ca387))
* Dedupe pnpm packages (#877) ([51044a9](51044a9c52bd98e62ed63076027ac5f2e0116e94))
* Disable eslint rule for v-html in ProjectInfo component ([12e1a90](12e1a90f79ff65714fde0ddc16388bb67f7c3c07))
* Disable vue/no-boolean-default lint rule ([7368f5c](7368f5c323c9f4d9c5d3162f3020a5fbdfbdbf47))
* Do not set default for required prop in EditAssignees ([cd9e5bd](cd9e5bddd93f55eeb7f4965380519005d36e355a))
* Do not set default for required prop in EditLabels ([7fe17b1](7fe17b1eb80ce7ceb321b2ecf9ea9e638b34b3f0))
* Do not set default for required value in Attachments ([80a54ec](80a54ecb824d564c07f820f25c0f2ed75bd40aec))
* Do not set default for required value in DatePickerWithValues ([a50fa13](a50fa13d7e429b9829cc78db28b46e22a54da93b))
* Do not set default for required value in Datepicker ([10161c0](10161c033a60c792bf6812d0aa660fd3414136b9))
* Do not set default for required value in DatepickerInline ([9d1750f](9d1750f1dbb27e791b3fb3607f84e10c184c4bd1))
* Do not set default for required value in FancyCheckbox ([886ab6b](886ab6b1d771d7a93c8a69f01c24c5013e3805b3))
* Do not set default for required value in Multiselect ([9edabb8](9edabb800f437d5d393be5bef3aaa21fe8801386))
* Do not set default for required value in PriorityLabel ([c71084d](c71084d64a23d622c7c7a571e5e5d58cfd5e4d58))
* Do not set default for required value in ProjectCardGrid ([b6435ad](b6435ad23f11f002cab3a96cec246a68dc70a4a8))
* Do not set default for required value in Subscription ([4e6a272](4e6a272a0f2b77993480328d125b3c7c36755c41))
* Do not set defaults for required props in Heading ([93d2021](93d20216e266097ade38d3d70f23cd7f76e52824))
* Do not set defaults for required props in RelatedTasks ([a71b6b6](a71b6b6ab3a68d7938e4e86ba61dac049caaa715))
* Enable debugging workaround in devenv ([92c58ab](92c58ab32faee4c2821a1161158c2d4cb3ddce57))
* Explicit function origin (#2945) ([f76970b](f76970b5a33c100d89f990a09482b68242c5d533))
* Fix comment ([1f3eb8f](1f3eb8f2a319e49cfb94fe6957950749040cbcf6))
* Fix indentation ([454418e](454418ee6360756a6be4a3c6690368643c5b84aa))
* Improve cypress factory types ([59aca9f](59aca9fd5dc016509c17228b64a3f5e7b0757064))
* Improve debug logging ([e9d9f04](e9d9f04763be7681c0cfb05b85f0da176c9f7df5))
* Improve error message ([bc5fd38](bc5fd380e55252f360059d90b894ddbe60688d94))
* Increase healthcheck interval ([675985b](675985b26c8ca192ced44ece32adfe5a1c11ce5f))
* Make disabled prop optional in RepeatAfter component ([7beec2e](7beec2ef5a54f3d133d2981f4c96cfb58ee3eaf3))
* Make prop optional in Comments ([60fcc67](60fcc67fbe1528b10216c02ff7bd72a1fc406514))
* Make prop optional in Datepicker ([fcaf5ab](fcaf5abbf473f0b10e6fe9148f822595e218ed5b))
* Make prop optional in Flatpickr ([22c1fa2](22c1fa2be1d3b34b3bb881326b009f453b485790))
* Make prop optional in PaginationEmit ([359b75f](359b75feacf4ec313a6902865cb415551f7b40fc))
* Make prop with default value optional in KanbanCard ([a8d60a4](a8d60a423a2beba6e2f5b0940a7b5c1c908457db))
* Migrate eslint config ([b601671](b601671395bd04cd279fd3a4054198188a9cd0ce))
* Rearrange cron registers ([4b2b8e3](4b2b8e3b83caad72bd9a1574f9967bed8e06d3d7))
* Refactor searching for link shares ([a571d42](a571d42f46aac8d6df4a95f6a928890e4e76e75f))
* Release preparation ([f16cd8d](f16cd8d81338340a8c76924ffe6e78243f22b522))
* Remove SelectProject component since it is not used ([6c85d12](6c85d12ee07ff159f77f78909ba7d5fae5a9b3ed))
* Remove browserslist:update script (#660) ([027240f](027240f08c72d8088da299c6a69514157ecde08a))
* Remove bulma spacing utiltities ([9e1ae2c](9e1ae2ce9c9cfe1a8f879ebcde5fc1640bb2e1c8))
* Remove console.log ([76f7797](76f7797e568e7a170cf78ecd973f3b256786f4a7))
* Remove console.log ([a110543](a1105434bff981e721640f31bc7f7cf3d1f34859))
* Remove default healthcheck in docker ([e56a01f](e56a01f42d9418a3cfff3b0622a33fad497b874c))
* Remove default value for prop in Description ([edc60c7](edc60c7c08d79a5cf8b72cb8074bdedf56119995))
* Remove deprecated config settings ([3509f1b](3509f1beb3d56145f51869635e0d5c47335c25d1))
* Remove lodash.debounce ([bcd306b](bcd306b84d189f7e0d4426121b428eec2637de78))
* Remove obsolete prop ([866a179](866a1791dad987ff2daaff4bd3c2b7543661e7a1))
* Remove the option modern-compiler ([8f5be72](8f5be7210482d8eaff208e7c044c52569767f37e))
* Remove unnecessary type prop from user share components in ProjectSettingsShare ([2f671ac](2f671ac187ac4ef332352aee4aa927127b3d98f6))
* Remove unused component ([5539591](5539591d9786cb351ad1ddc3f1dbcd57ed4d7f3f))
* Remove unused rushstack eslint patch ([d4a5d1e](d4a5d1ecdf255150d84060fbbd2a13fff17eb076))
* Rename API test suites (#938) ([6671ce3](6671ce38a807936e49334489f331c1822433f491))
* Rename user_id field to username ([2fd3046](2fd3046accfaf6a4e161ba978f940aa20d069bac))
* Replace all uses of bucket_id with the const ([d81f2db](d81f2db6ef3d34a7e20d2f4e05d2fe39f6538038))
* Set default prop value for Filters ([15530fc](15530fcb0a715e5de144c4daccc0ff4a4913c8b1))
* Set default values for optional props in Multiselect ([de26f32](de26f32758d9f3828fc1ccccd2c1c810c4559811))
* Set default values for optional props in Password ([ae0e893](ae0e8939edf1b082389a07c4cfc3ddb96b61a88b))
* Simplify sentry code ([8732837](8732837596dc60cf6c95949e9303e3c4baaaa18c))
* Use bulma sr-only styles instead of tailwind's ([5a406b2](5a406b2eccad77da168cf295dd171864d2606b31))
* Use nixpkgs unstable for more recent packages ([1d62461](1d624612ee5dc81f9fbaa00ef3bcfe7d0ebbe608))
* Use ref for new comment value ([a99518c](a99518c2b9d155a3e9dafa2f4cb78cacab5be4a6))

### Other

* *(other)* "feat: remove cypress install"
* *(other)* Add Issuer and Subject to user list command (#3063)
* *(other)* Add healthcheck command (#2856)
* *(other)* Allow filtering tests from mage (#1072)
* *(other)* Fixes typo in config comment: wheter --> whether (#673)
* *(other)* Remove concurrency from test workflow (#863)
* *(other)* Revert "feat: improve docker layers (#803)"
* *(other)* [skip ci] Updated swagger docs

### Refactor

* Use query parameter only when looking for password reset token ([3658cde](3658cde42fa73bf0982f757c2cf4ead1e55cb94a))
* Move test ([510b1f2](510b1f246ab00b6b65a077121d94e148d2856e96))
* Schedule user deletion ([9acba7a](9acba7a24544af32c62819f77cf8fb76ef04fc9a))
 
## [0.24.6] - 2024-12-22

### Bug Fixes

* *(export)* Update only current user export file id

## [0.24.5] - 2024-11-21

### Bug Fixes

* *(attachments)* Check permissions when accessing all attachments
* *(saved filters)* Check permissions when accessing tasks of a filter
* Pin xgo to 1.22.x ([87b2aac](87b2aaccb8cdcbe1ecb6092951a0bfe224ad7006))
* Upgrade xgo ([19b63c8](19b63c86c51f67614b867c75a58cda1774685edd))
* Upgrade xgo docker image everywhere ([04b40f8](04b40f8a7dcd01a86ddb8b27596073d1e50f9e97))
* *(ci)* Do not build linux 368 docker images
* Disable 368 releases ([73db10f](73db10fb02268e07d29842493df55f4d645ac503))
  - **BREAKING**: disable 368 releases

### Miscellaneous Tasks

* Sign drone config ([17c4878](17c487875b5771c0971ee8bf030807171de2dddc))
* Go mod tidy ([9639025](96390257e0911089ae33a9565e8be7fa954c772c))

## [0.24.4] - 2024-09-29

### Bug Fixes

* *(attachment)* Do not use image previews
* *(checkbox)* Use sibling css selector instead of has
* *(files)* Only use service rootpath for files when the files path is not absolute
* *(filters)* Explicitly search in json when using postgres
* *(task)* Paginate task comments
* *(task)* Do not show close button when the task was not opened via modal
* *(task)* Improve task delete modal on mobile
* *(test)* Use correct selector for modal header
* Partial fix to allow list tasks in ios reminders app (#2717)
* *(attachments)* Revert "chore(attachments): refactor building image preview"

### Dependencies

* *(deps)* Update desktop lockfile
* *(deps)* Update dependency vue to v3.5.7
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency vue-i18n to v10.0.2
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency vue to v3.5.8
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v5.1.0
* *(deps)* Update dependency vue-i18n to v10.0.3
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v5.2.0
* *(deps)* Update dependency @sentry/vue to v8.31.0
* *(deps)* Update dependency tailwindcss to v3.4.13
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency @sentry/vue to v8.32.0
* *(deps)* Update tiptap to v2.7.3
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency vue to v3.5.9
* *(deps)* Update dependency dompurify to v3.1.7
* *(deps)* Update tiptap to v2.7.4
* *(deps)* Update dependency vue to v3.5.10
* *(deps)* Update dev-dependencies

## [0.24.3] - 2024-09-20

### Bug Fixes

* *(a11y)* Hide unfocusable buttons
* *(api)* Return 404 response when using a token and the route does not exist
* *(auth)* Restrict max password length to 72 bytes
* *(caldav)* Make sure colors are correctly saved and returned
* *(caldav)* Reject invalid project id with error 400
* *(editor)* Restore the current value, not the one from a previous task
* *(files)* Use absolute path everywhere
* *(filter)* Do not replace labels keyword when the value is 'label'
* *(filter)* Make sure tasks are in a correct bucket and position when they are part of a date filter
* *(filters)* Immediately propagate changes
* *(filters)* Do not replace filter or project values when the id value resolves to undefined
* *(filters)* Correctly transform and populate saved filter when creating and editing
* *(home)* Explicitly use filter for tasks on home page when one is set
* *(kanban)* Save updated position to store
* *(kanban)* Make task creation loading spinner actually visible
* *(kanban)* Make kanban full width on mobile
* *(kanban)* Do not mark first bucked as done bucket in filter bucket mode
* *(kanban)* Correctly paginate filtered kanban buckets
* *(label)* Ignore existing ID during creation
* *(labels)* Trigger task.updated event when removing a label from a task
* *(labels)* Test error assertion
* *(labels)* Remove input interactivity when label edit is disabled
* *(labels)* Trigger task updated for bulk label task update
* *(modal)* Make sure modal and its content scrolls properly on mobile
* *(modal)* Do not prevent scrolling on mobile
* *(modal)* Make scrolling on iOS Safari work
* *(multiselect)* Make selectPlaceholder optional
* *(notifications)* Only add project subscription as task subscription when the user is not already subscribed to the task
* *(password)* Validate password before sending request to api
* *(project)* Show description in title attribute without html
* *(project)* Reset id before creating
* *(projects)* Do not hide 6th project on project overview
* *(projects)* Description not visible on mobile
* *(reminders)* Notify subscribed users as well
* *(service worker)* Use correct workbox version
* *(subscription)* Always return task subscription when subscribed to task and project
* *(subscriptions)* Ignore task subscription when the user is subscribed to the project
* *(subscriptions)* Correctly inherit subscriptions
* *(subscriptions)* Cleanup and simplify fetching subscribers for tasks and projects logic
* *(subscriptions)* Do not panic when a task does not have a subscription
* *(table)* Make sorting for two-word properties work
* *(task)* Set done at date when moving a task to the done bucket
* *(task)* Specify task index when creating multiple tasks at once
* *(task)* Cyclomatic complexity
* *(task)* Make print styles work when printing task detail view from kanban
* *(task)* Multiple overlapping defer due date popups
* *(task)* Align task title on mobile popup
* *(task)* Dragging and dropping on mobile
* *(task)* Add task to filter view after it was updated
* *(task)* Cleanup old task positions and task buckets when adding an updated or created task to filter
* *(task)* Mark related task as done from the task detail view
* *(task)* Open focused task when pressing enter
* *(test)* Cypress test selector
* *(typesense)* Only fail silently when a project was not found during indexing
* *(typesense)* Add new tasks to typesense properly
* *(typesense)* Make sure task positions are recreated properly when updating them
* *(typesense)* Use emplace instead of upsert to update documents
* *(typesense)* Index tasks one by one
* *(typesense)* Force position to always be float instead of auto-inferring
* *(typesense)* Use typesense bulk insert, log all errors
* *(user)* Do not create user with existing id
* *(view)* Do not crash when saving a view
* *(view)* Correctly resolve label for filtered views or buckets
* *(view)* Correctly resolve bucket filter when paginating
* *(view)* Correctly get paginated task results
* *(views)* Add migration for filtered kanban buckets* Lint ([53d62d3](53d62d35f4488940a96d755de93ded64b8ac34a3))
* Reset id before creating ([93f7dd6](93f7dd611ad288a149f5da5463867d224334815f))
* Test selector ([063aa7a](063aa7afec717c3ed05be9d2ca73bde3d0bd8d35))

### Dependencies

* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v5
* *(deps)* Update dependency @kyvg/vue3-notification to v3.3.0
* *(deps)* Update dependency @sentry/vue to v8.28.0
* *(deps)* Update dependency @sentry/vue to v8.29.0
* *(deps)* Update dependency @sentry/vue to v8.30.0
* *(deps)* Update dependency axios to v1.7.7
* *(deps)* Update dependency date-fns to v4
* *(deps)* Update dependency dayjs to v1.11.13
* *(deps)* Update dependency express to v4.20.0
* *(deps)* Update dependency express to v4.21.0
* *(deps)* Update dependency go to v1.23.1
* *(deps)* Update dependency pinia to v2.2.2
* *(deps)* Update dependency sortablejs to v1.15.3
* *(deps)* Update dependency tailwindcss to v3.4.10
* *(deps)* Update dependency tailwindcss to v3.4.11
* *(deps)* Update dependency tailwindcss to v3.4.12
* *(deps)* Update dependency vue to v3.5.3
* *(deps)* Update dependency vue to v3.5.4
* *(deps)* Update dependency vue to v3.5.5
* *(deps)* Update dependency vue to v3.5.6
* *(deps)* Update dependency vue-i18n to v10
* *(deps)* Update dependency vue-i18n to v10.0.1
* *(deps)* Update dependency vue-i18n to v9.14.0
* *(deps)* Update dependency vue-router to v4.4.3
* *(deps)* Update dependency vue-router to v4.4.4
* *(deps)* Update dependency vue-router to v4.4.5
* *(deps)* Update dependency vuemoji-picker to v0.3.1* Chore(deps): update goreleaser/nfpm docker tag to v2.40.0 (#2647)
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update github.com/wneessen/go-mail to v0.4.4
* *(deps)* Update golangci
* *(deps)* Update module dario.cat/mergo to v1.0.1
* *(deps)* Update module github.com/gabriel-vasile/mimetype to v1.4.5
* *(deps)* Update module github.com/getsentry/sentry-go to v0.29.0
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.23
* *(deps)* Update module github.com/prometheus/client_golang to v1.20.3
* *(deps)* Update module github.com/prometheus/client_golang to v1.20.4
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.6.1
* *(deps)* Update module github.com/threedotslabs/watermill to v1.3.7
* *(deps)* Update module github.com/typesense/typesense-go to v2
* *(deps)* Update module github.com/typesense/typesense-go to v2
* *(deps)* Update module golang.org/x/crypto to v0.27.0
* *(deps)* Update module golang.org/x/image to v0.20.0
* *(deps)* Update module golang.org/x/oauth2 to v0.23.0
* *(deps)* Update module golang.org/x/term to v0.24.0
* *(deps)* Update module golang.org/x/text to v0.18.0
* *(deps)* Update pnpm to v9.10.0
* *(deps)* Update tiptap to 2.6.6
* *(deps)* Update tiptap to v2.7.0
* *(deps)* Update tiptap to v2.7.1
* *(deps)* Update tiptap to v2.7.2
* *(deps)* Update vueuse to v11
* *(deps)* Update vueuse to v11.1.0
* *(deps)*: update dependency flexsearch to v0.7.43 (#2095)
* *(deps)*: update golangci/golangci-lint docker tag to v1.61.0 (#2678)

### Documentation

* *(api)* Use correct return type for the /user endpoint

### Features

* *(event)* Simplify dispatching task updated event from only a task id
* *(navigation)* Use focus-visible for nav items
* *(task)* Use focus-visible for task focus styles

### Miscellaneous Tasks

* *(attachments)* Refactor building image preview
* *(devenv)* Do not install cypress on darwin
* *(docker)* Use new env format
* *(docs)* Clarify usage of related model creation
* *(errors)* Always add internal error to echo error
* *(files)* Use absolute file path to retrieve and save files
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(logging)* Simplify log template string
* *(magefile)* Use tx.Sync instead of Sync2
* *(subscription)* Return subscription entity type using json Marshaler
* *(tasks)* Move drag options to direct attributes instead of v-bind
* *(typesense)* Add more debug logging
* *(web)* Move web handler package to Vikunja
* *(web)* Remove unused echo context
* *(web)* Use errors.As instead of type assertion
* *(web)* Remove redundant use of fmt.Sprintf
* *(web)* Directly use new db session
* *(web)* Use config directly
* *(web)* Use web auth factory directly
* *(web)* Use logger directly
* *(web)* Always set internal error* Remove console.log ([40105ee](40105ee4ced980f52565baec4c3219b0ddd4f6ec))
* Fix comment ([1df4a4e](1df4a4ea2e2ca4332347468e8973a2dcbab06ed7))
* Add go and direnv to recommended vscode extensions ([6ab12b9](6ab12b9dd133b52ed7267b6e9334081c2f9719ca))
* Remove console.log ([1e7d9c9](1e7d9c982d3d472e9b4082991b41e6567556f2b2))
* Rearrange cron registers ([4857bfb](4857bfbbdb8401b6ef02b1dc8de93f2a09e8bc3a))

### Other

* *(other)* [skip ci] Updated swagger docs

## [0.24.2] - 2024-08-12

### Bug Fixes

* *(i18n)* Change casing of Ukrainian language in selector
* *(kanban)* Always make cover image full width
* *(mail)* Do not fail testmail command when the connection could not be closed.
* *(migration)* Make sure tasks are associated to the correct view and bucket for data imported from Vikunja dump
* *(migration)* Ensure project background gets exported and imported
* *(projects)* Trigger only single mutation
* *(task)* Do not allow moving a task to the project the task already belongs to
* *(task)* Set current project after moving a task
* *(task)* Move task into new kanban bucket when moving between projects
* *(views)* Do not create task bucket and task position entries when duplicating a project* Emit for DatepickerWithValues ([3aaf363](3aaf3634134a6989337bad02ac99a9329d33b17f))
* Textarea autosize for LanguageTool ([d9f5555](d9f555554e5ecfa9d1243c565e2f42c77f7a7597))
* Remove console log ([0ca43dc](0ca43dc147acd04d9f9b566325ccde0a5782680f))

### Dependencies

* *(deps)* Update module github.com/coreos/go-oidc/v3 to v3.11.0
* *(deps)* Update flake
* *(deps)* Update go toolchain to 1.22.5
* *(deps)* Update dependency node to v20.16.0

### Documentation

* Clarify Todoist redirect url ([7117303](7117303d5705199bb39fb6661b20deebea96959f))

### Features

* *(editor)* Support custom protocol for links* Use withDefaults for Reminders ([8729c24](8729c24e1d2bdc783e750dc1166a8ab31a716107))
* Improve projects store ([d707e15](d707e1576a006baa73c529b432e694dee237db76))
* Improve label store ([a0e3efe](a0e3efe2d12521e773334a19ac985a2011575c3c))
* Improve priority visibility ([dddba4d](dddba4d64a9ddddd6f442c8180661d2d498de994))
* Add tailwind with prefix (#2513) ([d7c5451](d7c54517297e750d92618376f52045b11da6e82c))
* Improve ProjectSettingsViews ([811ccc1](811ccc1baa2fe1fb9fcc8b5515c8c3f3b591e96a))
* Add missing peer dependency ([d586f69](d586f691b7a245bbffd5df4e7a4937f0527047e0))
* Switch from nix flakes to devenv ([73f923b](73f923bc47d32a6a9a689e461901295b07646ebf))


### Miscellaneous Tasks

* *(i18n)* Update translations via Crowdin
* Remove lodash.debounce ([fc780a9](fc780a90ae856038db73d3ad63fcdd5627211c36))
* Improve error message ([6e38bcf](6e38bcf3498a15a1dc1f6fbef254c08513b408b3))
* Use nixpkgs unstable for more recent packages ([f5040ad](f5040ad2f4996fe1ccf03172b1a1ee6007068ee7))

### Other

* *(other)* [skip ci] Updated swagger docs

## [0.24.1] - 2024-07-18

### Bug Fixes

* *(api tokens)* Show error message when the user tries to create an api token without at least one permission selected
* *(filter)* Make sure filter values are properly escaped before executing them
* *(filters)* Add task to buckets of saved filters when creating the task
* *(filters)* Add tasks to filter buckets when updating the filter
* *(filters)* Do not create a default filter for list view when creating a saved filter
* *(filters)* Immediately emit filter query when editing saved filter
* *(filters)* Make sure filters are replaced case-insensitively before filtering on the server
* *(filters)* Only insert task buckets and positions when there are any
* *(filters)* Reload tasks silently when marking one done in the list
* *(filters)* Show actual error message from api when the filter query is invalid
* *(filters)* Trim spaces when parsing filter values
* *(kanban)* Dispatch task updated event when task is moved between buckets
* *(kanban)* Dispatch task updated event when task position is updated
* *(kanban)* Do not allow to create new tasks in saved filter
* *(kanban)* Do not move repeating task into a different bucket
* *(kanban)* Make sure tasks which changed their done status are moved around in buckets
* *(kanban)* Move repeating task back to old bucket when moved to the done bucket
* *(kanban)* Move task to done bucket in all views when moved to done bucket in one view
* *(kanban)* Move task to done bucket when it was marked done from the task detail view
* *(kanban)* Put task into correct bucket when creating via kanban board
* *(kanban)* Update task done status after moving it into done bucket
* *(kanban)* Use correct assertion in the test
* *(kanban)* Use correct text color for deletion button
* *(migration)* Correctly set bucket for related tasks
* *(migration)* Failed migration typo
* *(migration)* Revert to old path for migration routing
* *(project)* Do not use project id of nil project in error
* *(projects)* Do not create backlog bucket when duplicating views
* *(projects)* Do not create buckets in the original project when duplicating a project
* *(quick add magic)* Create the task even when it only contains quick add magic keywords
* *(settings)* Overflow of select on mobile
* *(task)* Use backdropView prop
* *(tasks)* Do not use typesense modified options to search with database
* *(tasks)* Explicitly add task position to select statement when looking up tasks with Typesense
* *(tasks)* Limit to max 250 entries when using typesense
* *(translation)* TOTP casing
* *(typesense)* Do not crash after creating a project when tasks are not yet indexed
* *(typesense)* Do not use modified opts for db fallback search
* *(typesense)* Reindex tasks when their position changed
* *(vscode)* I18n-ally locales path* ProjectSearch default value ([f08039b](f08039b23c1ff8ca208f5c5911f788780ecc80ee))
* Add info log message when starting to run migrations ([5e36bf7](5e36bf797e99d99f83cb084f8c1707d994d0559f))
* Add missing disabled prop ([ed0ef38](ed0ef385e9ba34524d3309bec87eeb44d9dba471))
* App bottom padding ([51660f7](51660f76779e189bca2a24f1e4fca34b8a1a2898))
* Disable button if loading ([a721d92](a721d9286bbb26e0a182992af6b2128e03d5670f))
* Dropdown item disabled prop ([3317280](3317280062c8506d092d5121b92eb4992177f9c2))
* Gitignore dist path ([7ef6ddf](7ef6ddf8f7630e6370bed762779b4ebabbf8962b))
* Lint ([7c42fb5](7c42fb5d75fd9a5aa01abaa852452875f07a1f61))
* Missing error handling ([744b40e](744b40e7f780851d5ce8a288271f78acc462ed14))
* Muliselect optional props ([0a81855](0a81855bc1b403233dec18f036c80f3cad70edee))
* Reorder mail options (#2533) ([136ef58](136ef58820b8e2d27ad0ca50bdd9adabd8e4a95d))
* Scss deprecation warning ([db81701](db81701d3841e12c39b751e919d70fdafc1869d6))
* Spelling mail ([2dc5415](2dc541571cbecc02a178182217d1e4596e4a62aa))
* Wrapped button ([af639a1](af639a180cf56519e9a3a31710da6fee5b305735))

### Dependencies

* *(deps)* Update dependency @github/hotkey to v3.1.1 (#2329)
* *(deps)* Update dependency @sentry/vue to v8.14.0
* *(deps)* Update dependency @sentry/vue to v8.15.0
* *(deps)* Update dependency @sentry/vue to v8.17.0
* *(deps)* Update dependency @sentry/vue to v8.18.0
* *(deps)* Update dependency dayjs to v1.11.12
* *(deps)* Update dependency dompurify to v3.1.6
* *(deps)* Update dependency ufo to v1.5.4
* *(deps)* Update dependency vue to v3.4.32
* *(deps)* Update dependency vue-tsc to v2.0.26
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update flake
* *(deps)* Update font awesome to v6.6.0
* *(deps)* Update goreleaser/nfpm docker tag to v2.38.0
* *(deps)* Update module github.com/arran4/golang-ical to v0.3.1 (#2606)
* *(deps)* Update module github.com/microcosm-cc/bluemonday to v1.0.27
* *(deps)* Update module golang.org/x/crypto to v0.25.0
* *(deps)* Update module golang.org/x/term to v0.22.0
* *(deps)* Update pnpm to v9.5.0
* *(deps)* Update tiptap to v2.5.4

### Features

* *(Multiselect)* Hide multiselect during loading
* *(project)* Add " - duplicate" suffix to duplicated projects title* Move powered by to bottom ([f9e0b43](f9e0b4370700b4cc9c8e99923737bb5aa022a1d0))
* Add withDefaults and emit types to PrioritySelect ([aaa2abc](aaa2abced4258645a9a47b228cece21741bfea25))
* Add withDefaults, defineEmits and defineSlots types for Dropdown ([5545b0e](5545b0e447d406837324186c765cf2e8c8ed47c4))
* Define prop and emit types DatepickerWithRange ([daeefeb](daeefeb487966bfc6535b3d5c8e1dd07ec8a0951))
* Define prop and emit types for FilterPopup ([9d2e79f](9d2e79f7253006d504bd916ab8d820187d50700b))
* DefineModel and withDefaults for PercentDoneSelect ([8ac0eb4](8ac0eb4aa4fbe5a3b147e7f31a29aa2c2f84a1ca))
* Improve BaseButtonEmits type ([c37fe49](c37fe4989015ee18b38bd89b9470c14484759cc8))
* Improve main nav spacing and open speed ([faa797f](faa797f461a57e8a9e115b760af92ab0024f6eed))
* Improve settings label casing ([20efacf](20efacfa59a7f7ff81fd17c2c3b3c9a1d5528e5d))
* Remove PropTypes helper from FilePreview ([0bc1832](0bc18320509e6db7ae484f7336282b7e8e5e841e))
* Remove PropTypes helper from ProjectInfo ([8ad7e7c](8ad7e7c9055e50768ea5b0ad5798d395bb5d3514))
* Remove PropTypes helper from ProjectSettingsEdit ([85889ff](85889fff5603f0b7c7a0f833cae7c32610eeede4))
* Remove eslint rule vue/no-required-prop-with-default ([df1f805](df1f805294e34f8bfaaf60ea65f18cabbd932d9c))
* Remove props destructuring ProjectSettingsViews ([20bdb01](20bdb011877fe17fe5c266eafe77b96dce1646d9))
* Remove props destructuring for DeferTask ([46aa2ff](46aa2fff0f01bf233e4edd39e6675c358c488794))
* Remove props destructuring for Filters ([3ff56d7](3ff56d79874d013ad97c8635fd580edf95290b39))
* Remove props destructuring for ProjectView ([553a97f](553a97f03de3e16593c006ca4a9043fadf5d6a6f))
* Remove props destructuring for ProjectWrapper ([38744df](38744dfd5db0acb53ae03db5e2021265876df2f7))
* Remove props destructuring from Attachments ([07130bd](07130bdc3a65bfb4ed1775e61312b68c62521a4a))
* Remove props destructuring from ColorPicker ([eb07be1](eb07be1a62a9af377b044a7e0268d9d7a8bab77c))
* Remove props destructuring from FilterDelete ([825d1ad](825d1add49595184593ff6fb3ab228cf3fbde9cb))
* Remove props destructuring from FilterEdit ([449e9a9](449e9a911c2fcca7712479cd78dbfe80d5a4bde2))
* Remove props destructuring from FilterInput ([fd6692e](fd6692ea1afce3950c2295a4e8fb551af80e52aa))
* Remove props destructuring from KanbanCard ([ddc18aa](ddc18aa17739963afa5d156241aa07b7100374f4))
* Remove props destructuring from NoAuthWrapper ([c9591fe](c9591fe464e24f38126d5562ef7d0e6060f14b39))
* Remove props destructuring from ProjectKanban ([5af908b](5af908b2e5508712004ee1c702915e4def0ca007))
* Remove props destructuring from ProjectList ([0c70aed](0c70aedeb1a5ca05db3cc2b87138d28427eac959))
* Remove props destructuring from ProjectTable ([99e90c0](99e90c0b02561f1c7b8f49b93ece11b878137e48))
* Remove props destructuring from ShowTasks ([d02c349](d02c349231f29326511755dcae5003d3b3c01ca0))
* Remove props destructuring from SingleTaskInProject ([a15831e](a15831eb33b5104e19f0b85b882dd8ec5336ba62))
* Remove props destructuring from TaskDetailView ([a10f9ca](a10f9ca22580f6fcc435f152776867b380d5be6c))
* Remove props destructuring from ViewEditForm ([2f92e40](2f92e407ccf42892675dc1583dfa057474cb2489))
* Remove unnecessary prop from Card ([1eb1aa2](1eb1aa257574733894f865913a4bb173e22774ea))
* Rename TheNavigation to AppHeader ([43e38fa](43e38fae17d987d96820dcdfd936fcf8d11899ff))
* Simplify playPopSound setting check ([42c458a](42c458a736d88bf4b0f18fa23c08273ddf9c7ca5))
* Type defineEmits for ApiConfig ([1966cc3](1966cc3c0e7a590651db2a5bd0acdcd83e948aa7))
* Type defineEmits for Card ([043a6dd](043a6dd049ba22fa1c87ed88aa0b2f8dba4f2593))
* Type defineEmits for CreateEdit ([098c99f](098c99fd2ede44ddf231e1b504b9cf6d8c0ed17b))
* Type defineEmits for Heading ([fd3f1de](fd3f1de861ac593c0f780c3f341a4b0c5d7b2f1b))
* Type defineEmits for ProjectSearch ([f38a5c9](f38a5c9220c6404152e261a78e4829e52cef5d1b))
* Use withDefaults for Comments ([f19f19b](f19f19bb75a457dc2073d8a06f064eae04a3b772))
* Use withDefaults for DateTableCell ([f586e51](f586e51aad5194043fec37e3b0af94a30c95f3a3))
* Use withDefaults for Datepicker ([78811d9](78811d916aeaa105560c5b7c091290f3b2e38683))
* Use withDefaults for DatepickerInline ([df6a9b6](df6a9b67fd2a462e5027538e521da174ecd24bc6))
* Use withDefaults for DatepickerWithValues ([cb70641](cb706416c61ba31d1fd54d8b9bd8549cc9a33528))
* Use withDefaults for EditLabels ([6e72606](6e72606d7454fbb3b17302d3ef1d2cdf186873ae))
* Use withDefaults for FancyCheckbox ([b4e9d94](b4e9d9437e4fb4b8e5e56c633fc4c74ee45a6225))
* Use withDefaults for Heading ([5cf57a5](5cf57a520cd7048aba0656ce00333cc7cfcc14f9))
* Use withDefaults for Pagination ([1083216](10832165c3798bef63590c292426feb67e178f15))
* Use withDefaults for Password ([577f5ae](577f5ae69a31e14cdff6e0aed5f3091d07484faa))
* Use withDefaults for PriorityLabel ([871e0ac](871e0acd8aa8ba6a9ad7fb08bda8badf20104f94))
* Use withDefaults for ProgressBar ([2f63384](2f6338484bda4670133f3db6c122dbaa6a231c15))
* Use withDefaults for ProjectCardGrid ([479b786](479b786761948b0a0427676c42e082e75579029c))
* Use withDefaults for Reactions ([d6c3b5a](d6c3b5a9a10a04c8fac47ad998ebf434e93dfa71))
* Use withDefaults for ReminderPeriod ([4b9b9da](4b9b9da122a2d085dcff8fbb8dd2693062689fd4))
* Use withDefaults for RepeatAfter ([c8585d1](c8585d1a691403e6331cd9aff852873c0f7771d3))
* Use withDefaults for SelectProject ([fd12c87](fd12c8705ea9b28db6947fd98f2d34ce46fc15b2))
* Use withDefaults for SelectUser ([b500981](b50098143474be8d86e0ae0543bab1787585dfe7))
* Use withDefaults for Subscription ([30769fb](30769fb6eadb21d45629dcf2c67cc1174d26e828))
* Use withDefaults for UserTeam ([f2fdbad](f2fdbad7d4bde5880694a611a4f9d68fe2017367))
* Use withDefaults in ReminderDetail ([a56331d](a56331d39d12c65838b13987fe2096c4ca2a703f))
* WithDefaults for EditAssignees ([f1481d7](f1481d702cc426cddfeda0edae6a6cd06d9f3380))
* WithDefaults for Multiselect ([413d1f9](413d1f9ad7e9ba0234d81a1564745cd03d9f1044))

### Miscellaneous Tasks

* *(i18n)* Update translations via Crowdin
* *(popup)* Trigger close function directly
* *(project)* Rename receiver
* 0.24.0 release preparations ([0b14c31](0b14c311b4d344aecf974489eee19ab939c676ad))
* Go mod tidy ([e640149](e640149a23c931838a93b737804287a9ed570268))
* Update golangci lint config ([d2602a7](d2602a7629ee3d8cd67beb8ef1a44ca3cd2dd7f9))
* *(other)* [skip ci] Updated swagger docs

## [0.24.0] - 2024-07-02

### Bug Fixes

* *(BaseButton)* Comment spelling (#2348)
* *(assignees)* Spacing of users
* *(attachment)* Correct spacing around creation date
* *(auth)* Test assertion
* *(auth)* Use (issuer, name) to check for uniqueness of oidc teams (#2152)
* *(auth)* Log user out when the current account does not exist
* *(backgrounds)* Return full project after uploading image
* *(buckets)* Return correct task count for tasks in buckets
* *(caldav)* Return more than 1000 tasks
* *(caldav)* Check if vtodo contains any components
* *(caldav)* Do not crash for wrong parameters
* *(ci)* Enable cors for api testing
* *(ci)* Use correct docker image for frontend testing
* *(ci)* Use correct docker image
* *(ci)* Use correct docker image for desktop rename
* *(ci)* Correctly set shell for rename command
* *(ci)* Escape bash script for drone variable substitution
* *(ci)* Run test db in memory
* *(ci)* Sign drone config
* *(ci)* Exclude tasks from cron runs
* *(comments)* Order comments by created timestamp instead of id
* *(comments)* Do not use whitespace as gap
* *(datepicker)* Emit date value changes as soon as they happen
* *(datepicker)* Make the date format in the picker consistent with the input field
* *(db migration)* Do not try to create a unique index
* *(docker)* Don't install cypress in docker image
* *(docs)* Openid docs whitespace formatting (#2186)
* *(docs)* Typos
* *(docs)* Correctly document filter query usage
* *(dump)* Do not export files which do not exist in storage
* *(dump)* Only allow imports from the same version they were dumped on
* *(editor)* Ensure task list clicks are only fired once
* *(editor)* Set default id of tasklist items
* *(editor)* Revert task list dependence on ids
* *(editor)* Don't allow image upload when it's not possible to do it
* *(editor)* Do not use Tiptap to open links when clicking on them, use the browser native attributes instead
* *(editor)* Do not prevent shift+enter to add a line break in text
* *(editor)* Use colors from color scheme to render table cells
* *(export)* Make export work with project views and new task positions
* *(extensions)* Remove typescript-vue-plugin from recommendations (#2353)
* *(favorites)* Make favorites work with configurable views
* *(favorites)* Allow marking favorite tasks as done from favorites pseudo project
* *(filter)* Make sure single filter condition works
* *(filter)* Don't crash on empty filter
* *(filter)* Allow filtering on "in" condition
* *(filter)* Allow filtering for "project"
* *(filter)* Translate all tests
* *(filter)* Correctly filter for buckets
* *(filter)* Add role=search to filter card
* *(filter)* Correctly pass down options
* *(filter)* Bubble filter query changes up on blur only
* *(filter)* Don't transform anything when input is empty
* *(filter)* Correctly replace project title in filter query
* *(filter)* Do not match join operator
* *(filter)* Do not show filter footer when creating a filter
* *(filter)* Add white background to filter input
* *(filter)* Make sure highlight works for doneAt attribute
* *(filter)* Move spaces out of button to after the matched filter value to prevent removal of spaces
* *(filter)* Clarify `in` filter syntax
* *(filter)* Trim search term before searching
* *(filter)* Do not add enter in input field
* *(filters)* Lint
* *(filters)* Use readable colors for dark and light mode
* *(filters)* Date filter value not populated
* *(filters)* Make the button look less like a button to avoid spacing problems
* *(filters)* Color
* *(filters)* Make sure spaces before and after are not removed
* *(filters)* Pass correct filter query to kanban and gantt loading
* *(filters)* Swagger docs for kanban buckets
* *(filters)* Correctly use filter in saved filter
* *(filters)* Remove footer when editing a saved filter
* *(filters)* Layout problems with assignee user avatar
* *(filters)* Lint
* *(filters)* Close filter popup when clicking on show results
* *(filters)* Test fixture
* *(filters)* Correctly use date filters in gantt chart
* *(filters)* Do not require string for in comparator
* *(filters)* Parse labels and projects correctly when using `in` filter operator
* *(filters)* Label highlighting and autocomplete fields now work with in operator
* *(filters)* Don't escape valid escaped in queries
* *(filters)* Invalid filter range when converting dates to strings
* *(filters)* Replace project titles at the match position, not anywhere in the filter string
* *(filters)* Set default filter value to only undone tasks
* *(filters)* Rework filter popup button
* *(filters)* Lint
* *(filters)* Persist filters in url
* *(filters)* Do not fire filter change immediately
* *(filters)* Do not watch debounced
* *(filters)* Correctly return project from filter
* *(filters)* Correctly replace values when clicking on an autocomplete result
* *(filters)* Clear autocomplete results when starting the next character
* *(filters)* Make sure the same filter attribute is transformed in all instances
* *(filters)* Enclose values with a slash in them as strings so that date math values work
* *(filters)* Always show filter values in a readable color
* *(filters)* Always persist filter or search in query path and load it correctly into filter query input when loading the page
* *(filters)* Explicitly use `tasks.id` as task id filter column
* *(filters)* Do not match partial labels
* *(filters)* Allow managing views for saved filters
* *(gantt)* Use color variables for gantt header so that it works in dark mode
* *(gantt)* Correctly show day in chart
* *(i18n)* Use correct title for background settings menu
* *(i18n)* Clarify from current date string
* *(i18n)* Typo
* *(i18n)* Adjust tests from 34780daab0af0c088d6484d5fa0ddfba01471e8b
* *(i18n)* Remove duplicate key
* *(kanban)* Pass active filters down to task lazy loading
* *(kanban)* Reset done and default bucket when the bucket itself is deleted
* *(kanban)* Do not use the bucket id saved on the task
* *(kanban)* Remove unused function
* *(kanban)* Make sure all saved taskBucket positions are saved with their project view id
* *(kanban)* Save done and default bucket on the view and not on the project
* *(kanban)* Do not focus kanban board
* *(kanban)* Do not add bottom spacing to view
* *(kanban)* Do not focus on task list in bucket when clicking on a task
* *(kanban)* Fetch project and view when checking permissions
* *(kanban)* Remove leftovers of kanban_position property
* *(labels)* Make sure labels are aligned in the middle
* *(labels)* Allow link shares to add existing labels to a task
* *(logo)* Use correct month for pride logo change
* *(logo)* Add width and height to pride logo svg
* *(metrics)* Typo
* *(migration)* Make sure to correctly check if a migration was already running
* *(migration)* Do not halt the whole migration when copying a background file failed
* *(migration)* Show correct help message when a migration was started
* *(migration)* Do not expire trello token
* *(migration)* Convert trello card descriptions from markdown to html
* *(migration)* Trello checklists (#2140)
* *(migration)* Updated Trello color map to import all labels (#2178)
* *(migration)* Import card covers when migrating from Trello
* *(migration)* Only download uploaded attachments
* *(migration)* Show correct message after starting a migration
* *(migration)* Trello: only fetch attachments when the card actually has attachments
* *(migration)* Import card comments from Trello when migrating
* *(migration)* Invalid field in organization struct
* *(migration)* Import task comments with original timestamps
* *(migration)* Remove buckets table name when dropping index
* *(migration)* Ensure tasks are put into the correct bucket when migrating from todoist
* *(migration)* Put "Import from other services" in settings
* *(modal)* Do not set p in modal card as flex
* *(navigation)* Do not hide shadows of dropdown menu
* *(navigation)* Scrolling when many projects are present
* *(notifications)* Only sanitze html content in notifications, do not convert it to markdown
* *(notifications)* Rendering of plaintext mails
* *(openid)* OIDC teams should not have admins (#2161)
* *(password)* Don't validate password min length on login page
* *(pnpm)* Remove obsolete settings
* *(project)* Don't allow archival or deletion of default projects in UI
* *(project)* Check for project nesting cycles with a single recursive cte instead of a loop
* *(project)* Typo in table name
* *(project)* Correctly show the number of tasks and projects when deleting a project
* *(project)* Load full project after creating a project
* *(project)* Save the last 6 projects in history, show only 5 on desktop
* *(project)* Return the full project when setting a background
* *(project)* Remove child projects from state when deleting a project
* *(project)* Do not crash when views were not loaded yet
* *(project)* Delete all related entities when deleting a project
* *(project)* Do not crash when duplicating a project with no tasks
* *(project)* Return full project after duplicating it
* *(project)* Add more spacing between filter button and view switcher on mobile
* *(project)* Bottom spacing in list view
* *(project)* Make sure gantt and kanban views shared with link share are full width
* *(project)* Do not remove project from navigation after removing background image
* *(project)* Show "remove background" button only when the project has a background set
* *(projects)* Return correct project pagination count
* *(projects)* Load all projects when first opening Vikunja
* *(projects)* Load projects only one when fetching subscriptions for a bunch of projects at once
* *(projects)* Remove done bucket id field from projects struct
* *(projects)* Allow arbitrary nesting of new projects
* *(projects)* Do not return parent project id of parents where the user does not have access
* *(projects)* Do not return parent project id when authenticating as link share
* *(quick actions)* Do not allow creating a task when the current project is a saved filter
* *(quick add magic)* Parse full month name as month, do not replace only the abbreviation
* *(quick add magic)* Assume today when no date was specified with time
* *(reactions)* Do not enable reaction picker when the current user does not have write access
* *(reminder)* Do not close the popup directly after changing the value
* *(reminders)* Emit reminder changes at the correct time (and make sure they are actually emitted)
* *(reminders)* Make debounce logic actually work
* *(reminders)* Do not fall back to hours when the reminder interval is minutes
* *(reminders)* Do not show relative reminders as minutes when they round to hours
* *(restore)* Transform json fields during restore
* *(semver)* Fix produced version number (#2378)
* *(sentry)* Send unwrapped error to sentry instead of http error
* *(sentry)* Do not send api errors to sentry
* *(sharing)* Show user display name and avatar when displaying search results
* *(table view)* Do not sort table column fields when the field in question is hidden
* *(task)* Move done tasks to the done bucket when they are moved between projects and the new project has a done bucket
* *(task)* Navigate back to project when the project was the last page in the history the user visited
* *(task)* Clear timeout for description save when closing the task detail
* *(task)* Do not crash when loading a task if parent projects are not loaded
* *(task)* Show repeating indicator in task list for monthly repeating tasks
* *(task)* Only count unique tasks in a bucket when checking bucket limit
* *(task)* Do not require admin permission to move tasks between buckets
* *(task)* Do not try to set bucket for filtered bucket configuration
* *(task)* Show correct success message when marking a repeating task as done
* *(task)* Do not move task dates when undoing a repeated task
* *(tasklist)* Migrate old tasklist format
* *(tasks)* Sort done tasks last in relations
* *(tasks)* Correctly show different project in related tasks
* *(tasks)* Use correct filter query when filtering
* *(tasks)* Index and order by task position when using typesense
* *(tasks)* Make fetching tasks in buckets via typesense work
* *(tasks)* Ambiguous column name error when fetching favorite tasks
* *(tasks)* Do not crash when order by id and position
* *(tasks)* Tests
* *(tasks)* Clarify usage of repeating modes available in quick add magic.
* *(teams)* Use the same color for border between teams in list
* *(teams)* Do not show leave button for OIDC teams (#2181)
* *(teams)* Fix duplicate teams being shown when new public team visibility feature is enabled (#2187)
* *(test)* Use correct selector in Cypress test
* *(test)* Correctly mock localstorage in unit tests
* *(test)* Visit one more project in project history test
* *(test)* Add task to bucket in test
* *(test)* Cast result before comparing
* *(tests)* Make filter tests work again
* *(tests)* Do not try to create tasks with bucket_id
* *(ts)* Align with create-vue setup
* *(typesense)* Fix reindexing views and positions in typesense
* *(typesense)* Make fetching task positions per view more efficient
* *(typesense)* Correctly incorporate existing filter when it is empty
* *(typesense)* Only return distinct tasks once
* *(typesense)* Correctly join task position table when sorting by it
* *(typesense)* Do not try to sort by position when searching in a saved filter
* *(typesense)* Correctly index assignee changes on tasks
* *(views)* Correctly fetch project when fetching tasks
* *(views)* Do not break filters when combining them with view filters
* *(views)* Make gantt view load tasks again
* *(views)* Make table view load tasks again
* *(views)* Make fetching tasks in kanban buckets through view actually work
* *(views)* Fetch buckets through view
* *(views)* Return tasks in their buckets
* *(views)* Return buckets when fetching tasks via kanban view
* *(views)* Return tasks directly or in buckets, no matter if accessing via user or link share
* *(views)* Make no initial view work in the frontend
* *(views)* Move to new project view when moving tasks
* *(views)* Do not load views async
* *(views)* Get tasks in saved filter
* *(views)* Make setting task position in saved filters work
* *(views)* Make bucket creation work again
* *(views)* Make bucket edit work
* *(views)* Do not return kanban tasks multiple times
* *(views)* Make parsing work
* *(views)* View deletion
* *(views)* Create view
* *(views)* Set correct default view
* *(views)* Set current project after modifying views
* *(views)* Make kanban tests work again
* *(views)* Move all tasks to the default bucket when deleting a bucket
* *(views)* Duplicate all views and related entities when duplicating a project
* *(views)* Update test fixtures for new structure
* *(views)* Test assertions
* *(views)* Count task buckets
* *(views)* Return correct error
* *(views)* Integration tests
* *(views)* Import
* *(views)* Lint
* *(views)* Lint
* *(views)* Make tests for project history kind of work again
* *(views)* Tests for kanban and gantt views
* *(views)* Correctly pass project id when loading more tasks in kanban views
* *(views)* Return only tasks when the bucket id was already specified
* *(views)* Reset bucket when moving tasks between projects
* *(views)* Make kanban cypress tests work again
* *(views)* Make list cypress tests work again
* *(views)* Always redirect to the first view when none was specified
* *(views)* Make table view cypress tests work again
* *(views)* Correctly save and retrieve last accessed project views
* *(views)* Make link share cypress tests work again
* *(views)* Make overview cypress tests work again
* *(views)* Make task cypress tests work again
* *(views)* Kanban test assertions
* *(views)* Update done status of recurring tasks
* *(views)* Include order by fields in distinct clause when sorting by task position
* *(views)* Stable assertion for bucket in tests
* *(views)* Redirect to project after authenticating with a link share
* *(views)* Intercept request
* *(views)* Create bucket in test
* *(views)* Create default bucket
* *(views)* Do not map bucket id from xorm
* *(views)* Add bottom spacing
* *(views)* Update all fields when updating a view
* *(views)* Use correct assertion in test
* *(views)* Correctly pass view id to wrapper when gantt view is active
* *(views)* Transform view filter before and after loading it from the api
* *(views)* Refactor filter button slot in wrapper
* *(views)* Remove default filter from frontend, apply by default to new list views instead (#2240)
* *(views)* Check if bucket index already exists before creating new index
* *(views)* Make sure the view is saved properly in localStorage
* *(views)* Make sure view changes are reflected in switcher
* *(views)* Only allow project admins to manage views
* *(views)* Transform bucket configurations
* *(views)* Edit views with filters
* *(views)* Do not allow moving tasks or editing board when bucket mode is filter
* *(views)* Move bucket update to extra endpoint
	- **BREAKING**: The bucket id of the task model is now only used internally and will not trigger a change in buckets when updating the task.
* *(vue)* ToValue instead of unref
* *(webhook)* Log errors in webhook response
* *(webhooks)* Fire webhooks set on parent projects as well* Never return frontend on routes starting with /api ([641fec1](641fec12157504b8ed2935ba9943828662a725f9))
* Do not send etag when serving the frontend index file ([a12c169](a12c169ce88c5cf6711a3239f1687a1dad24a241))
* Lint ([162741e](162741e94064ee199cd5ff021d1ed05f7f5d5ff1))
* Lint ([cc5f48e](cc5f48eb7411f2afa1b0bfb0fc975356b330399a))
* Lint ([ff1730e](ff1730e323b61c8c5ab6f9955bb067bc04e72c8f))
* Clarify preview deployment text and fix typo ([1ffb93b](1ffb93b63c6a57202c7154d09c1db749779b2fbd))
* Lint ([1275dfc](1275dfc260cfd3be98ebed652ef449f182ca42ff))
* Usage of limit and order by usage in recursive cte ([5b70609](5b70609ba760ea68b43f6d42a69a1b32eeb2abec))
* Open external migration service in current tab ([178cd8c](178cd8c3927759a5ca553b3ae76be5ff23e23d83))
* Add root ca to final docker image ([e42a605](e42a605597335507c71ac038f51a775df2775ebd))
* Lint ([6fc3d1e](6fc3d1e98fe28d7e561a4ebe1d00938f8346fae1))
* Lint ([49ab90f](49ab90fc19f9da7d1308c923d6dd99b8a6a355ef))
* Lint ([5e9edef](5e9edef3b36f6a4be5002b8aef4bc02c7649f7b6))
* Lint ([6f51b56](6f51b565895edd75ca26d96c08af26d85ce38f3a))
* Pick first available view if currently configured view got deleted (#2235) ([c4d3d99](c4d3d99cd49aa65d602327abcc5f848d81d6da4e))
* Do not try to fetch nonexistent bucket ([037022e](037022e8570f9b7b0d3053e2b20057b8f5630803))
* Update task in typesense when adding a label or assignee to them ([5213006](521300613f24f2ed585ca7da49a02b58f7d77fb8))
* Lint ([1cd5dd2](1cd5dd2b2fc06731c70721a42ca93966449fa3d2))
* Drop bucket index before recreating it ([ca33c0b](ca33c0b2bcaf9de018cecca1051bc4c3b176ce61))
* Lint ([af3b0bb](af3b0bbea1f31725910011a57bf8db81b8d73e43))
* Lint ([7d755fc](7d755fcb89bbdbbb4f5fe7f329903f1ffba96a29))
* License in cmd help text ([9a16f6f](9a16f6f817157316ce40c8c76f83a8e0d8c0e669))
* Do not push nil errors to sentry ([1460d21](1460d212ee4a0e2baddb297d52d91af69d58c881))
* Correctly return error and bubble up when the api could not be reached ([84197dd](84197dd9c14b7f016bad452f8d529b32f593683c))
* Do not remove empty openid teams when none are present ([66e9632](66e96322eabf009b25a1f7b9c4b2750b9cb47817))
* Use correct project title in project card ([d3a7d79](d3a7d79eb95595f7154b9fbf05e369941189cf5c))
* Ignore casing to check if file extensions can be previewed (#2341) ([81bdad4](81bdad4bebdc8ee19c01d8b44012e89daef6930b))
* Recommended vitest extension (#2351) ([d3d5df5](d3d5df5f62cbb61dd8bb9166500203b212173f28))
* Remove obsolete vscode plugin settings (#2354) ([666eef2](666eef248b5b328f51bda430098d4a6fa625e9db))
* Throw in warnHandler ([81bb49f](81bb49f83aa7878963bb21a436f96d766464188c))
* Use node20 typescript types ([abf912f](abf912f93f86f98855fd141d8c9e4deb447390c1))
* Remove jsxTemplates ([9fa8c54](9fa8c5429b1191e9705caf06d113ec196755e0ae))
* Remove wrong expression ([fe2c390](fe2c3906cabce98fe4dbe1bb7240424fc58e6a05))
* Remove obsolete types ([97a11d2](97a11d2e120f7819be8eed31765ab152fd69da35))
* UseTitle types (#2369) ([9fd17ac](9fd17aca1813087e98b4d5e5a758b7e62482d3d9))
* Remove obsolete vite reactivityTransform option (#2349) ([3718d09](3718d09f3573cfd122ccff92576ee4e03abdd0b6))
* Missing required prop BackgroundColor ([47143af](47143af9d1716dac8b75c29c5c26066a96ddc2e6))
* Use button icon prop ([18e23bf](18e23bf371ef2a6067bde0a976ddc546d0a7d73a))
* Remove uppercase transformation from username (#2445) ([ff5ee51](ff5ee515f9da78b506f8be124b9e803b494df49c))
* Disable vetur in case it's installed ([abdec17](abdec17d366b7dbfc2565a6e354e065256a7451e))
* Import PeriodUnit as type ([baaf612](baaf612239e4e53631c9148f2f3735d8a10ca1b4))
* Reset drag.value ([c90ee01](c90ee0142a959135546d5821f03f4615b5020f07))
* Import type in EditorToolbar ([9f375ec](9f375ecd7d8fb36ac74e4d45af19a271b4272551))
* Remove props prefix from template ([b224b33](b224b331f5df94a9a976bccd18d9a905298b9e54))
* Move types to dev dependencies ([7979884](79798847b2a095c3c89931cb3d0354441bee80d4))
* Typecheck ([142443c](142443c0a757968a9c8b2caeb7fddb6c6bc6dc76))
* Align spelling in config.yml.sample (#2499) ([6d79eb0](6d79eb00885c7f597e681ab2af339cdd4a11b807))

### Dependencies

* *(deps)* Update dependency vue to v3.4.18
* *(deps)* Update module golang.org/x/sys to v0.17.0
* *(deps)* Update module github.com/getsentry/sentry-go to v0.27.0
* *(deps)* Update module xorm.io/xorm to v1.3.8
* *(deps)* Pin dependencies
* *(deps)* Update module golang.org/x/oauth2 to v0.17.0
* *(deps)* Update tiptap to v2.2.2
* *(deps)* Update dev-dependencies to v7
* *(deps)* Update pnpm to v8.15.2
* *(deps)* Update dependency vue to v3.4.19
* *(deps)* Update sentry-javascript monorepo to v7.101.0
* *(deps)* Update module github.com/arran4/golang-ical to v0.2.5
* *(deps)* Update pnpm to v8.15.3
* *(deps)* Update module github.com/arran4/golang-ical to v0.2.6
* *(deps)* Update dependency vue-flatpickr-component to v11.0.4
* *(deps)* Update sentry-javascript monorepo to v7.101.1
* *(deps)* Update dependency @kyvg/vue3-notification to v3.2.0
* *(deps)* Update module github.com/go-testfixtures/testfixtures/v3 to v3.10.0
* *(deps)* Update tiptap to v2.2.3
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.5.0
* *(deps)* Update vueuse to v10.8.0
* *(deps)* Update sentry-javascript monorepo to v7.102.0
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.5.1
* *(deps)* Update dependency electron to v29
* *(deps)* Update dependency dompurify to v3.0.9
* *(deps)* Update dependency vue-router to v4.3.0
* *(deps)* Update tiptap to v2.2.4
* *(deps)* Update pnpm to v8.15.4
* *(deps)* Update sentry-javascript monorepo to v7.102.1
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency vue to v3.4.21
* *(deps)* Update module github.com/prometheus/client_golang to v1.19.0
* *(deps)* Update module golang.org/x/crypto to v0.20.0
* *(deps)* Update sentry-javascript monorepo to v7.103.0
* *(deps)* Update vueuse to v10.9.0
* *(deps)* Update dependency express to v4.18.3
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency vue-tsc to v2
* *(deps)* Update sentry-javascript monorepo to v7.105.0
* *(deps)* Update module github.com/stretchr/testify to v1.9.0
* *(deps)* Update dependency vue-i18n to v9.10.1
* *(deps)* Update module golang.org/x/sys to v0.18.0
* *(deps)* Update module github.com/arran4/golang-ical to v0.2.7
* *(deps)* Update module golang.org/x/term to v0.18.0
* *(deps)* Update module golang.org/x/crypto to v0.21.0
* *(deps)* Update dependency vue-flatpickr-component to v11.0.5
* *(deps)* Update module golang.org/x/oauth2 to v0.18.0
* *(deps)* Update dev-dependencies
* *(deps)* Update src.techknowlogick.com/xgo digest to 770b8ea
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v3
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v3.0.1
* *(deps)* Update dependency electron to v29.1.1
* *(deps)* Update github.com/go-jose/go-jose to 3.0.3
* *(deps)* Update sentry-javascript monorepo to v7.106.0
* *(deps)* Update module github.com/golang-jwt/jwt/v5 to v5.2.1
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency @vue/eslint-config-typescript to v13
* *(deps)* Update module github.com/go-sql-driver/mysql to v1.8.0
* *(deps)* Update dependency happy-dom to v13.7.1
* *(deps)* Go mod tidy
* *(deps)* Update dependency node to v20.11.1
* *(deps)* Sign drone config
* *(deps)* Update golangci/golangci-lint docker tag to v1.56.2 (#2099)
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency date-fns to v3.4.0
* *(deps)* Update sentry-javascript monorepo to v7.106.1
* *(deps)* Pin dependency vuemoji-picker to 0.2.1
* *(deps)* Update dependency happy-dom to v13.8.2
* *(deps)* Update dev-dependencies
* *(deps)* Update google.golang.org/protobuf from 1.32.0 to 1.33.0
* *(deps)* Update dev-dependencies
* *(deps)* Update sentry-javascript monorepo to v7.107.0
* *(deps)* Update dependency axios to v1.6.8
* *(deps)* Update dependency ufo to v1.5.0
* *(deps)* Update dependency vue-i18n to v9.10.2
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency ufo to v1.5.1
* *(deps)* Update dependency date-fns to v3.5.0
* *(deps)* Update module github.com/adlio/trello to v1.11.0
* *(deps)* Update dependency date-fns to v3.6.0
* *(deps)* Update dev-dependencies
* *(deps)* Update pnpm to v8.15.5
* *(deps)* Update module xorm.io/xorm to v1.3.9
* *(deps)* Update dependency happy-dom to v14
* *(deps)* Update dependency ufo to v1.5.2
* *(deps)* Update dependency @kyvg/vue3-notification to v3.2.1
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency dompurify to v3.0.10
* *(deps)* Update module github.com/coreos/go-oidc/v3 to v3.10.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency @infectoone/vue-ganttastic to v2.3.1
* *(deps)* Update dependency express to v4.19.0
* *(deps)* Update dependency ufo to v1.5.3
* *(deps)* Update sentry-javascript monorepo to v7.108.0
* *(deps)* Update dependency dompurify to v3.0.11
* *(deps)* Update dependency express to v4.19.2
* *(deps)* Update module github.com/go-sql-driver/mysql to v1.8.1
* *(deps)* Sign drone config
* *(deps)* Update sentry-javascript monorepo to v7.109.0
* *(deps)* Update dev-dependencies (#2229)
* *(deps)* Update dev-dependencies
* *(deps)* Update src.techknowlogick.com/xgo digest to e01c4fb
* *(deps)* Update golangci/golangci-lint docker tag to v1.57.2 (#2225)
* *(deps)* Update dev-dependencies
* *(deps)* Update pnpm to v8.15.6
* *(deps)* Update dependency @infectoone/vue-ganttastic to v2.3.2
* *(deps)* Update goreleaser/nfpm docker tag to v2.36.1
* *(deps)* Update font awesome to v6.5.2
* *(deps)* Update module github.com/yuin/goldmark to v1.7.1
* *(deps)* Update tiptap to v2.2.6
* *(deps)* Update dependency dompurify to v3.1.0
* *(deps)* Update dependency vue-i18n to v9.11.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency vue-i18n to v9.11.1
* *(deps)* Update tiptap to v2.3.0
* *(deps)* Update dev-dependencies
* *(deps)* Update module github.com/arran4/golang-ical to v0.2.8
* *(deps)* Update github.com/adlio/trello to v1.12.0
* *(deps)* Update sentry-javascript monorepo to v7.110.1
* *(deps)* Update pnpm to v8.15.7
* *(deps)* Update dependency vue to v3.4.23
* *(deps)* Update dependency @intlify/unplugin-vue-i18n to v4
* *(deps)* Update module golang.org/x/sync to v0.7.0 (#2258)
* *(deps)* Update dependency vue-router to v4.3.2
* *(deps)* Update module github.com/tkuchiki/go-timezone to v0.2.3
* *(deps)* Update module golang.org/x/oauth2 to v0.19.0
* *(deps)* Update sentry-javascript monorepo to v7.111.0
* *(deps)* Update module github.com/labstack/echo/v4 to v4.12.0
* *(deps)* Update dependency vue-i18n to v9.13.0
* *(deps)* Update pnpm to v9
* *(deps)* Update dependency node to v20.12.2 (#2238)
* *(deps)* Update dev-dependencies
* *(deps)* Update pnpm to v9.0.4
* *(deps)* Update dev-dependencies
* *(deps)* Update pnpm to v9.0.5
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency vue to v3.4.24
* *(deps)* Update dependency vue-i18n to v9.13.1
* *(deps)* Update github.com/dustinkirkland/golang-petname digest to 76c06c4
* *(deps)* Update dev-dependencies
* *(deps)* Update sentry-javascript monorepo to v7.112.0
* *(deps)* Update sentry-javascript monorepo to v7.112.1
* *(deps)* Update dependency workbox-precaching to v7.1.0
* *(deps)* Update dev-dependencies
* *(deps)* Update sentry-javascript monorepo to v7.112.2
* *(deps)* Update dependency vue to v3.4.25
* *(deps)* Update dependency vitest to v1.5.1
* *(deps)* Update pnpm to v9.0.6
* *(deps)* Update dependency vitest to v1.5.2
* *(deps)* Update dependency dompurify to v3.1.1
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency dayjs to v1.11.11
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency vue to v3.4.26
* *(deps)* Update tiptap to v2.3.1
* *(deps)* Update dependency dompurify to v3.1.2
* *(deps)* Update dev-dependencies
* *(deps)* Update sentry-javascript monorepo to v7.113.0
* *(deps)* Update dependency vite to v5.2.11
* *(deps)* Update dev-dependencies
* *(deps)* Update module golang.org/x/oauth2 to v0.20.0
* *(deps)* Update module golang.org/x/sys to v0.20.0
* *(deps)* Update module golang.org/x/text to v0.15.0
* *(deps)* Update dev-dependencies
* *(deps)* Update module golang.org/x/term to v0.20.0
* *(deps)* Update module golang.org/x/image to v0.16.0
* *(deps)* Update pnpm to v9.1.0
* *(deps)* Update module golang.org/x/crypto to v0.23.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency vue to v3.4.27
* *(deps)* Update dependency node to v20.13.0
* *(deps)* Update dependency go to v1.22.3
* *(deps)* Update dev-dependencies
* *(deps)* Update sentry-javascript monorepo to v7.114.0
* *(deps)* Update tiptap to v2.3.2
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency node to v20.14.0 (#2334)
* *(deps)* Update dependency go to v1.22.4
* *(deps)* Update dependency @fortawesome/vue-fontawesome to v3.0.8
* *(deps)* Update dependency dompurify to v3.1.5
* *(deps)* Update dependency vue-advanced-cropper to v2.8.9
* *(deps)* Update dependency vue-router to v4.3.3
* *(deps)* Update module github.com/gabriel-vasile/mimetype to v1.4.4
* *(deps)* Update module github.com/ganigeorgiev/fexpr to v0.4.1
* *(deps)* Update module github.com/prometheus/client_golang to v1.19.1
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.5.3
* *(deps)* Update pnpm to v9.3.0
* *(deps)* Update dev-dependencies
* *(deps)* Update tiptap to v2.4.0
* *(deps)* Update vueuse to v10.10.1
* *(deps)* Update dependency @sentry/vue to v7.117.0
* *(deps)* Update dependency axios to v1.7.2
* *(deps)* Update module github.com/arran4/golang-ical to v0.3.0
* *(deps)* Update module github.com/hashicorp/go-version to v1.7.0
* *(deps)* Update module github.com/go-testfixtures/testfixtures/v3 to v3.11.0
* *(deps)* Update module golang.org/x/oauth2 to v0.21.0
* *(deps)* Update module golang.org/x/sys to v0.21.0
* *(deps)* Update module github.com/spf13/viper to v1.19.0
* *(deps)* Update module golang.org/x/text to v0.16.0
* *(deps)* Update module golang.org/x/image to v0.17.0
* *(deps)* Update dependency @sentry/vue to v8
* *(deps)* Update module github.com/getsentry/sentry-go to v0.28.1
* *(deps)* Update dependency snake-case to v4
* *(deps)* Update flake
* *(deps)* Update dev-dependencies
* *(deps)* Update module golang.org/x/term to v0.21.0
* *(deps)* Update vueuse to v10.11.0
* *(deps)* Update dependency camel-case to v5
* *(deps)* Update dependency vue to v3.4.29
* *(deps)* Update module golang.org/x/crypto to v0.24.0
* *(deps)* Update module github.com/typesense/typesense-go to v1.1.0
* *(deps)* Update module github.com/yuin/goldmark to v1.7.2
* *(deps)* Update module github.com/spf13/cobra to v1.8.1
* *(deps)* Update pnpm to v9.4.0
* *(deps)* Update dev-dependencies
* *(deps)* Update goreleaser/nfpm docker tag to v2.37.1
* *(deps)* Update golangci-lint to 1.59.1
* *(deps)* Update module github.com/wneessen/go-mail to v0.4.1
* *(deps)* Update dependency @sentry/vue to v8.10.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency @types/node to v20.14.6
* *(deps)* Update dependency node to v20.15.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency vue-router to v4.4.0
* *(deps)* Update dependency @sentry/vue to v8.11.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency vue to v3.4.30
* *(deps)* Update module github.com/yuin/goldmark to v1.7.3
* *(deps)* Update dev-dependencies to v7.14.1
* *(deps)* Update dependency @sentry/vue to v8.12.0
* *(deps)* Update module github.com/yuin/goldmark to v1.7.4
* *(deps)* Update module golang.org/x/image to v0.18.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dependency caniuse-lite to v1.0.30001638
* *(deps)* Update module github.com/wneessen/go-mail to v0.4.2
* *(deps)* Update dependency vue to v3.4.31
* *(deps)* Update dependency @sentry/vue to v8.13.0
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies
* *(deps)* Update dev-dependencies

### Documentation

* *(filter)* Add filter query explanation* Mention installation of rpm packages ([eec53e8](eec53e8a5451143eefc53c861a26e601aed3e636))
* Add missing traefik label ([99856b2](99856b2031b8811ac08acfc773e1df294465cd9e))
* Add desktop packages ([1255bdc](1255bdc4abc57aa2a30eb6ced138a70ea94fcd08))
* Add healthcheck to docker compose examples ([001268a](001268a33eefa7dd2b14f401d6fb7bfbee04c21c))
* Fix healthcheck and mariadb password ([18374c2](18374c2e52a78d48698945981fb59ea3a86eff12))
* Fix database healthcheck command ([89e349f](89e349f2fd3c57674b3645ac682582b52e5c398f))
* Fix postgres example healthcheck ([7ae38c5](7ae38c5ac1e3eaa845753cfcbd2d06fd83418070))
* Remove outdated information ([5ab720d](5ab720d7093d14f1b57bf7317e16131e8ec766c0))
* Clarify public url usage in installation ([e532979](e532979101287eda13a479382432129bba85bf43))
* Mention how to support the project in readme ([4d11dd0](4d11dd0383ef4e4e0dc78d01eb24febee7661f96))
* Update publiccode.yml [skip ci] ([2e5c193](2e5c19352ec5abdf0c7281637066e96243f413ff))
* Update config docs ([09d4467](09d446765d9bb287084322b8a22977fe4355aaca))
* Add missing front matter ([6db8728](6db8728420296f2c76320b3efe71c38e848cefc9))
* Improve OpenID documentation (#2151) ([92d9c31](92d9c31101e5d3bc0d9d726c5075f0ca6b99f2d2))
* Clarify what to use for authurl ([6e52db7](6e52db76dcc020f4492731f1823bf7a94b4cb951))
* Add migrations setup doc (#2183) ([1d5517b](1d5517b53a4649bbe2e6aec8ae07687b0288cfd4))
* Fix broken link in migration docs (#2185) ([4b4a7f3](4b4a7f3c0afa2344e72e695918fdad20d97ddb2d))
* Add details about supported and required OIDC claims (#2201) ([be54a36](be54a361fd680e7bd63309db5338980badea86de))
* Add Korganizer to supported caldav clients ([e0417c8](e0417c8bdaa398f718f1f16522ff512276753201))
* Fix build-from-sources docs mistake (#2251) ([1adaa73](1adaa7314172e6c8dfa239a328c5899d58a842cf))
* Clarify transitioning from unstable to release ([0557d4b](0557d4b5bba1a111ed70a484f88cec1636e29371))
* Clarify automatic openid team creation ([4e49ec9](4e49ec9e16ba5176ce8596e88f96d4a8983922f9))
* Clarify vikunja cli usage in docker ([75fd17c](75fd17c7503b6a1189a3ede1c45be03a69b1fd72))
* Clarify version checkout when building from source ([73bf119](73bf119409acdd474999fe9a513691da61bbe429))
* Fix typo in README.md (#2271) ([aceaccb](aceaccbf117216a5127d88f78add575cd6ed6dcd))
* Clarify where to file issues ([8d1fc08](8d1fc08de628d721234cb22a230e04e3d037fb44))
* Remove superfluous yaml code block (#2342) ([8bc23b3](8bc23b3a54068da2d3cdd1aa7ace0d9e21bb72b4))
* Remove docs files ([7a29011](7a290116e86f73149f22a1443d4867a363c96580))

### Features

* *(FilterInput)* Use expandable
* *(XButton)* Merge script blocks
* *(api)* All usable routes behind authentication now have permissions
* *(api)* Add bulk endpoints to api tokens
* *(api tokens)* Add task attachment to api scopes
* *(auth)* Update team name in Vikunja when it was changed in the openid provider
* *(backgrounds)* Resize images to a maximum of 4K
* *(ci)* Rename unstable desktop packages
* *(ci)* Automatically create a gitea release when tagging
* *(components)* Align component name casing datemath (#2405)
* *(components)* Align component name casing ContentAuth
* *(components)* Align component name casing multiselect
* *(components)* Align component name casing apiconfig
* *(components)* Align component name casing CreateEdit
* *(components)* Align component name casing password
* *(components)* Align component name casing card
* *(components)* Align component name casing color-bubble
* *(components)* Align component name casing Error
* *(components)* Align component name casing Pagination
* *(components)* Align component name casing Dropdown
* *(components)* Align component name casing ContentLinkShare
* *(components)* Align component name casing Navigation
* *(components)* Align component name casing fancycheckbox
* *(components)* Align component name casing Legal
* *(components)* Align component name casing Loading
* *(components)* Align component name casing Message
* *(components)* Align component name casing NoAuthWrapper
* *(components)* Align component name casing Nothing
* *(components)* Align component name casing Button
* *(components)* Align component name casing Modal
* *(components)* Align component name casing Notification
* *(components)* Align component name casing Popup
* *(components)* Align component name casing Ready
* *(components)* Align component name casing Shortcut
* *(components)* Align component name casing Subscription
* *(components)* Align component name casing User
* *(components)* Align component name casing Notifications
* *(components)* Align component name casing FilterPopup
* *(components)* Align component name casing Filter
* *(components)* Align component name casing ViewEditForm
* *(components)* Align component name casing ProjectSettingsDropdown
* *(components)* Align component name casing QuickActions
* *(components)* Align component name casing LinkSharing
* *(components)* Align component name casing UserTeam
* *(components)* Align component name casing AssigneeList
* *(components)* Align component name casing Attachments
* *(components)* Align component name casing ChecklistSummary
* *(components)* Align component name casing Comments
* *(components)* Align component name casing CreatedUpdated
* *(components)* Align component name casing DateTableCell
* *(components)* Align component name casing DeferTask
* *(components)* Align component name casing Description
* *(components)* Align component name casing EditAssignees
* *(components)* Align component name casing EditLabels
* *(components)* Align component name casing FilePreview
* *(components)* Align component name casing Heading
* *(components)* Align component name casing KanbanCard
* *(components)* Align component name casing Label
* *(components)* Align component name casing Labels
* *(components)* Align component name casing PercentDoneSelect
* *(components)* Align component name casing PriorityLabel
* *(components)* Align component name casing PrioritySelect
* *(components)* Align component name casing ProjectSearch
* *(components)* Align component name casing QuickAddMagic
* *(components)* Align component name casing RelatedTasks
* *(components)* Align component name casing ReminderDetail
* *(components)* Align component name casing ReminderPeriod
* *(components)* Align component name casing Reminders
* *(components)* Align component name casing RepeatAfter
* *(components)* Align component name casing SingleTaskInlineReadonly
* *(components)* Align component name casing SingleTaskInProject
* *(components)* Align component name casing AddTask
* *(components)* Align component name casing ProjectSettings
* *(docker)* Use scratch as base image
	- **BREAKING**: use scratch as base image
* *(editor)* Checklist visual improvements (#2264)
* *(editor)* Add hotkeys to quickly edit and discard (#2265)
* *(filter)* More tests
* *(filter)* Nesting
* *(filter)* Migrate existing saved filters
* *(filter)* Add better error message when passing an invalid filter expression
* *(filter)* Add in keyword
* *(filter)* Add basic highlighting filter query component
* *(filter)* Add auto resize for filter query input
* *(filter)* Add autocompletion poc for labels
* *(filter)* Make the autocomplete look pretty
* *(filter)* Add actual label search when autocompleting
* *(filter)* Autocomplete for assignees
* *(filter)* Autocomplete for projects
* *(filter)* Emit filter query
* *(filter)* Remove now unused code
* *(filter)* Add button to show filter results
* *(filter)* Resolve labels and projects to ids before filtering
* *(filter)* Resolve label and project ids back to titles when loading a filter
* *(filter)* Fall back to simple search when filter query does not contain any filter inputs
* *(filter)* Make filter input label configurable
* *(filter)* Add unique id to filter input
* *(filters)* Very basic filter parsing
* *(filters)* Basic text filter works now
* *(filters)* Make new filter syntax work with Typesense
* *(filters)* Parse date properties to enable datepicker button
* *(filters)* Make date values in filter query editable
* *(filters)* Add date values
* *(filters)* Show user name and avatar for assignee filters
* *(filters)* Add basic autocomplete component
* *(filters)* Highlight label colors in filter
* *(filters)* Query-based filter logic (#2177)
* *(filters)* Pass timezone down when filtering with relative date math
* *(filters)* Make clear filters button less obvious
* *(i18n)* Add pt-br as selectable language in the frontend
* *(i18n)* Add Croatian to selectable languages
* *(i18n)* Add Ukrainian for language selection in UI
* *(kanban)* Debounce bucket limit setting
* *(kanban)* Do not remove focus from the input after creating a new bucket
* *(kanban)* Set task position to 0 (top) when it is moved into the done bucket automatically after marking it done
* *(migration)* Notify the user when a migration failed
* *(migration)* Trello organization based migration (#2211)
* *(migration)* Include non upload attachments from Trello (#2261)
* *(navigation)* Persist project open state in navigation
* *(registration)* Improve username and password validation
* *(subscription)* Use a recursive cte to fetch subscriptions of parent projects
* *(task)* Show attachment preview for image attachments (#2266)
* *(tasks)* Make done at column available for selection in table view
* *(tasks)* Expand subtasks (#2345)
* *(tasks)* Add tests for moving a task out of the done bucket
* *(teams)* Add public flags to teams to allow easier sharing with other teams (#2179)
* *(typesense)* Move partial reindex to a flag instead of a separate command
* *(views)* Add new model and migration
* *(views)* Add crud handlers and routes for views
* *(views)* Add new default views for filters
* *(views)* Return views with their projects
* *(views)* Create default 4 default view for projects
* *(views)* Return tasks in a view
* *(views)* Create default views when creating a filter
* *(views)* Do not override filters in view
* *(views)* Use project id when fetching views
* *(views)* Add bucket configuration mode
* *(views)* (un)marshal custom project view mode types
* *(views)* Return tasks in buckets by view
	- **BREAKING**: tasks in their bucket are now only retrievable via their view. The /project/:id/buckets endpoint now only returns the buckets for that project, which is more in line with the other endpoints
* *(views)* Move task position handling to its own crud entity
	- **BREAKING**: the position of tasks now can't be updated anymore via the task update endpoint. Instead, there is a new endpoint which takes the project view into account as well.
* *(views)* Sort tasks by their position relative to the view they're in
* *(views)* Decouple buckets from projects
* *(views)* Decouple bucket CRUD from projects
	- **BREAKING**: decouple bucket CRUD from projects
* *(views)* Move done and default bucket setting to view
	- **BREAKING**: move done and default bucket setting to view
* *(views)* Decouple bucket <-> task relationship
	- **BREAKING**: decouple bucket <-> task relationship
* *(views)* Make updating a bucket work again
	- **BREAKING**: make updating a bucket work again
* *(views)* Only update the bucket when necessary
* *(views)* Recalculate all positions when updating
* *(views)* Set default position
* *(views)* Save position in Typesense
* *(views)* Save view and position in Typesense
* *(views)* Sort by position
* *(views)* Fetch tasks via view context when accessing them through views
* *(views)* Generate swagger docs
* *(views)* Save task position
* *(views)* Return position when retrieving tasks
* *(views)* Save task position in list view
* *(views)* Load views when navigating with link share
* *(views)* Create task bucket relation when creating a new bucket
* *(views)* Show tasks on kanban board in saved filter
* *(views)* Crud in frontend
* *(views)* Hide view switcher when there is only one view
* *(views)* Lint
* *(views)* Allow reordering views
* *(views)* Add filter syntax docs to filter input in views* Allow using sqlite in memory database ([2dab2cc](2dab2ccedde96b2363e69ed14d026922c8883705))
* *(other)* Enter edit mode when double clicking
* Run frontend tests with api build from the same branch (#2137) ([5d127c2](5d127c28973fa58dfd97db055dcd215c4c9e30ed))
* Fetch all projects with a recursive cte instead of recursive query ([6b1e674](6b1e67485bda84e9229fc57bac3782aa598240ef))
* Assign users to teams via OIDC claims (#1393) ([ed4da96](ed4da96ab15fe11ced9383f7e7a25329207472ab))
* Nest api token permissions under their parents ([67f5551](67f55510bf70afbd0c82477004428549dfc35df9))
* Emoji reactions for tasks and comments (#2196) ([a5c51d4](a5c51d4b1ebf0a6bde33c0004c00eca5e0321038))
* Decouple views from projects (#2217) ([7230db1](7230db160355c6b67c3586bf7bf6da57444c76cb))
* New login image ([2d084c0](2d084c091ef759964ad31b19fc4bc7ac17b12d60))
* Do not save language on the server when in demo mode ([e1dcf2e](e1dcf2e8591c3a7482ba35b243ef1b2c88505420))
* Default view setting (#2306) ([aac01c7](aac01c7a35836421c17882b4f77334fc14bfeaec))
* Add pluralization rules for Russian (#2344) ([73780e4](73780e4b5007d3dfbd3b4f92d9cb1c38d603fe27))
* Update pnpm (#2355) ([50cf952](50cf952b011d97c792fe296b4b54888c13555e2e))
* Remove polyfills ([19a7605](19a760506cdd19eb465a202d2d4cb149e0ef4da7))
* Update packages (#2367) ([0523350](0523350f395067ca26ebd4cd920ec9e12d10f53b))
* Migrate to unplugin-inject-preload (#2357) ([50d6987](50d698794b1f4f9c55d7f9f00a80a87fe56ae400))
* Improve types (#2368) ([bc897a4](bc897a45037e1834adefba56bab229cf03238f57))
* Reduce eslint warnings (#2396) ([2004d12](2004d129c39b9d84abf79806822a1dbfb451eca5))
* Align sort icon color ([0061ec0](0061ec03f57b10ec83970c46f79d8e9b247e4d1a))
* Improve shortcut types ([6c113ea](6c113eaca1add681b8bcad74dbb4a537f29458e8))
* Improve popup ([92f2e0e](92f2e0e214e7d611820354e107ba6411e549d959))
* Improve user component ([fe21a2c](fe21a2c3daf193ac0381d6c84dd271cac835a8e2))
* Use withDefaults and defineOptions in Modal ([b1a8bbe](b1a8bbe760026d135baf8a84b054bfdde4381660))
* Improve subscription ([341b8d2](341b8d20450a339fc31f916e565084941821eb61))
* Eslint enforce vue/component-name-in-template-casing ([23707fc](23707fc493b0a335b0ddb4d3737b9be67fc0242d))
* Switch to change-case lib ([1268145](1268145f713dcb8b94bde5462382366c3913a623))
* Use withDefaults in CreateEdit ([5e4b9e3](5e4b9e38a64a73b44e9d4c5dc80931898b39e63f))
* Add default to custom transition ([1977a7b](1977a7bee0837b3c3c8b89055cebb255fcc16708))
* Camelcase global components ([f361158](f36115871c52ed6a9b733df06f530968cea94251))
* Add root tsconfig ([4546bd6](4546bd6986e360edc8eb222ad7ed28a8e8e58d5d))
* Set add tsconfigRootDir option ([9b43c13](9b43c130616ac1e08fc74bdda4b9d6e1e377c15d))
* Improve message types ([4c5bb3f](4c5bb3f114084b2ab14ef06da1893e5d1f92b4e2))
* Improve gantt filter label ([66be016](66be016a7f3737e6b8130a18464950cb38b96727))
* Use withDefaults for BaseCheckbox ([94a907b](94a907b009b2b78c6a1942380ab7f4e3cf9090a5))
* Remove props destructuring ProjectsNavigationItem ([4bd9c79](4bd9c79912dea02a2766134ad763e48106784e7d))
* Use defineOptions for Loadings.vue ([ff2644d](ff2644d1c516c31b3a90c95f3b71d25394ae2c13))
* Use withDefaults for ProjectSearch ([bd32f7a](bd32f7aef58fddedd7cffb864719c51038f94ea3))
* Remove props destructuring from SingleTaskInlineReadonly ([7c9f0b8](7c9f0b8ada6aeb3a06e5a7a6ccc754cffbe13f0b))
* Use withDefaults for Card ([5b0ce4e](5b0ce4e01c98feff5636a0fbad1b5c6457b31493))
* Use defineProps types for ProjectSettingsDropdown ([9e266f1](9e266f1e36dc1dab7ecab684f15177378bbb888b))
* Use defineProps types for ChecklistSummary ([1dbd8b6](1dbd8b6c3748fe4684f8ed716fb5e51f47e0108c))
* Use withDefaults for Labels ([dea0510](dea051010d724614a7f5568d0645397ffc3f431f))
* Use defineProps types in CreatedUpdated ([c81649c](c81649c139b310e31206d62aef7c2c24d7dab788))
* Use withDefaults for Done ([01a4ad9](01a4ad99ab5234530fdf15f8982f201f81992fad))
* Add getter support to useProjectBackground ([914fe09](914fe092e5858fdaaa7002b70b6e91f7c0143be3))
* Improve ProjectSettingsEdit reactivity ([fb449d7](fb449d7b29de9b5281368beab6f12d57b7f74901))
* Remove props destructuring EditorToolbar ([516f507](516f507ac42ec831189cd5b367ccd22c676401ae))
* Remove props destructuring from ProjectCard ([8a2c74a](8a2c74a702492d496d331f3bb8f9e2e1161667d4))
* Use withDefaults for AddTask ([7db9e64](7db9e64053d27de408a307637986e41ed6359508))


### Miscellaneous Tasks

* *(auth)* Refactor openid team creation
* *(auth)* Add oidc suffix to openid team name in db
* *(auth)* Refactor removing empty openid teams to cron job
* *(auth)* Show registration disabled message when registration is disabled
* *(desktop)* Switch from yarn to pnpm
* *(desktop)* Only build zip in ci to speed up smoke test builds
* *(dev)* Move nix flake to top level, add api tooling
* *(filter)* Cleanup
* *(filters)* Cleanup old variables
* *(filters)* Add histoire story file
* *(filters)* Copy datepicker
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Remove "new" from creation strings
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(i18n)* Update translations via Crowdin
* *(magefile)* Add aliases for lint
* *(mail)* Update logger to new interface
* *(views)* Add fixme
* *(views)* Remove old view routes
* *(views)* Use view id instead of passing whole view object
* *(views)* Move actual project views into their own folder* Don't send http errors with a status < 500 to sentry ([d4a3892](d4a389279c24dfdd860a0a90af5778bde64cb66c))
* Add publiccode.yml ([837360f](837360f1226bd63234eb8b0a51465ff68c180863))
* Remove unused docker entrypoint script ([f18cde2](f18cde269b8b2beaa99645517bf18b7ec550a62f))
* Update lockfile ([3548709](35487093c6ac7372f1a37195a0233f0d1b9c5016))
* Format ([356399f](356399f8539da43cedf085a7917a40705f9d2d09))
* Generate swagger docs ([32e1a20](32e1a2018a9607e13efbd55a91b463e041c65520))
* Cleanup leftover console.log ([cf6b476](cf6b476b7d750265026fde462128eb65a8cbff78))
* Remove old saved views migration ([337d289](337d289a396cea592e25f6118fea4e91df53e0cc))
* Do not import message dynamically ([6ad83c0](6ad83c06858bf7c1e115d82dd69f538418026738))
* UseDefineOptions for inheritAttrs in Error.vue ([2d358a5](2d358a57cc23ed0f188fe418de4cdc4a4fc4ecc7))

### Other

* *(other)* Cancel current edits and exit edit mode with escape
* *(other)* Add discardShortcutEnabled setting to opt into this feature
* *(other)* Use if to conditionally add escape hotkey
* *(other)* Merge import
* *(other)* Rename discardShortcutEnabled to enableDiscardShortcut
* *(other)* Rename discardShortcutEnabled to enableDiscardShortcut
* *(other)* Proof of concept for image preview
* *(other)* Extract img to FilePreview component
* *(other)* Replace table with grid
* *(other)* Adjust file preview style
* *(other)* Replace px with rem
* *(other)* Move object-fit to styles
* *(other)* Rename grid-item class to attachment
* *(other)* Attempt to fix attachment verification
* *(other)* Change attachment div to button
* *(other)* Fix test again
* *(other)* Add width: 100%
* *(other)* File name and cover label styling improvement
* *(other)* Add file image as fallback preview
* *(other)* Replace cover text links with icons
* *(other)* Only allow cover if a preview is available
* *(other)* Make fallback icon grey
* *(other)* Use file fallback icon
* *(other)* Add cover tooltips
* *(other)* Improve preview spacing
* *(other)* Set attachment width to 100%
* *(other)* [skip ci] Updated swagger docs

## [0.23.0] - 2024-02-10

### Bug Fixes

* *(assignees)* Use correct amount of spacing in assignee selection
* *(ci)* cd to frontend in frontend pipelines
* *(ci)* Deploy packages into the correct directory
* *(ci)* Swagger docs generate should use the correct url
* *(ci)* Typo
* *(ci)* Update shasum
* *(docs)* Old install pages redirect
* *(editor)* Don't set editor content initially
* *(export)* Don't crash when an exported file does not exist
* *(filters)* Add explicit check for string slice filter
* *(gantt)* Correctly import languages from dayjs
* *(kanban)* Assignee spacing
* *(kanban)* Bottom spacing of labels
* *(notifications)* Mark all notifications as read in ui directly when marking as read on the server
* *(progress)* Cleanup unused css
* *(progress)* Less rounding
* *(reminders)* Set reminder date on datepicker when editing a reminder
* *(task)* Make sure the drag handle is shown as intended
* *(task)* Move cover image setter to store
* *(task)* Remove default task color
* *(tasks)* Check for cycles during creation of task relations and prevent them
* *(tasks)* Show any errors happening during task load
* *(tests)* Adjust gantt rows identifier
* *(webhook)* Fetch all event details before sending the webhook

### Features
 
* Merge API, Frontend and Desktop repos
* *(ci)* Combine api and frontend drone configs
* *(ci)* Merge desktop ci config
* *(ci)* Save .tags file to generate release tags
* *(ci)* Run desktop build without waiting on the frontend when not doing release builds
* *(ci)* Run desktop pipeline only on PRs
* *(editor)* Use primary color for currently selected node
* *(filters)* Log type if unknown filter type
* *(progress)* Move customizations into progress bar component

### Dependencies

* *(deps)* Update dependency @4tw/cypress-drag-drop to v1.8.1 (#693)
* *(deps)* Update dependency @fortawesome/vue-fontawesome to v3.0.6
* *(deps)* Update dependency @kyvg/vue3-notification to v3.1.4
* *(deps)* Update dependency @types/node to v20.11.10
* *(deps)* Update dependency autoprefixer to v10.3.3 (#684)
* *(deps)* Update dependency autoprefixer to v10.3.4 (#697)
* *(deps)* Update dependency axios to v0.21.2 (#698)
* *(deps)* Update dependency axios to v0.21.3 (#700)
* *(deps)* Update dependency cypress to v8.3.1 (#689)
* *(deps)* Update dependency electron to v28.2.1 (#186)
* *(deps)* Update dependency electron to v28.2.2 (#187)
* *(deps)* Update dependency esbuild to v0.12.23 (#683)
* *(deps)* Update dependency esbuild to v0.12.24 (#688)
* *(deps)* Update dependency esbuild to v0.12.25 (#696)
* *(deps)* Update dependency esbuild to v0.14.53 (#2217)
* *(deps)* Update dependency eslint-plugin-vue to v7.17.0 (#686)
* *(deps)* Update dependency floating-vue to v5.2.1
* *(deps)* Update dependency floating-vue to v5.2.2
* *(deps)* Update dependency jest to v27.1.0 (#687)
* *(deps)* Update dependency marked to v3.0.1 (#677)
* *(deps)* Update dependency marked to v3.0.2 (#682)
* *(deps)* Update dependency postcss to v8.4.19 (#2673)
* *(deps)* Update dependency sass to v1.38.1 (#679)
* *(deps)* Update dependency sass to v1.38.2 (#690)
* *(deps)* Update dependency sass to v1.39.0 (#695)
* *(deps)* Update dependency typescript to v4.4.2 (#685)
* *(deps)* Update dependency ufo to v1.4.0
* *(deps)* Update dependency vite to v2.5.1 (#680)
* *(deps)* Update dependency vite to v2.5.2 (#692)
* *(deps)* Update dependency vite to v2.5.3 (#694)
* *(deps)* Update dependency vite-plugin-pwa to v0.11.2 (#681)
* *(deps)* Update dependency vue to v3.2.45
* *(deps)* Update dependency vue-i18n to v9.9.1
* *(deps)* Update goreleaser/nfpm docker tag to v2.35.3 (#1692)
* *(deps)* Update module github.com/arran4/golang-ical to v0.2.4
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.21
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.22
* *(deps)* Update module github.com/swaggo/swag to v1.16.3
* *(deps)* Update module github.com/yuin/goldmark to v1.7.0
* *(deps)* Update pnpm to v8.15.0
* *(deps)* Update pnpm to v8.15.1
* *(deps)* Update sentry-javascript monorepo to v7.100.1
* *(deps)* Update sentry-javascript monorepo to v7.17.2 (#2587)
* *(deps)* Update sentry-javascript monorepo to v7.19.0 (#2670)
* *(deps)* Update sentry-javascript monorepo to v7.99.0
* *(deps)* Update src.techknowlogick.com/xgo digest to 45b9ea6
* *(deps)* Update src.techknowlogick.com/xgo digest to 5aae655
* *(deps)* Update tiptap to v2.2.0
* *(deps)* Update tiptap to v2.2.1
* *(deps)* Update typescript-eslint monorepo to v4.29.3 (#676)
* *(deps)* Update typescript-eslint monorepo to v4.30.0 (#691)

### Miscellaneous Tasks

* *(Expandable)* Spelling ⛈
* *(deps)* Move renovate config
* *(deps)* Remove redundant renovate config
* *(quick actions)* Format

## [0.22.1] - 2024-01-28

### Bug Fixes

* *(api)* Make sure permission to read all tasks work for reading all tasks per project
* *(assignees)* Improve wording for assignee emails
* *(assignees)* Prevent double notifications for assignees
* *(assignees)* Subscribe assigned users directly to the task, not async
* *(assignees)* Make sure task assignee created event contains the full task
* *(auth)* Don't reset user settings when updating name or email from external auth provider
* *(migration)* Ignore tasks with empty titles
* *(openid)* Use the calculated redirect url when authenticating with openid providers
* *(projects)* Don't remove parent project id if the parent project is available in the same run
* *(relations)* Don't allow creating relations which already exist
* *(subscriptions)* Don't crash when a project is already deleted
* *(task)* Delete the task after all related attributes to prevent task not found errors
* *(typesense)* Update tasks in Typesense directly when the change happened
* *(user)* Make disable command actually work
* *(webhooks)* Make sure all events with tasks have the full task* Create webhooks table for fresh installation ([09696ae](09696aec1bea647a5bfc7be16b31054626d721e4))
* Lint ([2c84688](2c84688a4013a816eca02caabba8c634a03d3d57))
* Convert everything which looks like an url to a <a href html element ([27a5f68](27a5f6862b1748ec10ca9282e0fe1a64f9ccf910))
* Update function signatures ([4d48d81](4d48d814c95244f21454219c1004b6298744e076))
* Tests ([1630e4f](1630e4fc08bc5fccff191a6cc4afe936543635d8))
* Lint ([30a2dcd](30a2dcd04c8379291a2ae5068ec0cab07bc9a7fb))

### Dependencies

* *(deps)* Update dessant/repo-lockdown action to v4
* *(deps)* Update alpine docker tag to v3.19
* *(deps)* Update module github.com/arran4/golang-ical to v0.2.3 (#1669)
* *(deps)* Update module github.com/labstack/gommon to v0.4.2
* *(deps)* Update module xorm.io/xorm to v1.3.6
* *(deps)* Update module golang.org/x/term to v0.16.0
* *(deps)* Update module golang.org/x/image to v0.15.0
* *(deps)* Update module github.com/prometheus/client_golang to v1.18.0
* *(deps)* Update module github.com/typesense/typesense-go to v1
* *(deps)* Update module golang.org/x/oauth2 to v0.16.0
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.4.0
* *(deps)* Update module github.com/getsentry/sentry-go to v0.26.0
* *(deps)* Update goreleaser/nfpm docker tag to v2.35.2
* *(deps)* Update module github.com/labstack/echo/v4 to v4.11.4
* *(deps)* Update module golang.org/x/sync to v0.6.0
* *(deps)* Update module xorm.io/xorm to v1.3.7
* *(deps)* Update module github.com/google/uuid to v1.6.0
* *(deps)* Update src.techknowlogick.com/xgo digest to 77ac23f
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.20

### Features

* *(reminders)* Persist reminders in the db

### Miscellaneous Tasks

* Check if import zip contains a VERSION file ([ec6e3e9](ec6e3e99e0d6f2d8a9c889c7261e0d16b4ebea7d))
* Rename function ([0d24ba1](0d24ba12bb85078afd8c821bae61926fd81f163e))

## [0.22.0] - 2023-12-19

### Bug Fixes

* *(api tokens)* Make sure read one routes show up in routes endpoint
* *(api tokens)* Test
* *(api tokens)* Lint
* *(api tokens)* Make sure task create routes are available to use with the api
  - **BREAKING**: The api route to create a new task is now /projects/:project/tasks instead of /projects/:project
* *(build)* Don't run go mod commands when generating swagger docs
* *(build)* Don't generate swagger files when building
* *(build)* Don't require swagger to build
* *(build)* Don't remove swagger files when running build:clean step
* *(caldav)* Check for related tasks synced back from a caldav client
* *(caldav)* Do not update dates of tasks when repositioning them (#1605)
* *(ci)* Don't generate swagger docs in ci
* *(ci)* Use the same go image for everything
* *(ci)* Don't try to install when linting
* *(cmd)* Do not initialize async operations when running certain cli commands
* *(comments)* Make sure comment sort order is stable
* *(docs)* Add empty swagger file so that the package exists
* *(docs)* Remove duplicate paths (params) in swagger docs
* *(files)* Keyvalue init in tests
* *(filter)* Assignee search by partial username test
* *(filters)* Make "in" filter comparator work with Typesense
* *(import)* Don't fail when importing from dev exports
* *(import)* Ignore duplicate project identifier
* *(import)* Resolve task relations by old task ids
* *(import)* Correctly set child project relations
* *(import)* Create related tasks without an id
* *(import)* Make sure importing works if parent / child projects are created in a different order
* *(kanban)* Don't prevent setting a different bucket as done bucket
* *(kanban)* Create missing kanban buckets (#1601)
* *(kanban)* Filter for tasks in buckets by assignee should not modify the filter directly
* *(labels)* Make sure labels of shared sub projects are usable
* *(migration)* Use string for todoist project note id
* *(migration)* Make sub project hierarchy work when importing from other services
* *(openid)* Make sure usernames with spaces work
* *(project)* Duplicating a project should not create two backlog buckets
* *(project background)* Add more checks for whether a background file exists when duplicating or deleting a project
* *(projects)* Save done and default bucket when updating project
* *(projects)* Don't limit results to top-level projects when searching
* *(projects)* Don't return child projects multiple times
* *(projects)* Correctly set project's archived state if their parent was archived
* *(projects)* Delete child projects when deleting a project
* *(reminders)* Make sure reminders are only sent once per user
* *(swagger)* Add generated swagger docs to repo
* *(task)* Remove task relation in the other direction as well
* *(test)* Don't check for error
* *(tests)* Use string IDs in Todoist test
* *(tests)* Remove duplicate projects from assertions
* *(tests)* Pass the map
* *(typesense)* Upsert one document at a time
* *(typesense)* Add more error logging
* *(typesense)* Add more error logging
* *(typesense)* Pass the correct user when fetching task comments
* *(typesense)* Upsert all documents at once
* *(typesense)* Explicitly create typesense sync table
* *(typesense)* Don't try to index tasks if there are none
* *(typesense)* Add typesense sync to initial structs
* *(typesense)* Make sure searching works when no task has a comment at index time
* *(typesense)* Getting all data from typesense
* *(typesense)* Correctly convert date values for typesense
* *(user)* Don't crash when attempting to change a user's password
* *(user)* Allow deleting a user if they have a default project
* *(user)* Don't prevent deleting a user if their default project was shared
* *(user)* Allow openid users to request their deletion
* *(webhooks)* Routes should use the common schema used for other routes already
* *(webhooks)* Don't send the proxy auth header to the webhook target
* *(webhooks)* Lint
* *(webhooks)* Lint
* *(webhooks)* Add created by user object when creating a webhook
* *(webhooks)* Send application/json header* Typo ([49d8713](49d87133885b4fa660c300fc38768bd91f56340e))
* Lint ([29317b9](29317b980e68b7e10b127e7e93afff1dd56ace3e))
* Order by clause in task comments ([5811d2a](5811d2a13b5a1017cdd0b393599ffe01db95e836))
* Lint ([e4c7112](e4c71123ef91480d41284288bee38939cd17ae39))
* Validate usernames on registration ([11810c9](11810c9b3e1a4bb4c5fc1f4a3ac44e8552f6a937))
* Lint ([d6db498](d6db49885383ed3e4f98acf649dc302ed1411ccd))
* Lint ([b8e73f4](b8e73f4fa5821ce07b42667cf84c1ff9b87e0888))
* Lint ([424bf76](424bf7647baa34e0fa594c2c36eec542ebea531b))
* Lint ([e34f503](e34f503674c2aab06c7215cba9e2133037e96b6a))
* Lint ([56625b0](56625b0b90d659bd49fc95749691d0100e964dcd))
* Properly tag bucket-related operations ([a375223](a3752238729d50b38a5cf0b811e050c3d9f8985f))
* Lint ([6ef1bc3](6ef1bc3944980588238fb44295b520695a4ed19a))


### Dependencies

* *(deps)* Update module github.com/wneessen/go-mail to v0.4.0
* *(deps)* Update src.techknowlogick.com/xgo digest to 617d3b6
* *(deps)* Update module github.com/iancoleman/strcase to v0.3.0
* *(deps)* Update module github.com/labstack/echo/v4 to v4.11.0
* *(deps)* Update module github.com/labstack/echo/v4 to v4.11.1
* *(deps)* Update module xorm.io/builder to v0.3.13
* *(deps)* Update module golang.org/x/image to v0.11.0
* *(deps)* Update module github.com/getsentry/sentry-go to v0.23.0
* *(deps)* Update module github.com/arran4/golang-ical to v0.1.0
* *(deps)* Update src.techknowlogick.com/xgo digest to 1510ee0
* *(deps)* Update module github.com/yuin/goldmark to v1.5.6
* *(deps)* Update module xorm.io/xorm to v1.3.3
* *(deps)* Update module github.com/jinzhu/copier to v0.4.0
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.2.1
* *(deps)* Update module github.com/threedotslabs/watermill to v1.3.5
* *(deps)* Update module golang.org/x/oauth2 to v0.13.0
* *(deps)* Update lockfile
* *(deps)* Update lockfile
* *(deps)* Update github.com/dustinkirkland/golang-petname digest to 6a283f1
* *(deps)* Update module github.com/prometheus/client_golang to v1.17.0
* *(deps)* Update src.techknowlogick.com/xgo digest to 6fc6b16
* *(deps)* Update module github.com/getsentry/sentry-go to v0.25.0
* *(deps)* Update lockfile
* *(deps)* Update module github.com/spf13/viper to v1.17.0
* *(deps)* Update module github.com/spf13/afero to v1.10.0
* *(deps)* Update lockfile
* *(deps)* Update module github.com/swaggo/swag to v1.16.2
* *(deps)* Update module golang.org/x/image to v0.13.0
* *(deps)* Update module golang.org/x/sync to v0.4.0
* *(deps)* Update module github.com/labstack/echo/v4 to v4.11.2
* *(deps)* Update lockfile
* *(deps)* Update postgres docker tag to v16 (#1618)
* *(deps)* Update goreleaser/nfpm docker tag to v2.33.1 (#1560)
* *(deps)* Update mariadb docker tag to v11 (#1544)
* *(deps)* Update xgo to go 1.21
* *(deps)* Update module github.com/gabriel-vasile/mimetype to v1.4.3
* *(deps)* Update lockfile
* *(deps)* Update module github.com/coreos/go-oidc/v3 to v3.7.0
* *(deps)* Update src.techknowlogick.com/xgo digest to ecfba3d
* *(deps)* Update lockfile
* *(deps)* Update module src.techknowlogick.com/xormigrate to v1.6.0 (#1627)
* *(deps)* Update module github.com/google/uuid to v1.4.0
* *(deps)* Update module src.techknowlogick.com/xormigrate to v1.7.0
* *(deps)* Update lockfile
* *(deps)* Update module xorm.io/xorm to v1.3.4 (#1630)
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.3.0
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.18
* *(deps)* Update module golang.org/x/sync to v0.5.0
* *(deps)* Update module golang.org/x/sys to v0.14.0
* *(deps)* Update module github.com/spf13/cobra to v1.8.0
* *(deps)* Update module src.techknowlogick.com/xormigrate to v1.7.1
* *(deps)* Update module github.com/yuin/goldmark to v1.6.0
* *(deps)* Update lockfile
* *(deps)* Update github.com/go-jose/go-jose/v3
* *(deps)* Update module golang.org/x/text to v0.14.0
* *(deps)* Update module golang.org/x/oauth2 to v0.15.0
* *(deps)* Update module golang.org/x/crypto to v0.17.0
* *(deps)* Update module golang.org/x/term to v0.15.0
* *(deps)* Update module golang.org/x/image to v0.14.0
* *(deps)* Update module github.com/golang-jwt/jwt/v5 to v5.2.0
* *(deps)* Update src.techknowlogick.com/xgo digest to c7ed783
* *(deps)* Update module github.com/labstack/echo/v4 to v4.11.3
* *(deps)* Update golangci/golangci-lint docker tag to v1.55.2
* *(deps)* Update goreleaser/nfpm docker tag to v2.34.0
* *(deps)* Update lockfile
* *(deps)* Update golangci-lint rules
* *(deps)* Update sqlite bindings
* *(deps)* Update deps

### Documentation

* *(webhooks)* Add general docs about webhooks
* *(webhooks)* Add swagger docs for all webhook endpoints* Add Caddyfile to reverse proxies setup (#1580) ([665c046](665c04671739fd08e5b24e59749707ce5de83daa))
* *(webhooks)* Add webhook config to sample config
* Add Authentik example config (#1660) ([4615b4d](4615b4dbfbbf8514d9c41176e6e68a8ba3a453ce))
* Add config guide for NGINX Proxy Manager ([a1d0541](a1d0541a7a6926127ba0bac4df03ce62b74f0c84))
* Add n8n docs ([6a7aec2](6a7aec2e9ded619b074ef27f360c96c313e4449c))
* Add typesense setup ([70d1903](70d1903dcac67e33bdfdf54d0ba561af76dbf927))
* Clarify minimum required go version ([a2925cf](a2925cf55bee4c71ac5be1bad66cb3ec2230056d))
* Clarify required language code ([e1525fc](e1525fca6eb5af17afa332d2c76a37b288673c5b))
* Fix typo ([db0153a](db0153a7213a9b0bbafb43bc2762e2060f1ec9d1))


### Features

* *(api tokens)* Add api token struct and migration
* *(api tokens)* Add crud routes to manage api tokens
* *(api tokens)* Add tests
* *(api tokens)* Better error message for invalid tokens
* *(api tokens)* Check for expiry date
* *(api tokens)* Check for scopes
* *(api tokens)* Check if a provided token matched a hashed on in the database
* *(api tokens)* Check permissions when saving
* *(api tokens)* Move token validation middleware to new function
* *(api tokens)* Properly hash tokens
* *(api)* Enable notifications for api token routes
* *(caldav)* Add support for subtasks (i.e. `RELATED-TO` property) in CalDAV (#1634)
* *(cli)* Added --confirm/-c argument when deleting users to bypass prompt (#86)
* *(docs)* Update sample config and docs about Typesense config
* *(metrics)* Add active link share logins
* *(metrics)* Add total number of attachments metric
* *(metrics)* Add total number of files metric
* *(migration)* Migration from other services now happens in the background
* *(notifications)* Add endpoint to mark all notifications as read
* *(notify)* Don't notify disabled users
* *(reminders)* Include project in reminder notification
* *(tasks)* Add periodic resync of updated tasks to Typesense
* *(tasks)* Add searching via typesense
* *(tasks)* Add typesense indexing
* *(tasks)* Allow filtering for reminders, assignees and labels with Typesense
* *(tasks)* Find tasks by their identifier when searching with Typesense
* *(tasks)* Make sorting and filtering work with Typesense
* *(tasks)* Remove deleted tasks from Typesense
* *(typesense)* Add new tasks to typesense directly when they are created
* *(webhooks)* Add basic crud actions for webhooks
* *(webhooks)* Add basic sending of webhooks
* *(webhooks)* Add created by user object when returning all webhooks
* *(webhooks)* Add event listener to send webhook payload
* *(webhooks)* Add filter based on project id
* *(webhooks)* Add hmac signing
* *(webhooks)* Add index on project id
* *(webhooks)* Add route to get all available webhook events
* *(webhooks)* Add routes
* *(webhooks)* Add setting to enable webhooks
* *(webhooks)* Add support for webhook proxy
* *(webhooks)* Add timeout config option
* *(webhooks)* Expose whether webhooks are enabled
* *(webhooks)* Prevent link shares from managing webhooks
* *(webhooks)* Register task and project events as webhook
* *(webhooks)* Set user agent header to Vikunja
* *(webhooks)* Validate events and target url* Search improvements (#1598) ([6f825fa](6f825fa4133a3200dab8a46faa2932cf5633263c))
* Accept hex values which start with a # ([a1ea77f](a1ea77f7519efe7696bce018814071cbabaaa62c))
* Add demo mode flag ([97b5cd3](97b5cd306f44a23d5f8923b1cf750533c1ca3e10))
* Add setting for default bucket ([b99b323](b99b323c4c5a003c5b34e0196da566816469c608))
* Add very basic bruno collection ([7eb59f5](7eb59f577c32791af77770e5c4ca2e1d7c01ee04))
* Api tokens ([60cd125](60cd1250a0431f33748f83da3256f19ee8144dde))
* Convert all markdown content to html (#1625) ([8a4856a](8a4856ad8747dd590f61e80212f77fb6e41cfb4b))
* Endpoint to get all token routes ([1ca93a6](1ca93a678e6d931aa3afb3aaa654763ee8304d3b))
* Make default bucket configurable ([60bd5c8](60bd5c8a79af18b09cb87c650436d0eff771d670))
* Make unauthenticated user routes rate limit configurable ([c6c465c](c6c465c273037fd2c1f02360e647366834ab0cde))
* Move done bucket setting to project ([bbbb45d](bbbb45d22461ed88d744cc1d66f74a743a51b843))
* Webhooks (#1624) ([4d9baa3](4d9baa38d0861c082aa21713744927d520750fd6))


### Miscellaneous Tasks

* *(api tokens)* Add swagger docs about api token auth
* *(api tokens)* Remove updated date from tokens as it can't be updated anyway
* *(build)* Use our own goproxy to prevent issues with packages not found
* *(caldav)* Improve trimming .ics file ending
* *(ci)* Sign drone config
* *(ci)* Use golangci-lint docker image for lint step
* *(tasks)* Better error messages when indexing tasks into Typesense
* *(test)* Add task deleted assertion to project deletion test
* *(webhooks)* Remove WebhookEvent interface
* *(webhooks)* Reuse webhook client
* *(webhooks)* Simplify registering webhook events* Remove year from copyright headers ([e518fb1](e518fb1191c0a21180f91bf2defcef80e26f02a7))
* Add pr lockdown ([0abf686](0abf686f6630e052c43537cfcaf7b90eebcaa910))
* Assume username instead of id when parsing fails for user commands (#87) ([137f3bc](137f3bc151d6417ba3cc8362afec1e7457915ef5))
* Go mod tidy ([7c4b2c9](7c4b2c9b3911214d42ab9ab9a01605828013da55))
* Reverse the coupling of module log and config (#1606) ([ad04d30](ad04d302af94fe3cf8e5a70ebb87af9002da5610))
* Update contributing guidelines ([83f02b1](83f02b1ebc4ceda8226fb6d9c004241c0c47ae8d))


### Other

* *(other)* [skip ci] Updated swagger docs


## [0.21.0] - 2023-07-07

### Bug Fixes

* *(CalDAV)* Naming
* *(api)* License (#1457)
* *(build)* Make sure the docker image can access go tools
* *(caldav)* Do not create label if it exists by title (#1444)
* *(caldav)* Incoming tasks do not get correct time zone (#1455)
* *(ci)* Pipeline dependency
* *(cli)* Rename user project command
* *(docker)* Don't chown everything in Vikunja's default root folder
* *(docs)* Added Keycloak OpenID example (#1521)
* *(docs)* Clarify error codes in swagger docs
* *(docs)* Link to usage/api
* *(docs)* Semver link (#1470)
* *(filter)* Don't try to get the real subscription for a saved filter project
* *(filters)* Return all filters with all projects, not grouped under a pseudo project
* *(filters)* Sorting tasks from filters
* *(image)* Json type of struct property (#1469)
* *(import)* Don't try to load a nonexistent attachment file
* *(lint)* Disable misspell linter on redoc
* *(migration)* Don't try to fetch task details of tasks whose projects are deleted
* *(migration)* Enable insert from structure work recursively
* *(migration)* Make file migration work with new structure
* *(migration)* Remove unused is_deleted flag from Todoist api response
* *(migration)* Remove wunderlist leftovers
* *(migration)* Remove wunderlist leftovers
* *(migration)* Remove wunderlist leftovers
* *(migration)* Rename TickTick migration
* *(migration)* Revert wrongly changed url
* *(migration)* Use correct struct
* *(project)* Don't allow un-archiving a project when its parent project is archived
* *(project)* Don't check for namespaces in overdue reminders
* *(project)* Duplicate project into parent project
* *(project)* Recursively get all users from all parent projects
* *(project)* Remove comments, clarifications, notifications about namespaces
* *(project)* Remove namespaces checks
* *(project)* Remove namespaces from creating projects
* *(project)* Remove namespaces from getting projects
* *(projects)* Delete project in the correct order
* *(projects)* Don't allow making a project child of itself
* *(projects)* Don't check if new projects are archived
* *(projects)* Don't fail to fetch a task if there's a broken subscription record associated to it
* *(projects)* Don't return child projects twice
* *(projects)* Don't try to share for nonexisting namespace
* *(projects)* Permission check now works
* *(projects)* Properly check if a user or link share is allowed to create a new project
* *(projects)* Recalculate project's position after dragging when position would be 0
* *(projects)* Reset pagination limit when fetching subprojects
* *(projects)* Return subprojects which were shared from another user
* *(saved filters)* Don't let query parameters override saved sorting parameters
* *(spelling)* In config sample (#1489)
* *(task)* Don't build partial task identifier
* *(task)* Don't try to return a project identifier if there is no project
* *(tasks)* Don't check for namespaces in filters
* *(tasks)* Get all tasks from parent projects
* *(tasks)* Make sure task deleted notification actually has information about the deleted task
* *(tasks)* Read all tests
* *(tasks)* Return a correct task identifier if the list does not have a good one set
* *(tasks)* Sql for overdue reminders
* *(tasks)* Task relation test
* *(test)* Adjust fixture bucket and list ids
* *(test)* Adjust fixture id
* *(test)* Fixtures
* *(test)* Use correct filter id
* *(tests)* Adjust parent projects
* *(tests)* Make the tests compile again
* *(tests)* Permission tests for parent projects
* *(tests)* Subscription test fixtures
* *(tests)* Task collection fixtures
* *(tests)* Task permissions from parents
* Accept for migrations ([8edbca3](8edbca39cf9d771645d6feb05ee94eebc6403cbf))
* Add missing error code ([f2d943f](f2d943f5c4f1b13ef565692b893da05c6669c6d0))
* Add missing license header ([f4e12da](f4e12dab273474c0eb27f59c00faa828bb86522c))
* Align "ID" param for Delete and Update method of Task model ([b6d5605](b6d5605ef6b2799f939d016b1572b3d43e857d4d))
* Align "otherTaskID" param for Delete method of TaskRelation model ([ac377a7](ac377a7a5d708ef7543d99f716ceaa1ee8502649))
* Align namespaceID param ([7ada82e](7ada82ea926556ae39d106dc85d5a05f3c1c8cd3))
* Align task ID param ([f76bb2b](f76bb2b4a9c8a3b53bc73d0913ba94bba350f5da))
* Check if usernames contain spaces when creating a new user ([672fb35](672fb35bcbb47e4c0331813aa837fee28f372471))
* Compile errors ([a21bff3](a21bff3ffb8497d6e1b6c3bb50d9a9b2469f4eb0))
* Correctly pass unix socket to xorm ([7ad256f](7ad256f6cd3e15aeafce2bc29c28c458c3abdc0a))
* Docs auth openID method ([4f7d69a](4f7d69a108a2836e90b3c7ffe7f05247d80bfb85))
* Don't get favorite task projects filter multiple times ([a51bbd1](a51bbd1159fb1ada5980a5b27972ccf1404641af))
* Don't send bad request errors to sentry ([c0c523f](c0c523f0a8c83eb164febbc508ac98142d572d7a))
* Don't try to load subscriptions for nonexistent projects ([b519462](b5194624e021360ccdec20cb58bba57c23028c3f))
* Fetch all tasks for all projects ([353279c](353279cbff8fd6fa6b1bb81a8726a7a5a1b6b623))
* ILIKE helper ([dff4e01](dff4e01327907d42bf0b20a20912e5e9c69dd23e))
* Lint ([50c922b](50c922b7d1135b8f75478b89502fe0bb4c39547f))
* Lint ([ad06903](ad0690369f39dab3683ac5ef7664bd765fa1cb18))
* Lint ([e17b63b](e17b63b9201889946e91e7e295f31a80055c6ae4))
* Lint ([ef779e8](ef779e8730af169101bf1ebffb8d2522e5c6b7bc))
* Lint ([f0dcce7](f0dcce702f03f237ecde107a7ba62f61e2c3e313))
* Lint config ([9111db2](9111db2a16df6a4eec9e3cc2021bc6fdcace9ead))
* Lint errors ([ebc3dd2](ebc3dd2b3e72f56880320480829aead1bf554f67))
* Make it compile again ([d79c393](d79c393e5b4e880b8b09ce5944e8247ae07c4d58))
* Make sure Vikunja is buildable without swagger docs present ([47e4223](47e42238ef47ad6e4e90284593aae278e77c8631))
* Make sure projects are correctly sorted ([db3c7aa](db3c7aa8b04e828fafdf10bcfd5bde8cf19e6f10))
* Provide a proper error message when viewing a link share with an invalid token ([aa43127](aa43127e52aeb7412b13b4aaab091442dad534db))
* Reminder fixture ([4b00f22](4b00f224d92f0c6933f6cba14433538d64545eca))
* Remove old saved openid provider settings from cache when starting Vikunja ([9bf535d](9bf535d06f5b9bb455979b0bf3b6f0942daa1c9e))
* Rename after rebase ([e93a5ff](e93a5ff11fee7adac2897b3251db7abbbad4bcc5))
* Rename incorrectly named ProjectUsers method ([7e53a21](7e53a214070ee9b48fdffffcc42de9250c323e96))
* Rename project receiver variable ([f1cbe50](f1cbe50605b46e506c3233cc8da4b325f5727c87))
* Spelling ([fc2cc4a](fc2cc4a1555ca7e63ff902cde62380035a60ebb8))
* Test fixtures ([06f1d2e](06f1d2e91237195f8e720d4dd55b491b91e6547d))
* Test import ([fb818ea](fb818ea1867f8db813ff52622695fd206c21452e))
* Trello import tests ([61a3380](61a3380a9482312eac56f4cfd436517205f601aa))
* Typo ([4c698dc](4c698dc7c71418239e24b1756604371dcb6a2f74))
* Typo in email template ([2dad404](2dad4042170677af3db7be85cbe978ce6be721aa))
* Update redoc ([8916de0](8916de03666482c2319689e950d30a6fb737f239))
* Update xgo in dockerfile to 1.20.2 ([33f0d0f](33f0d0f85a7fdfd509bc8a4aad26df95c064468c))
* Upgrade jwt v5 ([359d051](359d0512cc7e73cdde9d4dd145332591c6743d11))
* Use rewrite when hosting frontend files via the api ([b56e45d](b56e45d74389d38c747887d3cb2a2b295bb549c7))
* Users_lists name in migration ([0a3fdc0](0a3fdc0344790f059140d8e482b028ffecdb3e4b))
* Using mysql via a socket ([0a6bbc2](0a6bbc2efd6bb4468c72cff2a70cd29350a50b75))


### Dependencies

* *(deps)* Update module github.com/imdario/mergo to v0.3.14
* *(deps)* Update github.com/arran4/golang-ical digest to 19abf92
* *(deps)* Update goreleaser/nfpm docker tag to v2.27.1 (#1438)
* *(deps)* Update module github.com/swaggo/swag to v1.8.11
* *(deps)* Update module github.com/imdario/mergo to v0.3.15 (#1443)
* *(deps)* Update golangci-lint to 1.52.1
* *(deps)* Update module github.com/wneessen/go-mail to v0.3.9
* *(deps)* Update github.com/gocarina/gocsv digest to 9a18a84
* *(deps)* Update module github.com/swaggo/swag to v1.8.12
* *(deps)* Update module github.com/getsentry/sentry-go to v0.20.0
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.0.3
* *(deps)* Update goreleaser/nfpm docker tag to v2.28.0 (#1475)
* *(deps)* Update src.techknowlogick.com/xgo digest to bff48e4 (#1474)
* *(deps)* Update module golang.org/x/sys to v0.7.0
* *(deps)* Update github.com/gocarina/gocsv digest to 6445c2b
* *(deps)* Update module golang.org/x/term to v0.7.0
* *(deps)* Update module github.com/spf13/cobra to v1.7.0
* *(deps)* Update module golang.org/x/image to v0.7.0
* *(deps)* Update module golang.org/x/oauth2 to v0.7.0
* *(deps)* Update module golang.org/x/crypto to v0.8.0
* *(deps)* Update module github.com/prometheus/client_golang to v1.15.0
* *(deps)* Update module github.com/lib/pq to v1.10.8
* *(deps)* Update module github.com/go-sql-driver/mysql to v1.7.1
* *(deps)* Update module github.com/lib/pq to v1.10.9
* *(deps)* Update src.techknowlogick.com/xgo digest to e65295a
* *(deps)* Update github.com/arran4/golang-ical digest to f69e132
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.0.4
* *(deps)* Update module github.com/go-testfixtures/testfixtures/v3 to v3.9.0
* *(deps)* Update module github.com/prometheus/client_golang to v1.15.1
* *(deps)* Update module golang.org/x/term to v0.8.0
* *(deps)* Update src.techknowlogick.com/xgo digest to 52d704d
* *(deps)* Update module github.com/swaggo/swag to v1.16.1
* *(deps)* Update module golang.org/x/sync to v0.2.0
* *(deps)* Update module github.com/getsentry/sentry-go to v0.21.0
* *(deps)* Update module golang.org/x/oauth2 to v0.8.0
* *(deps)* Update module golang.org/x/crypto to v0.9.0
* *(deps)* Update alpine docker tag to v3.18
* *(deps)* Update github.com/gocarina/gocsv digest to 7f30c79
* *(deps)* Update module github.com/magefile/mage to v1.15.0
* *(deps)* Update github.com/gocarina/gocsv digest to 9ddd7fd
* *(deps)* Update module github.com/coreos/go-oidc/v3 to v3.6.0
* *(deps)* Update module github.com/stretchr/testify to v1.8.3
* *(deps)* Update module github.com/labstack/echo-jwt/v4 to v4.2.0
* *(deps)* Update goreleaser/nfpm docker tag to v2.29.0 (#1528)
* *(deps)* Update module github.com/ulule/limiter/v3 to v3.11.2
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.0.5
* *(deps)* Update module github.com/imdario/mergo to v0.3.16
* *(deps)* Update module github.com/stretchr/testify to v1.8.4
* *(deps)* Update module github.com/spf13/viper to v1.16.0
* *(deps)* Update github.com/vectordotdev/go-datemath digest to 640a500 (#1532)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.17
* *(deps)* Update klakegg/hugo docker tag to v0.110.0 (#1538)
* *(deps)* Update golangci
* *(deps)* Update klakegg/hugo docker tag to v0.111.0 (#1539)
* *(deps)* Update klakegg/hugo docker tag to v0.111.3 (#1542)
* *(deps)* Update src.techknowlogick.com/xgo digest to 494bc06
* *(deps)* Update goreleaser/nfpm docker tag to v2.30.1 (#1540)
* *(deps)* Update module golang.org/x/sys to v0.9.0
* *(deps)* Update module golang.org/x/term to v0.9.0
* *(deps)* Update module golang.org/x/image to v0.8.0
* *(deps)* Update module golang.org/x/crypto to v0.10.0
* *(deps)* Update module golang.org/x/oauth2 to v0.9.0
* *(deps)* Update module golang.org/x/sync to v0.3.0
* *(deps)* Update github.com/gocarina/gocsv digest to 2696de6
* *(deps)* Update module github.com/prometheus/client_golang to v1.16.0
* *(deps)* Update module github.com/getsentry/sentry-go to v0.22.0
* *(deps)* Update github.com/gocarina/gocsv digest to 99d496c
* *(deps)* Update module github.com/imdario/mergo to v1 (#1559)
* *(deps)* Update github.com/dustinkirkland/golang-petname digest to e794b93
* *(deps)* Update module golang.org/x/sys to v0.10.0
* *(deps)* Update module golang.org/x/image to v0.9.0
* *(deps)* Update module golang.org/x/term to v0.10.0
* *(deps)* Update module golang.org/x/crypto to v0.11.0
* *(deps)* Update module golang.org/x/oauth2 to v0.10.0


### Documentation

* Add docs for installing with sqlite in docker (#70) ([a16fd67](a16fd67b51c02e09ef6709bee9ad2b341d80cd73))
* Add information about our Helm Chart ([22f89c1](22f89c1ccc3a281a75db9e42702604f88eb0568b))
* Fix menu links ([1f13b5d](1f13b5d7b4041042ea3b26ac2a850784b11ac377))
* Remove all traces of namespaces ([3b0935d](3b0935d033c6b5060f18e955acf4a647eb10721b))
* Remove outdated information ([327bb3b](327bb3bed99e0a4c5664251e3af15accf1a13062))
* Update error references to list ([259cf7d](259cf7d25bbb7a289fe9569c81c6f7d3855543bf))
* Update prometheus docs for clarity (#1458)
* Update references to list ([8dc6c95](8dc6c95333b38eb83c8053c628d05599e79dd27e))


### Features

* *(caldav)* Sync Reminders / VALARM (#1415)
* *(docs)* Change order of sections in nav (#1471)
* *(docs)* Various improvements
* *(kanban)* Return the total task count per bucket
* *(migration)* Ignore namespace changes
* *(migration)* Use new structure for migration
* *(projects)* Add parent project, migrate namespaces
* *(projects)* Check all parent projects for permissions
* *(projects)* Check parent project when checking archived status
* *(projects)* Cleanup namespace leftovers
* *(projects)* Don't allow deleting or archiving the default project
* *(projects)* Get all projects recursively
* *(projects)* Remove namespaces
* *(projects)* Return a favorites pseudo project when the user has favorite tasks
* *(subscriptions)* Make sure all subscriptions are inherited properly
* *(users)* Don't hide user email if it was the search request* Rename lists to projects ([349e6a5](349e6a59050a0beba82a7f626c2f72f6b8c88dde))
* Add logging options to mailer settings ([9590b82](9590b82c11852666524eeab562988226574a1b1c))
* Add relative Reminders (#1427) ([3f5252d](3f5252dc24a3dea89b2e049ccb1f9d0a59a89a88))
* Add token example ([4417223](441722372af3349b677dc013b1863e678b0e7158))
* Allow saving frontend settings via api ([04e2c51](04e2c51fac24a045abe1a85c8b661b6bc628686c))
* Allow to find users with access to a project more freely ([a7231e1](a7231e197e3d86d3ef27fad89ae60863d25b5df0))
* Check for cycles when creating or updating a project's parent ([9011894](9011894a2975d9d112dc3db453739e13261c0716))
* Generate swagger docs at build time ([efa24ce](efa24cec44865c5a8ab42a106deeb331ad1bed91))
* Improve relation kinds docs ([b826c13](b826c13f385b24ed1b33b8890cc5cdd5fe8b8f22))
* Make the new inbox project the default ([0110f93](0110f933134af0460d9fed9d652148c98e94b6cd))
* Migrate lists to projects in db identifiers ([2fba7bd](2fba7bdf02983e5cf7def09803def4cbf830f53b))
* Remove ChildProjects project property ([edcb806](edcb806421c2181a8b85aed5b53e8da6350b9630))
* Remove namespaces, make projects infinitely nestable (#1362) ([82beb3b](82beb3bf671ca0670b714160f0b4d9c186dfe120))
* Rename all list files ([8f4abd2](8f4abd2fe86e7a23d80bc5ebc4fc1ae75e1b78fb))
* Rename lists to projects ([47c2da7](47c2da7f1856e95956cdb968fa95295d3441a9f6))
* Rename lists to projects ([96a0f5e](96a0f5e169c9e8f8d20e3fe1d9de5eecead53ac9))
* Rename lists to projects ([fc73c84](fc73c84bf2b9a7cbd2f6cbd2a83ea9ccc3fd58fd))
* Rename lists to projects everywhere (#1318) ([869d4a3](869d4a336cb122df894acf040e02b6b2ba786fdb))


### Miscellaneous Tasks

* *(changelog)* Fix spelling
* *(docs)* Add info about `/buckets` sorting
* *(docs)* Move login and register routes to auth category in api docs
* *(docs)* Update error docs
* *(docs)* Update list -> project
* *(docs/translation)* Remove mention of weblate
* *(export)* Remove unused events
* *(project)* Fmt
* *(projects)* use a slice again ([3e8d1b3](3e8d1b3667ccfb2960650a4506771ec3c9b3a970))
* *(test)* Show table content when db assertion failed
* Cleanup ([7a9611c](7a9611c2daa41ec2da135a2a4e804551e4ab8ff2))
* Disable false-positive linter for generated docs ([076e857](076e857507a4cf59e0b0399a2e51a8d8baa03065))
* Fix comment url ([5856f21](5856f21f31fe7b81e7ffd203f70460785955411c))
* Fix spelling ([cd90db3](cd90db3117a7fa40175ecebd3ca37cc94a46e1ee))
* Generate swagger docs ([55410ea](55410ea73d50f5bc124eaf411c77125024b6fefa))
* Go mod tidy ([93056da](93056da792dafa70f91f7d114669997b3f93f7f1))
* Go mod tidy ([e5dde31](e5dde315fb6a7163546b9f88ebafacc886744db3))
* Remove cache options ([d83e3a0](d83e3a0a037b9a4d40ce22c8c51932eb23963ac2))
* Remove reminderDates after frontend is migrated to reminders (#1448) ([4a4ba04](4a4ba041e0f3e9c71dd4844d5191c9cbe4e4e3b7))
* Rename files (fix typo) ([6aadaaa](6aadaaaffc1fff4a94e35e8fa3f6eab397cbc3ce))


## [0.20.4] - 2023-03-12

### Bug Fixes

* *(docker)* Allow non-unique group id

### Documentation

* Add link to tutorial for installing Vikunja on Synology ([4de0efe](4de0efec1dd7da95dbf936728d7e23791396a63a))


## [0.20.3] - 2023-03-10

### Bug Fixes

* *(build)* Downgrade xgo to 1.19.2 so that builds work again
* *(caldav)* Add Z suffix to dates make it clear dates are in UTC
* *(caldav)* Use const for repeat modes
* *(caldav)* Make sure only labels where the user has permission to use them are used
* *(ci)* Pipeline dependency
* *(ci)* Pin nfpm container version and binary location
* *(ci)* Set release path to /source
* *(ci)* Tagging logic for release docker images
* *(ci)* Save generated .tags file to correctly tag docker releases
* *(ci)* Sign drone config
* *(docd)* Update Subdirectory Documentation (#1363)
* *(docker)* Cross compilation with buildx
* *(docker)* Re-add expose
* *(docker)* Passing environment variables into the container
* *(docker)* Make sure the vikunja user always exists and only modify the uid instead of recreating the user
* *(docs)* Add docs about cli user delete
* *(docs)* Old helm charts url (#1344)
* *(docs)* Fix a few minor typos (#59)
* *(docs)* Fix traefik v2 example (#65)
* *(docs)* Clarify support for caldav recurrence
* *(drone)* Add type, fix pull, remove group (#1355)
* *(dump)* Make sure null dates are properly set when restoring from a dump
* *(export)* Ignore file size for export files
* *(list)* Return lists for a namespace id even if that namespace is deleted
* *(list)* When list background is removed, delete file from file system and DB (#1372)
* *(mailer)* Forcessl config (#60)
* *(migration)* Use Todoist v9 api to migrate tasks from them
* *(migration)* Import TickTick data by column name instead of index (#1356)
* *(migration)* Use the proper authorization method for Todoist's api, fix issues with importing deleted items
* *(migration)* Remove unused todoist parameters
* *(migration)* Todoist pagination now avoids too many loops
* *(migration)* Don't try to add nonexistent tasks as related
* *(migration)* Make sure trello checklists are properly imported
* *(reminders)* Overdue tasks join condition
* *(reminders)* Make sure an overdue reminder is sent when there is only one overdue task
* *(reminders)* Prevent duplicate reminders when updating task details
* *(restore)* Check if we're really dealing with a string
* *(task)* Make sure the task's last updated timestamp is always updated when related entities changed
* *(task)* Correctly load tasks by id and uuid in caldav
* *(tasks)* Don't include undone overdue tasks from archived lists or namespaces in notification mails
* *(tasks)* Don't reset the kanban bucket when updating a task and not providing one
* *(tasks)* Don't set a repeating task done when moving it do the done bucket
* *(tasks)* Recalculate position of all tasks in a list or bucket when it would hit 0
* *(tasks)* Make sure tasks are sorted by position before recalculating them
* *(user)* Make reset the user's name to empty actually work
* Swagger docs ([96b5e93](96b5e933796275e87f3007e31db0623688dbdb3a))
* Restore notifications table from dump when it already had the correct format ([8c67be5](8c67be558f697ab52740c51ab453092c0f8f7c14))
* Make sure labels are always exported as caldav (#1412) ([1afc72e](1afc72e1906c02b093bb6d9748235b93ab0eb181))
* Lint ([491a142](491a1423788b76f236d070071cb46f5b2f5d3fd0))
* Lint ([20a5994](20a5994b1717e7751750f14a9a164825a8e6ade6))
* Lint ([077baba](077baba2eaff2f10b97384f07375ece7f51ec0fa))
* Lint ([9f14466](9f14466dfa8660362a4e51b3c8c6810bf8d66a22))


### Dependencies

* *(deps)* Update module github.com/yuin/goldmark to v1.5.3 (#1317)
* *(deps)* Update module golang.org/x/crypto to v0.2.0 (#1315)
* *(deps)* Update module github.com/spf13/afero to v1.9.3 (#1320)
* *(deps)* Update module golang.org/x/crypto to v0.3.0 (#1321)
* *(deps)* Update github.com/arran4/golang-ical digest to a677353 (#1323)
* *(deps)* Update module github.com/wneessen/go-mail to v0.3.5 (#1325)
* *(deps)* Update github.com/arran4/golang-ical digest to 1093469 (#1326)
* *(deps)* Update module github.com/golang-jwt/jwt/v4 to v4.4.3 (#1328)
* *(deps)* Update module github.com/go-sql-driver/mysql to v1.7.0 (#1332)
* *(deps)* Update module golang.org/x/sys to v0.3.0 (#1333)
* *(deps)* Update module golang.org/x/term to v0.3.0 (#1336)
* *(deps)* Update module golang.org/x/image to v0.2.0 (#1335)
* *(deps)* Update module golang.org/x/oauth2 to v0.2.0 (#1316)
* *(deps)* Update module golang.org/x/oauth2 to v0.3.0 (#1337)
* *(deps)* Update module github.com/getsentry/sentry-go to v0.16.0 (#1338)
* *(deps)* Update module golang.org/x/crypto to v0.4.0 (#1339)
* *(deps)* Update module github.com/pquerna/otp to v1.4.0 (#1341)
* *(deps)* Update module github.com/swaggo/swag to v1.8.9 (#1327)
* *(deps)* Update module github.com/wneessen/go-mail to v0.3.6 (#1342)
* *(deps)* Update module github.com/labstack/echo/v4 to v4.10.0 (#1343)
* *(deps)* Update module github.com/wneessen/go-mail to v0.3.7 (#1348)
* *(deps)* Update module github.com/coreos/go-oidc/v3 to v3.5.0 (#1349)
* *(deps)* Update module golang.org/x/sys to v0.4.0 (#1351)
* *(deps)* Update module golang.org/x/image to v0.3.0 (#1350)
* *(deps)* Update module golang.org/x/term to v0.4.0 (#1352)
* *(deps)* Update module golang.org/x/crypto to v0.5.0 (#1353)
* *(deps)* Update goreleaser/nfpm docker tag to v2.23.0 (#1347)
* *(deps)* Update module github.com/wneessen/go-mail to v0.3.8 (#1357)
* *(deps)* Update module src.techknowlogick.com/xgo to v1.6.0+1.19.5 (#1358)
* *(deps)* Update klakegg/hugo docker tag to v0.107.0 (#1272)
* *(deps)* Update module github.com/getsentry/sentry-go to v0.17.0 (#1361)
* *(deps)* Update module src.techknowlogick.com/xgo to v1.7.0+1.19.5 (#1364)
* *(deps)* Update module github.com/spf13/viper to v1.15.0 (#1365)
* *(deps)* Update module github.com/labstack/echo-jwt/v4 to v4.0.1 (#1369)
* *(deps)* Update module golang.org/x/oauth2 to v0.4.0 (#1354)
* *(deps)* Update github.com/gocarina/gocsv digest to 763e25b (#1370)
* *(deps)* Update goreleaser/nfpm docker tag to v2.24.0 (#1367)
* *(deps)* Update module github.com/swaggo/swag to v1.8.10 (#1371)
* *(deps)* Update module github.com/go-redis/redis/v8 to v9 (#1377)
* *(deps)* Update module github.com/labstack/echo-jwt/v4 to v4.1.0
* *(deps)* Update module github.com/ulule/limiter/v3 to v3.11.0 (#1378)
* *(deps)* Update module github.com/redis/go-redis/v9 to v9.0.2
* *(deps)* Update goreleaser/nfpm docker tag to v2.25.0 (#1382)
* *(deps)* Upgrade golangci-lint to 1.51.0
* *(deps)* Update module github.com/yuin/goldmark to v1.5.4
* *(deps)* Update module go to 1.20
* *(deps)* Update xgo to 1.20
* *(deps)* Update module golang.org/x/sys to v0.5.0
* *(deps)* Update module github.com/getsentry/sentry-go to v0.18.0 (#1386)
* *(deps)* Update module golang.org/x/term to v0.5.0
* *(deps)* Update module golang.org/x/crypto to v0.6.0
* *(deps)* Update module golang.org/x/oauth2 to v0.5.0
* *(deps)* Update module golang.org/x/image to v0.4.0
* *(deps)* Update goreleaser/nfpm docker tag to v2.26.0 (#1394)
* *(deps)* Update github.com/arran4/golang-ical digest to 07c6aad
* *(deps)* Update module github.com/threedotslabs/watermill to v1.2.0 (#1384)
* *(deps)* Update module golang.org/x/image to v0.5.0 (#1396)
* *(deps)* Update golang.org/x/net to 0.7.0
* *(deps)* Update module github.com/golang-jwt/jwt/v4 to v4.5.0 (#1399)
* *(deps)* Update github.com/gocarina/gocsv digest to bcce7dc
* *(deps)* Update golangci-lint to 1.51.2
* *(deps)* Update module github.com/labstack/echo/v4 to v4.10.1
* *(deps)* Update github.com/gocarina/gocsv digest to bee85ea
* *(deps)* Update module github.com/labstack/echo/v4 to v4.10.2
* *(deps)* Update module github.com/spf13/afero to v1.9.4
* *(deps)* Update github.com/gocarina/gocsv digest to dc4ee9d
* *(deps)* Update module github.com/stretchr/testify to v1.8.2
* *(deps)* Update github.com/gocarina/gocsv digest to 70c27cb
* *(deps)* Update module golang.org/x/sys to v0.6.0
* *(deps)* Update module golang.org/x/term to v0.6.0
* *(deps)* Update module golang.org/x/crypto to v0.7.0
* *(deps)* Update module golang.org/x/oauth2 to v0.6.0
* *(deps)* Update module golang.org/x/image to v0.6.0
* *(deps)* Update github.com/kolaente/caldav-go digest to 2a4eb8b
* *(deps)* Remove fsnotify replacement
* *(deps)* Update github.com/vectordotdev/go-datemath digest to f3954d0
* *(deps)* Update src.techknowlogick.com/xgo digest to 44f7e66
* *(deps)* Update module github.com/getsentry/sentry-go to v0.19.0
* *(deps)* Update module github.com/spf13/afero to v1.9.5
* *(deps)* Update module github.com/ulule/limiter/v3 to v3.11.1
* *(deps)* Update src.techknowlogick.com/xgo digest to b607086
* *(deps)* Update module github.com/gabriel-vasile/mimetype to v1.4.2

### Features

* *(background)* Add Last-Modified header (#1376)
* *(caldav)* Add support for repeating tasks
* *(caldav)* Export Labels to CalDAV (#1409)
* *(caldav)* Import caldav categories as Labels (#1413)
* *(migrators)* Remove wunderlist (#1346)
* *(release)* Use compressed binaries for package releases
* Use docker buildx to build multiarch images ([a6e214b](a6e214b654f28836cc8b93683dbfd5999282d11c))
* Provide logout url for openid providers (#1340) ([a79b1de](a79b1de2d0247a424f49cecaa267d30e8fa70a83))
* Refactored Dockerfile (#1375) ([522bf7d](522bf7d2fc3cc1704f58299b6435baccc7add533))
* Disable events log by default ([da9d25c](da9d25cf727c56acd7394b4b74e17a2959ee5242))
  - **BREAKING**: events log level is now off unless explicitly enabled


### Miscellaneous Tasks

* *(docs)* Adjust docs about frontend docker container
* *(docs)* Remove sponsors
* *(task)* Add test to check if a task's reminders are duplicated
* Remove custom gitea bug template in favor of githubs ([4fa45bf](4fa45bf9dcbaa8a41a53fc2305c4c2c1aa15691c))
* 0.20.2 release preparations ([d19fc80](d19fc80b8be08673136d84e10187cadb293822bf))
* Update funding links ([aa25ccd](aa25ccdc917684583a9bff4b7cb272004386f0fa))


### Other

* *(other)* Added Google & Google Workspace to OpenId examples (#1319)


## [0.20.2] - 2023-01-24

### Bug Fixes

* *(build)* Downgrade xgo to 1.19.2 so that builds work again
* *(caldav)* Add Z suffix to dates make it clear dates are in UTC
* *(caldav)* Use const for repeat modes
* *(ci)* Pipeline dependency
* *(ci)* Pin nfpm container version and binary location
* *(ci)* Set release path to /source
* *(ci)* Tagging logic for release docker images
* *(docs)* Add docs about cli user delete
* *(docs)* Old helm charts url (#1344)
* *(docs)* Fix a few minor typos (#59)
* *(drone)* Add type, fix pull, remove group (#1355)
* *(dump)* Make sure null dates are properly set when restoring from a dump
* *(export)* Ignore file size for export files
* *(list)* Return lists for a namespace id even if that namespace is deleted
* *(mailer)* Forcessl config (#60)
* *(migration)* Use Todoist v9 api to migrate tasks from them
* *(migration)* Import TickTick data by column name instead of index (#1356)
* *(migration)* Use the proper authorization method for Todoist's api, fix issues with importing deleted items
* *(reminders)* Overdue tasks join condition
* *(reminders)* Make sure an overdue reminder is sent when there is only one overdue task
* *(reminders)* Prevent duplicate reminders when updating task details
* *(restore)* Check if we're really dealing with a string
* *(tasks)* Don't include undone overdue tasks from archived lists or namespaces in notification mails
* *(tasks)* Don't reset the kanban bucket when updating a task and not providing one
* *(tasks)* Don't set a repeating task done when moving it do the done bucket
* *(user)* Make reset the user's name to empty actually work* Swagger docs ([41c9e3f](41c9e3f9a47280887b56941280904aea6ef31f85))
* Restore notifications table from dump when it already had the correct format ([15811fd](15811fd4d4485cd25cf8d2f8fdd04ebfea8e6663))


### Dependencies

* *(deps)* Update module github.com/yuin/goldmark to v1.5.3 (#1317)
* *(deps)* Update module golang.org/x/crypto to v0.2.0 (#1315)
* *(deps)* Update module github.com/spf13/afero to v1.9.3 (#1320)
* *(deps)* Update module golang.org/x/crypto to v0.3.0 (#1321)
* *(deps)* Update github.com/arran4/golang-ical digest to a677353 (#1323)
* *(deps)* Update module github.com/wneessen/go-mail to v0.3.5 (#1325)
* *(deps)* Update github.com/arran4/golang-ical digest to 1093469 (#1326)
* *(deps)* Update module github.com/golang-jwt/jwt/v4 to v4.4.3 (#1328)
* *(deps)* Update module github.com/go-sql-driver/mysql to v1.7.0 (#1332)
* *(deps)* Update module golang.org/x/sys to v0.3.0 (#1333)
* *(deps)* Update module golang.org/x/term to v0.3.0 (#1336)
* *(deps)* Update module golang.org/x/image to v0.2.0 (#1335)
* *(deps)* Update module golang.org/x/oauth2 to v0.2.0 (#1316)
* *(deps)* Update module golang.org/x/oauth2 to v0.3.0 (#1337)
* *(deps)* Update module github.com/getsentry/sentry-go to v0.16.0 (#1338)
* *(deps)* Update module golang.org/x/crypto to v0.4.0 (#1339)
* *(deps)* Update module github.com/pquerna/otp to v1.4.0 (#1341)
* *(deps)* Update module github.com/swaggo/swag to v1.8.9 (#1327)
* *(deps)* Update module github.com/wneessen/go-mail to v0.3.6 (#1342)
* *(deps)* Update module github.com/labstack/echo/v4 to v4.10.0 (#1343)
* *(deps)* Update module github.com/wneessen/go-mail to v0.3.7 (#1348)
* *(deps)* Update module github.com/coreos/go-oidc/v3 to v3.5.0 (#1349)
* *(deps)* Update module golang.org/x/sys to v0.4.0 (#1351)
* *(deps)* Update module golang.org/x/image to v0.3.0 (#1350)
* *(deps)* Update module golang.org/x/term to v0.4.0 (#1352)
* *(deps)* Update module golang.org/x/crypto to v0.5.0 (#1353)
* *(deps)* Update goreleaser/nfpm docker tag to v2.23.0 (#1347)
* *(deps)* Update module github.com/wneessen/go-mail to v0.3.8 (#1357)
* *(deps)* Update module src.techknowlogick.com/xgo to v1.6.0+1.19.5 (#1358)
* *(deps)* Update klakegg/hugo docker tag to v0.107.0 (#1272)
* *(deps)* Update module github.com/getsentry/sentry-go to v0.17.0 (#1361)
* *(deps)* Update module src.techknowlogick.com/xgo to v1.7.0+1.19.5 (#1364)
* *(deps)* Update module github.com/spf13/viper to v1.15.0 (#1365)
* *(deps)* Update module github.com/labstack/echo-jwt/v4 to v4.0.1 (#1369)

### Features

* *(migrators)* Remove wunderlist (#1346)
* *(release)* Use compressed binaries for package releases
* Use docker buildx to build multiarch images ([9bd6795](9bd6795266fd54ae42664c20ed7633ac7daf6199))

### Miscellaneous Tasks

* Remove custom gitea bug template in favor of githubs ([7b1e1c7](7b1e1c79e358f3fcecb217259491f016402cdcc7))

### Other

* *(other)* Added Google & Google Workspace to OpenId examples (#1319)

## [0.20.1] - 2022-11-11

### Bug Fixes

* *(docs)* Add explanation on how to run the cli in docker
* *(filter)* Also check for 0 values if the filter should include nulls
* *(filter)* Only check for 0 values in filter fields with numeric values
* *(filters)* Try to parse date filter fields of the provided dates are not valid iso dates
* *(filters)* Try parsing dates without time
* *(filters)* Try parsing invalid dates like 2022-11-1
* *(metrics)* Make currently active users actually work
* *(task)* Duplicate reminders when adding different ones between winter / summer time
* *(tasks)* Allow sorting by task index* Make sure task indexes are calculated correctly when moving tasks between lists ([c495096](c4950964443a9bffc4cdd8fc25004ad951520f20))
* Look for the default bucket based on the position instead of the index ([622f2f0](622f2f0562bd8e3a5c97ec0b001c646a33a86c2b))
* Usage with postgres over unix socket (#1308) ([641a9da](641a9da93d24a18d6cbad2929eea1be6c1e0d0b2))

### Dependencies

* *(deps)* Update module github.com/prometheus/client_golang to v1.13.1 (#1307)
* *(deps)* Update module github.com/spf13/viper to v1.14.0 (#1309)
* *(deps)* Update module golang.org/x/sys to v0.2.0 (#1311)
* *(deps)* Update module golang.org/x/term to v0.2.0 (#1312)
* *(deps)* Update module github.com/prometheus/client_golang to v1.14.0 (#1313)
* *(deps)* Update module github.com/getsentry/sentry-go to v0.15.0 (#1314)

### Features

* *(docs)* Add release checklist

### Other

* *(other)* Necessary is a common misspelling of necessary (#1304)

## [0.20.0] - 2022-10-28

### Bug Fixes

* *(caldav)* Make sure duration and due date follow rfc5545
* *(caldav)* No failed login emails for tokens (#1252)
* *(ci)* Make sure release zip files have a .zip ending
* *(ci)* Make sure release os packages are properly named
* *(docs)* Clarify using port 25 as mail port when mail does not work
* *(docs)* Document pnpm instead of yarn
* *(docs)* Fix redirect_url example (#50)
* *(lists)* Return correct max right for lists where the user has created the namespace
* *(mail)* Pass mail server timeout (#1253)
* *(migration)* Properly parse duration
* *(migration)* Expose ticktick migrator to /info
* *(migration)* Make sure importing works when the csv file has errors and don't try to parse empty values as dates
* *(namespaces)* Add list subscriptions (#1254)
* *(todoist)* Properly import all done tasks* Properly log extra message ([c194797](c19479757a20d72484b4e071b45055746ff2b67e))
* Don't try to compress riscv64 binaries in releases ([d8f387f](d8f387f7967ffb94035de2fcfc4578247ae1023e))
* Preserve dates for repeating tasks (#47) ([090c671](090c67138a16258480b866b05c6fdc2e02d12c89))
* Tasks with the same assignee as doer should not appear twice in overdue task mails ([45defeb](45defebcf435cade4b72763236e1e2dfdac770cc))
* Don't allow setting a list namespace to 0 ([96ed1e3](96ed1e33e38beec1bb1ab0813074b035dd02fade))
* Make sure pseudo namespaces and lists always have the current user as owner ([878d19b](878d19beb81869392e33a8ffc1ec247d1cf1e4d6))
* Use connection string for postgres ([fcb205a](fcb205a842a4e828e6e933339b23f5aa8b297125))
* Make sure user searches are always case-insensitive ([c076f73](c076f73a87bc9b39b17389e25d0186ab71aa24bf))
* Make cover image id actually updatable ([0e1904d](0e1904d50b8576a2e9ea5812314aa3c8f304edb5))
* Make cover image id actually updatable ([0eb4709](0eb47096db02ceb5032c7439b3b901fbadd0d1bb))
* Make sure a user can only be assigned once to a task ([008908e](008908eb49eeb50a554c416422feb3b465efa165))
* Make sure list subscriptions are set correctly when their namespace has a subscription already ([2fc690a](2fc690a783f5b702fad71da627aa616017727f56))


### Dependencies

* *(deps)* Update klakegg/hugo docker tag to v0.101.0
* *(deps)* Update golang.org/x/sync digest to 8fcdb60
* *(deps)* Update golang.org/x/oauth2 digest to f213421
* *(deps)* Update module src.techknowlogick.com/xgo to v1.5.0+1.19
* *(deps)* Update module github.com/coreos/go-oidc/v3 to v3.4.0
* *(deps)* Update golang.org/x/image digest to e7cb969
* *(deps)* Update golang.org/x/term digest to 7a66f97
* *(deps)* Update module github.com/lib/pq to v1.10.7
* *(deps)* Update module github.com/spf13/viper to v1.13.0 (#1260)
* *(deps)* Update dependency golang to v1.19 (#1228)
* *(deps)* Update module github.com/wneessen/go-mail to v0.2.8 (#1258)
* *(deps)* Update module github.com/yuin/goldmark to v1.5.2 (#1261)
* *(deps)* Update module src.techknowlogick.com/xormigrate to v1.5.0 (#1262)
* *(deps)* Update module github.com/magefile/mage to v1.14.0 (#1259)
* *(deps)* Update module github.com/swaggo/swag to v1.8.6 (#1243)
* *(deps)* Update module github.com/wneessen/go-mail to v0.2.9 (#1264)
* *(deps)* Update dependency klakegg/hugo to v0.102.3 (#1265)
* *(deps)* Update module github.com/getsentry/sentry-go to v0.14.0 (#1266)
* *(deps)* Update module github.com/labstack/gommon to v0.4.0 (#1269)
* *(deps)* Update golang.org/x/crypto digest to 4161e89 (#1268)
* *(deps)* Update golang.org/x/oauth2 digest to b44042a (#1270)
* *(deps)* Update golang.org/x/sys digest to 84dc82d (#1271)
* *(deps)* Update dependency klakegg/hugo to v0.104.2 (#1267)
* *(deps)* Update golang.org/x/crypto digest to d6f0a8c (#1275)
* *(deps)* Update golang.org/x/sys digest to 090e330 (#1276)
* *(deps)* Update module github.com/spf13/cobra to v1.6.0 (#1277)
* *(deps)* Update module github.com/wneessen/go-mail to v0.3.0 (#1278)
* *(deps)* Update golang.org/x/crypto digest to 56aed06 (#1280)
* *(deps)* Update golang.org/x/text to v0.3.8
* *(deps)* Update module github.com/wneessen/go-mail to v0.3.1 (#1281)
* *(deps)* Update module github.com/labstack/echo/v4 to v4.9.1 (#1282)
* *(deps)* Update golang.org/x/sys digest to 95e765b (#1283)
* *(deps)* Update golang.org/x/oauth2 digest to 6fdb5e3 (#1284)
* *(deps)* Update golang.org/x/image digest to ffcb3fe (#1288)
* *(deps)* Update module golang.org/x/sync to v0.1.0 (#1291)
* *(deps)* Update module github.com/swaggo/swag to v1.8.7 (#1290)
* *(deps)* Update golang.org/x/term digest to 8365914 (#1289)
* *(deps)* Update module github.com/coreos/go-systemd/v22 to v22.4.0 (#1287)
* *(deps)* Update module golang.org/x/oauth2 to v0.1.0 (#1296)
* *(deps)* Update module golang.org/x/crypto to v0.1.0 (#1295)
* *(deps)* Update module golang.org/x/image to v0.1.0 (#1293)
* *(deps)* Update module github.com/wneessen/go-mail to v0.3.2 (#1297)
* *(deps)* Update module github.com/stretchr/testify to v1.8.1 (#1298)
* *(deps)* Update module github.com/spf13/cobra to v1.6.1 (#1299)
* *(deps)* Update module github.com/wneessen/go-mail to v0.3.3 (#1300)
* *(deps)* Update module github.com/wneessen/go-mail to v0.3.4 (#1302)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.16 (#1301)

### Features

* *(docs)* Add docs about how to deploy Vikunja in a subdirectory
* *(docs)* Document pnpm (#1251)
* *(migration)* Add TickTick migrator
* *(migration)* Add routes for TickTick migrator
* *(migration)* Generate swagger docs
* *(task)* Add cover image attachment id property
* *(task)* Add cover image attachment id property (#1263)* Add sponsor to readme (realm) ([f814dd0](f814dd03eb7f1ae08ea67ae0e3e89b8b4e684ce3))
* Upgrade xorm ([b1fd13b](b1fd13bbcbc551d1bbfe78d91fe6209369709df5))
* Upgrade xorm ([4323803](4323803fd6801e21121eac0d9f9cd62879f090f7))
* Upgrade xorm (#1197) ([5341918](53419180be386d675b4513e7ec70aca85b5ac99b))
* Add github issue templates ([9c4bb5a](9c4bb5a24429dec686e3ccdcd2b920ce5528031c))
* Remove gitea issue template so that only the form is used ([ce621ee](ce621ee5d6b47a0776628073bbd53312a97d116b))
* Add gitea issue template ([0612f4d](0612f4d0e03fbe85018f51056c4833557e78cd3f))
* Provide default user settings for new users via config ([5a40100](5a40100ac5be33d2cbce3c25e355d4036b9b4d3f))
* Add proper checks and errors to see if an attachment belongs to the task it's being used as cover image in ([631a265](631a265d2de9a6196faf28574023fc3cdcc0bfc7))
* Allow a user to remove themselves from a team ([b8769c7](b8769c746ceddc9818f91d6a8a404293ea2e837e))
* TickTick migrator (#1273) ([df2e36c](df2e36c2a378d4bd1b81d959da180b6e9b9a37b9))


### Miscellaneous Tasks

* Upgrade echo ([86ee827](86ee8273bce36c7b4767a34e0d878d63b37ea1b4))
* Go mod tidy ([903b8ff](903b8ff43871234f41f706d571ee2caaba5f4232))
* Generate swagger docs ([e113fe3](e113fe34d074f698f4b0cb237821f359976daa5c))
* Remove unused dependencies ([f5fd849](f5fd849a0b93ff3bba53ac4907bb3fb04fa8692b))

## [0.19.2] - 2022-08-17

### Bug Fixes

* Don't fail a migration if there is no filter saved ([10ded56](10ded56f6697ef47910ec68d37f26ed47cbe9180))
* Don't override saved filters ([beb4d07](beb4d07cf95fc25f7cc5f7471b46bdab49f95fe0))

## [0.19.1] - 2022-08-17

### Bug Fixes

* Prevent moving a list into a pseudo namespace ([3ccc636](3ccc6365a6892f37ee54b0750a34a61e52f6dba1))
* Make sure generating blur hashes for bmp, tiff and webp images works ([8bf0f8b](8bf0f8bb571ddff69a7142be1acaa2e4e0c38e3b))
* Add debian-based docker image for arm 32 builds ([c9e044b](c9e044b3ad60d25e9641d22d84571a7db83a26ac))
* Only list all users when allowed ([9ddd7f4](9ddd7f48895f508539d591aeebde450a86987024))
* Lint ([0c8bed4](0c8bed4054649de8510e5a636d1a14b65d52c402))

### Dependencies

* *(deps)* Update golang.org/x/sys digest to 6e608f9 (#1229)
* *(deps)* Update golang.org/x/sync digest to 886fb93 (#1221)
* *(deps)* Update golang.org/x/sys digest to 8e32c04 (#1230)
* *(deps)* Update golang.org/x/term digest to a9ba230 (#1222)
* *(deps)* Update module github.com/prometheus/client_golang to v1.13.0
* *(deps)* Update module github.com/prometheus/client_golang to v1.13.0 (#1231)
* *(deps)* Update golang.org/x/sys digest to 1c4a2a7
* *(deps)* Update golang.org/x/oauth2 digest to 128564f (#1220)
* *(deps)* Update golang.org/x/image digest to 062f8c9 (#1219)
* *(deps)* Update golang.org/x/crypto digest to 630584e (#1218)
* *(deps)* Update module github.com/labstack/echo/v4 to v4.8.0 (#1233)
* *(deps)* Update golang.org/x/sys digest to fbc7d0a (#1234)
* *(deps)* Update module github.com/wneessen/go-mail to v0.2.6 (#1235)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.15 (#1238)

### Features

* *(docs)* Add k8s docs* Add openid examples ([dbb0f54](dbb0f5473269fb29c4a484cd233a5b76484c4ca7))
* Search by assignee username instead of id ([7f28865](7f28865903740d6dde15ee005323fbdee3072166))
* Add migration to change user ids to usernames in saved filters ([3047ccf](3047ccfd4af8fee55d9ebff49138911ab80cb3d2))

## [0.19.0] - 2022-08-03

### Bug Fixes

* *(caldav)* Make sure the caldav tokens of non-local accounts are properly checked
* *(caldav)* Properly parse durations when returning VTODOs
* *(caldav)* Make sure description is parsed correctly when multiline
* *(ci)* Sign drone config
* *(ci)* Make sure the linter actually runs
* *(ci)* Install git in lint step
* *(docker)* Switch to debian base image
* *(docker)* Use official go image instead of our own to build
* *(docs)* Update minimum required go version
* *(docs)* Use up-to-date hugo image for building
* *(docs)* Don't use cannonify url
* *(docs)* Image urls in synology setup explanation
* *(docs)* Clarify frontend requirements to use Vikunja
* *(dump)* Don't try to save a config file if none was provided and dump vikunja env variables
* *(mage)* Handle different types of errors
* *(mail)* Don't set a username by default
* *(mail)* Don't try to authenticate against the mail server when no credentials are provided
* *(mail)* Set server name in tls config so that sending mail works with skipTlsVerify set to false
* *(restore)* Properly decode notifications json data
* *(restore)* Make sure to reset sequences after importing a dump when using postgres
* *(restore)* Use the correct initial migration* Generate swagger docs ([4de8ec5](4de8ec56a62caef22c2061376383de1fe53ca4c3))
* Make sure the full task is available in notifications ([c2b6119](c2b6119434e6e806785d2c259c3ca3d25496ec75))
* Don't try to load the namespace of a list if it is a shared list ([d7e47a2](d7e47a28d4bb04d4c7c3ed85a263134180da447a))
* Correctly load and pass the user when deleting it ([50b65a5](50b65a517da6869dc6a48fec40323e254ba4c032))
* Updating a list might remove its background ([cf05de1](cf05de19b317bd99c30de4c6a149a0d8a4ff4f49))
* Sorting for saved filters ([57e5d10](57e5d10eee4c45a04e9e1aaeaf41dd44eb8ce788))
* Importing trello attachments ([c3e0e64](c3e0e6405a634894a30dbf9c0506d1691ae4d443))
* Lint ([0b77625](0b7762590f6a0a82090ef74e9e7e32b37142d343))
* Deleting users with no namespaces ([f8a0a7e](f8a0a7e9539a44b2f790a08eb1b03028b56eaac3))
* Importing tasks from todoist without a due time set ([fd0d462](fd0d462bf4dd8225c67ba34958e5148f6167d264))
* User deletion never happens ([72d3c54](72d3c54efd3dda6ae846a069415688391cb1c9ae))
* User deletion reminder emails counting up ([f581885](f581885e65ada15439ec02f1d18d825b03581523))
* User not actually deleted ([70e005e](70e005e7ce5cf1dd25ec9ddfde3cfbbd258fadb6))
* User deletion schedule ([5c88dfe](5c88dfe88eab442724f22c3b29741e78939deae2))
* Friendly name not getting synced on first login from openid ([190a9f2](190a9f2a4c1a59bc68b839c465bb2536532c0e96))
* Importing archived lists or namespaces ([8bb3f8d](8bb3f8d37c78dc704ff4316c750e143528151b48))
* Lint ([a31086a](a31086a7a9ca7723f61a826bccbea125243478f1))
* Microsoft todo migration not importing all tasks ([43f1daf](43f1daf40c388a0aa40f7fd6a8db4c78308d4efd))
* Clarify which config file is used on startup ([44aaf0a](44aaf0a4eccebb1d1a25f5563e928bd1bb82d351))
* Disabling logging completely now works ([22e3f24](22e3f242a396aa9cf54e9426077816f97a0da36f))
* Restoring dumps with no config file saved in them ([8bf2254](8bf2254f4b87446ab0a39080cb0b7d32ccec7c0a))
* Validate email address when creating a user via cli ([75f74b4](75f74b429eea7ae3a75cb10def1ca658af35086a))
* Checking for error types ([ac6818a](ac6818a4769a162c458553944509fe64357370f9))
* Lint ([7fa0865](7fa086518800243385d8cc4696eeea9bf093e5b3))
* Return BlurHash in unsplash search results ([6b51fae](6b51fae0931308464038f55b25e81e68d014c49c))
* Go mod tidy ([e19ad11](e19ad1184662dc9ac9aa89a44abdffc091e2a1b8))
* Decoding images for blurHash generation ([d3bdafb](d3bdafb717b1ad3e2165097ef0b0c2dd47e1502e))
* Lint ([de97fcb](de97fcbd121b1d56b74175fd79ef594ef34e71c8))
* Broken link (#27) ([96e519e](96e519ea96c9537222d0b455037e11fbe9660c31))
* Add more methods to figure out the current binary location ([9845fcc](9845fcc1708431f8f736d36e7e19a1067b0e0e52))
* Set derived default values only after reading config from file or env ([f5ebada](f5ebada91351faf1e5602f0260908defaaabd810))
* Sort tasks logically and consistent across dbms (#1177) ([e52c45d](e52c45d5aabb74ea7b472e8d5b44491cdd7e9489))
* VIKUNJA_SERVICE_JWT_SECRET should be VIKUNJA_SERVICE_JWTSECRET (#1184) ([172a621](172a6214d7c30278017129b950339c78a6ddb7bc))
* Add missing migration ([d837f8a](d837f8a6248b5ff2700a4bfc300d7f9d466cb918))
* Revert renaming Attachments to Embeds everywhere ([c62e26b](c62e26b6fe9d9f362fcfb1df2d5664d7f6854c31))
* Set the correct go version in go.mod ([bc7f6a8](bc7f6a858693b0e61fff7d03b5c2b40b6ae1a55d))
* Go mod tidy ([7a30294](7a30294407843693f6c3a7414b3b9d7093359194))
* Tests ([d0e09d6](d0e09d69d048e62ee7c5b666c2f56761b03e68e6))
* Go mod tidy ([951d74b](951d74b272b1e881faa10095f47b6598bb076273))
* Prevent logging openid provider errors twice ([25ffa1b](25ffa1bc2e2f1108f20b0336708d2410bb61c9e1))
* Remove credential escaping for postgres connections to allow for passwords with special characters ([230478a](230478aae947c86f4c6f1f251dcb30aeb1293283))
* Cycles in tasks array when memory caching was enabled ([f5a4c13](f5a4c136fbca6fc5770476e6de8d81173f007df2))
* Add missing error check ([5cc4927](5cc4927b9ef97667bf763772beb36225fdbeded8))
* Properly set tls config for mailer ([5743a4a](5743a4afe51de221beeeabe66552ae4d92eed1a6))
* Return 9:00 as default time for reminders if none was set ([79b3167](79b31673e2a79eaa124976840e85757d2bebb887))
* Reset id sequence when importing a dump from postgres ([0f555b7](0f555b7ec74ad493d2f70a4f4040db333943dc1c))
* Add validation for negative repeat after values ([dd46174](dd461746a655d716ef142d96a2bcef5615de3dd9))
* Lint ([1feb62c](1feb62cc458e939d46d16d24347557e7959ddfb9))
* Make sure to use user discoverability settings when searching list users ([382a788](382a7884be1f37da5c8f657c4b17316d8691dd59))
* Properly decode params in url ([8f27e7e](8f27e7e619ac73716211d838f52c73d7d97aead5))
* Return all users on a list when no search param was provided ([c51ee94](c51ee94ad1d552d69c71adfc2180c7ad0d23235d))
* Don't return email addresses from user search results ([3688bbd](3688bbde20e989397353ea4f7e872b00a53099c2))
* Lint ([77fafd5](77fafd5dc32aee464961be40d5d0ccf82490d02a))
* Increase test timeout ([26e2d0b](26e2d0bddeaea902dba055baf7a4c866a44ba7f1))
* Switch to buster for build image ([59796fd](59796fd4905fca74d26c5541878379cda143a30e))
* Use our own build image as base build image ([b6d7323](b6d7323cdfac958c9740feba1342114ab13a0afd))
* Use golang build image to test migrations ([84bcdbf](84bcdbf937c3be7823fcf8d5fef52e3cbb1c9bde))
* Switch back to alpine for everything, disable arm 32 docker builds ([7ffe9b6](7ffe9b625e441202a704db2774dd66fc38244c6d))


### Dependencies

* *(deps)* Update golang.org/x/sys commit hash to a851e7d (#972)
* *(deps)* Update golang.org/x/sys commit hash to aa78b53 (#973)
* *(deps)* Update golang.org/x/sys commit hash to 528a39c (#974)
* *(deps)* Update golang.org/x/sys commit hash to 437939a (#975)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.1 (#976)
* *(deps)* Update module github.com/coreos/go-oidc/v3 to v3.1.0 (#985)
* *(deps)* Update module github.com/spf13/viper to v1.9.0 (#987)
* *(deps)* Update golang.org/x/crypto commit hash to 089bfa5 (#979)
* *(deps)* Update golang.org/x/term commit hash to 140adaa (#983)
* *(deps)* Update module github.com/labstack/echo/v4 to v4.6.0 (#988)
* *(deps)* Update golang.org/x/sys commit hash to b8560ed (#989)
* *(deps)* Update module github.com/golang-jwt/jwt/v4 to v4.1.0 (#991)
* *(deps)* Update golang.org/x/sys commit hash to 92d5a99 (#992)
* *(deps)* Update module github.com/swaggo/swag to v1.7.3 (#990)
* *(deps)* Update module github.com/labstack/echo/v4 to v4.6.1 (#993)
* *(deps)* Update golang.org/x/sys commit hash to 1cf2251 (#994)
* *(deps)* Update golang.org/x/sys commit hash to 39ccf1d (#995)
* *(deps)* Update golang.org/x/term commit hash to 03fcf44 (#996)
* *(deps)* Update golang.org/x/oauth2 commit hash to 6b3c2da (#1000)
* *(deps)* Update golang.org/x/sys commit hash to 69063c4 (#1001)
* *(deps)* Update module github.com/gabriel-vasile/mimetype to v1.4.0 (#1004)
* *(deps)* Update postgres docker tag to v14 (#1005)
* *(deps)* Update module github.com/go-redis/redis/v8 to v8.11.4 (#1003)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.9 (#1008)
* *(deps)* Update golang.org/x/sys commit hash to 9d821ac (#1009)
* *(deps)* Update golang.org/x/sys commit hash to 0ec99a6 (#1010)
* *(deps)* Update golang.org/x/sys commit hash to 9d61738 (#1011)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.2 (#1012)
* *(deps)* Update golang.org/x/sys commit hash to 8e51046 (#1016)
* *(deps)* Update golang.org/x/sys commit hash to d6a326f (#1017)
* *(deps)* Update module github.com/swaggo/swag to v1.7.4 (#1018)
* *(deps)* Update golang.org/x/sys commit hash to 711f33c (#1019)
* *(deps)* Update golang.org/x/sys commit hash to 69cdffd (#1020)
* *(deps)* Update golang.org/x/oauth2 commit hash to ba495a6 (#1022)
* *(deps)* Update golang.org/x/image commit hash to 6944b10 (#1023)
* *(deps)* Update golang.org/x/sys commit hash to 6e78728 (#1024)
* *(deps)* Update golang.org/x/sys commit hash to b3129d9 (#1025)
* *(deps)* Update golang.org/x/sys commit hash to 611d5d6 (#1026)
* *(deps)* Update golang.org/x/sys commit hash to 39c9dd3 (#1027)
* *(deps)* Update golang.org/x/sys commit hash to a2f17f7 (#1028)
* *(deps)* Update golang.org/x/sys commit hash to 4dd7244 (#1029)
* *(deps)* Update golang.org/x/sys commit hash to ae416a5 (#1030)
* *(deps)* Update golang.org/x/sys commit hash to 7861aae (#1031)
* *(deps)* Update golang.org/x/oauth2 commit hash to d3ed0bb (#1032)
* *(deps)* Update module github.com/labstack/gommon to v0.3.1 (#1033)
* *(deps)* Update golang.org/x/sys commit hash to c75c477 (#1034)
* *(deps)* Update golang.org/x/sys commit hash to ebca88c (#1035)
* *(deps)* Update golang.org/x/sys commit hash to e0b2ad0 (#1037)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.3 (#1038)
* *(deps)* Update golang.org/x/crypto commit hash to ceb1ce7 (#1041)
* *(deps)* Update module github.com/lib/pq to v1.10.4 (#1040)
* *(deps)* Update golang.org/x/sys commit hash to 51b60fd (#1042)
* *(deps)* Update golang.org/x/sys commit hash to 99a5385 (#1043)
* *(deps)* Update golang.org/x/sys commit hash to f221eed (#1044)
* *(deps)* Update golang.org/x/sys commit hash to 0c823b9 (#1045)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.4 (#1046)
* *(deps)* Update golang.org/x/sys commit hash to 0a5406a (#1048)
* *(deps)* Update golang.org/x/crypto commit hash to b4de73f (#1047)
* *(deps)* Update module github.com/ulule/limiter/v3 to v3.9.0 (#1049)
* *(deps)* Update golang.org/x/crypto commit hash to ae814b3 (#1050)
* *(deps)* Update golang.org/x/sys commit hash to dee7805 (#1051)
* *(deps)* Update golang.org/x/sys commit hash to ef496fb (#1052)
* *(deps)* Update golang.org/x/sys commit hash to fe61309 (#1054)
* *(deps)* Update module github.com/swaggo/swag to v1.7.6 (#1055)
* *(deps)* Update golang.org/x/crypto commit hash to 5770296 (#1056)
* *(deps)* Update module github.com/golang-jwt/jwt/v4 to v4.2.0 (#1057)
* *(deps)* Update golang.org/x/sys commit hash to 94396e4 (#1058)
* *(deps)* Update golang.org/x/sys commit hash to 97ca703 (#1059)
* *(deps)* Update golang.org/x/crypto commit hash to 4570a08 (#1062)
* *(deps)* Update golang.org/x/sys commit hash to 798191b (#1061)
* *(deps)* Update golang.org/x/sys commit hash to af8b642 (#1063)
* *(deps)* Update module github.com/spf13/viper to v1.10.0 (#1064)
* *(deps)* Update golang.org/x/sys commit hash to 03aa0b5 (#1067)
* *(deps)* Update golang.org/x/sys commit hash to 3b038e5 (#1068)
* *(deps)* Update module github.com/spf13/cobra to v1.3.0 (#1070)
* *(deps)* Update golang.org/x/sys commit hash to 4825e8c (#1071)
* *(deps)* Update module github.com/spf13/viper to v1.10.1 (#1072)
* *(deps)* Update golang.org/x/crypto commit hash to e495a2d (#1073)
* *(deps)* Update golang.org/x/sys commit hash to 4abf325 (#1074)
* *(deps)* Update golang.org/x/sys commit hash to 1d35b9e (#1075)
* *(deps)* Update module github.com/magefile/mage to v1.12.0 (#1076)
* *(deps)* Update module github.com/magefile/mage to v1.12.1 (#1077)
* *(deps)* Update module github.com/getsentry/sentry-go to v0.12.0 (#1079)
* *(deps)* Update module github.com/swaggo/swag to v1.7.8 (#1080)
* *(deps)* Update module github.com/spf13/afero to v1.7.0 (#1078)
* *(deps)* Update module github.com/spf13/afero to v1.7.1 (#1081)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.10 (#1082)
* *(deps)* Update module github.com/spf13/afero to v1.8.0 (#1083)
* *(deps)* Update module github.com/labstack/echo/v4 to v4.6.2 (#1084)
* *(deps)* Update module github.com/labstack/echo/v4 to v4.6.3 (#1089)
* *(deps)* Update golang.org/x/sys commit hash to a018aaa (#1088)
* *(deps)* Update golang.org/x/sys commit hash to 5a964db (#1090)
* *(deps)* Update golang.org/x/crypto commit hash to 5e0467b (#1091)
* *(deps)* Update golang.org/x/sys commit hash to da31bd3 (#1093)
* *(deps)* Update module github.com/prometheus/client_golang to v1.12.0 (#1094)
* *(deps)* Update golang.org/x/crypto commit hash to e04a857 (#1097)
* *(deps)* Update golang.org/x/crypto commit hash to aa10faf (#1098)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.11 (#1099)
* *(deps)* Update golang.org/x/crypto commit hash to 198e437 (#1100)
* *(deps)* Update golang.org/x/sys commit hash to 99c3d69 (#1101)
* *(deps)* Update module github.com/prometheus/client_golang to v1.12.1 (#1102)
* *(deps)* Update klakegg/hugo docker tag to v0.92.0 (#1103)
* *(deps)* Update klakegg/hugo docker tag to v0.92.1 (#1104)
* *(deps)* Update golang.org/x/crypto commit hash to 30dcbda (#1105)
* *(deps)* Update module github.com/swaggo/swag to v1.7.9 (#1106)
* *(deps)* Update golang.org/x/sys commit hash to 1c1b9b1 (#1107)
* *(deps)* Update module github.com/spf13/afero to v1.8.1 (#1108)
* *(deps)* Update golang.org/x/sys commit hash to 5739886 (#1110)
* *(deps)* Update golang.org/x/crypto commit hash to 20e1d8d (#1111)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.5 (#1112)
* *(deps)* Update golang.org/x/crypto commit hash to bba287d (#1113)
* *(deps)* Update golang.org/x/crypto commit hash to dad3315 (#1114)
* *(deps)* Update module github.com/golang-jwt/jwt/v4 to v4.3.0 (#1117)
* *(deps)* Update golang.org/x/sys commit hash to 3681064 (#1116)
* *(deps)* Update golang.org/x/crypto commit hash to db63837 (#1115)
* *(deps)* Update golang.org/x/crypto commit hash to f4118a5 (#1118)
* *(deps)* Update golang.org/x/crypto commit hash to 8634188 (#1121)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.6 (#1122)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.7 (#1123)
* *(deps)* Update module github.com/swaggo/swag to v1.8.0 (#1124)
* *(deps)* Update golang.org/x/sys commit hash to 0005352 (#1125)
* *(deps)* Update golang.org/x/sys commit hash to f242548 (#1126)
* *(deps)* Update klakegg/hugo docker tag to v0.92.2 (#1127)
* *(deps)* Update golang.org/x/sys commit hash to dbe011f (#1129)
* *(deps)* Update golang.org/x/sys commit hash to 95c6836 (#1130)
* *(deps)* Update golang.org/x/oauth2 commit hash to ee48083 (#1128)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.12 (#1132)
* *(deps)* Update golang.org/x/sys commit hash to 4e6760a (#1131)
* *(deps)* Update golang.org/x/image commit hash to 723b81c (#1133)
* *(deps)* Update module github.com/labstack/echo/v4 to v4.7.0 (#1134)
* *(deps)* Update klakegg/hugo docker tag to v0.93.0 (#1135)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.8 (#1136)
* *(deps)* Update klakegg/hugo docker tag to v0.93.2 (#1137)
* *(deps)* Update golang.org/x/sys commit hash to 22a9840 (#1138)
* *(deps)* Update golang.org/x/crypto commit hash to efcb850 (#1139)
* *(deps)* Update golang.org/x/oauth2 commit hash to 6242fa9 (#1140)
* *(deps)* Update golang.org/x/sys commit hash to b874c99 (#1141)
* *(deps)* Update klakegg/hugo docker tag to v0.93.3 (#1142)
* *(deps)* Update module github.com/labstack/echo/v4 to v4.7.1 (#1146)
* *(deps)* Update module github.com/stretchr/testify to v1.7.1 (#1148)
* *(deps)* Update module github.com/swaggo/swag to v1.8.1 (#1156)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.11 (#1143)
* *(deps)* Update module github.com/spf13/cobra to v1.4.0 (#1145)
* *(deps)* Update module github.com/lib/pq to v1.10.5 (#1157)
* *(deps)* Update module github.com/spf13/viper to v1.11.0 (#1159)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.12 (#1162)
* *(deps)* Update module github.com/prometheus/client_golang to v1.12.2 (#1166)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.13 (#1165)
* *(deps)* Update module github.com/coreos/go-oidc/v3 to v3.2.0 (#1164)
* *(deps)* Update module github.com/swaggo/swag to v1.8.2 (#1167)
* *(deps)* Update module github.com/lib/pq to v1.10.6 (#1169)
* *(deps)* Update module gopkg.in/yaml.v3 to v3.0.1 (#1179)
* *(deps)* Update module github.com/imdario/mergo to v0.3.13 (#1178)
* *(deps)* Update module github.com/stretchr/testify to v1.7.2 (#1182)
* *(deps)* Update module github.com/swaggo/swag to v1.8.3 (#1185)
* *(deps)* Update module github.com/spf13/cobra to v1.5.0 (#1192)
* *(deps)* Update module github.com/golang-jwt/jwt/v4 to v4.4.2 (#1193)
* *(deps)* Update module github.com/stretchr/testify to v1.8.0 (#1191)
* *(deps)* Update module github.com/go-testfixtures/testfixtures/v3 to v3.8.0 (#1168)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.14 (#1194)
* *(deps)* Update golang.org/x/term digest to 065cf7b (#1207)
* *(deps)* Update golang.org/x/image digest to 41969df (#1203)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.13 (#1209)
* *(deps)* Update golang.org/x/crypto digest to 0559593 (#1202)
* *(deps)* Update module github.com/spf13/afero to v1.9.0 (#1210)
* *(deps)* Update module github.com/gabriel-vasile/mimetype to v1.4.1 (#1208)
* *(deps)* Update golang.org/x/sync digest to 0de741c (#1205)
* *(deps)* Update github.com/c2h5oh/datasize digest to 859f65c (#1201)
* *(deps)* Update golang.org/x/oauth2 digest to 2104d58 (#1204)
* *(deps)* Update golang.org/x/sys digest to c0bba94 (#1206)
* *(deps)* Update golang.org/x/oauth2 digest to c8730f7 (#1214)
* *(deps)* Update module github.com/spf13/afero to v1.9.2 (#1215)
* *(deps)* Update module github.com/swaggo/swag to v1.8.4 (#1216)
* *(deps)* Update module github.com/spf13/viper to v1.12.0 (#1180)
* *(deps)* Update golang.org/x/sys digest to 1609e55 (#1217)
* *(deps)* Update module github.com/go-testfixtures/testfixtures/v3 to v3.8.1 (#1226)
* *(deps)* Update module go to 1.18 (#1225)

### Documentation
* Add docker-compose example with no proxy ([4255bc3](4255bc3a945b6fe4314e3cd3f62908dd1be1ff4a))
* Add another youtube tutorial ([dbd6f36](dbd6f36da6e56355993cc1411379997e26c88b36))
* Fix api url in docker examples without a proxy ([68998e9](68998e90a446569869fb150bd5fc0739f496b066))
* Make sure all links to vikunja pages are https ([cc612d5](cc612d505f22e5d895b6ebda61fe62498634cec5))
* Update backup instructions ([4829c89](4829c899400544ad27cacfb7d19b40988302a413))
* Add postgres to docker-compose examples ([2aea169](2aea1691cf33b7d9e03fbe2c711af7d8f76d9724))
* Improve development docs ([9bf32aa](9bf32aae99a7e69cce0cd4477e8fc8ddcaea25ea))
* Add another tutorial link ([1fa74cb](1fa74cba6407c2b694b14f8439f1492476433d62))
* Improve wording for systemd ([13561f2](13561f211493903b17c856b3010345ea9df725d4))
* Update testing ([da318e3](da318e3db15121ba864db8450a76ba9ed18b9fd5))
* Add guide for Synology NAS ([049ae39](049ae39c62079f77921b7a9fad5023b2c1c0c1c5))


### Features

* *(docs)* Add details of using NGINX Proxy Manager to the Reverse Proxy docs (#13)
* *(docs)* Add versions explanation
* *(mail)* Don't try to authenticate when no username and password was provided* Add better error logs for mage commands ([bb086eb](bb086eb9f87669f844c283d42ea9ca9f3f5a7877))
* Expose if task comments are enabled or not in /info ([ae8db17](ae8db176db57fa6176e00b87924f70352332ca66))
* Improve account deletion email grammar (#1006) ([dcb52c0](dcb52c00f1c6b3217e2b508d7799fc83adb3b055))
* Add more debug logging when deleting users ([8f55af0](8f55af07c936218487ec94e65c6673fbddd0cdb5))
* Don't require a password for data export from users authenticated with third-party auth ([9eca971](9eca971c938699d481915fb6e14c765aea1fa3b5))
* Expose if a user is a local user through its jwt token ([516c812](516c812043e77be7f834ae1326d13d39e156ef77))
* Expose if a user is a local user through the /user endpoint ([2683ef2](2683ef23d538eb846d5d799798fa82cca70dc017))
* Enable rate limit for unauthenticated routes ([093d0c6](093d0c65ca6338358dbd1df904daadd7808f2817))
* Use wallpaper topic for default unsplash background list ([88a2ced](88a2cede19f1844814530af948c3cc5a0b026419))
* Gravatar - Lowercase emails before MD5 hash (#10) ([36bf3d2](36bf3d216a7be28e917e2816a9e5da43439f2c20))
* Add marble avatar (#1060) ([73ee696](73ee696fc3cf941af2d2c2cf81224aa01f93234e))
* Save user language in the settings ([a98119f](a98119f2d670a11efab6008129b767f9208f8113))
* Add time zone setting for reminders (#1092) ([61d49c3](61d49c3a56a59e52ce407b858ddd4aa573dbee9d))
* Add long-lived api tokens (#1085) ([1322cb1](1322cb16d76a40ad90631e3e091da0f0d44957a9))
* Upgrade golangci-lint to 1.45.2 ([5cf263a](5cf263a86f954a38cbfafb6b0857bf591f82a811))
* Add date math for filters (#1086) ([0a1d8c9](0a1d8c940410b03a78016ac6110883ca05484816))
* Add migration to create BlurHash strings for all list backgrounds ([362706b](362706b38d52720b5a1615e185a985b7708168f7))
* Generate a BlurHash when uploading a new image ([f83b09a](f83b09af59ed25425a16824ccf48d903c81e861a))
* Save BlurHash from unsplash when selecting a photo from unsplash ([2ec7d7a](2ec7d7a8a85cc12c07d20cfab9b90a78a7857eb6))
* Return BlurHash for unsplash search results ([6df8658](6df865876df961f2bec476126bf6e7fbe5d43e0e))
* Add caldav tokens (#1065) ([e4b50e8](e4b50e84a44f809cc829c2fdb6f52b03b40a367b))
* Ability to serve static files (#1174) ([acaa850](acaa85083f2bebbc67608ae0f454ed5e9a3ef8a0))
* Restrict max avatar size ([2f25b48](2f25b48869f59256bf7d692c4486c64c30b85e5e))
* Send overdue tasks email notification at 9:00 in the user's time zone ([7eb3b96](7eb3b96a4465ca6648572b07c506c06f2c28c375))
* Add setting to change overdue tasks reminder email time ([8869adf](8869adfc276f674b686bf68f949d7efbb417e55b))
* Allow only the authors of task comments to edit them ([01271c4](01271c4c0111b3b040dcb9a0d502d31078ad6d4b))
* Migrate away from gomail ([30e0e98](30e0e98f7738e36698990523377f47edcbf6806c))
* Embed the vikunja logo as inline attachment ([f4f8450](f4f8450d166f1a836eea202dd0340d2156d3dfe9))
* Use embed fs directly to embed the logo in mails ([73c4c39](73c4c399e5d610bb713f1e9feab543e0425ee959))
* Use actual uuids for tasks ([62325de](62325de9cd5da5b70987081956a28e7baa907081))
* Add issue template ([117f6b3](117f6b38e1d35c09f2657975ea75dcfedcd8425d))


### Miscellaneous Tasks

* *(ci)* Use latest version of s3 plugin
* *(ci)* Sign drone config
* *(docs)* Update docs about compiling from source
* *(docs)* Redirect properly from /docs/docs
* *(docs)* Add new mailer option to docs
* *(docs)* Clarify openid setup with environment variables
* *(docs)* Add frontendurl to all example configs
* *(mage)* Don't set api packages when they are not used* Sign drone config ([1d8d0f1](1d8d0f140e4f2a59947167bd597e5f12b84b009d))
* Cleanup namespace creation ([b60c69c](b60c69c5a8c004a780b989cf0bb8ab6455086b0f))
* Generate swagger docs ([ba2bdff](ba2bdff39109db9ecc4b525e39e2642b41ac03b8))
* Go mod tidy ([726a517](726a517bec731f1af8e3186e280718fef02cadf7))
* Upgrade trello api wrapper and remove fork ([7e99618](7e99618319547c7e7dfa2cc063f654300f7074fb))
* Use our custom build image to build docker image ([251b877](251b877015761fdd2b8dbd18cd8ec696dc374103))
* Update golangci-lint ([430057a](430057a404b04e75c62a15693f479c6fc8e63189))


### Other

* *(other)* Healthcheck endpoint (#998)
* *(other)* Added the ability to configure the JWT expiry date using a new server.jwtttl config parameter. (#999)
* *(other)* Enable a list to be moved across namespaces (#1096)
* *(other)* A bunch of dependency updates at once (#1155)
* *(other)* Add client-cert parameters of the Go pq driver to the Vikunja config (#1161)
* *(other)* Add exec to run script to run app as PID 1 (#1200)

## [0.18.1] - 2021-09-08

### Fixed

* Docs: Add another third-party tutorial link
* Don't try to export items which do not have a parent
* fix(deps): update golang.org/x/sys commit hash to 6f6e228 (#970)
* fix(deps): update golang.org/x/sys commit hash to c212e73 (#971)
* Fix exporting tasks from archived lists
* Fix lint
* Fix tasks not exported
* Fix tmp export file created in the wrong path

## [0.18.0] - 2021-09-05

### Added

* Add default list setting (#875)
* Add menu link to Vikunja Cloud in docs
* Add more logging and better error messages for openid authentication + clarify docs
* Add more logging for test data api endpoint
* Add searching for tasks by index
* Add setting for first day of the week
* Add support of Unix socket (#912)
* Add truncate parameter to test fixtures setup
* Notify the user after three failed login attempts
* Reorder tasks, lists and kanban buckets (#923)
* Send a notification on failed TOTP
* Task mentions (#926)
* Try to get more information about the user when authenticating with openid
* User account deletion (#937)
* User Data Export and import (#967)

### Changed

* Allow running migration 20210711173657 multiple times to fix issues when it didn't completely run through previously
* Better logging for errors while importing a bunch of tasks
* Change task title to TEXT instead of varchar(250) to allow for longer task titles
* Disable the user account after 10 failed password attempts
* Docs: Add a note about default password
* Docs: Add another youtube tutorial
* Docs: Add ios to the list of not working caldav clients
* Docs: Add k8s-at-home Helm Chart for Vikunja
* Docs: Add other installation resources
* Docs: Add translation docs
* Docs: Fix rewrite rules in apache example configs
* Docs: Translation now happening at crowdin
* Docs: Update translation guidelines
* Don't fail when removing the last bucket in migration from other services
* Don't notify the user who created the team
* Don't use the mariadb root user in docker-compose examples
* Ensure case insensitive search on postgres (#927)
* Increase test timeout
* Only filter out failing openid providers if multiple are configured and one of them failed
* Only send an email about failed totp after three failed attempts
* Rearrange setting frontend url in config
* Refactor user email confirmation + password reset handling (#919)
* Rename and sign drone config
* Replace jwt-go with github.com/golang-jwt/jwt
* Reset failed totp attempts when logging in successfully
* Save user tokens as text and not varchar
* Save user tokens as varchar(450) and not text to fix mysql indexing issues
* Set todoist migration redirect url to the frontend url by default
* Show config full paths and env variables with config options
* Switch the :latest docker image tag to contain the latest release instead of the latest unstable
* Tune test db server settings to speed up tests (#939)

### Fixed

* Fix authentication callback
* Fix duplicating empty lists
* Fix error handling when deleting an attachment file
* Fix error when searching for a namespace returned no results
* Fix error when searching for a namespace with subscribers
* Fix goimports
* Fix importing archived projects and done items from todoist
* Fix jwt middleware
* Fix lint
* Fix mapping task priorities from Vikunja to calDAV
* Fix moving the done bucket around
* Fix old references to master in docs
* Fix panic on invalid smtp config
* Fix parsing openid config when using a json config file
* Fix saving pointer values to memory keyvalue
* Fix saving reminders of repeating tasks
* Fix setting a saved filter as favorite
* Fix setting task favorite status of related tasks
* Fix setting up keyvalue storage in tests
* Fix swagger docs for create requests
* Fix task relations not getting properly cleaned up when deleting them
* Fix tests & lint
* Make sure a bucket exists or use the default bucket when importing tasks
* Make sure all associated entities of a task are deleted when the task is deleted
* Make sure list / task favorites are set per user, not per entity (#915)
* Make sure the configured frontend url always has a / at the end
* Refactor & fix storing struct-values in redis keyvalue
* Todoist migration: don't panic if no reminder was found for task

### Dependency updates

* fix(deps): update golang.org/x/sys commit hash to 63515b4 (#959)
* fix(deps): update golang.org/x/sys commit hash to 97244b9 (#965)
* fix(deps): update golang.org/x/sys commit hash to f475640 (#962)
* fix(deps): update golang.org/x/sys commit hash to f4d4317 (#961)
* fix(deps): update module github.com/lib/pq to v1.10.3 (#963)
* Update alpine Docker tag to v3.13 (#884)
* Update alpine Docker tag to v3.14 (#889)
* Update golang.org/x/crypto commit hash to 0a44fdf (#944)
* Update golang.org/x/crypto commit hash to 0ba0e8f (#943)
* Update golang.org/x/crypto commit hash to 32db794 (#949)
* Update golang.org/x/crypto commit hash to 5ff15b2 (#891)
* Update golang.org/x/crypto commit hash to a769d52 (#916)
* Update golang.org/x/image commit hash to 775e3b0 (#880)
* Update golang.org/x/image commit hash to a66eb64 (#900)
* Update golang.org/x/image commit hash to e6eecd4 (#893)
* Update golang.org/x/net commit hash to 37e1c6af
* Update golang.org/x/oauth2 commit hash to 14747e6 (#894)
* Update golang.org/x/oauth2 commit hash to 2bc19b1 (#955)
* Update golang.org/x/oauth2 commit hash to 6f1e639 (#931)
* Update golang.org/x/oauth2 commit hash to 7df4dd6 (#952)
* Update golang.org/x/oauth2 commit hash to a41e5a7 (#902)
* Update golang.org/x/oauth2 commit hash to a8dc77f (#896)
* Update golang.org/x/oauth2 commit hash to bce0382 (#895)
* Update golang.org/x/oauth2 commit hash to d040287 (#888)
* Update golang.org/x/oauth2 commit hash to f6687ab (#862)
* Update golang.org/x/oauth2 commit hash to faf39c7 (#935)
* Update golang.org/x/sys commit hash to 00dd8d7 (#953)
* Update golang.org/x/sys commit hash to 15123e1 (#946)
* Update golang.org/x/sys commit hash to 1e6c022 (#947)
* Update golang.org/x/sys commit hash to 30e4713 (#945)
* Update golang.org/x/sys commit hash to 41cdb87 (#956)
* Update golang.org/x/sys commit hash to 7d9622a (#948)
* Update golang.org/x/sys commit hash to bfb29a6 (#951)
* Update golang.org/x/sys commit hash to d867a43 (#934)
* Update golang.org/x/sys commit hash to e5e7981 (#933)
* Update golang.org/x/sys commit hash to f52c844 (#954)
* Update golang.org/x/term commit hash to 6886f2d (#887)
* Update module getsentry/sentry-go to v0.11.0 (#869)
* Update module github.com/coreos/go-oidc to v3 (#885)
* Update module github.com/gabriel-vasile/mimetype to v1.3.1 (#904)
* Update module github.com/golang-jwt/jwt to v3.2.2 (#928)
* Update module github.com/golang-jwt/jwt to v4 (#930)
* Update module github.com/go-redis/redis/v8 to v8.11.0 (#903)
* Update module github.com/go-redis/redis/v8 to v8.11.1 (#925)
* Update module github.com/go-redis/redis/v8 to v8.11.2 (#932)
* Update module github.com/go-redis/redis/v8 to v8.11.3 (#942)
* Update module github.com/iancoleman/strcase to v0.2.0 (#918)
* Update module github.com/labstack/echo/v4 to v4.4.0 (#917)
* Update module github.com/labstack/echo/v4 to v4.5.0 (#929)
* Update module github.com/mattn/go-sqlite3 to v1.14.8 (#921)
* Update module github.com/spf13/cobra to v1.2.0 (#905)
* Update module github.com/spf13/cobra to v1.2.1 (#906)
* Update module github.com/spf13/viper to v1.8.0 (#890)
* Update module github.com/spf13/viper to v1.8.1 (#899)
* Update module github.com/swaggo/swag to v1.7.1 (#936)
* Update module github.com/yuin/goldmark to v1.3.8 (#892)
* Update module github.com/yuin/goldmark to v1.3.9 (#901)
* Update module github.com/yuin/goldmark to v1.4.0 (#908)
* Update module go-redis/redis/v8 to v8.10.0 (#882)
* Update module go-redis/redis/v8 to v8.7.1 (#807)
* Update module go-testfixtures/testfixtures/v3 to v3.6.1 (#868)
* Update module lib/pq to v1.10.2 (#865)
* Update module prometheus/client_golang to v1.11.0 (#879)
* Update module yuin/goldmark to v1.3.6 (#863)
* Update module yuin/goldmark to v1.3.7 (#867)
* Update monachus/hugo Docker tag to v0.75.1 (#940)

## [0.17.1] - 2021-06-09

### Fixed

* Fix parsing openid config when using a json config file

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
* Make the debian repo structure for buster instead of stretch
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
* Add separate docker pipeline for amd64 and arm
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

* Add 2fa for authentication (#383)
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
* Update xorm redis cacher to use the xorm logger instead of a special separate one
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
* Better efficiency for loading teams (#128)
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
* Fixed a bug where adding assignees or reminders via an update would re-create them and not respect already inserted
  ones, leaving a lot of garbage
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

* Better CalDAV support (#73)
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
* Properly init tables Redis
* unexpected EOF when using metrics (#35)
* Task sorting in lists (#36)
* Various user fixes (#38)
* Fixed a bug where updating a list would update it with the same values it had

### Changed

* Simplified list rights check (#50)
* Refactored some structs to not expose unneeded values via json (#52)

### Misc

* Updated libraries
* Updated drone to version 1
* Releases are now signed with our pgp key (more info about this
  on [the download page](https://vikunja.io/en/download/)).

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
