# TypeScript Fixes Plan

## Summary of Issues Found (166 total: 160 errors, 6 warnings)

### By Category:

1. **Explicit `any` types (Primary focus)**: ~80+ instances
   - Need to replace with proper TypeScript interfaces
   - Most critical for type safety

2. **Indentation issues**: ~10+ instances
   - Expected tab indentation vs spaces
   - Vue template indentation issues

3. **Unused variables/imports**: ~8+ instances
   - Remove unused imports and variables
   - Follow eslint @typescript-eslint/no-unused-vars rule

4. **Vue reactivity issues**: ~5+ instances
   - vue/no-ref-object-reactivity-loss
   - vue/no-setup-props-reactivity-loss

5. **Missing trailing commas**: ~10+ instances
   - Add trailing commas per ESLint rules

6. **Other Vue/ESLint issues**: ~15+ instances
   - Attribute hyphenation
   - Deprecated filters
   - Case declarations in switch blocks
   - Event hyphenation

## Fix Strategy:
1. Start with explicit `any` type fixes (highest impact)
2. Fix formatting/indentation issues (easy wins)
3. Remove unused variables/imports
4. Fix Vue reactivity issues
5. Address remaining ESLint issues
6. Run tests after each batch of changes

## Files with Most Issues:
- UserTeam.vue (6 explicit any)
- TaskDetailView.vue (5 explicit any)
- Multiple other components with 1-3 explicit any issues