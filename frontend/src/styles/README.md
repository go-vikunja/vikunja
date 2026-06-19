# Frontend Styles

This directory contains all global styling for the Vikunja web client. Component-scoped
styles live next to their `.vue` file; what lives here is the material every component
relies on: the Bulma base, the design-token system (CSS custom properties), typography,
and a handful of cross-cutting theme overrides.

If you are wondering **"where do I add X?"** — jump to [Where do I add …?](#where-do-i-add-) at
the bottom.

## Directory map

```
styles/
├── common-imports.scss     SCSS variables/mixins auto-injected into every .scss/.vue <style lang="scss">
├── global.scss             Entry point: pulls in Bulma partials + theme + components + tokens
├── fonts.scss              @font-face declarations for Quicksand and Open Sans
├── transitions.scss        Vue <Transition> classes (fade, width)
├── tailwind.css            Tailwind v4 entry (utilities only, `tw` prefix)
│
├── custom-properties/      CSS custom-property token definitions
│   ├── colors.scss         Color tokens (--scheme-*, --grey-*, --primary, …) + dark mode
│   └── shadows.scss        Shadow tokens (--shadow-xs/sm/md/lg) + dark mode
│
├── theme/                  Global theme rules — selectors that style the whole app
│   ├── theme.scss          Base resets, focus-visible ring, .box / .is-fullwidth / etc.
│   ├── typography.scss     Heading styles using $vikunja-font
│   ├── navigation.scss     .menu / .menu-list styling used by the sidebar
│   ├── form.scss           Field / control / button add-on tweaks
│   ├── scrollbars.scss     Custom scrollbar colors
│   ├── link-share.scss     Tweaks for the public link-share layout
│   ├── flatpickr.scss      Overrides to make flatpickr use our custom properties
│   ├── loading.scss        .loader-container / .is-loading spinner styling
│   ├── background.scss     Optional project background image layer
│   ├── content.scss        .content overrides (Bulma's rich-text container)
│   ├── helpers.scss        Print utilities (.d-print-none, …)
│   └── logical-spacing.scss Margin/padding utilities using logical properties
│
└── components/             Global styling that couldn't (yet) be scoped to a .vue file
    ├── task.scss           Styles for .task-view on the public share page
    ├── tasks.scss          Legacy .tasks tree — flagged for untangling
    ├── labels.scss         .labels-list styles
    └── tooltip.scss        v-popper tooltip theme
```

Files marked with a `FIXME:` comment are styles that *should* live in a component's
`<style scoped>` block but haven't been moved yet. Prefer fixing the component rather
than extending these files.

## The `common-imports.scss` contract

`common-imports.scss` is prepended to **every** SCSS stylesheet in the project — every
`*.scss` file and every `<style lang="scss">` block in a `.vue` component. This is wired up
in `vite.config.ts` via `css.preprocessorOptions.scss.additionalData`:

```ts
// vite.config.ts
const PREFIXED_SCSS_STYLES = `@use "sass:math";
@import "${pathSrc}/styles/common-imports.scss";`

css: {
  preprocessorOptions: {
    scss: { additionalData: PREFIXED_SCSS_STYLES, … },
  },
}
```

**Because of that, `common-imports.scss` must produce zero CSS output.** It only defines
SCSS variables (`$vikunja-font`, `$navbar-height`, `$transition`, …) and imports Bulma's
utilities partial — pure functions, mixins, and variables with no selectors. If you
accidentally add a selector here, its rules will be duplicated into every compiled
stylesheet. The file has a prominent header comment calling this out; keep it that way.

Add a variable to `common-imports.scss` when you need it available in multiple components'
`<style lang="scss">` blocks. For a one-off, keep it local to the component.

## Bulma: which variant and why

Vikunja uses [`bulma-css-variables`](https://www.npmjs.com/package/bulma-css-variables), a
fork of Bulma that emits CSS custom properties (`--primary`, `--text`, `--scheme-main`, …)
instead of inlining SCSS variables. This is what makes runtime theming (including dark
mode) possible without recompiling SCSS.

`global.scss` imports Bulma partials one-by-one rather than pulling in the whole
`bulma.sass`. Each intentionally-excluded partial has an explanatory comment. The current
exclusions fall into three buckets:

1. **Moved to Vue components** — e.g. `elements/button` lives in `Button.vue`,
   `components/dropdown` lives in the dropdown component.
2. **Not used** — `breadcrumb`, `level`, `panel`, `tabs`, `notification`, `progress`,
   `message`, `hero`, `section`, `footer`, `tiles`.
3. **Replaced by a custom implementation in `theme/`** — most notably
   `helpers/spacing` is replaced by `theme/logical-spacing.scss` so that spacing
   utilities use logical properties (margin-inline/block) for RTL support.

If you need a Bulma feature that is currently excluded, re-enable the import in
`global.scss` rather than duplicating the rules.

## CSS custom-property token system

All design tokens are defined as CSS custom properties on `:root` in
`custom-properties/colors.scss` and `custom-properties/shadows.scss`.

### The HSL-with-alpha pattern

Most colors are declared as HSL *components* — separate `-h`, `-s`, `-l`, `-a`
properties — plus a composed `hsla()` value:

```scss
--primary-h: 217deg;
--primary-s: 98%;
--primary-l: 53%;
--primary-a: 1;
--primary-hsl: var(--primary-h), var(--primary-s), var(--primary-l);
--primary: hsla(var(--primary-h), var(--primary-s), var(--primary-l), var(--primary-a));
```

This lets consumers change **one dimension** at a time. A hover state can soften the
same color without redefining it:

```scss
box-shadow: 0 0 0 2px hsla(var(--primary-hsl), 0.5);
```

Dark mode then tweaks only `--primary-l` (to `58%`) to keep sufficient contrast — the
composed `--primary` re-derives automatically.

### Grey scale

`--grey-50` … `--grey-900` form a Tailwind-style neutral ramp. The `-hsl` companions
(`--grey-100-hsl`, `--grey-500-hsl`, `--grey-900-hsl`, …) expose the raw HSL tuple for
the alpha-composition trick above. Dark mode reverses the scale: `--grey-900` becomes the
light-mode `--grey-50` value, and so on down the ladder.

### Bulma `--scheme-*` tokens

The `--scheme-main`, `--scheme-main-bis`, `--scheme-invert`, … block at the top of
`colors.scss` mirrors Bulma's own defaults and exists as a workaround for a
`bulma-css-variables` scoping bug (see [vikunja/frontend#1064][1]). Don't touch those
lines except to update Bulma. The Vikunja-specific tokens and overrides start below the
`// Vikunja specific variables` comment.

[1]: https://kolaente.dev/vikunja/frontend/issues/1064

### Adding a new token

1. Decide whether it's a color, shadow, or something else — add it to the matching file.
2. If it's a color that should shift in dark mode, declare HSL components (or reference an
   existing `--grey-*` token) so the dark-mode override only has to change one dimension.
3. Add the dark-mode override inside the `&.dark { @media screen { … } }` block.
4. Consume it with `var(--token-name)` — never re-declare the token in a component.

## Tailwind's limited role

Tailwind v4 is loaded via `styles/tailwind.css` and imported exactly once in `App.vue`.
Every utility is **prefixed with `tw`** (e.g. `class="tw-flex tw-gap-2"`) to avoid
collisions with the Bulma class names used throughout the app:

```css
@import "tailwindcss/theme.css" layer(theme) prefix(tw);
@import "tailwindcss/utilities.css" layer(utilities) prefix(tw);
```

Tailwind is meant for quick layout fixes inside `.vue` templates. For anything reusable —
especially anything that needs dark mode — prefer a CSS custom property or a scoped
`<style>` block.

## Theming and dark mode

There is no SCSS-level light/dark split. Instead:

1. `custom-properties/colors.scss` defines the light-mode values on `:root`.
2. The same file defines overrides inside `:root.dark { @media screen { … } }`.
3. A `dark` class is toggled on `<html>` at runtime (see the color scheme composable).
4. All components reference the tokens via `var(--…)`, so switching the class is enough
   to retheme the entire UI without touching the stylesheet.

The `@media screen` wrapper exists so the dark-mode overrides don't apply to
`@media print`, where we always want a light background.

## Where do I add …?

| I want to …                                                   | Put it in                                          |
| ------------------------------------------------------------- | -------------------------------------------------- |
| A new color / shadow token                                     | `custom-properties/colors.scss` or `shadows.scss` |
| A tweak that only affects one component                        | That component's `<style scoped>` block            |
| A SCSS variable reused by multiple components                  | `common-imports.scss` (no selectors!)              |
| A new `@font-face`                                             | `fonts.scss`                                       |
| A new Vue `<Transition>` class pair                            | `transitions.scss`                                 |
| A global rule targeting a Bulma class we can't scope yet       | The matching file in `theme/`                      |
| Re-enabling a Bulma partial                                    | Uncomment the `@import` in `global.scss`           |
| A Tailwind utility                                             | Use it inline with the `tw-` prefix                |
