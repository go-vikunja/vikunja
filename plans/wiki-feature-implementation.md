# Wiki Feature Implementation Plan

## Overview
Add a new "Wiki" view to projects allowing users to create and organize markdown documents with code highlighting and Mermaid diagram support.

## ✅ Existing Infrastructure We Can Leverage

### Backend (Already Implemented)
1. **Hierarchical Project Structure** - Projects already support parent/child relationships
   - `ParentProjectID` field in Project model
   - Recursive queries with `project_hierarchy` CTE
   - Cycle detection and validation logic
   - `GetAllParentProjects()`, `GetChildProjects()` methods
   
2. **File/Attachment System** - Fully implemented file handling
   - `TaskAttachment` model with CRUD operations
   - `files.Create()`, `files.Delete()` in `pkg/files/`
   - File size validation, thumbnails, blur hash support
   - Permission checks integrated
   
3. **Permission System** - Complete 3-tier system
   - `CanRead`, `CanWrite`, `CanCreate`, `CanDelete` interfaces
   - Project-level permissions with team/user support
   - Permission inheritance for child projects
   - `checkPermissionsForProjects()` with recursive hierarchy checking

4. **Search Infrastructure** - Advanced search capabilities
   - Typesense integration for full-text search
   - Task search with filters already implemented
   - Document indexing patterns in `task_search.go`

5. **Events System** - Notification/webhook infrastructure
   - Event listeners in `pkg/models/listeners.go`
   - Webhook support for CRUD operations
   - Subscription system for notifications

### Frontend (Already Implemented)
1. **TipTap Rich Text Editor** - Fully configured
   - Code blocks with syntax highlighting (`lowlight`)
   - Links, images, tables, mentions
   - Bubble menu, toolbar (`EditorToolbar.vue`)
   - `TipTap.vue` component ready to reuse
   
2. **Tree Navigation with Drag-Drop** - Complete implementation
   - `ProjectsNavigationItem.vue` - collapsible tree items
   - `ProjectsNavigation.vue` - drag-drop reordering with `zhyswan-vuedraggable`
   - Collapsible folders with state persistence (`useStorage`)
   - Position calculation and saving logic
   
3. **Hierarchical Store Logic** - Already in place
   - `getChildProjects()` computed in projects store
   - `getAncestors()` for breadcrumb navigation
   - Parent/child relationship management
   - Recursive deletion of child items

4. **Markdown/Content Support**
   - `marked` library for markdown parsing
   - `dompurify` for XSS protection
   - Task descriptions already use rich content

## Simplified Implementation Approach

**We DON'T need to build:**
- ❌ Hierarchical structure from scratch (Projects already have it!)
- ❌ Permission system (Already complete)
- ❌ File attachment system (Already working)
- ❌ Rich text editor (TipTap fully configured)
- ❌ Tree navigation UI (Can reuse ProjectsNavigationItem pattern)
- ❌ Drag-and-drop reordering (Already implemented)
- ❌ Search infrastructure (Typesense ready)

**We ONLY need to build:**
- ✅ New `WikiPage` model (similar to Task but simpler)
- ✅ Wiki API endpoints (standard CRUD + search)
- ✅ WikiPage service and store (follow existing patterns)
- ✅ Wiki view components (reuse TipTap, tree navigation patterns)
- ✅ Mermaid extension for TipTap

## Feature Requirements
- New "Wiki" project view (alongside List, Kantt, Table, Kanban)
- Create/Edit/Delete markdown pages
- Organize pages into folders (hierarchical structure)
- Markdown rendering with:
  - Code syntax highlighting
  - Mermaid.js diagram support
- Permission system integration (Read/Write/Admin)
- Search within wiki pages
- Page history/versioning (future enhancement)

## Architecture Overview

### Backend Changes (Go)

