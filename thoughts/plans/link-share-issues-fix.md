# Link Share Issues Fix Implementation Plan

## Overview

Fix inconsistent behavior in Vikunja's link share functionality, specifically addressing authentication feedback, project name display, and navigation issues identified in [GitHub issue #1380](https://github.com/go-vikunja/vikunja/issues/1380).

## Current State Analysis

Based on comprehensive codebase analysis, the link share system consists of:

- **Backend Authentication**: `/pkg/routes/api/v1/link_sharing_auth.go:55-89` handles JWT token generation
- **Frontend Authentication**: `/frontend/src/views/sharing/LinkSharingAuth.vue` manages password entry and auth flow  
- **Frontend Display**: `/frontend/src/components/home/ContentLinkShare.vue` renders the link share UI
- **State Management**: Auth and project state managed through Pinia stores

### Key Discoveries:
- Authentication errors (403) are handled for password validation but not for subsequent API calls at `LinkSharingAuth.vue:134-152`
- Project title loading depends on `baseStore.currentProject` which may not load reliably in link share mode at `ContentLinkShare.vue:15-21`
- Project title is displayed but not clickable, unlike normal navigation patterns found in `ProjectsNavigationItem.vue:26-49`
- No explicit error handling for post-authentication API failures that could cause blank screens

## Desired End State

After implementation:
- **Clear Error Feedback**: All 403 errors show user-friendly messages with retry options
- **Reliable Project Display**: Project names always appear correctly after successful authentication
- **Intuitive Navigation**: Project titles are clickable and provide clear navigation back to project views
- **Consistent UX**: Link share behavior matches main application navigation patterns

### Verification Criteria:
- No more blank screens on authentication failures
- Project names always visible after successful auth
- Clicking project title navigates back to main project view
- All error states provide actionable feedback

## What We're NOT Doing

- Changing the underlying JWT authentication mechanism
- Modifying database schema or backend API contracts
- Redesigning the overall link share UI layout
- Adding new permission levels or sharing types

## Implementation Approach

Fix the issues through targeted improvements to error handling, state management, and navigation without disrupting the existing architecture. Focus on reliability and user feedback rather than structural changes.

## Phase 1: Enhanced Error Handling & Feedback

### Overview
Improve error handling for authentication and post-authentication failures to prevent blank screens and provide clear user feedback.

### Changes Required:

#### 1. Global Error Handling for Link Share Context
**File**: `frontend/src/views/sharing/LinkSharingAuth.vue`
**Changes**: Add error boundary and improved error handling for post-authentication failures

```vue
<!-- Add after line 152 in the authenticate() function -->
} catch (e) {
    // Existing error handling...
    
    // Handle generic 403 errors that might occur after initial auth
    if (e?.response?.status === 403 && !e?.response?.data?.code) {
        errorMessage.value = t('sharing.accessDenied')
        authenticateWithPassword.value = false
        return
    }
    
    // Handle network/server errors
    if (e?.response?.status >= 500 || !e?.response) {
        errorMessage.value = t('sharing.serverError')
        authenticateWithPassword.value = false
        return
    }
    
    // Log unexpected errors for debugging
    console.error('Link share authentication error:', e)
    
    // Existing error handling...
}
```

#### 2. Add Error Handling to Content Display
**File**: `frontend/src/components/home/ContentLinkShare.vue`
**Changes**: Add error boundary for project loading failures

```vue
<!-- Add error display after line 21 -->
<Message
    v-if="projectLoadError"
    variant="danger"
    class="mbe-4"
>
    {{ $t('sharing.projectLoadError') }}
    <BaseButton
        variant="secondary"
        class="mls-2"
        @click="retryProjectLoad"
    >
        {{ $t('sharing.retry') }}
    </BaseButton>
</Message>

<!-- Modify the title display to handle error states -->
<h1
    v-if="!projectLoadError"
    :class="{'m-0': !logoVisible}"
    :style="{ 'opacity': currentProject?.title === '' ? '0': '1' }"
    class="title"
>
    {{ currentProject?.title === '' ? $t('misc.loading') : currentProject?.title }}
</h1>
```

#### 3. Add Required Translations  
**File**: `frontend/src/i18n/lang/en.json`
**Changes**: Add new translation strings for error messages

```json
{
  "sharing": {
    "accessDenied": "Access denied. Please check your permissions and try again.",
    "serverError": "Server error occurred. Please try again later.", 
    "projectLoadError": "Failed to load project information.",
    "retry": "Retry"
  }
}
```

### Success Criteria:

#### Automated Verification:
- [ ] Frontend linting passes: `cd frontend && pnpm lint`
- [ ] Frontend type checking passes: `cd frontend && pnpm typecheck`
- [ ] Link share E2E tests pass: `cd frontend && pnpm test:e2e --spec="cypress/e2e/sharing/linkShare.spec.ts"`

