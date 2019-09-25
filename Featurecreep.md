# Featurecreep

This is the place where I write down ideas to work on at some point. 
Sorry for some of them being in German, I'll tranlate them at some point.

## Feature Ideas

* [x] Priorities
* [x] Repeating tasks
* [x] Get all tasks which are due between two given dates
* [x] Subtasks

## Anderes

* [x] Refactor!!!! Delete everything not being used anymore, simplify. 
* [x] Drone
* [x] Tests
* [x] Find a nme
* [x] Move packages to a better structure
* [x] Swagger UI
+ [x] Fix CORS
* [x] Use echo.NewHTTPError instead of c.JSON(Message{})
* [x] Better error messages when the model which is sent to the server is wrong
* [x] Better error handling to show useful error messages and status codes
* [x] Viper for config instead of ini
* [x] Docs for installing
* [x] Tests for rights managemnt
* [x] Rights checks:
      * [x] Create lists
      * [x] Edit lists
      * [x] Add tasks
      * [x] Edit tasks
* [x] The -1 namespace should also be accessible seperately

### Short Term

* [x] Cacher configurable
* [x] Should throw an error when an id < 1
* [x] /users should also return the rights
* [x] Extra endpoint /teams/members /list/users to update rights without needing to remove and re-add them
* [x] namespaces & listen update does not work, returns 500
* [x] Logging for all errors somewhere
* [x] Ne extra funktion für list exists machen, damit die nicht immer über GetListByID gehen, um sql-abfragen zu sparen
* [x] Rausfinden warum xorm teilweise beim einfügen IDs mit einfügen will -> Das schlägt dann wegen duplicate fehl
* [x] Bei den Structs "AfterLoad" raus, das verbraucht bei Gruppenabfragen zu viele SQL-Abfragen -> Die sollen einfach die entsprechenden Read()-Methoden verwenden (Krassestes bsp. ist GET /namespaces mit so ca 50 Abfragen)
* [x] General search endpoints
* [x] Validation der ankommenden structs, am besten mit https://github.com/go-validator/validator oder mit dem Ding von echo
* [x] Pagination
	* Sollte in der Config definierbar sein, wie viel pro Seite angezeigt werden soll, die CRUD-Methoden übergeben dann ein "gibt mir die Seite sowieso" an die CRUDable-Funktionenen, die müssen das dann Auswerten. Geht leider nicht anders, wenn man erst 2342352 Einträge hohlt und die dann nachträglich auf 200 begrenzt ist das ne massive Ressourcenverschwendung.
* [x] Testen, ob man über die Routen methode von echo irgendwie ein swagger spec generieren könnte -> Andere Swagger library
* [x] CalDAV
  * [x] Basics
  * [x] Reminders
  * [x] Discovery, stichwort PROPFIND 
* [x] Wir brauchen noch ne gute idee, wie man die listen kriegt, auf die man nur so Zugriff hat (ohne namespace)
    * Dazu am Besten nen pseudonamespace anlegen (id -1 oder so), der hat das dann alles
* [x] Testing mit locust: https://locust.io/
* [x] Endpoint to get all users who have access to a list - regardless of via team, user share or via namespace

#### Userstuff

* [x] Userstuff aufräumen
	-> Soweit es geht und Sinnvoll ist auf den neuen Handler umziehen
		-> Login/Register/Password-reset geht natürlich nicht
		-> Bleibt noch Profile abrufen und Einstellungen -> Macht also keinen Sinn das auf den neuen Handler umzuziehen
* [x] Email-Verifizierung beim Registrieren
* [x] Password Reset
* [x] New field in user model which holds a url of an avatar image - for now just Gravatar, later more
* [ ] Settings
  * [ ] Password update
  * [ ] Email update
  * [ ] Ob man über email oder Benutzernamen gefunden werden darf

### Bugfixes

* [x] Panic wenn mailer nicht erreichbar -> Als workaround mailer deaktivierbar machen, bzw keine mails verschicken
* [x] "unexpected EOF"
* [x] Beim Login & Password reset gibt die API zurück dass der Nutzer nicht existiert
* [x] Re-check rights checks to see if all information which is compared against is properly read from the db and not only based on user input
  * [x] Lists
  * [x] List users
  * [x] List Teams
  * [x] Labels
  * [x] Tasks
  * [x] Namespaces
  * [x] Namespace users
  * [x] Namespace teams
  * [x] Teams
  * [x] Team member handling
  * [x] Also check `ReadOne()` for unnessecary database operations since the inital query is already done in `CanRead()`
