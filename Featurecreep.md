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
* [ ] CalDAV
  * [x] Basics
  * [x] Reminders
  * [ ] Discovery, stichwort PROPFIND 
* [x] Wir brauchen noch ne gute idee, wie man die listen kriegt, auf die man nur so Zugriff hat (ohne namespace)
    * Dazu am Besten nen pseudonamespace anlegen (id -1 oder so), der hat das dann alles
* [x] Testing mit locust: https://locust.io/
* [ ] Endpoint to get all users who have access to a list - regardless of via team, user share or via namespace

#### Userstuff

* [x] Userstuff aufräumen
	-> Soweit es geht und Sinnvoll ist auf den neuen Handler umziehen
		-> Login/Register/Password-reset geht natürlich nicht
		-> Bleibt noch Profile abrufen und Einstellungen -> Macht also keinen Sinn das auf den neuen Handler umzuziehen
* [x] Email-Verifizierung beim Registrieren
* [x] Password Reset -> Link via email oder so
* [ ] Settings
  * [ ] Password update
  * [ ] Email update
  * [ ] Ob man über email oder Benutzernamen gefunden werden darf

### Bugfixes

* [x] Panic wenn mailer nicht erreichbar -> Als workaround mailer deaktivierbar machen, bzw keine mails verschicken
* [x] "unexpected EOF"
* [x] Beim Login & Password reset gibt die API zurück dass der Nutzer nicht existiert
* [ ] Re-check rights checks to see if all information which is compared against is properly read from the db and not only based on user input
  * [ ] Lists
  * [ ] List users
  * [ ] List Teams
  * [ ] Labels
  * [ ] Tasks
  * [ ] Namespaces
  * [ ] Namespace users
  * [ ] Namespace teams
  * [ ] Teams
  * [ ] Team member handling

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
* [ ] Attachments
* [ ] Task-Templates innerhalb namespaces und Listen (-> Mehrere, die auswählbar sind)
* [ ] Ein Task muss von mehreren Assignees abgehakt werden bis er als done markiert wird
* [ ] Besseres Rechtesystem, damit man so fine-graded sachen machen kann wie "Der da darf aber nur Tasks hinzufügen, aber keine abhaken"
  * [ ] Roles which enable or disable chaning certain fields of a task -> includes custm fields
* [ ] Custom fields: Templates at List > Namespace > Global level, overwriting each other
* [ ] Related tasks -> settable with a "kind" of relation like blocked, or just related or so
* [ ] Description should be longtext

### General features

* [x] Deps nach mod umziehen
* [x] Performance bei rechtchecks verbessern
  * User & Teamright sollte sich für n rechte in einer Funktion testen lassen
* [ ] Endpoint um die Rechte mit Beschreibung und code zu kriegen
* [ ] "Smart Lists", Listen nach bestimmten Kriterien gefiltert -> speichern und im pseudonamespace
* [ ] "Performance-Statistik" -> Wie viele Tasks man in bestimmten Zeiträumen so geschafft hat etc
* [ ] IMAP-Integration -> Man schickt eine email an Vikunja und es macht daraus dann nen task -> Achtung missbrauchsmöglichkeiten
* [ ] In und Out webhooks, mit Templates vom Payload
* [ ] Reminders via mail
* [ ] Activity Feed, so à la "der und der hat das und das gemacht etc"
  * [ ] Per list
  * [ ] For the current user
* [ ] ~~Websockets~~ SSE https://github.com/kljensen/golang-html5-sse-example
  * User authenticates (with jwt)
  * When updating/creating/etc an event struct is sent to the broker
  * The broker has a list of subscribed users
  * It then checks who is allowed to the see the event it recieved and sends it
  * [ ] Being able to define filters for notifications or turn them silent completely -> Probably frontend only
