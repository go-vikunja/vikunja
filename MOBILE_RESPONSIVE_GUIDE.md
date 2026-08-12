# Vikunja Mobile Responsive Design Guide

## Overview

Vikunja has been enhanced with comprehensive mobile-responsive design improvements to provide an excellent user experience across all device sizes, from small phones (320px) to large desktops (1920px+).

## Changes Made

### 1. **Modal Component** (`frontend/src/components/misc/Modal.vue`)
✅ **Improvements:**
- Added touch-friendly scrolling (`-webkit-overflow-scrolling: touch`)
- Improved mobile padding with `max()` function to respect safe areas
- Better text wrapping and word-breaking
- Responsive button layout (stacked on mobile, horizontal on desktop)
- Font size responsive using `clamp()` for smooth scaling
- Proper close button sizing (44x44px minimum) and positioning for mobile
- Better modal content margins for small screens
- Full-width modals on mobile with proper padding

### 2. **Login Page** (`frontend/src/views/user/Login.vue`)
✅ **Improvements:**
- Input fields set to 16px font size (prevents iOS auto-zoom)
- Minimum 44x44px touch targets for all interactive elements
- Full-width buttons on mobile with proper spacing
- Responsive label and link layouts
- Better form field margins and padding
- Password reset link responsive positioning
- Checkbox and radio inputs with proper hit areas

### 3. **Register Page** (`frontend/src/views/user/Register.vue`)
✅ **Improvements:**
- Added comprehensive mobile styling section
- Touch-friendly form fields with proper sizing
- Responsive button layout
- Single-column form on mobile
- Improved message and notification display
- Proper font scaling on small screens

### 4. **Project List Page** (`frontend/src/views/project/ListProjects.vue`)
✅ **Improvements:**
- Responsive header with proper wrapping
- Single-column layout for buttons on mobile
- Full-width buttons with touch targets
- Better spacing and padding for small screens
- Improved text wrapping for project names

### 5. **Task List View** (`frontend/src/views/tasks/ShowTasks.vue`)
✅ **Improvements:**
- Responsive task options with full-width buttons
- Stacked button layout on mobile
- Improved empty states for small screens
- Responsive title sizing with `clamp()`
- Better label filter display with word wrapping
- Proper spacing adjustments for mobile

### 6. **Home Dashboard** (`frontend/src/views/Home.vue`)
✅ **Comprehensive Mobile Overhaul:**
- **Topbar:** Responsive title and subtitle, touch-friendly icon buttons
- **Quick Add:** Stacked input and button on mobile, horizontal on desktop
- **Stats:** Single-column grid on mobile instead of 3-column
- **Layout:** Single-column layout instead of two-column on mobile
- **Panels:** Responsive padding and text sizing
- **Task Items:** Proper hit areas and text wrapping
- **Modal:** Full-height with proper bottom positioning on mobile
- **Project List:** Single column with proper spacing
- Multiple breakpoints for extra-small devices (480px)

### 7. **Global Form Styles** (`frontend/src/styles/global.scss`)
✅ **Added comprehensive mobile form improvements:**
- **Input Fields:** 44px minimum height, 16px font size, proper padding
- **Labels:** 1rem font size with proper margins
- **Buttons:** 44x44px minimum, full-width option, proper spacing
- **Checkboxes/Radios:** 44px minimum height, larger touch targets
- **Tables:** Responsive with horizontal scroll, stacked layout on mobile
- **Modals:** Proper padding and margins
- **Navigation:** Touch-friendly items with proper sizing
- **Text:** Word-wrap and overflow handling
- **Layout:** Safe area insets, proper containers
- **Safe Area Support:** Using `env(safe-area-inset-*)` for notches

### 8. **Common Issues Fixed**
✅ **Responsiveness:**
- Fixed horizontal scrolling issues
- Proper text wrapping with `word-wrap` and `overflow-wrap`
- Responsive typography with `clamp()`
- Proper flex-wrap on mobile

✅ **Touch Targets:**
- All interactive elements minimum 44x44px
- Proper padding around clickable areas
- Keyboard accessible

✅ **Spacing:**
- Reduced padding on mobile (1rem instead of 2rem+)
- Proper margins between elements
- Stack layouts instead of side-by-side

## Mobile-First Design Principles Applied

### 1. **Touch-Friendly Design**
```scss
// All buttons and links have minimum 44x44px hit area
.button {
  min-height: 44px;
  min-width: 44px;
}
```

