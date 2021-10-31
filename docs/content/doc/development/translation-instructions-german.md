---
title: "German Translation Instructions"
date: 2021-06-23T23:47:34+02:00
draft: false
---

# German Translation Instructions

<div class="notification is-warning">
<b>NOTE:</b> This document contains translation instructions specific to the german translation of Vikunja.
For instructions applicable to all languages, check out the <a href="{{< ref "./translations.md">}}">general translation instructions</a>.
</div>

{{< table_of_contents >}}

## Allgemein

Anrede: Wenig förmlich:

* “Du”-Form
* Keine “Amtsdeusch“-Umschreibungen, einfach so als ob man den Nutzer direkt persönlich ansprechen würde

Genauer definiert:

* “falsch” anstatt “nicht korrekt/inkorrekt”
* “Wende dich an …” anstatt “kontaktiere …”
* In derselben Zeit übersetzen (sonst wird aus dem englischen “is“ das deutsche “war”)
* Richtige Anführungszeichen verwenden. Also `„“` statt `''` oder `'` oder ` oder ´
	* `„` für beginnende Anführungszeichen, `“` für schließende Anführungszeichen

Es gelten Artikel und Worttrennungen aus dem [Duden](https://duden.de).

## Formulierungen

* `Account` statt `Konto`.
* `TOTP` immer als ein Wort und Groß.
* `CalDAV` immer so.
* `löschen` oder `entfernen` je nach Kontext. Wenn etwas *gelöscht* wird, existiert das gelöschte Objekt und danach
  nicht mehr und hat evtl. andere Objekte mitgelöscht (z.B. eine Aufgabe). Wird etwas *entfernt*, bezieht sich das
  meistens auf die Beziehung zu einem anderen Objekt. Das entfernte Objekt existiert danach immernoch, z.B. beim
  Entfernen eine:r Nutzer:in aus einem Team.
* Analog zu `löschen` oder `entfernen` gilt ähnliches für `hinzufügen` oder `erstellen`. Eine Aufgabe wird *erstellt*,
  aber ein:e Nutzer:in nur zu einem Team *hinzugefügt*.
* `Anmeldename` anstatt `Benutzer:innenname`

## Formulierungen in Modals und Buttons

Es sollten die gleichen Formulierungen auf Buttons und Modals verwendet werden.

Beispiel: Wenn der Button mit `löschen` beschriftet ist, sollte im Modal die Frage
lauten `Willst du das wirklich löschen?` und nicht `Willst du das wirklich entfernen?`. Gleiches gilt für
Erfolgs/Fehlermeldungen nach der Aktion.

## Gendern

Wo möglich, sollte eine geschlechtsneutrale Anrede verwendet werden. Falls diese sehr umständlich würden (siehe oben
„Amtsdeutsch-Umschreibungen“), soll mit *Doppelpunkt* gegendert werden.

Beispiel: „Benutzer:in“

## Trennungen

* E-Mail-Adresse (siehe Duden)

## Wörter und Ausdrücke

| Englisches Original | Verwendung in deutscher Übersetzung |
| ------------------- | -------------------- |
| Bucket | Spalte |
| Namespace | Namespace |
| Link Share | Linkfreigabe |
| Username | Anmeldename |

## Weiterführende Links

* https://docs.translatehouse.org/projects/localization-guide/en/latest/guide/translation_guidelines_german.html
