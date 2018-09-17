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
* [ ] Endpoint um nach Usern zu suchen, erstmal nur mit Nutzernamen, später mit setting ob auch mit email gesucht werden darf

## Feature-Ideen

* [ ] Labels
* [ ] Priorities
* [ ] Assignees
* [ ] Subtasks
* [ ] Attachments
* [ ] Repeating tasks
* [ ] Tagesübersicht ("Was ist heute/diese Woche due?") -> Machen letztenendes die Clients, wir brauchen nur nen endpoint, der alle tasks auskotzt, der Client macht dann die Sortierung.

## Clients

* [ ] Webapp (vue.js) + Bulma
* [ ] "Native" Clients (auf dem Rechner installiert (mit elektron oder so? Oder native mit qt oder so?)
* [ ] Android (Flutter)
* [ ] iOS (mit Framework???? (Ging das nich auch mit Flutter?))

## Anderes

* [ ] Refactor!!!! Alle Funktionen raus, die nicht mehr grbaucht werden + Funktionen vereinfachen/zusammenführen.
      Wenn ein Objekt 5x hin und hergereicht wird, und jedesmal nur geringfügig was dran geändert wird sollte das
      doch auch in einer Funktion machbar sein.
      * [ ] ganz viel in eigene neue Dateien + Packages auslagern, am besten eine package pro model mit allen methoden etc.
      * [ ] Alle alten dinger die nicht mehr gebraucht werden, weg.
            * [x] Die alten handlerfunktionen alle in eine datei packen und erstmal "lagern", erstmal brauchen wir die noch für swagger.
* [x] Drone aufsetzen
* [x] Tests schreiben
* [x] Namen finden
* [x] Alle Packages umziehen
* [x] Swagger UI aufsetzen
+ [x] CORS fixen
* [x] Überall echo.NewHTTPError statt c.JSON(Message{}) benutzen
* [x] Bessere Fehlermeldungen wenn das Model was ankommt falsch ist und nicht geparst werden kann
* [ ] Fehlerhandling irgendwie besser machen. Zb mit "World error messages"? Sprich, die Methode ruft einfach auf obs die entsprechende Fehlermeldung gibt und zeigt sonst 500 an.
* [ ] Endpoints neu organisieren? Also zb `namespaces/:nID/lists/:lID/items/:iID` statt einzelnen Endpoints für alles
* [x] Viper für config einbauen und ini rauswerfen
* [x] Docs für installationsanleitung
* [x] Tests für Rechtekram
* [x] "Apiformat" Methoden, damit in der Ausgabe zb kein Passwort drin ist..., oder created/updated von Nutzern... oder ownerID nicht drin ist sondern nur das ownerobject
* [x] Rechte überprüfen:
      * [x] Listen erstellen
      * [x] Listen bearbeiten (nur eigene im Moment)
      * [x] Listenpunkte hinzufügen
      * [x] Listenpunkte bearbeiten

### Short Term

* [x] Cacher konfigurierbar
* [ ] Validation der ankommenden structs, am besten mit https://github.com/go-validator/validator
* [x] Wenn die ID bei irgendeiner GetByID... Methode < 1 ist soll ein error not exist geworfen werden
* [ ] Bei den Structs "AfterLoad" raus, das verbraucht bei Gruppenabfragen zu viele SQL-Abfragen -> Die sollen einfach die entsprechenden Read()-Methoden verwenden (Krassestes bsp. ist GET /namespaces mit so ca 50 Abfragen)
* [ ] Methode einbauen, um mit einem gültigen token ein neues gültiges zu kriegen
* [ ] Wir brauchen noch ne gute idee, wie man die listen kriegt, auf die man nur so Zugriff hat
* [ ] /users sollte die Rechte mit ausgeben
* [ ] Nen endpoint um /teams/members /list/users etc die Rechte updazudaten ohne erst zu löschen und dann neu einzufügen

### Later/Nice to have

* [ ] An "accepted" für post/put payloads schrauben, man soll da zb keine id/created/updated/etc übergeben können.
* [ ] Globale Limits für anlegbare Listen + Namespaces
* [ ] Mgl., dass die Instanz geschlossen ist, also sich keiner registrieren kann, und man sich einloggen muss
* [ ] mgl. zum Emailmaskieren haben (in den Nutzereinstellungen, wenn man seine Email nicht an alle Welt rausposaunen will)
* [ ] Mgl. zum Accountlöschen haben (so richtig krass mit emailverifiezierung und dass alle Privaten Listen gelöscht werden und man alle geteilten entweder wem übertragen muss oder  auf privat stellen)
* [ ] Deps nach mod (dem nachfolger von dep) umziehen, blocked by Go 1.11
