# Featurecreep

This is the place where I write down ideas to work on at some point. 
Sorry for some of them being in German, I'll tranlate them at some point.

## Feature-Ideen

* [x] Priorities
* [x] Repeating tasks
* [x] Tagesübersicht ("Was ist heute/diese Woche due?") -> Machen letztenendes die Clients, wir brauchen nur nen endpoint, der alle tasks auskotzt, der Client macht dann die Sortierung.
* [x] Subtasks

## Anderes

* [x] Refactor!!!! Alle Funktionen raus, die nicht mehr grbaucht werden + Funktionen vereinfachen/zusammenführen.
      Wenn ein Objekt 5x hin und hergereicht wird, und jedesmal nur geringfügig was dran geändert wird sollte das
      doch auch in einer Funktion machbar sein.
      * [x] ganz viel in eigene neue Dateien + Packages auslagern, am besten eine package pro model mit allen methoden etc.
      * [x] Alle alten dinger die nicht mehr gebraucht werden, weg.
      * [x] Die alten handlerfunktionen alle in eine datei packen und erstmal "lagern", erstmal brauchen wir die noch für swagger.
* [x] Drone aufsetzen
* [x] Tests schreiben
* [x] Namen finden
* [x] Alle Packages umziehen
* [x] Swagger UI aufsetzen
+ [x] CORS fixen
* [x] Überall echo.NewHTTPError statt c.JSON(Message{}) benutzen
* [x] Bessere Fehlermeldungen wenn das Model was ankommt falsch ist und nicht geparst werden kann
* [x] Fehlerhandling irgendwie besser machen. Zb mit "World error messages"? Sprich, die Methode ruft einfach auf obs die entsprechende Fehlermeldung gibt und zeigt sonst 500 an.
* [x] Viper für config einbauen und ini rauswerfen
* [x] Docs für installationsanleitung
* [x] Tests für Rechtekram
* [x] "Apiformat" Methoden, damit in der Ausgabe zb kein Passwort drin ist..., oder created/updated von Nutzern... oder ownerID nicht drin ist sondern nur das ownerobject
* [x] Rechte überprüfen:
      * [x] Listen erstellen
      * [x] Listen bearbeiten (nur eigene im Moment)
      * [x] Listenpunkte hinzufügen
      * [x] Listenpunkte bearbeiten
* [x] Der -1 namespace sollte auch seperat angesprochen werden können, gibt sonst probleme mit der app.

### Short Term

* [x] Cacher konfigurierbar
* [x] Wenn die ID bei irgendeiner GetByID... Methode < 1 ist soll ein error not exist geworfen werden
* [x] /users sollte die Rechte mit ausgeben
* [x] Nen endpoint um /teams/members /list/users etc die Rechte updazudaten ohne erst zu löschen und dann neu einzufügen
* [x] namespaces & listen updaten geht nicht, gibt nen 500er zurück
* [x] Logging für alle Fehler irgendwohin, da gibts bestimmt ne coole library für
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

### Docs

* [ ] Bauanleitung in die Readme/docs
  * [x] Auch noch nen "link" zum Featurecreep
* [ ] Anleitung zum Makefile
* [ ] Struktur erklären
* [ ] Backups
* [ ] Deploy in die docs
  * [ ] Docker
  * [ ] Native (systemd + nginx/apache)
* [ ] Docs aufsetzen

### Tasks

* [x] Start/Enddatum für Tasks
* [x] Timeline/Calendar view -> Dazu tasks die in einem Bestimmten Bereich due sind, macht dann das Frontend
* [x] Tasks innerhalb eines definierbarem Bereich, sollte aber trotzdem der server machen, so à la "Gib mir alles für diesen Monat"
* [x] Bulk-edit -> Transactions
* [ ] Labels
* [ ] Assignees
* [ ] Attachments
* [ ] Task-Templates innerhalb namespaces und Listen (-> Mehrere, die auswählbar sind)
* [ ] Ein Task muss von mehreren Assignees abgehakt werden bis er als done markiert wird
* [ ] Besseres Rechtesystem, damit man so fine-graded sachen machen kann wie "Der da darf aber nur Tasks hinzufügen, aber keine abhaken"

### General features

* [x] Deps nach mod umziehen
* [ ] Globale Limits für anlegbare Listen + Namespaces
* [ ] "Smart Lists", Listen nach bestimmten Kriterien gefiltert -> nur UI?
* [ ] "Performance-Statistik" -> Wie viele Tasks man in bestimmten Zeiträumen so geschafft hat etc
* [ ] IMAP-Integration -> Man schickt eine email an Vikunja und es macht daraus dann nen task -> Achtung missbrauchsmöglichkeiten
* [ ] In und Out webhooks, mit Templates vom Payload
* [ ] Reminders via mail
* [ ] Activity Feed, so à la "der und der hat das und das gemacht etc"
* [ ] ~~Websockets~~ SSE https://github.com/kljensen/golang-html5-sse-example
      * User authenticates (with jwt)
      * When updating/creating/etc an event struct is sent to the broker
      * The broker has a list of subscribed users
      * It then checks who is allowed to the see the event it recieved and sends it
* [ ] Mgl., dass die Instanz geschlossen ist, also sich keiner registrieren kann, und man sich einloggen muss
* [ ] mgl. zum Emailmaskieren haben (in den Nutzereinstellungen, wenn man seine Email nicht an alle Welt rausposaunen will)
* [ ] Mgl. zum Accountlöschen haben (so richtig krass mit emailverifiezierung und dass alle Privaten Listen gelöscht werden und man alle geteilten entweder wem übertragen muss oder  auf privat stellen)
* [ ] /info endpoint, in dem dann zb die limits und version etc steht

### Linters

* [x] goconst
* [ ] Gosimple -> waiting for mod
* [ ] Staticcheck -> waiting for mod
* [ ] unused -> waiting for mod
* [ ] gosec -> waiting for mod