# Future Refactoring: API URL State Management

This document outlines the steps for a future refactoring task to improve how the frontend application manages the API's base URL.

## Current Situation

The application currently relies on a global variable, `window.API_URL`, to store the location of the Vikunja API. This approach has a significant drawback: `window.API_URL` is not a reactive data source for Vue and Pinia. This has led to a subtle bug where parts of the application (specifically, computed properties in Pinia stores) do not update when the global variable is changed, resulting in stale data and incorrect API requests.

The current implementation uses workarounds to bypass this issue, but a full refactor would be the "best practice" solution.

## Proposed Refactoring

The goal is to move the API URL from the global `window` object into a reactive Pinia state property. This will make the state management more robust, predictable, and aligned with modern Vue architecture.

### Step 1: Modify the `config` Store

**File:** `frontend/src/stores/config.ts`

1.  **Add a new state property:** Introduce a new reactive state property to hold the API URL.
    ```typescript
    export const useConfigStore = defineStore('config', () => {
        const apiUrl = ref('') // New reactive state
        // ... other state properties
    ```

2.  **Update the `apiBase` computed property:** Change `apiBase` to be based on the new `apiUrl` state property instead of the global `window.API_URL`.
    ```typescript
    const apiBase = computed(() => {
        if (!apiUrl.value) return ''
        // The parsing logic can be reused
        const {host, protocol, href} = parseURL(apiUrl.value)
        const cleanHref = href ? (href.endsWith('/') ? href.slice(0, -1) : href) : ''
		return `${protocol}//${host}${cleanHref ? `/${cleanHref}` : ''}`
    })
    ```

3.  **Create a `setApiUrl` action:** Add a new action to the store that allows other parts of the application to change the API URL state.
    ```typescript
    function setApiUrl(newUrl: string) {
        apiUrl.value = newUrl
    }
    ```

### Step 2: Update URL Handling Logic

**File:** `frontend/src/helpers/checkAndSetApiUrl.ts`

1.  **Use the new action:** Modify the `checkAndSetApiUrl` function to use the new `setApiUrl` action instead of directly manipulating `window.API_URL`.
    ```typescript
    // Inside checkAndSetApiUrl.ts, after a valid URL is found
    const configStore = useConfigStore()
    configStore.setApiUrl(foundUrl) // Instead of window.API_URL = foundUrl
    ```

### Step 3: Update Application Initialization

**File:** `frontend/src/main.ts` (or wherever the app is initialized)

1.  **Initialize the store state:** The initial API URL (from `localStorage` or default values) needs to be set in the Pinia store when the application first loads.
    ```typescript
    // Example logic in main.ts
    const configStore = useConfigStore()
    const initialUrl = localStorage.getItem('API_URL') || '/api/v1' // Or other default
    configStore.setApiUrl(initialUrl)
    ```

### Step 4: Remove Reliance on `window.API_URL`

1.  **Code Cleanup:** Perform a global search for `window.API_URL` across the entire frontend codebase.
2.  **Refactor:** Replace any remaining usages with the reactive `configStore.apiBase` computed property. This ensures the entire application uses the single, reactive source of truth from the Pinia store.

## Benefits of this Refactor

- **Improved Reactivity:** Eliminates bugs caused by stale data. The entire application will react automatically when the API URL changes.
- **Centralized State Management:** Consolidates application state within Pinia, which is the intended architectural pattern.
- **Removes Global State:** Reduces reliance on the global `window` object, preventing potential conflicts and making the code easier to reason about and test.
- **Improved Testability:** Components and stores will be easier to test in isolation without needing to manipulate a global `window` object.
