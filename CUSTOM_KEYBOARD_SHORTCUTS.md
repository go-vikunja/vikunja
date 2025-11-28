# Custom Keyboard Shortcuts Feature

## Overview

This feature allows users to customize keyboard shortcuts for various actions in Vikunja. Users can modify shortcuts for task operations, general app functions, and more through a dedicated settings page.

## Features

### ✅ Implemented

- **Customizable Action Shortcuts**: Users can customize shortcuts for task operations (mark done, assign, labels, etc.) and general app functions (toggle menu, quick search, etc.)
- **Fixed Navigation Shortcuts**: Navigation shortcuts (j/k for list navigation, g+key sequences) remain fixed and cannot be customized
- **Conflict Detection**: Prevents users from assigning the same shortcut to multiple actions
- **Individual and Bulk Reset**: Users can reset individual shortcuts or entire categories to defaults
- **Persistent Storage**: Custom shortcuts are saved to user settings and sync across devices
- **Real-time Updates**: Changes apply immediately without requiring a page refresh
- **Comprehensive UI**: Dedicated settings page with organized categories and intuitive editing

### 🔧 Architecture

#### Frontend Components

1. **useShortcutManager Composable** (`frontend/src/composables/useShortcutManager.ts`)
   - Core logic for managing shortcuts
   - Validation and conflict detection
   - Persistence through auth store
   - Reactive updates

2. **ShortcutEditor Component** (`frontend/src/components/misc/keyboard-shortcuts/ShortcutEditor.vue`)
   - Individual shortcut editing interface
   - Key capture functionality
   - Real-time validation feedback

3. **KeyboardShortcuts Settings Page** (`frontend/src/views/user/settings/KeyboardShortcuts.vue`)
   - Main settings interface
   - Category organization
   - Bulk operations

4. **Enhanced v-shortcut Directive** (`frontend/src/directives/shortcut.ts`)
   - Supports both old format (direct keys) and new format (actionIds)
   - Backwards compatible

#### Data Models

- **ICustomShortcut.ts**: TypeScript interfaces for custom shortcuts
- **IUserSettings.ts**: Extended to include `customShortcuts` field
- **shortcuts.ts**: Enhanced with metadata (actionId, customizable, category, contexts)

#### Storage

Custom shortcuts are stored in the user's `frontendSettings.customShortcuts` object:

```typescript
{
  "general.toggleMenu": ["alt", "m"],
  "task.markDone": ["ctrl", "d"]
}
```

## Usage

### For Users

1. **Access Settings**: Navigate to User Settings → Keyboard Shortcuts
2. **Customize Shortcuts**: Click "Edit" next to any customizable shortcut
3. **Capture Keys**: Press the desired key combination in the input field
4. **Save Changes**: Click "Save" to apply the new shortcut
5. **Reset Options**: Use "Reset to default" for individual shortcuts or "Reset Category" for bulk operations

### For Developers

#### Adding New Customizable Shortcuts

1. **Define the shortcut** in `shortcuts.ts`:
```typescript
{
  actionId: 'myFeature.doSomething',
  title: 'myFeature.doSomething.title',
  keys: ['ctrl', 'x'],
  customizable: true,
  contexts: ['/my-feature/*'],
  category: ShortcutCategory.GENERAL,
}
```

2. **Add translation keys** in `en.json`:
```json
{
  "myFeature": {
    "doSomething": {
      "title": "Do Something"
    }
  }
}
```

3. **Use in components**:
```vue
<template>
  <button v-shortcut="'.myFeature.doSomething'" @click="doSomething">
    Do Something
  </button>
</template>
```

#### Using the Shortcut Manager

```typescript
import { useShortcutManager } from '@/composables/useShortcutManager'

const shortcutManager = useShortcutManager()

// Get effective shortcut
const keys = shortcutManager.getShortcut('task.markDone')

// Get hotkey string for @github/hotkey
const hotkeyString = shortcutManager.getHotkeyString('task.markDone')

// Validate shortcut
const result = shortcutManager.validateShortcut('task.markDone', ['ctrl', 'd'])

// Set custom shortcut
await shortcutManager.setCustomShortcut('task.markDone', ['ctrl', 'd'])
```

## Implementation Details

### Phase 1: Infrastructure Setup ✅
- Created TypeScript interfaces and models
- Built core shortcut manager composable
- Developed UI components
- Added routing and translations

### Phase 2: Integration ✅
- Updated v-shortcut directive for backwards compatibility
- Refactored existing components to use new system
- Updated help modal to show effective shortcuts

### Phase 3: Polish and Testing ✅
- Added comprehensive unit tests
- Verified all translation keys
- Created documentation

## Testing

### Unit Tests
- `useShortcutManager.test.ts`: Tests for the core composable
- `ShortcutEditor.test.ts`: Tests for the editor component

### Manual Testing Checklist
- [ ] Can access keyboard shortcuts settings page
- [ ] Can customize individual shortcuts
- [ ] Conflict detection works correctly
- [ ] Reset functionality works (individual and bulk)
- [ ] Changes persist across browser sessions
- [ ] Help modal shows effective shortcuts
- [ ] All existing shortcuts continue to work

## Future Enhancements

### Potential Improvements
- **Import/Export**: Allow users to backup and restore their custom shortcuts
- **Profiles**: Multiple shortcut profiles for different workflows
- **Advanced Sequences**: Support for more complex key sequences
- **Context Awareness**: Different shortcuts for different views/contexts
- **Accessibility**: Better support for screen readers and alternative input methods

### Technical Debt
- Improve test coverage for complex scenarios
- Add E2E tests for the complete workflow
- Consider performance optimizations for large shortcut sets

## Migration Notes

This feature is fully backwards compatible. Existing shortcuts continue to work without any changes required. The new system runs alongside the old system until all shortcuts are migrated to use actionIds.

## Support

For issues or questions about custom keyboard shortcuts:
1. Check the help modal (Shift+?) for current shortcuts
2. Visit the keyboard shortcuts settings page for customization options
3. Reset to defaults if experiencing issues
4. Report bugs with specific key combinations and browser information