#### 1. Database Schema
**New Tables:**
- `wiki_pages`
  - `id` (bigint, primary key)
  - `project_id` (bigint, foreign key to projects)
  - `parent_id` (bigint, nullable, foreign key to wiki_pages for folder structure)
  - `title` (varchar)
  - `content` (text, markdown content)
  - `path` (varchar, full path for easier querying)
  - `is_folder` (bool)
  - `position` (int, for ordering within parent)
  - `created_by` (bigint, foreign key to users)
  - `created` (timestamp)
  - `updated` (timestamp)

**Indexes:**
- project_id + parent_id (for listing pages in folder)
- project_id + path (for path lookups)

#### 2. Models (`pkg/models/`)
**Create `wiki_page.go`:**
```go
type WikiPage struct {
    ID        int64     `xorm:"bigint autoincr not null unique pk" json:"id"`
    ProjectID int64     `xorm:"bigint not null INDEX" json:"project_id"`
    ParentID  *int64    `xorm:"bigint null INDEX" json:"parent_id"`
    Title     string    `xorm:"varchar(250) not null" json:"title"`
    Content   string    `xorm:"longtext null" json:"content"`
    Path      string    `xorm:"varchar(500) not null INDEX" json:"path"`
    IsFolder  bool      `xorm:"bool default false" json:"is_folder"`
    Position  int       `xorm:"int default 0" json:"position"`
    CreatedBy int64     `xorm:"bigint not null" json:"-"`
    
    Created time.Time `xorm:"created not null" json:"created"`
    Updated time.Time `xorm:"updated not null" json:"updated"`
    
    CreatedByUser *user.User `xorm:"-" json:"created_by"`
}
```

**Implement interfaces:**
- `CRUDable` for standard CRUD operations
- `Rights` for permission checking (delegates to project permissions)

