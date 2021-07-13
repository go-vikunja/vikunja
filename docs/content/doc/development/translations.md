---
title: "Translations"
date: 2021-06-23T22:52:06+02:00
draft: false
menu:
  sidebar:
    parent: "development"
---

# Translations

This document provides documentation about how to translate Vikunja.

{{< table_of_contents >}}

## Where to translate

Translation happens at [crowdin](https://crowdin.com/project/vikunja).

Currently, only the frontend (and by extension, the desktop app) is translatable.

## Translation Instructions

> These are the instructions for translating Vikunja in another language. 
> For information about how to add new translation strings, see below.

For all languages these translation guidelines should be applied when translating:

* Use a less-formal style, as if you were talking to a friend.
* If the source string contains characters like `&` or `…`, the translated string should contain them as well.

More specific instructions for some languages can be found below.

### Wrong translation strings

If you encounter a wrong original translation string while translating, please don't correct it in the translation.
Instead, translate it to reflect the original meaning in the translated string but add a comment under the source string to discuss potential changes.

### Language-specific instructions

* [German]({{< ref "./translation-instructions-german.md">}})

## How to add new translation strings

All translation strings are stored in `src/i18n/lang/`.
New strings should be added only in the `en.json` file.
Strings in other languages will be synced through weblate and should not be added directly as a PR/commit in the frontend repo.

## Requesting a new language

If you want to start translating Vikunja in a language not yet available in Vikunja, please request the language through the weblate interface.
If you have issues with this or need a discussion before doing so, pleace [contact us](https://vikunja.io/contact/) or [start a discussion in the forum](https://community.vikunja.io).

Once at least 50% of all translation strings are translated and approved, they will be added and distributed with the Vikunja frontend for users to select and use Vikunja with them.
