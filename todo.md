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
* [x] Google fonts raus (sollen von lokal geladen werden)
* [x] Ladeanimationen erst nach 100ms anzeigen, sonst wird das überflüssigerweise angezeigt
* [x] Btns für Teams und neuer Namespace nach oben in die Leiste verschieben
* [x] Fancy Scrollbars
* [ ] Card-like overview of all lists with the first 3-5 tasks, undone first
* [ ] Be able to collapse all lists in a namespace by clicking on the menu entry

## Funktionales

* [x] Den Sharing-Updateshit mit der neuen methode machen (post)
* [x] User suchen einbauen, mit neuem endpoint
* [x] Fertige Tasks schöner visualisieren
  * [x] Alles abgehakte ausblenden, mit btn zum wieder einblenden
* [x] Wenn man den Namen einer Liste updated wird der Name in der List nicht upgedated
* [x] Links an den Freigewordenen Platz Menüpunkte machen à la "Heute"/Morgen/Diese Woche etc. Da kommt dann alles rein was dann due ist.
* [x] Wenn ein Task due ist das auch in der Übersicht anzeigen
  * [x] Overdue rot anzeigen
* [x] Beim Team bearbeiten Nutzer suchen einbauen
* [ ] Keyboard shortcuts
* [x] Gantt chart
  * [x] Basics
  * [x] Add tasks without dates set
  * [x] Edit tasks with a kind of popup when clicking on them - needs refactoring edit task into an own component
  * [x] Add a new task with a button or so
  * [x] Be able to choose the range for the chart
  * [x] Show task priority
  * [x] Show task done or not done
  * [x] Colors - needs api change before 
  * [x] More view modes
    * [x] Month: "The big picture"
    * [x] Day: 3-hour slices of a day
* [ ] Table view (list view, bit with more details)
* [ ] Calender view
* [ ] Kanaban
* [ ] Group list view by almost all fields

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
* [x] Subtasks
* [x] Start/Enddatum für Tasks
* [x] Tasks in time range
* [ ] Search everything
  * [ ] Lists
  * [ ] Tasks
  * [ ] Namespaces
  * [ ] Teams
  * [ ] Users with access on a list
  * [ ] Users with access to a namespace
  * [ ] Teams with access to a list
  * [ ] Teams with access to a namespace
* [x] Priorities
  * [x] Highlight tasks with high priority
* [x] Assignees
* [x] Labels
  * User should be able to search for a label
  * if none is found, "enter" should create and add it to the task
  * multiselect -> action dispatcher + styling
  * [x] Label overview + edit
  	* [x] Only be able to edit labels where the user has the right, disable the others
  * [x] Delay when searching to not search for the character I entered 5 minutes ago
* [ ] Timeline/Calendar view -> Get and show tasks in a range

## Other features

* [ ] Copy lists
* [ ] "Move to Vikunja" -> Migrator von Wunderlist/todoist/etc

## Refactor

* [x] Move everything to models
  * [x] Make sure all loading properties are depending on its service
* [x] Fix the first request afer login being made with an old token
* [x] Team sharing
  * [x] Refactor team sharing to not make a new request every time something was changed
  * [x] Team sharing should be able to search for a team instead of its ID, like it's the case with users
  * [x] Dropdown for rights
* [x] Same improvements also for user sharing
* [x] Use rights const everywhere
* [x] Styling of the search dropdown to match the rest of the theme
* [x] Use query params when getting tasks in a range

## Waiting for backend

* [ ] In und Out webhooks, mit Templates vom Payload
* [ ] "Smart Lists", Listen nach bestimmten Kriterien gefiltert -> nur UI?
* [ ] "Performance-Statistik" -> Wie viele Tasks man in bestimmten Zeiträumen so geschafft hat etc
* [ ] Activity Feed, so à la "der und der hat das und das gemacht etc"
* [ ] Attachments for tasks
* [ ] Search for users at new task assignees only in users who have access to the list
