# Dark Theme Specification

## Scope

This document defines the dark theme feature for MesaScore. Dark mode was explicitly deferred in `docs/04-frontend/spec.md` (Out of Scope, v1). This spec promotes it to a supported, first-class feature.

The dark theme covers:
- A user-controlled theme preference (light / dark / system).
- Persistence of that preference across sessions.
- Complete visual coverage of all app surfaces.

---

## Design Goals

1. **User control**: the user can explicitly choose light, dark, or system (OS preference). System is the default.
2. **Persistence**: the chosen preference is saved and restored on every visit.
3. **No flash**: the theme is applied before first paint — no visible light-to-dark flash on load.
4. **Accessible**: dark theme colors meet WCAG AA contrast requirements.
5. **Coherent palette**: dark theme is not a simple color inversion — it uses purpose-built dark surface colors that feel native to the app's visual language.

---

## Feature Definition

### TH1 — Theme Preference

**As a user, I want to choose between light mode, dark mode, and system default so the app matches my environment.**

Acceptance criteria:
- A theme toggle is accessible from the user avatar menu (top bar) on all authenticated pages.
- On unauthenticated pages (login, register, verify-email), the toggle is available in the top-right corner.
- Three options: **Light**, **Dark**, **System** (follows OS `prefers-color-scheme`).
- The active selection is visually indicated (highlighted / checked).
- Changing the preference takes effect immediately without a page reload.

---

### TH2 — Preference Persistence

**As a user, I want my theme preference to be remembered so I don't have to set it every time.**

Acceptance criteria:
- The preference is stored in `localStorage` under the key `mesascore_theme`.
- On subsequent visits, the stored preference is restored.
- If no stored preference exists, the default is `system`.
- Clearing browser storage resets to `system`.

---

### TH3 — No Flash on Load

**As a user, I should not see a flash of the wrong theme when the page loads.**

Acceptance criteria:
- The correct theme class (`dark` or `light`) is applied to `<html>` before any visible rendering.
- This is achieved via an inline script in `app.html` that runs synchronously before page paint.
- The inline script reads `localStorage` and applies the class; no external dependency.

---

### TH4 — Full Visual Coverage

**All app surfaces must be themed correctly in both light and dark mode.**

Acceptance criteria:
- Surfaces covered: backgrounds, cards, top bar, bottom nav, sidebar, modals, toasts, form inputs, dropdowns, badges, avatars, skeletons, charts.
- No surface falls back to browser-default white or black.
- Images and game cover thumbnails are not altered; a subtle overlay or border is sufficient to integrate them visually in dark mode.
- Medal colors (gold/silver/bronze) on the leaderboard remain legible in both modes.
- Status colors (success green, error red, info blue) are adjusted for dark mode to maintain contrast without being harsh.
- The activity chart uses theme-aware colors for bars and axis labels.

---

### TH5 — System Preference Sync

**As a user with `system` selected, I want the app to update automatically if I change my OS dark mode setting.**

Acceptance criteria:
- When the stored preference is `system`, the app listens to `window.matchMedia('(prefers-color-scheme: dark)')` change events.
- If the OS setting changes while the app is open, the theme updates immediately without a reload.
- When a user has explicitly set `light` or `dark`, OS changes are ignored.

---

## Color Palette

The following semantic color tokens are defined for both themes. Exact hex values are set during implementation; these are the required token names and their purpose.

| Token | Light | Dark | Purpose |
|---|---|---|---|
| `--color-bg` | near-white | dark gray | Page background |
| `--color-surface` | white | slightly lighter dark | Card / panel background |
| `--color-surface-raised` | light gray | lighter dark | Elevated surface (modal, dropdown) |
| `--color-border` | gray-200 | gray-700 | Dividers and outlines |
| `--color-text-primary` | gray-900 | gray-100 | Primary text |
| `--color-text-secondary` | gray-500 | gray-400 | Labels, metadata, placeholders |
| `--color-text-disabled` | gray-300 | gray-600 | Disabled state text |
| `--color-primary` | brand color | same (adjusted saturation if needed) | Primary action color |
| `--color-primary-hover` | darker primary | lighter primary | Hover state |
| `--color-danger` | red-600 | red-400 | Destructive action |
| `--color-success` | green-600 | green-400 | Success state |
| `--color-info` | blue-600 | blue-400 | Info / neutral state |
| `--color-overlay` | black/40% | black/60% | Modal backdrop |

---

## Theme Toggle Component

The toggle is a compact segmented control with three options: a sun icon (Light), a monitor icon (System), and a moon icon (Dark). Selecting an option:
1. Updates the `mesascore_theme` key in `localStorage`.
2. Recomputes the active theme class on `<html>`.
3. Updates the reactive store so all components re-render.

On mobile (< 640px), the toggle can be a single icon button that cycles Light → Dark → System → Light, with a tooltip or aria-label identifying the current mode.

---

## Accessibility

- The theme toggle button has an `aria-label` that describes the current mode and what clicking will do (e.g., "Switch to dark mode").
- Color contrast in dark mode meets WCAG AA: minimum 4.5:1 for normal text, 3:1 for large text and UI components.
- Focus rings are visible in dark mode (must not disappear against dark backgrounds).
- The dark theme is not the sole means of conveying status — icons and text always accompany color.

---

## Out of Scope

- Per-party or per-page theme overrides.
- High-contrast mode (separate from dark mode).
- Custom color themes / theming beyond light/dark.
- Syncing theme preference to the backend (it is client-only).
