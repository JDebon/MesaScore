# Dark Theme — Technical Tasks

## Stack Notes

- **Styling**: Tailwind CSS 4 with CSS custom properties for semantic color tokens.
- **Theme switching**: Tailwind's `dark:` variant triggered by a `.dark` class on `<html>`.
- **State**: a dedicated `theme.svelte.ts` store using Svelte 5 runes.
- **No-flash**: inline `<script>` in `src/app.html`.

---

## T11.1 — CSS Token Definitions

In `src/app.css`, define semantic color tokens as CSS custom properties inside `:root` and `.dark` selectors. Tailwind 4 supports referencing these via `theme()` or directly in utilities.

```css
@import 'tailwindcss';

:root {
  --color-bg:               #f9fafb;   /* gray-50  */
  --color-surface:          #ffffff;
  --color-surface-raised:   #f3f4f6;   /* gray-100 */
  --color-border:           #e5e7eb;   /* gray-200 */
  --color-text-primary:     #111827;   /* gray-900 */
  --color-text-secondary:   #6b7280;   /* gray-500 */
  --color-text-disabled:    #d1d5db;   /* gray-300 */
  --color-primary:          #4f46e5;   /* indigo-600 */
  --color-primary-hover:    #4338ca;   /* indigo-700 */
  --color-danger:           #dc2626;   /* red-600   */
  --color-success:          #16a34a;   /* green-600 */
  --color-info:             #2563eb;   /* blue-600  */
  --color-overlay:          rgb(0 0 0 / 0.4);
}

.dark {
  --color-bg:               #0f172a;   /* slate-900 */
  --color-surface:          #1e293b;   /* slate-800 */
  --color-surface-raised:   #334155;   /* slate-700 */
  --color-border:           #475569;   /* slate-600 */
  --color-text-primary:     #f1f5f9;   /* slate-100 */
  --color-text-secondary:   #94a3b8;   /* slate-400 */
  --color-text-disabled:    #475569;   /* slate-600 */
  --color-primary:          #6366f1;   /* indigo-500 */
  --color-primary-hover:    #818cf8;   /* indigo-400 */
  --color-danger:           #f87171;   /* red-400   */
  --color-success:          #4ade80;   /* green-400 */
  --color-info:             #60a5fa;   /* blue-400  */
  --color-overlay:          rgb(0 0 0 / 0.6);
}
```

Register the tokens as Tailwind theme extensions in `vite.config.ts` or a `tailwind.config.ts` (Tailwind 4 uses `@theme` in CSS):

```css
@theme {
  --color-bg: var(--color-bg);
  --color-surface: var(--color-surface);
  --color-surface-raised: var(--color-surface-raised);
  --color-border: var(--color-border);
  --color-text-primary: var(--color-text-primary);
  --color-text-secondary: var(--color-text-secondary);
  --color-text-disabled: var(--color-text-disabled);
  --color-primary: var(--color-primary);
  --color-primary-hover: var(--color-primary-hover);
  --color-danger: var(--color-danger);
  --color-success: var(--color-success);
  --color-info: var(--color-info);
}
```

This lets components use `bg-surface`, `text-text-primary`, `border-border`, etc. directly as Tailwind utilities.

---

## T11.2 — Theme Store

`src/lib/stores/theme.svelte.ts`:

```ts
import { browser } from '$app/environment';

type ThemePreference = 'light' | 'dark' | 'system';

const STORAGE_KEY = 'mesascore_theme';

let preference = $state<ThemePreference>('system');
let resolvedDark = $state(false);

function getSystemDark(): boolean {
  return browser && window.matchMedia('(prefers-color-scheme: dark)').matches;
}

function applyTheme(pref: ThemePreference) {
  const isDark = pref === 'dark' || (pref === 'system' && getSystemDark());
  resolvedDark = isDark;
  if (browser) {
    document.documentElement.classList.toggle('dark', isDark);
  }
}

// Initialize from localStorage
if (browser) {
  const stored = localStorage.getItem(STORAGE_KEY) as ThemePreference | null;
  preference = stored ?? 'system';
  applyTheme(preference);

  // Listen for OS preference changes when in system mode
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    if (preference === 'system') applyTheme('system');
  });
}

export function getThemePreference(): ThemePreference { return preference; }
export function isResolvedDark(): boolean { return resolvedDark; }

export function setThemePreference(pref: ThemePreference) {
  preference = pref;
  if (browser) localStorage.setItem(STORAGE_KEY, pref);
  applyTheme(pref);
}
```

---

## T11.3 — No-Flash Inline Script

In `src/app.html`, add an inline `<script>` as the first child of `<head>`. It must run synchronously before any CSS or JS is parsed.

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    %sveltekit.head%
    <script>
      (function() {
        var pref = localStorage.getItem('mesascore_theme') || 'system';
        var isDark = pref === 'dark' || (pref === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);
        if (isDark) document.documentElement.classList.add('dark');
      })();
    </script>
  </head>
  <body data-sveltekit-preload-data="hover">
    <div style="display: contents">%sveltekit.body%</div>
  </body>
