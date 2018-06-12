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
* [ ] "Apiformat" Methoden, damit in der Ausgabe zb kein Passwort drin ist..., oder created/updated von Nutzern
* [ ] Swaggerdocs !!!!

#### v0.2

* [ ] Listen teilbar
  * [ ] Mit anderen Nutzern
  * [ ] Mit Link
    * [ ] Offen
    * [ ] Passwortgeschützt

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
* [ ] Ne Instanz mit den docs aufsetzen
* [ ] Namen finden
* [ ] Alle Packages umziehen

* [ ] mgl. zum Emailmaskieren haben (in den Nutzereinstellungen, wenn man seine Email nicht an alle Welt rausposaunen will)
* [ ] Mgl. zum Accountlöschen haben (so richtig krass mit emailverifiezierung und dass alle Privaten Listen gelöscht werden und man alle geteilten entweder wem übertragen muss oder  auf provat stellen)