### 2. **Readable Typography**
```scss
// Inputs must be 16px to prevent iOS zoom
input {
  font-size: 16px !important;
  min-height: 44px;
}
```

### 3. **Responsive Scaling**
```scss
// Font sizes scale smoothly across screen sizes
.title {
  font-size: clamp(1.25rem, 5vw, 2rem);
}
```

### 4. **Safe Area Support**
```scss
// Handle notches on iOS
padding-block-start: env(safe-area-inset-top);
padding-block-end: env(safe-area-inset-bottom);
```

### 5. **Proper Stack Layouts**
```scss
// Stack to single column on mobile
.grid {
  @media (max-width: $tablet) {
    grid-template-columns: 1fr;
  }
}
```

## Breakpoints Used

```scss
$mobile: 320px;      // Minimum mobile width
$tablet: 768px;      // Tablet breakpoint
$desktop: 1024px;    // Desktop breakpoint
$widescreen: 1216px; // Wide desktop
$fullhd: 1408px;     // Full HD
```

### Media Query Patterns

```scss
// Mobile-first approach
@media screen and (max-width: $tablet) {
  // Mobile specific styles
}

// Desktop enhancement
@media screen and (min-width: $tablet) {
  // Desktop specific styles
}

// Extra small devices
@media screen and (max-width: 480px) {
  // Very small screen adjustments
}
```

## Testing Checklist

### Device Testing
- [ ] iPhone SE (375px)
- [ ] iPhone 12 (390px)
- [ ] iPhone 14 Pro Max (430px)
- [ ] Android Small (360px)
- [ ] Android Large (412px)
- [ ] iPad (768px)
- [ ] iPad Pro (1024px)
- [ ] Desktop (1920px)

### Orientation Testing
- [ ] Portrait mode
- [ ] Landscape mode
- [ ] Device rotation transitions

### Interaction Testing
- [ ] Touch targets (44x44px minimum)
- [ ] Tap interactions
- [ ] Swipe gestures
- [ ] Long-press actions
- [ ] Double-tap zoom disable

### Screen Reader Testing
- [ ] VoiceOver (iOS)
- [ ] TalkBack (Android)
- [ ] NVDA/JAWS (Desktop)

### Performance Testing
- [ ] Load time < 3s on 4G
- [ ] No layout shift (CLS < 0.1)
- [ ] Smooth animations (60fps)
- [ ] No horizontal scroll

### Specific Feature Testing

#### Login/Register
- [ ] Form fills comfortably on mobile
- [ ] Error messages display properly
- [ ] Forgot password link accessible
- [ ] No keyboard covering inputs
- [ ] Submit button easy to tap

#### Task Creation
- [ ] Quick add input responsive
- [ ] Button accessible on mobile
- [ ] Modal displays correctly
- [ ] Project selection works
- [ ] Success message visible

#### Task List
- [ ] Tasks display clearly
- [ ] Checkboxes are touch-friendly
- [ ] Dates are readable
- [ ] No text overflow
- [ ] Scroll is smooth

#### Home Dashboard
- [ ] All stats visible on mobile
- [ ] Panels stack properly
- [ ] Import banner not overwhelming
- [ ] Quick add works well
- [ ] Modal displays full-height

#### Projects
- [ ] Grid responsive on mobile
- [ ] Project cards readable
- [ ] Buttons full-width on mobile
- [ ] Archived toggle works
- [ ] No horizontal scroll

#### Settings
- [ ] Form fields responsive
- [ ] Toggles and checkboxes proper size
- [ ] Buttons easy to tap
- [ ] No form overflow
- [ ] Success messages visible

## Best Practices for Developers

### When Adding New Components

1. **Always consider mobile first:**
   ```scss
   // Mobile first approach
   .component {
     // Base mobile styles
   }
   
   @media screen and (min-width: $tablet) {
     .component {
       // Desktop enhancements
     }
   }
   ```

2. **Use proper breakpoints:**
   ```scss
   @import "@/styles/common-imports.scss"; // Imports breakpoint variables
   ```

3. **Ensure touch targets:**
   ```scss
   .button,
   .link,
   .checkbox {
     min-height: 44px;
     min-width: 44px;
   }
   ```

4. **Test inputs for iOS:**
   ```scss
   input {
     font-size: 16px !important; // Prevent zoom
   }
   ```

5. **Handle text wrapping:**
   ```scss
   .text {
     word-wrap: break-word;
     overflow-wrap: break-word;
     word-break: break-word;
   }
   ```