* [x] Add a `User.AfterLoad()` which obfuscates the email address
* [x] Fix priority not updating to 0

### Docs

* [x] Readme
  * [x] Auch noch nen "link" zum Featurecreep
  * [x] ToC
  * [x] Logo
  * [x] How to build -> Docs
  * [x] How to dev -> Docs
  * [x] License
  * [x] Contributing
* [x] Redocs
* [x] Swaggerdocs verbessern
  * [x] Descriptions in structs
  * [x] Maxlength specify etc. (see swaggo docs)
* [x] Rights
* [x] API
* [x] Anleitung zum Makefile
* [x] How to build from source
* [x] Struktur erklären
* [x] Deploy in die docs
  * [x] Docker
  * [x] Native (systemd + nginx/apache)
* [x] Backups
* [x] Docs aufsetzen

#### Better caldav support

* [x] VTODO
* [x] Fix organizer prop
* [x] Depricate the events thing for now
* [x] PROPFIND/OPTIONS : caldav discovery
* [x] Create new tasks
* [x] Save uid from the client
* [x] Update tasks
* [x] Marking as done
* [x] Fix OPTIONS Requests to the rest of the api being broken
* [x] Parse all props defined in rfc5545
* [x] COMPLETED -> Need to actually save the time the task was completed
* [x] Whenever a task ist created/updated, update the `updated` prop on the list so the etag changes and clients get notified
* [x] Fix not all tasks being displayed (My guess: Something about that f*cking etag)
* [x] Delete tasks
* [x] Last modified
* [x] Content Size
* [x] Modify the caldav lib as proposed in the pr
* [x] Improve login performance, each request taking > 1.5 sec is just too much, maybe just use the default value for hash iterations in the login/register function
* [x] Only show priority when we have one
* [x] Show a proper calendar title
* [x] Fix home principal propfind stuff
* [x] Docs
* [x] Setting to disable caldav completely
* [ ] Make it work with the app
* [ ] Cleanup the whole mess I made with the handlers and storage providers etc -> Probably a good idea to create a seperate storage provider etc for lists and tasks
* [ ] Tests
* [ ] Check if only needed things are queried from the db when accessing dav (for ex. no need to get all tasks when we act

### General features

* [x] Deps nach mod umziehen
* [x] Performance bei rechtchecks verbessern
  * User & Teamright sollte sich für n rechte in einer Funktion testen lassen
* [x] Colors for tasks
* [x] /info endpoint, in dem dann zb die limits und version etc steht
* [x] Bindata for templates
* [x] User struct should have a field for the avatar url (-> gravatar md5 calculated by the backend)
* [x] Middleware to have configurable rate-limiting per user
* [ ] Colors for lists and namespaces -> Up to the frontend to implement these
* [ ] Reminders via mail
* [ ] Be able to "really" delete the account -> delete all lists and move ownership for others
* [ ] All `ReadAll` methods should return the number of items per page, the number of items on this page, the total pages and the items
      -> Check if there's a way to do that efficently. Maybe only implementing it in the web handler.
* [ ] Move lists between namespaces -> Extra endpoint

### Infra 

* [ ] Debian package should have a service file
* [ ] Downloads should be served via nginx (with theme?), minio should only be used for pushing artifacts.

### Refactor 

* [x] ListTaskRights, sollte überall gleich funktionieren, gibt ja mittlerweile auch eine Methode um liste von nem Task aus zu kriegen oder so
* [x] Re-check all `{List|Namespace}{User|Team}` if really all parameters need to be exposed via json or are overwritten via param anyway.
* [x] Things like list/task order should use queries and not url params
* [x] Fix lint errors
* [x] Add settings for max open/idle connections and max connection lifetime
* [x] Reminders should use an extra table so we can make reverse lookups aka "give me all tasks with reminders in this period" which we'll need for things like email reminders notifications
* [x] When giving a user access to a list/namespace, they should be reffered to by uuid, not numeric id
* [x] Adding users to a team should also use uuid
* [x] Check if the team/user really exist before updating them on lists/namespaces
* [x] Refactor config handling: Custom type "key" or so which holds the viper const and then mixins on that type to get the values from viper
* [x] Less files, but with some kind of logic
* [x] Have extra functions for logging to call so it is possible to call `log.Info` instead of `log.Log.Info`  
* [x] `GetUserByID` and the likes should return pointers
* [x] `ListTask` should be just `Task`
* [x] Move `teams_{namespace|list}_*` to `{namespace|list}_teams_*` for better consistency

### Linters

* [x] goconst
* [x] Staticcheck
* [x] gocyclo-check
* [ ] gosec-check -> waiting for mod
* [x] goconst-check
* [ ] golangci -> docker in drone, will probably make all other linters obsolete

### Tasks

* [x] Start/Enddatum für Tasks
* [x] Timeline/Calendar view -> Dazu tasks die in einem Bestimmten Bereich due sind, macht dann das Frontend
* [x] Tasks innerhalb eines definierbarem Bereich, sollte aber trotzdem der server machen, so à la "Gib mir alles für diesen Monat"
* [x] Bulk-edit -> Transactions
* [x] Assignees
  * [x] Check if something changed at all before running everything
  * [x] Don't use `list.ReadOne()`, gets too much unnessecary shit
  * [x] Wegen Performance auf eigene endpoints umziehen, wie labels
  * [x] "One endpoint to rule them all" -> Array-addable
* [x] Labels
  * [x] Check if something changed at all before running everything
  * [x] Editable via task edit, like assignees
  * [x] "One endpoint to rule them all" -> Array-addable
* [x] Description should be longtext
* [x] Percent done - For now just a float, may later depend on how many sub tasks are done or so
* [ ] Attachments
* [ ] Related tasks -> settable with a "kind" of relation like blocked, subtask, or just related or so
  * [x] Should be possible to relate tasks which are not in the same list
  * [ ] Replace the old subtask implementation
  * New Struct for the relation
  * Endpoint to get all full related tasks for a task
  * When using task.ReadOne() or ReadAll() only get the relation kind + title etc, not everything
  * Only check if the user has the right to see the task when creating it, not when viewing it -> put a warning in frontend
  * Relation Kinds (for now): Subtask/Parent Task (from subtask, no extra relation), Related, later also Blocked/Blocking
  * Should also be possible to create a new task with relations directly
  * For everything else dedicated endpoints to manage relations
* [ ] Move tasks between lists
* [ ] "Status" field for things like "New", "In Progress", "Done", etc (customizable statuses)

#### Events

* [ ] Whenever something happens an event should be registered in the db, sse, mail, etc. -> Abstract with implementations for various things 
* [ ] Activity Feed, so à la "der und der hat das und das gemacht etc"
  * [ ] Per list
  * [ ] For the current user
* [ ] ~~Websockets~~ SSE https://github.com/kljensen/golang-html5-sse-example
  * User authenticates (with jwt)
  * When updating/creating/etc an event struct is sent to the broker
  * The broker has a list of subscribed users
  * It then checks who is allowed to the see the event it recieved and sends it
  * [ ] Being able to define filters for notifications or turn them silent completely -> Probably frontend only
* [ ] Settings for which notifications to recieve over which channel
  * [ ] Different time-based filters
* [ ] Subscriptions to tasks or whole project/namespace

### More server settings

* [x] Caldav disable/enable
* [ ] Assignees disable/enable
* [ ] Max number of assignees
* [ ] List/Namespace limits
* [ ] Attachements disable/enable
* [ ] Attachements size
* [ ] Templates disable/enable
* [ ] Stats disable/enable
* [ ] Activity notifications disable/enable
* [ ] IMAP integration disable/enable
* [ ] Reminders via mail disable/enable

### Later

* [ ] Public lists
* [ ] Sorting lists by members, tasks, teams, last modified, etc (all possible fields)
* [ ] Backgrounds for lists -> needs uploading and storing and so on
* [ ] Plugins
* [ ] Rename Namespaces to collections (or spaces?)? Or discard the whole namespace concept alltogether since it may be too confusing? -> Projects only
* [ ] Collections n-n (one list can be in multiple collections)?
* [ ] Rename lists to projects
* [ ] "Performance-Statistik" -> Wie viele Tasks man in bestimmten Zeiträumen so geschafft hat etc
* [ ] List stats to see how many tasks are done, how many are there in total, how many people have acces to a list etc
* [ ] IMAP-Integration -> Man schickt eine email an Vikunja und es macht daraus dann nen task -> Achtung missbrauchsmöglichkeiten
* [ ] In und Out webhooks, mit Templates vom Payload
* [ ] Some kind of milestones for tasks
* [ ] Get Weekly/Daily summary of the tasks to come per mail
* [ ] Boards (Boards)
* [ ] Sprints
* [ ] Epics
* [ ] Stories
  * Maybe we can do both epics and stories with a new field "kind" on tasks, since epics and stories are basically high-level-overview tasks
  * Needs proper subtask stuff / relatsions
* [ ] Per-User limits of lists/namespaces
* [ ] Admin-Interface to do stuff like settings and user management
  * [ ] Enable/Disable users
  * [ ] Manage user groups -> Creating new roles and defining what they're allowed to do etc.
  * [ ] Manage custom fields task templates
  * [ ] Enable/disable allowing user adding to lists/namespaces for specific lists or namespaces
  * [ ] Admins should be able to see and mange all the boards
  * [ ] Admin interface also usable via cli
* [ ] `dump` and `restore` cli commands
* [ ] Better rights system with user roles, to be able to manage fine-graded permissions on fields -> Giving read/write right for each field/action
  * [ ] Roles which enable or disable chaning certain fields of a task -> includes custm fields
  * [ ] Pre-defined roles to make it easier to set up (creatable via admin)
* [ ] Limit registration to users with a defined email domain
* [ ] Close the instance, either no registration or only one with defined email
* [ ] 2fa
* [ ] Custom fields for tasks: Templates at List > Namespace > Global level, overwriting each other
* [ ] Task-Templates in namespaces and lists (-> Multiple which are selectable)
  * The frontend would then GET the template, offer them in a drop down kind of thing and prefill the values, the api won't do anything 
    about creating a task based on a template except providing it to the frontend
* [ ] "Favourite lists" -> A user can favourize boards which will then show up in a pseudonamespace
* [ ] Internal lists -> Only registered users can see the list, but without explicit sharing
* [ ] Rights management for both public and internal lists
* [ ] Export all data from Vikunja to json (related: `dump` cli command)
  * [ ] Per user and for the whole instance
* [ ] Watch a (n internal) list -> Will get notification for everything
  * [ ] Automatically subscribe when mentioning/commenting/assigning a user
* [ ] Archive a task/project/namespace instead of deleting -> When archived, everything is read only
* [ ] Task dependencies -> Can be implemented by using relations
* [ ] Time tracking (possible plugin)
* [ ] IFTTT
* [ ] Task automation: Being able to create own flows "if this event is fired, do this" and so on
* [ ] User online status -> Maybe just take that from metrics?
* [ ] Add new users via to a list which don't have an account yet, they'd get a link to sign up for vikunja.
  * [ ] Respect registration email domain limits
* [ ] More sharing features (all of these with the already existing permissions)
  * [x] Share a link with/without password
  * [ ] Invite users per mail -> can easily be based off the link sharing stuff
* [ ] Comments on tasks
* [ ] @mention users in tasks or comments to get them notified
* [ ] Summary of tasks to do in a configurable interval (every day/week or so)
* [ ] Disable/enable task fields in a list
  * [ ] With inheritence from namespaces
* [ ] Custom statuses for tasks and projects, configurable in the list settings
  * [ ] With inheritence from namespaces
* [ ] (sub-)Project start/end/deadline dates 
* [ ] Better filters
  * [ ] by lables
  * [ ] Due dates
  * [ ] Start/End dates
  * [ ] Assignees
  * [ ] Priorities
  * [ ] Status
* [ ] "Smart Lists", filtered lists, saved in some kind of pseudonamespace
  * [ ] Global and per list
* [ ] Bulk-create lists/tasks -> needed for importer
* [ ] Importer (maybe frontend only)
  * [ ] Trello
  * [ ] Wunderlist
  * [ ] Zenkit
  * [ ] Asana
  * [ ] Microsoft Todo
  * [ ] Nozbe
  * [ ] Lanes
  * [ ] Nirvana
  * [ ] Any.do
  * [ ] Good ol' Caldav (Tasks)
  * [ ] ClickUp
  * [ ] Plutoio
  * [ ] Clubhouse
  * [ ] Roadmap
  * [ ] Joplin
  * [ ] A markdown file like this one
  * [ ] Vikunja json export
* [ ] More auth providers
  * [ ] LDAP/AD
  * [ ] Kerberos
  * [ ] SAML (what?)
  * [ ] smtp
  * [ ] OpenID
