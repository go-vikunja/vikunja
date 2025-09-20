# TypeScript/Linting Issues Fixing Plan

## Summary
TypeScript compilation passes but ESLint reports 246 problems (242 errors, 4 warnings):
- Primary issue: `@typescript-eslint/no-explicit-any` (majority of errors)
- Secondary issues: unused variables, formatting issues, deprecated features

## Categories of Issues

### 1. `@typescript-eslint/no-explicit-any` (Most Critical)
Replace `any` types with proper TypeScript types in:
- Components (BaseButton.story.vue, ProjectsNavigation.vue, etc.)
- Input components (AutocompleteDropdown, Button, DatepickerInline, Multiselect, etc.)
- Editor components (commands.ts, suggestion.ts, setLinkInEditor.ts)
- Services (various API services)
- Views (multiple views with any types)

### 2. Unused Variables (`@typescript-eslint/no-unused-vars`)
- Remove or prefix with underscore unused variables
- Files: MigrationHandler.vue, ProjectSettingsBackground.vue, EditTeam.vue

### 3. Vue-specific Issues
- `vue/define-macros-order`: defineProps order issues
- `vue/v-on-event-hyphenation`: event naming
- `vue/no-deprecated-filter`: deprecated filter usage
- `vue/html-indent` and `vue/html-closing-bracket-newline`: formatting

### 4. Other Formatting
- `comma-dangle`: missing trailing commas

## Fixing Strategy
1. Start with most common issues (no-explicit-any)
2. Work file by file systematically
3. Test after each significant batch
4. Commit frequently with conventional commits
5. Ensure tests still pass throughout