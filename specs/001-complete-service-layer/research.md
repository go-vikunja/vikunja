# Research: Complete Service-Layer Refactor Stabilization and Validation

## Current System Analysis

### Decision: Focus on Three-Phase Approach
**Rationale**: The partially-refactored state requires systematic stabilization before continuing. The failing test suite indicates incomplete service layer implementations, particularly for task-related functionality. Breaking into phases allows for validation at each step.

**Alternatives considered**: 
- Complete rewrite from scratch (rejected - too risky, loses working functionality)
- Continue refactoring without stabilization (rejected - unstable foundation leads to cascading failures)

### Decision: Development Environment Only for Phase 1
**Rationale**: Clarification confirmed no uptime requirements during stabilization. This allows for iterative testing and debugging without production concerns.

**Alternatives considered**:
- Production-ready deployment pipeline (rejected - unnecessary complexity for refactor)
- Staging environment requirements (rejected - adds overhead without value)

## Architecture Pattern Analysis

### Decision: Strict "Chef, Waiter, Pantry" Architecture
**Rationale**: Existing REFACTORING_GUIDE.md provides proven patterns. Services contain ALL business logic, handlers are thin glue layer, models are data-only. This separation enables testability and maintainability.

**Alternatives considered**:
- Domain-driven design layers (rejected - would require complete restructure)
- Microservices architecture (rejected - overkill for monolithic application)

### Decision: Dependency Inversion for Backward Compatibility  
**Rationale**: Breaking model -> service import cycles requires function variable injection pattern documented in guide. Allows gradual migration without breaking existing code.

**Alternatives considered**:
- Interface-based dependency injection (rejected - more complex than needed)
- Complete model method removal (rejected - too disruptive)

## Testing Strategy Analysis

### Decision: TDD with 90% Service Layer Coverage
**Rationale**: Constitutional requirement enhanced to 90% for refactored services (vs 80% general). Failing tests indicate incomplete implementations, TDD ensures complete functionality before deployment.

**Alternatives considered**:
- Test-after approach (rejected - violates constitution)
- Lower coverage thresholds (rejected - refactor quality requires higher standards)

### Decision: Automated Test Parity Analysis
**Rationale**: Must ensure no test cases lost during refactor. Systematic comparison between vikunja/ and vikunja_original_main/ test suites prevents functionality regression.

**Alternatives considered**:
- Manual test review (rejected - error-prone, incomplete)
- No parity checking (rejected - violates functional parity requirement)

## Prioritization Strategy Analysis

### Decision: Dependency-First, Then Complexity-Based Ordering
**Rationale**: Clarification specified dependencies must be resolved before dependent features. Within dependency groups, simplest features first reduces risk and builds confidence.

**Alternatives considered**:
- Risk-based prioritization (rejected - dependencies create blocking issues)
- Usage-based prioritization (rejected - ignores technical constraints)

## Validation Framework Analysis

### Decision: Original System Behavior Takes Precedence
**Rationale**: Clarification established that when differences are found, original system behavior is authoritative. Refactor must achieve functional parity, not functional improvement.

**Alternatives considered**:
- Refactored system takes precedence (rejected - violates parity requirement)
- Case-by-case evaluation (rejected - adds decision overhead)

### Decision: AI Review + Human Final Approval
**Rationale**: Combines systematic analysis capabilities of AI with human judgment for final architectural validation. Balances thorough review with practical approval process.

**Alternatives considered**:
- Human-only review (rejected - may miss systematic issues)
- AI-only review (rejected - lacks architectural judgment)

## Current Failure Analysis

### Failed Tests Root Cause: Incomplete Service Layer Data Population
**Evidence**: Test failures show missing RelatedTasks, Labels, Attachments, and Assignees in query results
**Rationale**: Service layer query methods not properly populating related data structures that model layer previously handled

### UI Bug Root Cause: Service Layer Integration Issues  
**Evidence**: 404 errors on label creation, empty task detail views
**Rationale**: Handler layer not properly calling service methods or service methods not implementing complete business logic

## Technology Stack Validation

### Decision: Retain Existing Tech Stack
**Rationale**: Go + Echo + XORM + Vue.js stack is proven and working. Refactor focuses on architecture, not technology change.

**Alternatives considered**:
- Migrate to different ORM (rejected - adds unnecessary complexity)
- Replace Echo framework (rejected - working system needs architectural refactor, not framework change)

## Routing Architecture Analysis

### Decision: Modern Declarative APIRoute Pattern
**Rationale**: Project is transitioning from legacy WebHandler pattern (implicit permissions, scattered in routes.go) to modern declarative pattern (explicit permissions, organized in api/v1/ files). Current state has 9 features migrated, 15 still using legacy pattern - creating architectural inconsistency.

**Alternatives considered**:
- Keep mixed patterns (rejected - violates code quality standards, creates technical debt)
- Complete migration in separate refactor (rejected - loses context, adds future coordination overhead)
- Migrate only new features (rejected - leaves permanent inconsistency)

### Decision: Add Phase 2.4 for Route Modernization
**Rationale**: Completing route migration while service refactor context is fresh ensures complete architectural consistency, eliminates technical debt, and sets foundation for future API versioning.

