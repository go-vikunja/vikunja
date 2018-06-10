# Featurecreep

* Listen erstellen, ändern, löschen
* Todopunkte zu Listen hinzufügen, bearbeiten, löschen
* Listen teilen (Email/Benutzername angeben, oder öffentlicher link (+einstellbar ob mit registrierung oder nicht, oder passwortgeschützt)
* Rechtemanagement

### Todopunkte

* ID
* Text
* Description
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

* [ ] Listen erstellen/bearbeiten/löschen
* [ ] Todopunkte hinzufügen/abhaken/löschen

#### v0.2

* [ ] Listen teilbar

#### v0.3

* [ ] Rechtemanagement (Und damit Unterscheidung zwischen Ownern und Mitgleidern)

#### v0.4 

* [ ] Websocket?

## Clients

* [ ] Webapp (vue.js)
* [ ] "Native" Clients (auf dem Rechner installiert (mit elektron oder so? Oder native?)
* [ ] Android (Flutter oder React Native)
* [ ] iOS (mit Framework????)