#### Manual Verification:
- [ ] 403 errors show proper error messages instead of blank screen
- [ ] Server errors display retry option
- [ ] Error messages are user-friendly and actionable
- [ ] Retry functionality works correctly

---

## Phase 2: Reliable Project Information Loading

### Overview
Ensure project titles and information load consistently after successful link share authentication.

### Changes Required:

#### 1. Add Project Loading Logic to ContentLinkShare
**File**: `frontend/src/components/home/ContentLinkShare.vue`
**Changes**: Add explicit project loading and error handling

```vue
<!-- Add to script setup section -->
<script lang="ts" setup>
import {computed, ref, watch, onMounted} from 'vue'
import {useRoute} from 'vue-router'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'
import {useAuthStore} from '@/stores/auth'

import Logo from '@/components/home/Logo.vue'
import PoweredByLink from './PoweredByLink.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import Message from '@/components/misc/Message.vue'

const baseStore = useBaseStore()
const projectStore = useProjectStore()
const authStore = useAuthStore()
const route = useRoute()

const currentProject = computed(() => baseStore.currentProject)
const background = computed(() => baseStore.background)
const logoVisible = computed(() => baseStore.logoVisible)
const projectLoadError = ref(false)

// Ensure project is loaded for link share
async function ensureProjectLoaded() {
    if (!authStore.authLinkShare || !route.params.projectId) {
        return
    }
    
    try {
        projectLoadError.value = false
        
        // Load project if not already loaded
        const projectId = Number(route.params.projectId)
        if (!currentProject.value || currentProject.value.id !== projectId) {
            await projectStore.loadProject(projectId)
        }
    } catch (e) {
        console.error('Failed to load project for link share:', e)
        projectLoadError.value = true
    }
}

async function retryProjectLoad() {
    await ensureProjectLoaded()
}

// Watch for route changes and ensure project is loaded
watch(() => route.params.projectId, ensureProjectLoaded, { immediate: true })

onMounted(ensureProjectLoaded)

// Existing code...
</script>
```

#### 2. Add Project Loading to Project Store
**File**: `frontend/src/stores/projects.ts`
**Changes**: Ensure project loading works for link share users

```typescript
// Add method to ensure single project loading works for link shares
async function loadProject(projectId: number) {
    const project = projects.value[projectId]
    if (project) {
        return project
    }
    
    try {
        const loadedProject = await projectService.get({id: projectId})
        projects.value[projectId] = loadedProject
        return loadedProject
    } catch (e) {
        console.error(`Failed to load project ${projectId}:`, e)
        throw e
    }
}
```

### Success Criteria:

#### Automated Verification:
- [ ] Frontend tests pass: `cd frontend && pnpm test:unit`
- [ ] Frontend linting passes: `cd frontend && pnpm lint`

#### Manual Verification:
- [ ] Project titles appear consistently after authentication
- [ ] Loading states are shown appropriately
- [ ] Project information displays correctly across all view types
- [ ] Error states handle project loading failures gracefully

---

## Phase 3: Clickable Project Title Navigation

### Overview
Make project titles clickable to enable intuitive navigation back to the main project view, matching patterns used elsewhere in the application.

### Changes Required:

#### 1. Add Clickable Project Title
**File**: `frontend/src/components/home/ContentLinkShare.vue`
**Changes**: Make project title clickable with proper routing

```vue
<!-- Replace the h1 title section (lines 15-21) -->
<BaseButton
    v-if="!projectLoadError && currentProject"
    :to="getProjectRoute()"
    variant="text"
    class="project-title-button"
    :class="{'m-0': !logoVisible}"
>
    <h1 class="title clickable-title">
        {{ currentProject?.title === '' ? $t('misc.loading') : currentProject?.title }}
    </h1>
</BaseButton>
<h1
    v-else-if="!projectLoadError"
    :class="{'m-0': !logoVisible}"
    class="title"
>
    {{ $t('misc.loading') }}
</h1>
```

```vue
<!-- Add to script setup -->
function getProjectRoute() {
    if (!currentProject.value) return null
    
    const hash = route.hash // Preserve link share hash
    
    // Default to the first available view or list view
    const projectId = currentProject.value.id
    const firstView = projectStore.projects[projectId]?.views?.[0]
    
    if (firstView) {
        return {
            name: 'project.view',
            params: { projectId, viewId: firstView.id },
            hash
        }
    }
    
    return {
        name: 'project.index', 
        params: { projectId },
        hash
    }
}
```

#### 2. Add Styling for Clickable Title
**File**: `frontend/src/components/home/ContentLinkShare.vue`
**Changes**: Add CSS to make the clickable title look appropriate

