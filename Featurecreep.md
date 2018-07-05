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

* [ ] Teams
  * [ ] Erstellen
  * [ ] Ansehen
  * [ ] Bearbeiten
  * [ ] Löschen
  
  Ein zu lösendes Problem: Wie regelt man die Berechtigungen um Teams zu verwalten?
  
* [ ] Namespaces
  * [x] Erstellen
  * [x] Ansehen
  * [x] Bearbeiten
  * [x] Löschen
  * [ ] Teams hinzufügen. Der Nutzer kriegt nur Teams angezeigt die er erstellt hat.
  * [x] Alle Listen eines Namespaces anzeigen
* [x] Listen
  * [x] Listen zu einem Namespace hinzufügen

#### v0.2

* [ ] Listen teilbar
  * [ ] Mit anderen Nutzern
  * [ ] Mit Link
    * [ ] Offen
    * [ ] Passwortgeschützt
    
    Wenn man Listen mit nem Nutzer teilt, wird ein Team für diesen Nutzer erstellt, falls er nicht bereits in einem ist.

#### v0.3

* [ ] Rechtemanagement (Und damit Unterscheidung zwischen Ownern und Mitgleidern)

#### v0.4 

* [ ] Websocket?

## Clients

* [ ] Webapp (vue.js) + Bulma
* [ ] "Native" Clients (auf dem Rechner installiert (mit elektron oder so? Oder native?)
* [ ] Android (Flutter)
* [ ] iOS (mit Framework???? (Ging das nich auch mit Flutter?))

## Anderes

* [ ] CI aufsetzen
* [ ] Tests schreiben
* [ ] Namen finden
* [ ] Alle Packages umziehen
* [x] Swagger UI aufsetzen
* [ ] Bessere Fehlermeldungen wenn das Model was ankommt falsch ist und nicht geparst werden kann
* [ ] Endpoints neu organisieren? Also zb `namespaces/:nID/lists/:lID/items/:iID` statt einzelnen Endpoints für alles

* [ ] "Apiformat" Methoden, damit in der Ausgabe zb kein Passwort drin ist..., oder created/updated von Nutzern... oder ownerID nicht drin ist sondern nur das ownerobject
* [ ] Rechte überprüfen (in extra Funktion auslagern, dann wird das einfacher später):
  * [ ] Listen erstellen
  * [ ] Listen bearbeiten (nur eigene im Moment)
  * [ ] Listenpunkte hinzufügen
  * [ ] Listenpunkte bearbeiten


* [ ] Globale Limits für anlegbare Listen + Namespaces
* [ ] Mgl., dass die Instanz geschlossen ist, also sich keiner registrieren kann, und man sich einloggen muss
* [ ] mgl. zum Emailmaskieren haben (in den Nutzereinstellungen, wenn man seine Email nicht an alle Welt rausposaunen will)
* [ ] Mgl. zum Accountlöschen haben (so richtig krass mit emailverifiezierung und dass alle Privaten Listen gelöscht werden und man alle geteilten entweder wem übertragen muss oder  auf provat stellen)
