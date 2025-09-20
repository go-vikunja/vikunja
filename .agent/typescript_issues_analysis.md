# TypeScript Issues Analysis

## Overview
There are approximately 400+ TypeScript errors in the Vikunja frontend. The issues fall into several main categories:

## Categories of Issues

### 1. Union Type Complexity (TS2590)
- **Files affected**: Button.vue, CreateEdit.vue, Dropdown.vue
- **Issue**: Expression produces a union type that is too complex to represent
- **Priority**: High - These are likely blocking other type inference

### 2. Null/Undefined Safety Issues (TS18047, TS18046, TS18048)
- **Files affected**: ProjectKanban.vue, ProjectList.vue, many user views, team views
- **Issue**: Variables potentially null/undefined being accessed without checks
- **Priority**: High - Common throughout codebase

### 3. Type Mismatch Issues (TS2345, TS2741, TS2322)
- **Files affected**: ProjectKanban.vue, many component files
- **Issue**: Arguments/assignments with incompatible types
- **Priority**: High - Indicates interface mismatches

### 4. Missing Properties (TS2741, TS2339)
- **Files affected**: Various model files, team components
- **Issue**: Objects missing required properties from interfaces
- **Priority**: Medium-High - Interface compliance issues

### 5. Implicit Any Types (TS7006)
- **Files affected**: ProjectKanban.vue, various event handlers
- **Issue**: Parameters without explicit types
- **Priority**: Medium - Easy to fix

### 6. Model/Interface Mismatches
- **Files affected**: Multiple files referencing models vs interfaces
- **Issue**: Backend models don't match frontend interface expectations
- **Priority**: High - Core type system issue

## Fix Strategy

### Phase 1: Core Type System Issues (High Priority)
1. Fix union type complexity issues in Button.vue, CreateEdit.vue, Dropdown.vue
2. Address model/interface mismatches (ProjectKanban.vue bucket/task issues)
3. Fix null safety in critical components

### Phase 2: Null Safety and Type Guards (Medium-High Priority)
1. Add null checks and type guards throughout
2. Fix undefined access patterns
3. Add optional chaining where appropriate

### Phase 3: Interface Compliance (Medium Priority)
1. Fix missing properties in object assignments
2. Update model types to match interfaces
3. Fix method signature mismatches

### Phase 4: Cleanup and Polish (Low-Medium Priority)
1. Fix implicit any types
2. Fix typos (e.g., "tertary" → "tertiary")
3. Add proper error typing

## Estimated Effort
- **Phase 1**: ~2-3 hours (foundational fixes)
- **Phase 2**: ~3-4 hours (systematic null safety)
- **Phase 3**: ~2-3 hours (interface alignment)
- **Phase 4**: ~1-2 hours (cleanup)

Total: ~8-12 hours of focused work across multiple commits