### CSS Utilities Available

```scss
// From global.scss - mobile responsive utilities
.is-hidden-mobile   // Hide on mobile
.is-hidden-tablet   // Hide on tablet
.is-hidden-desktop  // Hide on desktop

// Responsive visibility helpers
@media screen and (max-width: $tablet) { }
@media screen and (min-width: $tablet) { }
```

### Common Responsive Patterns

**Stacking on Mobile:**
```vue
<div class="columns">
  <div class="column">Content 1</div>
  <div class="column">Content 2</div>
</div>
```

```scss
.columns {
  @media (max-width: $tablet) {
    .column {
      width: 100% !important;
    }
  }
}
```

**Responsive Typography:**
```vue
<h1 class="title">{{ title }}</h1>
```

```scss
.title {
  font-size: clamp(1.5rem, 5vw, 2rem);
}
```

**Touch-Friendly Buttons:**
```vue
<button class="button">Action</button>
```

```scss
.button {
  min-height: 44px;
  min-width: 44px;
  padding: 0.75rem 1.5rem;
}
```

## Performance Considerations

1. **CSS Media Queries:** Prefer mobile-first with `max-width` media queries
2. **Font Sizing:** Use `clamp()` to avoid excessive media queries
3. **Images:** Use responsive images with `srcset` for mobile
4. **JavaScript:** Load mobile-specific JavaScript only when needed
5. **Animations:** Keep animations smooth (60fps) on mobile devices

## Known Limitations

1. **Older iOS devices** (iOS 12 and below) may not support all CSS features
2. **Android browsers** on older devices may have rendering issues
3. **Notch support** uses CSS `env()` which isn't supported in all browsers
4. **Safe area insets** require proper viewport meta tag setup

## Compatibility Matrix

| Feature | iOS 12+ | Android 5+ | Desktop | Notes |
|---------|---------|-----------|---------|-------|
| Safe Area Insets | ✅ | ⚠️ | N/A | Android support varies |
| Flexbox | ✅ | ✅ | ✅ | Full support |
| CSS Variables | ✅ | ✅ | ✅ | Full support |
| Clamp() | ✅ | ✅ | ✅ | Full support |
| 100dvh | ✅ | ⚠️ | ✅ | May not work on older Android |
| Touch Events | ✅ | ✅ | N/A | Full support |

## Debugging Mobile Issues

### Common Issues and Solutions

**Horizontal Scroll:**
```scss
body {
  overflow-x: hidden;
}
```

**Font Size Auto-Zoom on iOS:**
```scss
input, textarea, select {
  font-size: 16px !important;
}
```

**Notch Overlap:**
```scss
padding-block-start: env(safe-area-inset-top);
```

**Unresponsive Buttons:**
```scss
.button {
  min-height: 44px;
  min-width: 44px;
  cursor: pointer;
}
```

**Text Overflow:**
```scss
.text {
  word-wrap: break-word;
  overflow-wrap: break-word;
  word-break: break-word;
}
```

### Dev Tools

Use Chrome/Firefox/Safari DevTools to:
1. Toggle device toolbar (`Ctrl+Shift+M`)
2. Test responsive breakpoints
3. Simulate slow 3G connections
4. Test touch interactions
5. Debug CSS media queries

## Resources

- [MDN: Responsive Design](https://developer.mozilla.org/en-US/docs/Learn/CSS/CSS_layout/Responsive_Design)
- [Web.dev: Mobile UX](https://web.dev/mobile-ux-checklist/)
- [Apple: Human Interface Guidelines](https://developer.apple.com/design/human-interface-guidelines/)
- [Google: Material Design](https://material.io/design/)
- [WCAG: Mobile Accessibility](https://www.w3.org/WAI/WCAG21/Understanding/target-size.html)

## Contributing

When contributing mobile improvements:

1. Test on at least 2 real devices
2. Use the mobile-first approach
3. Maintain 44x44px touch targets
4. Ensure text wrapping works
5. Update this guide with new patterns
6. Add comments for non-obvious mobile code
7. Run linters before committing

## Future Improvements

- [ ] Implement progressive web app (PWA) features
- [ ] Add native app notifications
- [ ] Optimize for slow networks
- [ ] Add offline support
- [ ] Implement gesture controls
- [ ] Add dark mode mobile testing
- [ ] Performance monitoring for mobile
- [ ] Mobile-specific animations

---

**Last Updated:** April 2026
**Maintained by:** Vikunja Development Team
