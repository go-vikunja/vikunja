# Featurecreep

* Listen erstellen, ändern, löschen
* Todopunkte zu Listen hinzufügen, bearbeiten, löschen
* Listen teilen (Email/Benutzername angeben, oder öffentlicher link (+einstellbar ob mit registrierung oder nicht, oder passwortgeschützt)
* Rechtemanagement

### Todopunkte

* ID
* Text
* Description
* Status (done, not done)
* Fälligkeitsdatum
* Erinnerungsdatum (und zeit)
* Zuständig (später, mit teilen)
* Liste wo der Punkt drauf ist
* Timestamps

### Websockets

Das ganze soll als Websocket zur verfg gestellt werden, der dann automatisch bescheidsagt wenn sich was ändert. Benachrichtigungen machen clients.

## API-Roadmap

Ab v0.3 können wir mit clients anfangen.

#### v0.1

* [x] Listen erstellen/bearbeiten/löschen

      * [x] Ansehen
      * [x] Übersicht
      * [x] Einzelne liste mit allen todopunkten
      * [x] Erstellen
      * [x] Bearbeiten
      * [x] Löschen

* [x] Todopunkte hinzufügen/abhaken/löschen

      * [x] Erstellen
      * [x] Bearbeiten (abhaken)
      * [x] Löschen

* [x] Überall nochmal überprüfen dass der Nutzer auch das Recht hat die Liste zu löschen

* [x] Swaggerdocs !!!!

Neues Konzept: _Namespaces_

Ein Namespace kann Listen haben, es gibt mindestens einen Besiter pro Namespace. Wenn ein neuer Nutzer angelegt wird,
wird automatisch einer für den Nutzer erstellt.

Es gibt Lese- und Schreibrechte pro Namespace und Nutzer.

Namespace: 

* ID
* Name
* OwnerID
* Timestamps

Teams:

* ID
* Name
* Description
* Rights (Selbsthochzählende Konstanten als json-array abspeichern)
* CreatedByUser
* Timestamps

TeamMembers:

* ID
* TeamID
* MemberID
* Timestamps

TeamNamespaces:

* ID
* TeamID
* NamespaceID
* Timestamps

TeamLists: 

* ID
* TeamID
* ListID
* Timestamps

(+Check ob das Team schon Zugriff auf den Namespace hat und dafür sorgen dass das sich nicht überschneidet)
Bsp: wenn ein Namespace-Team Schreibrechte hat, soll es nicht möglich sein dieses Team mit Schreibrechten
zur Liste hinzuzufügen. Wenn das Team im Namespace aber nur Leserechte Hat soll es möglich sein dieses Team 
als Schreibend zur Liste hinzuzufügen.

Oder noch Besser: Man kann globale Rechte pro Namespace vergeben, die man dann wieder feinjustieren kann pro Liste.
                  Es soll aber nicht mgl. sein, ein Team zu einer Liste hinzuzufügen was nicht im Namespace ist. 
                  Es muss also möglich sein, Teams zum Namespace hinzuzufügen die keinerlei Rechte haben (damit man
                  denen dann wieder pro Liste welche geben kann) 

Rechte:
  Erstmal nur 3: Lesen, Schreiben, Admin. Admins dürfen auch Namen ändern, Teams verwalten, neue Listen anlegen, etc.
  Owner haben immer Adminrechte. Später sollte es auch möglich sein, den ownership an andere zur übertragen.s

Teams sind global, d.h. Ein Team kann mehrere Namespaces verwalten.

#### Neues Todo

* [x] Teams

      * [x] Erstellen
      * [x] Ansehen
      * [x] Bearbeiten
      * [x] Löschen

      ~~Ein zu lösendes Problem: Wie regelt man die Berechtigungen um Teams zu verwalten?~~

* [x] Namespaces

      * [x] Erstellen
      * [x] Ansehen
      * [x] Bearbeiten
      * [x] Löschen
      * [x] Teams hinzufügen. Der Nutzer kriegt nur Teams angezeigt die er erstellt hat.
      * [x] Alle Listen eines Namespaces anzeigen

* [x] Listen

      * [x] Listen zu einem Namespace hinzufügen

#### v0.2

* [x] Listen teilbar
      * [x] Mit anderen Nutzern
	    * [x Namespaces
      * [x] Teams
      * [ ] Mit Link
            * [ ] Offen
            * [ ] Passwortgeschützt

* [x] Rechtemanagement (Und damit Unterscheidung zwischen Ownern und Mitgleidern)
* [x] Mange Team members
      * [x] Hinzufügen
      * [x] Löschen

*Routen*

* [x] `namespaces/:id/teams`
      * [x] Create
      * [x] ReadAll
      * [x] Delete
* [x] `lists/:id/teams`
      * [x] Create
      * [x] ReadAll
      * [x] Delete

* [x] /namespaces soll zumindest auch die namen (+id) der dazugehörigen Listen rausgeben

## Feature-Ideen

* [x] Priorities
* [x] Repeating tasks
* [x] Tagesübersicht ("Was ist heute/diese Woche due?") -> Machen letztenendes die Clients, wir brauchen nur nen endpoint, der alle tasks auskotzt, der Client macht dann die Sortierung.
* [x] Subtasks

## Clients

* [ ] Webapp (vue.js) + Bulma
* [ ] "Native" Clients (auf dem Rechner installiert (mit elektron oder so? Oder native mit qt oder so?)
* [ ] Android (Flutter)
* [ ] iOS (mit Framework???? (Ging das nich auch mit Flutter?))

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
* [ ] Methode einbauen, um mit einem gültigen token ein neues gültiges zu kriegen

#### Userstuff

* [x] Userstuff aufräumen
	-> Soweit es geht und Sinnvoll ist auf den neuen Handler umziehen
		-> Login/Register/Password-reset geht natürlich nicht
		-> Bleibt noch Profile abrufen und Einstellungen -> Macht also keinen Sinn das auf den neuen Handler umzuziehen
* [x] Email-Verifizierung beim Registrieren
* [x] Password Reset -> Link via email oder so
* [ ] Settings

### Bugfixes

* [ ] Panic wenn mailer nicht erreichbar -> Als workaround mailer deaktivierbar machen, bzw keine mails verschicken
* [ ] "unexpected EOF"

### Docs

* [ ] Bauanleitung in die Readme/docs
  * [ ] Auch noch nen "link" zum Featurecreep
* [ ] Anleitung zum Makefile
* [ ] Struktur erklären
* [ ] Deploy in die docs
  * [ ] Docker
  * [ ] Native
* [ ] Docs aufsetzen

### Later/Nice to have

* [x] Deps nach mod umziehen
* [ ] Websockets
    * Nur lesend? (-> Updates wie bisher)
    * sollen den geupdaten Kram an alle anderen user schicken
    * Ein Channel in dem socket pro liste ... oder pro user?
    * Erst an die anderen Schicken wenn der write in die Datenbank erfolgeich war
    * Ein Nutzer authentifiziert sich mit jwt und bekommt dann zugriff auf alle rooms mit listen auf die er Zugriff hat
        * Unterscheidung nach lesen und Schreiben 
* [ ] Globale Limits für anlegbare Listen + Namespaces
* [ ] Mgl., dass die Instanz geschlossen ist, also sich keiner registrieren kann, und man sich einloggen muss
* [ ] mgl. zum Emailmaskieren haben (in den Nutzereinstellungen, wenn man seine Email nicht an alle Welt rausposaunen will)
* [ ] Mgl. zum Accountlöschen haben (so richtig krass mit emailverifiezierung und dass alle Privaten Listen gelöscht werden und man alle geteilten entweder wem übertragen muss oder  auf privat stellen)
* [ ] IMAP-Integration -> Man schickt eine email an Vikunja und es macht daraus dann nen task -> Achtung missbrauchsmöglichkeiten
* [ ] In und Out webhooks, mit Templates vom Payload
* [ ] Start/Enddatum für Tasks
* [ ] Timeline/Calendar view -> Dazu tasks die in einem Bestimmten Bereich due sind, macht dann das Frontend
* [ ] "Smart Lists", Listen nach bestimmten Kriterien gefiltert -> nur UI?
* [ ] "Performance-Statistik" -> Wie viele Tasks man in bestimmten Zeiträumen so geschafft hat etc
* [ ] Activity Feed, so à la "der und der hat das und das gemacht etc"
* [ ] Assignees
* [ ] Attachments
* [ ] Labels
* [ ] Tasks innerhalb eines definierbarem Bereich, sollte aber trotzdem der server machen, so à la "Gib mir alles für diesen Monat"
