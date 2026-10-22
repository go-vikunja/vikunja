# Translations

User-facing strings must be translatable.

- Frontend: source strings in `frontend/src/i18n/lang/en.json`. In components: `const {t} = useI18n({useScope: 'global'})`, then `t('your.key')`.
- API (mostly notifications): source strings in `pkg/i18n/lang/en.json`. Use `i18n.T(lang, "your.key", params...)`.
- Edit only `en.json`. Never add a language or translate strings yourself; translations are managed at https://vikunja.io/docs/translations/. If asked to, point the user there.
- Before adding a new string, check if a string with the same value already exists. If it does, use the existing key.
