# TypeScript Issues Analysis

## Overview
The Vikunja frontend has numerous TypeScript issues that need to be systematically fixed. Based on the typecheck output, I've identified several categories of issues:

## Error Categories

### 1. Union Type Complexity Issues
- `Button.vue(65,15)`: Expression produces union type too complex to represent
- `Button.vue(65,56)`: ButtonProps type incompatibility with IconProp

### 2. Implicit Any Type Parameters
- `TipTap.vue(690,32)`: event parameter implicitly has 'any' type
- `suggestion.ts(222,14)` & `(234,15)`: props parameter implicitly has 'any' type
- Multiple files with implicit any types on parameters

### 3. Type Assignment Issues
- `FilterAutocomplete.ts`: Multiple issues with SuggestionItem type assignments
- Array type mismatches (ILabel[] vs SuggestionItem[], etc.)
- Missing required properties on types

### 4. Null/Undefined Safety Issues
- Multiple "Object is possibly 'undefined'" errors
- "Type 'undefined' is not assignable to type 'string'" errors
- Null assignment to non-nullable types

### 5. Property Access Issues
- Object literal property mismatches
- Missing properties in type definitions
- Index signature issues with dynamic property access

### 6. Model/Interface Mismatches
- `ApiTokenModel` vs `IApiToken` incompatibilities
- Permission type mismatches
- Missing required properties in interfaces

## Priority Order for Fixes

1. **High Priority - Core Component Issues**
   - Button.vue (affects entire UI)
   - TipTap editor (core functionality)
   - Filter components (search/filtering)

2. **Medium Priority - Type Definitions**
   - Model interfaces and their implementations
   - API token handling
   - User authentication types

3. **Low Priority - Minor Type Issues**
   - Implicit any parameters
   - Property access refinements
   - Optional type improvements

## Files Requiring Immediate Attention

1. `src/components/input/Button.vue` - Union type complexity
2. `src/components/input/editor/TipTap.vue` - Event typing
3. `src/components/input/filter/FilterAutocomplete.ts` - Multiple type mismatches
4. `src/views/user/settings/ApiTokens.vue` - Model interface issues
5. `src/components/input/filter/FilterInput.vue` - Configuration issues
6. `src/components/input/filter/highlighter.ts` - Node type issues

## Strategy
- Fix files in order of impact (components used widely first)
- Ensure tests pass after each batch of fixes
- Make incremental commits with conventional commit messages
- Focus on type safety without breaking existing functionality