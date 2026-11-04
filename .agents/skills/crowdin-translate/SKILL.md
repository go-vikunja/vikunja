---
name: crowdin-translate
description: Translate Vikunja strings in Crowdin through the Crowdin MCP. Use when asked which translations are open for a language, to fill missing translations, or to fix an existing translation.
---

# Crowdin translation

Project `vikunja` (id `462614`). Crowdin files map to:

| fileId | Crowdin file | Source in repo | Usage |
|---|---|---|---|
| 4 | `en.json` | `frontend/src/i18n/lang/en.json` | `$t('key')` / `t('key')` in `frontend/src` |
| 16 | `en-api.json` | `pkg/i18n/lang/en.json` | `i18n.T(lang, "key")` in `pkg/` (mostly notifications) |
| 22 | `app_en.arb` | not in this repo (mobile app) | only Crowdin `context` |

## Steps

1. **Load the guidelines.** WebFetch https://vikunja.io/docs/translations/ and the language-specific page linked from it (German: https://vikunja.io/docs/german-translation-instructions). Ask for every rule and the full word list. They change, so fetch them every run. Done when you hold the general rules, form of address, quotation marks, and word list for the target language.
2. **Find the open strings.** `get_language_progress` gives counts per file. `list_strings` with croql `count of translations where (language = @language:"<lang>") = 0` returns the strings. Done when the string count matches the untranslated phrases in the progress output.
3. **Read each string's context in code.** For every string, grep its key outside `i18n/lang` and read the surrounding template or Go code. Work out where the text appears (button, badge, screen-reader-only label, email line, error toast), what it refers to, and how much space it has.
   - `error.<code>` keys are built dynamically (`frontend/src/message/index.ts`). Grep `= <code>` in `pkg/models/error.go` (or `pkg/**/error.go`) to see which error it is.
   - Sibling keys in the same `en.json` block and their existing translations show tone and terms.
4. **Match existing terminology.** `search_translations` (with `projectIds: 462614` and `languageIds`) and `list_language_translations` with `stringIds` show how the project already translates the terms in each string. If the guidelines' word list contradicts an existing translation, the word list wins. Note the conflict for the report.
5. **Add the translations** with `add_string_translation`. Leave them unapproved unless the user asks you to approve. If a source string is wrong, translate what it means and file `add_string_comment` with `type: issue`, `issueType: source_mistake`.
6. **Report** a table with key, source text, and translation. Below it, list the existing translations that conflict with the guidelines and offer to suggest fixes. A fix is another `add_string_translation` on that string. The old approved translation stays live until someone approves the new one.
