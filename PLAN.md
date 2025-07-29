# Plan to add customizable date display

## Overview
Implement a new frontend setting to control how all dates are rendered. The value should be available through the auth store and persisted together with the other `frontendSettings` in the user settings JSON. Existing date formatting helpers will read the setting and format dates accordingly.

## Tasks
1. **Define setting type**
   - Create an enum `DateDisplay` in `src/constants` with options:
     - `relative` (current behaviour)
     - `mm-dd-yyyy`
     - `dd-mm-yyyy`
     - `yyyy-mm-dd`
     - `mm/dd/yyyy`
     - `dd/mm/yyyy`
     - `yyyy/mm/dd`
     - `dayMonthYear` (`25th July 2025`)
     - `weekdayDayMonthYear` (`Friday, 25th July 2025`)
2. **Extend settings models**
   - Add `dateDisplay: DateDisplay` to `IFrontendSettings` (`frontend/src/modelTypes/IUserSettings.ts`).
   - Provide default value in `UserSettingsModel` (`frontend/src/models/userSettings.ts`).
   - When loading settings in the auth store, merge defaults similarly to other settings (`frontend/src/stores/auth.ts`).
3. **Expose in auth store**
   - The `settings` state already contains `frontendSettings`. After the change consumers can read `useAuthStore().settings.frontendSettings.dateDisplay`.
4. **Create composable**
   - Add `useDateDisplay` composable returning the current value from the auth store. This mirrors `useColorScheme` for easy access.
5. **Update date helpers**
   - Extend `formatDate`, `formatDateShort`, `formatDateLong` and `formatDateSince` (`frontend/src/helpers/time/formatDate.ts`) to read the selected format via the composable. When `relative` is selected keep the current relative output; for other options use `dayjs` or `Intl.DateTimeFormat` with the configured pattern.
6. **Update components**
   - Replace direct calls to `formatDateSince` or `formatDateShort` where dates are shown to instead call a new helper `formatDisplayDate`. This function will delegate to the correct formatter based on `dateDisplay`.
7. **Settings UI**
   - In `UserSettings/General.vue`, add a `<select>` allowing users to choose a date display format. Use the translation keys for each option.
   - Add new strings to `frontend/src/i18n/lang/en.json` under `user.settings`.
8. **Persistence**
   - Saving settings already calls `authStore.saveUserSettings`. The backend stores `frontend_settings` as JSON so no API change is required.
9. **Migration considerations**
   - Existing users will get `relative` as default via the auth store defaults.
10. **Testing**
   - Verify that changing the option updates date formatting across tasks, reminders, and other date displays.

This plan touches the files around the default settings definitions such as `frontend/src/models/userSettings.ts` lines 16‑27 where defaults are declared【F:frontend/src/models/userSettings.ts†L16-L27】 and the setting interface `IFrontendSettings` around lines 9‑18【F:frontend/src/modelTypes/IUserSettings.ts†L9-L18】. The auth store merges defaults at lines 120‑137 of `frontend/src/stores/auth.ts`【F:frontend/src/stores/auth.ts†L120-L137】.
