# Plan for Displaying User Data Exports

This document outlines how to show existing user data exports inside the settings.

## Overview

Currently the UI only provides a form to request a data export (`DataExport.vue`) and a separate download page reached via a link in the mail notification. The API does not provide a way to query if a user already has an export ready.

The goal is to list available exports in the settings so a user can download them directly and request new ones if needed. Because a user only stores one export file at a time and old exports are cleaned up automatically, we only need to display a single entry with creation date and expiry information.

## Backend

1. **New endpoint** – `GET /api/v1/user/export` returns the current export if one exists. The handler should:
   - Look up the logged‑in user and read `ExportFileID`.
   - If `ExportFileID` is `0`, return an empty response `{}`.
   - Otherwise load the corresponding file meta data (created date, size, id).
   - Calculate the expiry time based on the cleanup interval (7 days) and include it in the response.
   - Add swagger docs and register the route in `pkg/routes/routes.go`.
2. **Permissions** – reuse existing authentication (JWT/API token). Only the current user may access this endpoint.
3. **Unit tests** – cover cases with and without an export file.

## Frontend

1. **Service update** – extend `frontend/src/services/dataExport.ts` with a `status()` method calling `GET /user/export`.
2. **UI page** – extend the existing `DataExport.vue` to also display the current export status at the top:
   - On mount call `status()` and store the result.
   - If an export is available, show creation date, size and expiry and provide a button linking to `DataExportDownload.vue` for the actual download.
   - If no export is available, show a message and the existing request form.
3. **Navigation** – keep the route name `user.settings.data-export`; no additional settings page is required. The button on the download page should route back to this settings page if the user wants to request another export.
4. **Translations** – add strings for "Your export is ready", "Created", "Expires" and the link back to the request page.

## Migration

No database changes are necessary. The new endpoint only reads existing data.

## Summary

By enhancing the backend with an export status endpoint and extending the existing settings view, users can request exports and see when an export is available for download on the same page. The dedicated download page continues to be used for retrieving the file.