**Alternatives considered**:
- Defer to post-Phase 2 (rejected - technical debt lingers, loses momentum)
- Skip entirely (rejected - violates constitutional principles on technical debt)

**Impact**: Adds 2-3 days effort, 7-8 tasks, but achieves 100% architectural consistency.

## Architectural Compliance Findings (October 2025)

### Critical Discovery: Business Logic Duplication Pattern
**Evidence**: Post-implementation audit of T011 (Projects), T012 (Project-Users), T013 (Project-Teams) revealed business logic was DUPLICATED instead of MOVED from models to services.

**Violation Pattern**:
```go
// WRONG - DUPLICATION (what was done)
// Model file still has full business logic:
func (p *Project) ReadAll(s *xorm.Session, a web.Auth, ...) {
    // 40+ lines of business logic
    prs, resultCount, totalItems, err := GetAllRawProjects(...)
    err = AddProjectDetails(s, prs, a)
    // More business logic...
}

// Service file has IDENTICAL business logic:
func (ps *ProjectService) ReadAll(s *xorm.Session, a web.Auth, ...) {
    // SAME 40+ lines of business logic
    prs, resultCount, totalItems, err := GetAllRawProjects(...)
    err = AddProjectDetails(s, prs, a)
    // Same business logic...
}

// CORRECT - DELEGATION (should have done)
// Model file delegates to service:
func (p *Project) ReadAll(s *xorm.Session, a web.Auth, ...) {
    // DEPRECATED: Use ProjectService.ReadAll instead
    service := services.NewProjectService(db.GetEngine())
    return service.ReadAll(s, a, p.search, page, perPage, p.IsArchived, p.Expand)
}

// Service file has business logic ONLY:
func (ps *ProjectService) ReadAll(s *xorm.Session, a web.Auth, ...) {
    // Business logic HERE only
    prs, resultCount, totalItems, err := GetAllRawProjects(...)
    err = AddProjectDetails(s, prs, a)
    // More business logic...
}
```

**Root Cause**: Misunderstood "refactor service" as "add service layer" instead of "MOVE logic FROM models TO services"

**Impact**: 
- Violates DRY principle (two sources of truth)
- Violates FR-007 requirement (move logic from models to services)
- Creates maintenance burden (changes must be made in two places)
- Increases technical debt

**Affected Tasks**:
- 🔴 T011 (Projects): MIXED state - Delete deprecated, ReadAll/Create still have full model logic
- 🔴 T012 (Project-Users): FULL duplication - all methods have logic in both model and service
- 🔴 T013 (Project-Teams): FULL duplication - all methods have logic in both model and service
- ✅ T006 (User Mentions): COMPLIANT - uses dependency inversion pattern correctly
- ⚠️ T005, T007-T010: Status unknown - audit needed

### Correct Reference Pattern: T006 User Mentions
**Evidence**: T006 correctly implemented dependency inversion pattern documented in REFACTORING_GUIDE.md

**Pattern**: Model has function variable that points to service implementation:
```go
// In model file (pkg/models/task_comment.go):
var GetUserMentionsFromText func(s *xorm.Session, content string) ([]*user.User, error)

func (tc *TaskComment) Create(s *xorm.Session, a web.Auth) error {
    // Business logic delegated via function variable:
    mentionedUsers, err := GetUserMentionsFromText(s, tc.Comment)
    // ...
}

// In service file (pkg/services/user_mentions.go):
func init() {
    // Service sets the function variable:
    models.GetUserMentionsFromText = GetMentionsFromText
}

func GetMentionsFromText(s *xorm.Session, content string) ([]*user.User, error) {
    // Actual business logic HERE
    matches := mentionRegex.FindAllStringSubmatch(content, -1)
    // ...
}
```

**Why This Works**: Avoids import cycles (model -> service) while allowing model to delegate to service layer.

### Architectural Compliance Definition
**A task is COMPLIANT when**:
1. Service layer contains ALL business logic (validation, business rules, database operations)
2. Model methods either:
   - Are DEPRECATED with service delegation, OR
   - Use dependency inversion (function variables) to delegate to services
3. Routes call service layer directly (not model methods)
4. No business logic duplication exists between model and service

**Verification Commands**:
```bash
# Check if model has business logic (should return 0 after compliance)
grep -c "s.Where\|s.Insert\|s.Delete" pkg/models/[feature].go

# Check if model delegates to service (should return > 0 after compliance)
grep -c "Service\|services.New" pkg/models/[feature].go

# Check route integration (should find service calls, not model calls)
grep -rn "[Feature]Service" pkg/routes/
```

**Prevention**: Use pre-task checklist, reference task review (T006 for dependency inversion), and post-task compliance verification before marking any task complete (see plan.md "Prevention Process").

## Conclusion

All research confirms the expanded four-phase approach is sound. The technical foundation exists, patterns are documented (both service and routing layers), and clarifications provide clear decision criteria. Route modernization phase ensures complete architectural transformation.

**CRITICAL UPDATE**: Architectural compliance audit revealed systemic issue requiring follow-up tasks T011A-C, T012D-F, T013A-C to properly migrate business logic from models to services. Prevention process added to plan.md ensures future tasks achieve compliance on first attempt. Ready to proceed with T014 after compliance tasks are completed.