* [ ] mgl. zum Emailmaskieren haben (in den Nutzereinstellungen, wenn man seine Email nicht an alle Welt rausposaunen will)
* [ ] Mgl. zum Accountlöschen haben (so richtig krass mit emailverifiezierung und dass alle Privaten Listen gelöscht werden und man alle geteilten entweder wem übertragen muss oder  auf privat stellen)
* [ ] /info endpoint, in dem dann zb die limits und version etc steht
* [ ] Deprecate /namespaces/{id}/lists in favour of namespace.ReadOne() <-- should also return the lists
* [ ] Bindata for templates
* [ ] `GetUserByID` and the likes should return pointers
* [ ] Colors for lists and namespaces -> Up to the frontend to implement these
* [ ] Some kind of milestones for tasks
* [ ] Create tasks from a text/markdown file (probably frontend only)
* [ ] Label-view: Get a bunch of tasks by label
* [ ] Better caldav support (VTODO)
* [ ] Debian package should have a service file
* [ ] Downloads should be served via nginx (with theme?), minio should only be used for pushing artifacts.

### Refactor 

* [x] ListTaskRights, sollte überall gleich funktionieren, gibt ja mittlerweile auch eine Methode um liste von nem Task aus zu kriegen oder so
* [x] Re-check all `{List|Namespace}{User|Team}` if really all parameters need to be exposed via json or are overwritten via param anyway.
* [x] Things like list/task order should use queries and not url params
* [x] Fix lint errors
* [ ] Reminders should use an extra table so we can make reverse lookups aka "give me all tasks with reminders in this period" which we'll need for things like email reminders notifications
* [ ] Teams and users should also have uuids (for users these can be the username)
* [ ] When giving a team or user access to a list/namespace, they should be reffered to by uuid, not numeric id
* [ ] Adding users to a team should also use uuid
* [ ] Check if the team/user really exist before updating them on lists/namespaces

### Fixes

* [ ] Fix priority not updating to 0

### Linters

* [x] goconst
* [x] Staticcheck -> waiting for mod
* [x] gocyclo-check
* [ ] gosec-check -> waiting for mod
* [x] goconst-check -> waiting for mod

### More server settings

* [ ] Caldav disable/enable
* [ ] Assignees disable/enable
* [ ] List/Namespace limits
* [ ] Attachements disable/enable
* [ ] Attachements size
* [ ] Templates disable/enable
* [ ] Stats disable/enable
* [ ] Activity notifications disable/enable
* [ ] IMAP integration disable/enable
* [ ] Reminders via mail disable/enable

### Later

* [ ] Plugins
* [ ] Rename Namespaces?
* [ ] Namespaces to collections and n-n (one list can be in multiple collections)?
* [ ] Per-User limits of lists/namespaces
* [ ] Admin-Interface to do stuff like settings and user management
  * [ ] Enable/Disable users
  * [ ] Better rights, fine-graded
  * [ ] Enable/disable allowing user adding to lists/namespaces for specific lists or namespaces
  * [ ] Admins should be able to see and mange all the boards
* [ ] Limit registration to users with a defined email domain
* [ ] Close the instance, either no registration or only one with defined email
* [ ] 2fa
* [ ] Custom fields for tasks
* [ ] Sorting lists by members, tasks, teams, last modified, etc
* [ ] "Favourite lists" -> A user can favourize boards which will then show up in a pseudonamespace
* [ ] Public lists
* [ ] Internal lists -> Only registered users can see the list
* [ ] Rights management for both public and internal lists
* [ ] Add new users via to a list which don't have an account yet, they'd get a link to sign up for vikunja.
  * [ ]  Respect registration email domain limits
* [ ] Export all data from Vikunja to json
* [ ] Watch a (n internal) list -> Will get notification for everything
* [ ] Archive a task instead of deleting
* [ ] Task dependencies
* [ ] Time tracking (possible plugin)
* [ ] IFTTT
* [ ] More sharing features (all of these with the already existing permissions)
  * [ ] Invite users per mail
  * [ ] Share a link with/without password
* [ ] Comments on tasks
* [ ] @mention users in tasks or comments to get them notified
* [ ] Summary of tasks to do in a configurable interval (every day/week or so)
* [ ] Importer (maybe frontend only)
  * [ ] Trello
  * [ ] Wunderlist
  * [ ] Zenkit
  * [ ] Asana
  * [ ] Microsoft Todo
  * [ ] Nozbe
  * [ ] Lanes
  * [ ] Nirvana
  * [ ] Good ol' Caldav (Tasks)
* [ ] More auth providers
  * [ ] LDAP/AD
  * [ ] Kerberos
  * [ ] SAML (what?)
  * [ ] smtp
  * [ ] OpenID