</html>
```

The script is intentionally plain JS (no module, no TypeScript) to execute synchronously. It mirrors the logic in the store.

---

## T11.4 — Theme Toggle Component

`src/lib/components/ui/ThemeToggle.svelte`:

```svelte
<script lang="ts">
  import { getThemePreference, setThemePreference } from '$stores/theme.svelte';

  type Preference = 'light' | 'dark' | 'system';

  const options: { value: Preference; label: string; icon: string }[] = [
    { value: 'light',  label: 'Light',  icon: '☀️' },
    { value: 'system', label: 'System', icon: '💻' },
    { value: 'dark',   label: 'Dark',   icon: '🌙' }
  ];

  let current = $derived(getThemePreference());
</script>

<div role="group" aria-label="Theme preference" class="flex rounded-lg border border-border overflow-hidden">
  {#each options as opt}
    <button
      type="button"
      aria-label="{opt.label} theme"
      aria-pressed={current === opt.value}
      class="px-3 py-1.5 text-sm transition-colors
             {current === opt.value
               ? 'bg-primary text-white'
               : 'bg-surface text-text-secondary hover:bg-surface-raised'}"
      onclick={() => setThemePreference(opt.value)}
    >
      <span aria-hidden="true">{opt.icon}</span>
      <span class="sr-only">{opt.label}</span>
    </button>
  {/each}
</div>
```

Use real SVG icons (sun, monitor, moon) in the final implementation instead of emoji. The component is self-contained and does not require props.

---

## T11.5 — Integration Points

### Top Bar

Add `<ThemeToggle />` to the user avatar dropdown menu in `src/lib/components/layout/TopBar.svelte`. Place it above the "Profile" and "Logout" links.

### Unauthenticated Shell

Add `<ThemeToggle />` to `src/routes/(auth)/+layout.svelte` in the top-right corner, positioned absolutely or within a flex header row.

---

## T11.6 — Component Migration

Update all existing components to use the semantic token utilities (`bg-surface`, `text-text-primary`, etc.) instead of hard-coded Tailwind color classes. Work through components in this order:

1. **Layouts**: `TopBar.svelte`, `BottomNav.svelte`, `PartyHeader.svelte`, root layout backgrounds.
2. **UI primitives**: `Button.svelte`, `Input.svelte`, `Modal.svelte`, `Toast.svelte`, `Badge.svelte`, `Skeleton.svelte`.
3. **Cards**: `PartyCard.svelte`, `InviteCard.svelte`, `GameCard.svelte`, `SessionCard.svelte`.
4. **Tables & lists**: `LeaderboardTable.svelte`, `HeadToHeadTable.svelte`, `PerGameTable.svelte`, `MemberRow.svelte`.
5. **Stats & charts**: `StatsStrip.svelte`, `ActivityChart.svelte`, `PlayerHighlight.svelte`.
6. **Forms**: all form pages — inputs, textareas, selects, radio buttons, checkboxes.
7. **Misc**: `EmptyState.svelte`, `Avatar.svelte`, `Spinner.svelte`, confirmation modals, dropdown menus.

For each component: replace `bg-white` → `bg-surface`, `bg-gray-50` → `bg-bg`, `text-gray-900` → `text-text-primary`, `text-gray-500` → `text-text-secondary`, `border-gray-200` → `border-border`, and so on.

---

## T11.7 — Chart Theme Awareness

`src/lib/components/stats/ActivityChart.svelte`:

The chart must react to the resolved theme. Derive the bar color and axis label color from CSS custom properties at render time.

```ts
// Read the current resolved color from the CSS variable
function getCSSColor(variable: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(variable).trim();
}
```

On theme change, the chart must re-render with updated colors. Use a `$effect` that watches `isResolvedDark()` from the theme store and calls the chart's update method.

---

## T11.8 — Medal Colors

The leaderboard top-3 medal rows use gold/silver/bronze highlights. These colors must work in both themes.

Recommended approach — use Tailwind's opacity modifier rather than separate dark tokens:

| Rank | Light | Dark |
|------|-------|------|
| 1st (gold)   | `bg-amber-100 text-amber-800` | `bg-amber-900/40 text-amber-300` |
| 2nd (silver) | `bg-slate-100 text-slate-600` | `bg-slate-700/40 text-slate-300` |
| 3rd (bronze) | `bg-orange-100 text-orange-700` | `bg-orange-900/40 text-orange-300` |

Use Tailwind's `dark:` variant classes since medal styles are self-contained and do not map to a semantic token.

---

## T11.9 — Focus Ring Visibility

All interactive elements need a visible focus ring in dark mode. Add to `src/app.css`:

```css
:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}
```

This replaces Tailwind's default focus utilities globally. Components using `focus:ring-*` should be updated to rely on this rule instead, unless a component-specific override is needed.

---

## T11.10 — Implementation Order

1. `T11.1` — CSS tokens in `app.css`
2. `T11.2` — Theme store
3. `T11.3` — No-flash script in `app.html`
4. `T11.4` — `ThemeToggle` component
5. `T11.5` — Wire toggle into `TopBar` and auth layout
6. `T11.9` — Focus ring CSS
7. `T11.6` — Component migration (layouts → primitives → cards → tables → forms → misc)
8. `T11.7` — Chart theme awareness
9. `T11.8` — Medal color classes
10. Manual visual pass through all routes in both themes