**Methods needed:**
- `Create()` - Create new page/folder (**similar to `TaskAttachment.NewAttachment()`**)
- `ReadOne()` - Get single page
- `ReadAll()` - List pages in project/folder (**similar to Project's child queries**)
- `Update()` - Update page content/title
- `Delete()` - Delete page/folder with cascade (**reuse project deletion pattern**)
- `CanRead()`, `CanWrite()`, `CanCreate()`, `CanDelete()` - **Delegate to project permissions** (same as TaskAttachment)
- `Move()` - Move page to different folder (**reuse project parent change logic**)
- `Reorder()` - Change position in folder (**copy from project position logic**)

#### 3. Services (`pkg/services/`)
**Create `wiki_service.go`:**
- `GetWikiTree(projectID)` - Get full tree structure
- `SearchWikiPages(projectID, query)` - Full-text search
- `ValidateWikiPath(projectID, path)` - Ensure unique paths
- `GetBreadcrumbs(pageID)` - Get parent hierarchy

#### 4. API Routes (`pkg/routes/api/v1/`)
**Add to router:**
```
POST   /projects/:project/wiki                    - Create page/folder
GET    /projects/:project/wiki                    - List all pages (tree)
GET    /projects/:project/wiki/:id                - Get single page
PUT    /projects/:project/wiki/:id                - Update page
DELETE /projects/:project/wiki/:id                - Delete page
PUT    /projects/:project/wiki/:id/move           - Move to different folder
PUT    /projects/:project/wiki/:id/reorder        - Change position
GET    /projects/:project/wiki/search?q=:query    - Search pages
```

#### 5. Database Migration
Run: `mage dev:make-migration WikiPage`

Edit the generated migration to:
- Create `wiki_pages` table with all columns
- Add indexes
- Add foreign key constraints

#### 6. Events
Create events for:
- `WikiPageCreated`
- `WikiPageUpdated`
- `WikiPageDeleted`

These integrate with notification system and webhooks.

### Frontend Changes (Vue.js)

#### 1. TypeScript Interfaces (`src/modelTypes/`)
**Create `IWikiPage.ts`:**
```typescript
export interface IWikiPage {
    id: number
    projectId: number
    parentId: number | null
    title: string
    content: string
    path: string
    isFolder: boolean
    position: number
    created: string
    updated: string
    createdBy: IUser
    children?: IWikiPage[]  // For tree structure
}
```

#### 2. API Service (`src/services/`)
**Create `wikiPage.ts`:**
```typescript
export default class WikiPageService extends AbstractService<IWikiPage> {
    constructor() {
        super({
            getAll: '/projects/{projectId}/wiki',
            get: '/projects/{projectId}/wiki/{id}',
            create: '/projects/{projectId}/wiki',
            update: '/projects/{projectId}/wiki/{id}',
            delete: '/projects/{projectId}/wiki/{id}',
        })
    }
    
    move(projectId: number, pageId: number, newParentId: number | null)
    reorder(projectId: number, pageId: number, position: number)
    search(projectId: number, query: string)
}
```

#### 3. Pinia Store (`src/stores/`)
**Create `wikiPages.ts`:**
- **Copy structure from `projects.ts` store**
- Store wiki pages for current project
- Tree structure management (reuse `getChildProjects` pattern)
- CRUD operations
- Search functionality
- `getAncestors()` for breadcrumbs (same as projects store)
- Loading states

#### 4. Components (`src/components/project/wiki/`)
**Create new components:**
- `WikiView.vue` - Main wiki view container
- `WikiSidebar.vue` - Tree navigation sidebar
- `WikiPageContent.vue` - **Reuses existing `TipTap.vue` component**
- `WikiPageItem.vue` - Tree item (similar to `ProjectsNavigationItem.vue`)
- `WikiToolbar.vue` - Action buttons (new page, new folder, etc.)
- `WikiBreadcrumbs.vue` - **Use `getAncestors()` from store** for page hierarchy
- `WikiMermaidBlock.vue` - Custom TipTap node view for Mermaid diagrams

**Implementation notes:**
- Copy drag-drop logic from `ProjectsNavigation.vue` (uses `zhyswan-vuedraggable`)
- Copy collapsible folder logic from `ProjectsNavigationItem.vue`
- Reuse position calculation from project store
- Follow same permission checking patterns

#### 5. Markdown & Mermaid Integration
**✅ Already available in frontend:**
- ✅ TipTap editor (`@tiptap/vue-3`) - fully configured
- ✅ Code blocks with syntax highlighting (`lowlight`)
- ✅ Markdown support (`marked`)
- ✅ XSS protection (`dompurify`)
- ✅ Links, images, tables, mentions all supported

**Only need to add:**
```json
{
  "mermaid": "^10.6.0"          // Diagram support
}
```

**Implementation:**
- Create custom TipTap extension for Mermaid code blocks
- Extend existing `TipTap.vue` component for wiki use case
- Reuse existing `EditorToolbar.vue` and bubble menu

#### 6. Views (`src/views/project/`)
**Update `ProjectView.vue`:**
- Add "Wiki" tab alongside List, Gantt, Table, Kanban
- Route to wiki view when selected

**Create `ProjectWiki.vue`:**
- Main wiki view component
- Layout: sidebar (tree) + content area
- Handle routing: `/projects/:projectId/wiki/:pageId?`

#### 7. Router (`src/router/`)
Add route:
```typescript
{
    path: '/projects/:projectId/wiki/:pageId?',
    name: 'project.wiki',
    component: () => import('@/views/project/ProjectWiki.vue'),
}
```

#### 8. Styling
- Markdown content styling (code blocks, headings, etc.)
- Mermaid diagram theming (light/dark mode)
- Tree sidebar styling (indentation, icons)
- Split-pane resizing

### Implementation Phases

#### Phase 1: Backend Foundation (Week 1)
1. Create database migration
2. Implement WikiPage model with CRUD
3. Add basic API endpoints
4. Add permissions integration
5. Write unit tests

#### Phase 2: Frontend Structure (Week 1-2)
1. Create TypeScript interfaces
2. Create API service
3. Create Pinia store
4. Add wiki view tab to projects
5. Create basic routing

#### Phase 3: Core UI (Week 2)
1. Build WikiSidebar with tree navigation
2. Build WikiRenderer with markdown + Mermaid
3. Build WikiEditor with TipTap
4. Connect to API/store

#### Phase 4: Advanced Features (Week 3)
1. Drag-and-drop reorganization
2. Search functionality
3. Context menus
4. Breadcrumbs navigation
5. Folder operations

#### Phase 5: Polish & Testing (Week 3-4)
1. Add loading states and error handling
2. Write E2E tests
3. Add translations (en.json)
4. Update documentation
5. Polish UI/UX

### Testing Strategy

**Backend Tests:**
- Model CRUD operations
- Permission checks (read/write from different users)
- Path validation and uniqueness
- Folder deletion (cascade)
- Move/reorder operations

**Frontend Tests:**
- Unit tests for store actions
- Component tests for WikiEditor, WikiRenderer
- E2E tests for:
  - Creating pages/folders
  - Editing content
  - Deleting pages
  - Moving pages
  - Searching

### Documentation Updates
- Add wiki feature to README.md
- Update API documentation (Swagger)
- Add user guide for wiki usage
- Update AGENTS.md with wiki development patterns

### Security Considerations
- Sanitize markdown content (XSS prevention with DOMPurify)
- Validate Mermaid diagram syntax to prevent code injection
- Enforce project permissions on all wiki operations
- Rate limit wiki operations to prevent abuse
- Validate path traversal attacks in folder structure

### Performance Considerations
- Lazy load wiki pages (don't load all at once)
- Cache rendered markdown on backend
- Debounce autosave in editor
- Optimize tree rendering for large wikis
- Use virtual scrolling for large folder lists

### Future Enhancements (Not in MVP)
- Page templates
- Version history/revisions
- Page linking/backlinks
- Export wiki to PDF/HTML
- Import from external wikis
- Collaborative editing (real-time)
- Comments on pages
- Page attachments/embeds

## Development Workflow

### Starting Backend Development
```bash
# Create migration
mage dev:make-migration WikiPage

# Edit migration file in pkg/migration/

# Run tests
mage test:feature

# Run linter
mage lint:fix
```

### Starting Frontend Development
```bash
cd frontend

# Install dependencies if needed
pnpm install marked highlight.js mermaid dompurify

# Run dev server
pnpm dev

# Run tests
pnpm test:unit

# Lint
pnpm lint:fix
pnpm lint:styles:fix
```

## Estimated Timeline (REVISED)
- **Backend**: 3-5 days (model + API, reusing existing patterns)
- **Frontend**: 1 week (components reusing TipTap + tree navigation)
- **Mermaid Integration**: 1-2 days
- **Testing & Polish**: 3-4 days
- **Total**: 2-3 weeks (down from 4-6 weeks!)

## Open Questions
1. Should wiki pages be version controlled (git-like history)?
2. Do we need wiki page templates?
3. Should we support page attachments beyond inline images?
4. Real-time collaborative editing or just autosave?
5. Import/export formats (Confluence, Notion, etc.)?
6. Maximum page size limits?
7. Should folders have their own content/landing pages?

## Dependencies
- No new backend dependencies (use existing XORM, Echo, etc.)
- Frontend: **Only need to add `mermaid`** - all other deps already exist:
  - ✅ TipTap editor ecosystem
  - ✅ `lowlight` for code highlighting
  - ✅ `marked` for markdown parsing
  - ✅ `dompurify` for XSS protection
  - ✅ `sortablejs` for drag-and-drop

## Risks & Mitigation
- **Risk**: Large markdown files causing performance issues
  - **Mitigation**: Implement page size limits, lazy loading
  
- **Risk**: Mermaid diagrams with malicious code
  - **Mitigation**: Sandbox Mermaid rendering, validate syntax
  
- **Risk**: Complex folder structures becoming slow
  - **Mitigation**: Limit nesting depth, optimize queries with path indexing
  
- **Risk**: Concurrent editing conflicts
  - **Mitigation**: Implement optimistic locking with version checks

## Success Metrics
- Users can create and organize wiki pages
- Markdown renders correctly with code and Mermaid
- Permissions work correctly (can't edit read-only wikis)
- Page load time < 500ms
- No XSS vulnerabilities in markdown rendering