```scss
<style lang="scss" scoped>
// Existing styles...

.project-title-button {
    background: none !important;
    border: none !important;
    padding: 0 !important;
    text-decoration: none !important;
    
    &:hover .clickable-title {
        opacity: 0.8;
        cursor: pointer;
    }
}

.clickable-title {
    text-shadow: 0 0 1rem var(--white);
    margin: 0;
    
    &:hover {
        text-decoration: underline;
    }
}
</style>
```

### Success Criteria:

#### Automated Verification:
- [ ] Frontend linting passes: `cd frontend && pnpm lint`
- [ ] Frontend style linting passes: `cd frontend && pnpm lint:styles`

#### Manual Verification:
- [ ] Project title is clickable and shows hover effects
- [ ] Clicking project title navigates to the correct project view
- [ ] Navigation preserves link share authentication
- [ ] Title styling is consistent with design system

---

## Phase 4: Improved Task-to-Project Navigation

### Overview
Add breadcrumb-style navigation and improve the user experience when navigating from task detail back to project views.

### Changes Required:

#### 1. Add Breadcrumb Navigation to Task Detail
**File**: `frontend/src/views/tasks/TaskDetailView.vue`
**Changes**: Enhance breadcrumb navigation for link share context

```vue
<!-- Find the existing breadcrumb section and enhance it for link share -->
<nav
    v-if="currentProject && !isLinkShareAuth"
    class="breadcrumb"
    aria-label="breadcrumbs"
>
    <!-- Existing breadcrumb code -->
</nav>

<!-- Add link share specific breadcrumb -->
<nav
    v-if="currentProject && isLinkShareAuth"
    class="breadcrumb"
    aria-label="breadcrumbs"
>
    <ul>
        <li>
            <BaseButton
                :to="{ 
                    name: 'project.index', 
                    params: { projectId: currentProject.id },
                    hash: $route.hash 
                }"
                variant="text"
                class="breadcrumb-link"
            >
                <Icon icon="arrow-left" class="mie-2" />
                {{ getProjectTitle(currentProject) }}
            </BaseButton>
        </li>
        <li class="is-active">
            <a>{{ task.title }}</a>
        </li>
    </ul>
</nav>
```

#### 2. Add Link Share Detection
**File**: `frontend/src/views/tasks/TaskDetailView.vue`
**Changes**: Add computed property to detect link share mode

```vue
<!-- Add to script setup -->
import {useAuthStore} from '@/stores/auth'

const authStore = useAuthStore()
const isLinkShareAuth = computed(() => authStore.isLinkShareAuth)
```

### Success Criteria:

#### Automated Verification:
- [ ] Frontend linting passes: `cd frontend && pnpm lint`
- [ ] Task detail E2E tests pass: `cd frontend && pnpm test:e2e --spec="cypress/e2e/task/task.spec.ts"`

#### Manual Verification:
- [ ] Task detail view shows clear navigation back to project
- [ ] Back navigation preserves link share context
- [ ] Breadcrumb navigation is intuitive and visually clear
- [ ] Navigation works from all project view types

---

## Testing Strategy

### Unit Tests:
- Test error handling in `LinkSharingAuth.vue` for various error scenarios
- Test project loading logic in `ContentLinkShare.vue`
- Test navigation routing functions

### Integration Tests:
- Test complete authentication flow with error scenarios
- Test project loading after successful authentication
- Test navigation between task detail and project views

### Manual Testing Steps:
1. **Test Error Scenarios**:
   - Access invalid link share hash → should show proper error
   - Enter wrong password → should show "invalid password" message
   - Simulate server error → should show retry option

2. **Test Project Display**:
   - Access valid link share → project title should appear
   - Test across different view types (List, Kanban, Gantt, Table)
   - Test with and without logo visibility

3. **Test Navigation**:
   - Click project title → should navigate to main project view
   - Navigate to task detail → should show back navigation
   - Test navigation preservation of link share context

## Performance Considerations

- Project loading is optimized to only load when needed
- Error states don't repeatedly attempt failed operations
- Navigation maintains link share tokens efficiently

## Migration Notes

No database migrations required. Changes are purely frontend enhancements that maintain backward compatibility with existing link shares.

## References

- Original ticket: `thoughts/tickets/link-share-issue.md`
- Backend auth implementation: `pkg/routes/api/v1/link_sharing_auth.go:55-89`
- Frontend auth flow: `frontend/src/views/sharing/LinkSharingAuth.vue:110-152`
- Project loading patterns: `frontend/src/stores/projects.ts`
- Navigation patterns: `frontend/src/components/home/ProjectsNavigationItem.vue:26-49`