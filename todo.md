# Todo

* [x] Header-menu
    * [x] Logout nach rechts, mit icon statt button
    * [x] Logo oben links
    * [x] Benutzernamen neben logout
* [x] Die Listenauswahl sollte highlighten welche Liste man grade ausgewählt hat
* [x] Namespaces
    * [x] Bei jedem Namespace sollte rechts neben dem Namen ein Zahnrad zum Bearbeiten sein, das tauscht dann den view mit der aktuellen Liste
    * [x] Über Namespaces btn zum neuen Namespace anlegen mit popup zum Namen eingeben
    * [x] Namespace löschen btn bei bearbeiten
* [x] Listen
    * [x] Btn zum Liste hinzufügen
    * [x] Zahnrad zum Liste bearbeiten
    * [x] Btn zum Liste löschen bei bearbeiten
* [x] Tasks:
    * [x] Oben großes Eingabefeld zum Punkte hinzufügen
    * [x] Tasks in voller Breite drunter anzeigen
    * [x] Tasks bearbeiten geht mit Zahnrad rechts, da druffklicken, dann geht von Links eine card rein (halbe breite der Tasklist) mit den Optionen
      * [x] Datetimepicker einbauen für Daten etc. (flatpickr)
    * [x] Bug fixen der auf try dafür sorgt dass beim Abhaken die checkbox nicht geupdated wird
    * [x] Task löschen btn bei bearbeiten
* [x] Hintergrund durch das mit den Lamas von Freepick austauschen
* [x] Badges einfügen
* [x] Lizenz einfügen!
* [x] Runterladelink erwähnen
* [x] Den Kram für Teams & user managen in ne eigene Komponente auslagern, das ist ja fast das selbe

* [ ] Erklärungen zu was wie funktioniert -> wiki?

## Eye-Candy

* [x] Zurück zu Home (wenn man auf das Logo klickt)
* [ ] Google fonts raus (sollen von lokal geladen werden)
* [ ] Ladeanimationen erst nach 100ms anzeigen, sonst wird das überflüssigerweise angezeigt
* [ ] Btns für Teams und neuer Namespace nach oben in die Leiste verschieben
* [ ] Links an den Freigewordenen Platz Menüpunkte machen à la "Heute"/Morgen/Diese Woche etc. Da kommt dann alles rein was dann due ist.
* [ ] Andere Icons? --> Freepikkram

## Funktionales

* [ ] Beim Team bearbeiten Nutzer suchen einbauen
* [ ] Fertige Tasks schöner visualisieren
  * [ ] Alles abgehakte ausblenden, mit btn zum wieder einblenden
* [x] Den Sharing-Updateshit mit der neuen methode machen (post)
* [x] User suchen einbauen, mit neuem endpoint
* [ ] Wenn man den Namen einer Liste updated wird der Name in der List nicht upgedated

## Funktionen aus der API

* [x] Sharingshit
    * [x] Listen für Nutzer
        * [x] freigeben
        * [x] entfernen
        * [x] Einstellmglkt für Rechte
    * [x] Listen für Teams
        * [x] freigeben
        * [x] entfernen
        * [x] Einstellmglkt für Rechte
    * [x] Namespaces für Nutzer
        * [x] freigeben
        * [x] entfernen
        * [x] Einstellmglkt für Rechte
    * [x] Namespaces für Teams
        * [x] freigeben
        * [x] entfernen
        * [x] Einstellmglkt für Rechte
* [x] Userstuff
    * [x] Email-Verification
    * [x] Password forgot
* [x] Teams
    * [x] Mglkt zum Erstellen von neuen Teams
    * [x] Alle Teams auflisten, auf die der Nutzer Zugriff hat
        * [x] In der UI klarmachen, wenn der Nutzer admin ist (möglicherweise braucht das noch ne Änderung im Backend)
        * [x] Einzelne Teams ansehbar
            * [x] In den Teams, in denen der Nutzer admin ist, Bearbeitung ermöglichen
	    * [x] Löschen ermöglichen
* [ ] Search everything
  * [ ] Lists
  * [ ] Tasks
  * [ ] Namespaces
  * [ ] Teams
  * [ ] Users with access on a list
  * [ ] Users with access to a namespace
  * [ ] Teams with access to a list
  * [ ] Teams with access to a namespace
* [x] Subtasks
* [ ] Priorities
  * [ ] Sachen mit hoher Prio irgendwie hervorheben (rotes Dreieck zb)
* [ ] Listen Kopieren
* [ ] "Move to Vikunja" -> Migrator von Wunderlist/todoist/etc

## Waiting for backend

* [ ] Assignees
* [ ] Timeline/Calendar view
* [ ] In und Out webhooks, mit Templates vom Payload
* [ ] Start/Enddatum für Tasks
* [ ] Timeline/Calendar view -> Dazu tasks die in einem Bestimmten Bereich due sind, macht dann das Frontend
* [ ] "Smart Lists", Listen nach bestimmten Kriterien gefiltert -> nur UI?
* [ ] "Performance-Statistik" -> Wie viele Tasks man in bestimmten Zeiträumen so geschafft hat etc
* [ ] Activity Feed, so à la "der und der hat das und das gemacht etc"
* [ ] Attachments for tasks
* [ ] Labels