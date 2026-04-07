# Frontend Specification

## Overview

MesaScore's frontend is a SvelteKit single-page application optimized for mobile-first use. It consumes the Go backend REST API exclusively. All data fetching happens client-side via authenticated API calls — SvelteKit server-side rendering is used only for the initial shell and SEO-irrelevant pages (the app is behind auth).

---

## Design Principles

1. **Mobile-first**: all layouts designed for 360px+ screens, scaled up for tablet/desktop.
2. **Minimal navigation depth**: most actions reachable within 2 taps from a dashboard.
3. **Immediate feedback**: optimistic UI for simple actions (add to collection, accept invite), spinners for network-dependent views.
4. **No offline support**: requires active connection. Page refresh is sufficient for stale data (no real-time updates per v1 scope).
5. **Consistent visual language**: card-based layouts, bottom navigation within party context, top bar for global context.

---

## Layout Structure

### Global Shell (authenticated)

```
+----------------------------------+
|  Top Bar: logo | app name | user avatar menu  |
+----------------------------------+
|                                  |
|         Page Content             |
|                                  |
+----------------------------------+
```

- Top bar is persistent across all authenticated pages.
- User avatar menu (dropdown): profile link, logout.
- On mobile, the top bar collapses to logo + hamburger or avatar only.

### Party Shell (within `/parties/:id/*`)

```
+----------------------------------+
|  Top Bar: ← back | party name | gear (admin)  |
+----------------------------------+
|                                  |
|         Page Content             |
|                                  |
+----------------------------------+
|  Bottom Nav: Dashboard | Sessions | Leaderboard | Members  |
+----------------------------------+
```

- Bottom navigation tabs for party sub-pages. Active tab highlighted.
- Back arrow returns to the global dashboard.
- Gear icon visible only to the party admin, links to party settings.
- On desktop (768px+), bottom nav becomes a sidebar.

### Unauthenticated Shell

```
+----------------------------------+
|  Centered: logo + app name      |
+----------------------------------+
|                                  |
|       Auth Form Content          |
|                                  |
+----------------------------------+
```

- No navigation, no top bar. Clean centered layout.
- Used for `/login`, `/register`, `/verify-email`.

---

## Pages

### P1 — Registration (`/register`)

**Layout**: unauthenticated shell.

**Content**:
- Form fields: username, display name, email, password, confirm password.
- Username field: real-time availability check with debounce (300ms). Shows green check or red X inline.
- Password: minimum 8 characters, validated client-side with inline feedback.
- Confirm password: must match, validated on blur.
- Submit button: disabled until all fields valid.
- Link to login page: "Already have an account? Log in".

**On submit**:
- POST to `/api/auth/register`.
- On success: show confirmation screen ("Check your email to verify your account").
- On `409` (duplicate): show field-level error on username or email.
- On `400`: show field-level validation errors from `fields` object.

---

### P2 — Email Verification (`/verify-email`)

**Layout**: unauthenticated shell.

**Content**:
- On mount: extract `token` from URL query params, call `GET /api/auth/verify-email?token=<token>`.
- Loading state: spinner with "Verifying your email...".
- Success: green check icon + "Email verified! You can now log in." + link to `/login`.
- Error (invalid/expired token): red icon + "This link is invalid or has expired." + link to `/login` with note about resending.

---

### P3 — Login (`/login`)

**Layout**: unauthenticated shell.

**Content**:
- Form fields: email, password.
- Submit button.
- Link to register: "Don't have an account? Register".

**On submit**:
- POST to `/api/auth/login`.
- On success: store JWT and user data, redirect to `/` (or saved redirect path from invite flow).
- On `401`: show "Invalid email or password" (generic).
- On `403` with `email_not_verified`: show "Your email is not verified" + "Resend verification email" button.

**Resend flow**:
- Clicking "Resend" prompts for email (pre-filled if available), POSTs to `/api/auth/resend-verification`.
- Always shows "If an account exists, we sent a new verification email" (no information leak).

---

### P4 — Global Dashboard (`/`)

**Layout**: global shell.

**Content**:
- **Pending invites section** (if any): cards showing party name, invited by whom, Accept/Decline buttons. Highlighted with a notification-style background.
- **My parties list**: cards for each party the user belongs to. Each card shows:
  - Party name
  - Member count
  - Last session date (or "No sessions yet")
  - Tap to navigate to `/parties/:id`
- **Global stats strip**: total sessions | total wins | current streak. Compact horizontal layout.
- **Create party FAB**: floating action button (bottom-right on mobile) linking to `/parties/new`.

**Empty state**: if no parties and no invites, show illustration + "Create your first party or ask a friend for an invite link."

**Data source**: `GET /api/users/me/dashboard`.

---

### P5 — Create Party (`/parties/new`)

**Layout**: global shell.

**Content**:
- Form fields: party name (required), description (optional, textarea).
- Submit button: "Create Party".

**On submit**:
- POST to `/api/parties`.
- On success: redirect to `/parties/:id` (the new party's dashboard).

---

### P6 — Join via Invite Link (`/join/:code`)

**Layout**: unauthenticated shell (works pre-login).

**Content**:
- On mount: call `GET /api/parties/join/:invite_code` to get party preview.
- Shows: party name, member count.
- If not logged in: "Log in or register to join this party" with links. Save invite code to `localStorage` for post-auth redirect.
- If logged in: "Join [Party Name]?" button.

**On join**:
- POST to `/api/parties/join/:invite_code`.
- On success: redirect to `/parties/:id`.
- On `409` (already member): redirect to `/parties/:id` with info toast "You're already in this party".
- On `404`: show "This invite link is no longer valid."

---

### P7 — Party Dashboard (`/parties/:id`)

**Layout**: party shell.

**Content**:
- **Party header**: party name. Gear icon links to settings (admin only).
- **Stats strip**: total sessions | unique games | members. Compact horizontal cards.
- **Current leader**: avatar + name + win count. Highlighted card. Taps to player stats. Shows "No sessions yet" if no data.
- **Most played game**: game cover thumbnail + name + session count.
- **Recent sessions feed**: last 5 sessions. Each row shows: game name, date, winner(s). Tap to session detail. "View all" link to `/parties/:id/sessions`.
- **Activity chart**: bar chart of sessions per month (last 12 months). Use a lightweight chart library (Chart.js or a Svelte-native option like `layerchart`).

**Data source**: `GET /api/parties/:id/dashboard`.

---

### P8 — Session History (`/parties/:id/sessions`)

**Layout**: party shell.

**Content**:
- **Filter bar**: game (dropdown), player (dropdown), session type (dropdown), date range (from/to inputs). Filters apply immediately (debounced).
- **Session list**: paginated cards. Each card shows:
  - Game name + cover thumbnail
  - Date
  - Session type badge (competitive/team/co-op/score)
  - Winner(s) or result
  - Participant count
  - Tap to navigate to session detail
- **Pagination**: "Load more" button at bottom (not full pagination controls — mobile-friendly).
- **Admin FAB**: "Log session" floating button (visible to admin only) linking to `/parties/:id/sessions/new`.

**Data source**: `GET /api/parties/:id/sessions` with query params.

---

### P9 — Session Detail (`/parties/:id/sessions/:session_id`)

**Layout**: party shell (no bottom nav on this page — full-screen detail).

**Content**:
- **Header**: game name + cover image.
- **Meta**: date, duration (if set), session type badge, notes (if any), "brought by" (if set), "logged by".
- **Results section**: varies by type:
  - **Competitive**: ranked list of players (1st, 2nd, ...) with optional scores. Winner highlighted.
  - **Team**: teams grouped, ranked. Players listed under team names with optional scores.
  - **Cooperative**: "Victory" or "Defeat" banner. All players listed.
  - **Score**: ranked list by score, winner highlighted.
- **Admin actions** (admin only): "Edit" button, "Delete" button (with confirmation modal).

**Data source**: `GET /api/parties/:id/sessions/:session_id`.

---

### P10 — Log Session (`/parties/:id/sessions/new`)

**Layout**: party shell (no bottom nav — full-screen form).

**Access**: party admin only. Non-admins redirected to party dashboard.

**Content — multi-step form**:

**Step 1: Game & basics**
- Game: searchable dropdown populated from `GET /api/parties/:id/available-games`.
- Session type: radio buttons (competitive / team / cooperative / score).
- Date played: date picker, default today, cannot be future.
- Duration (optional): number input in minutes.
- Brought by (optional): dropdown of party members who own the selected game.
- Notes (optional): textarea.

**Step 2: Participants**
- List of party members with checkboxes. At least 2 must be selected.
- Members fetched from `GET /api/parties/:id/members`.

**Step 3: Results** (varies by session type)
- **Competitive**: drag-to-rank or number input per player. Optional score field per player. Ties allowed (same rank).
- **Team**: assign players to teams (free-text team name or add to existing). Rank teams. Optional score per team.
- **Cooperative**: single toggle: Win or Loss. Applied to all.
- **Score**: number input per player. Ranks auto-calculated (highest score = rank 1). Ranks editable for override.

**Step 4: Review & submit**
- Summary of all entered data.
- "Save session" button.

**On submit**: POST to `/api/parties/:id/sessions`. On success: redirect to session detail page.

**Edit mode**: same form, pre-filled from `GET /api/parties/:id/sessions/:session_id`. Uses PATCH on submit.

---

### P11 — Party Leaderboard (`/parties/:id/leaderboard`)

**Layout**: party shell.

**Content**:
- **Leaderboard table**: rows for each party member.
  - Columns: rank, avatar, display name, wins, sessions, win rate.
  - Default sort: by wins descending.
  - Sortable columns: wins, win rate, sessions (tap column header to sort).
- Top 3 highlighted with medal colors (gold/silver/bronze).
- Each row taps to `/parties/:id/users/:userId` (player stats in party).

**Data source**: `GET /api/parties/:id/stats/leaderboard`.

---

### P12 — Party Members (`/parties/:id/members`)

**Layout**: party shell.

**Content**:
- **Members list**: each row shows avatar, display name, username, join date. Admin has a crown icon.
  - Tap member to go to `/parties/:id/users/:userId`.
  - Admin sees "Remove" button on each non-admin member (with confirmation modal).
- **Invite section** (admin only):
  - "Invite by username" button: opens modal with user search input.
    - Search calls `GET /api/users/search?q=<query>` (debounced).
    - Results exclude current members and users with pending invites.
    - Selecting a user sends invite via `POST /api/parties/:id/invites`.
  - "Share invite link" button: copies invite URL to clipboard. Shows toast on copy.
  - **Pending invites list**: shows invited users with status (pending/declined). Admin can re-invite declined users.

**Data source**: `GET /api/parties/:id/members`.

---

### P13 — Party Settings (`/parties/:id/settings`)

**Layout**: party shell (no bottom nav).

**Access**: party admin only. Non-admins redirected to party dashboard.

**Content**:
- **Edit party**: name + description form. Save button.
- **Invite link management**: current invite link displayed + "Copy" button + "Regenerate" button (with confirmation: "This will invalidate the current link").
- **Transfer ownership**: dropdown of party members + "Transfer" button (with confirmation modal: "Are you sure? You will lose admin privileges.").
- **Leave party** (visible only after transferring ownership — i.e., when current user is no longer admin, which means they already transferred): "Leave party" button with confirmation.

For non-admin members, this page shows only the "Leave party" button.

---

### P14 — Game Catalog (`/games`)

**Layout**: global shell.

**Content**:
- **Search bar**: text input with debounce. Filters the catalog list.
- **Sort dropdown**: name (default), rating, sessions.
- **Game grid/list**: cards showing:
  - Cover image thumbnail (placeholder if none)
  - Game name
  - Player count range (e.g., "2-4 players")
  - BGG rating (if available)
  - "In collection" badge if user owns it
  - Tap to `/games/:id`
- **Add game button**: top-right or FAB, links to `/games/new`.

**Data source**: `GET /api/games`.

---

### P15 — Game Detail (`/games/:id`)

**Layout**: global shell.

**Content**:
- **Header**: cover image (large), game name.
- **Details**: description, player count, BGG rating, BGG link (external, if `bgg_id` exists).
- **Collection toggle**: "Add to collection" or "Remove from collection" button.
- **BGG refresh** (visible to owners only): "Refresh BGG data" button. Shows spinner during fetch.
- **Owners section**: list of users who own this game (avatar + display name). Each taps to user profile.
- **Session count**: total sessions played globally with this game.

**Data source**: `GET /api/games/:id`.

---

### P16 — Add Game (`/games/new`)

**Layout**: global shell.

**Content**:

**BGG search tab** (default):
- Search input: queries `GET /api/bgg/search?q=<query>` (debounced, min 3 chars).
- Results list: name, year, thumbnail. Tap to select.
- On select: pre-fills form with BGG data. If `bgg_id` already exists in catalog, show message "This game is already in the catalog" + "Add to collection" button.
- Confirm form: name (editable), description, cover image URL (pre-filled). Submit adds to catalog + user's collection.

**Manual entry tab**:
- Form: name (required), description, cover image URL, min/max players.
- Submit adds to catalog + user's collection.

**On submit**: POST to `/api/games`. On success: redirect to `/games/:id`.

---

### P17 — User Profile (`/users/:id`)

**Layout**: global shell.

**Content**:
- **Profile header**: avatar (large), display name, username, member since date.
- **Edit button** (own profile only): opens inline edit for display name and avatar URL.
- **Tabs**: Stats | Collection

**Stats tab** (default):
- Total sessions, total wins, win rate.
- Current streak, best streak.
- Most played game (name + session count).
- Best win rate game (name + win rate, min 3 sessions).
- Nemesis: avatar + name + "You lost N times against them."
- Punching bag: avatar + name + "You beat them N times."
- Per-game breakdown: table (game, sessions, wins, win rate).
- Head-to-head: table (opponent avatar+name, sessions together, your wins, their wins).

**Collection tab**:
- Grid of game cards (cover + name).
- Own profile: shows "Remove" button on each card.
- Other users: view-only.

**Data source**: `GET /api/users/:id`, `GET /api/users/:id/stats`, `GET /api/users/:id/collection`.

---

### P18 — Player Stats in Party (`/parties/:id/users/:userId`)

**Layout**: party shell.

**Content**:
- **Profile header**: avatar, display name, username.
- **Tabs**: Party Stats | Global Stats

**Party Stats tab** (default):
- Same metrics as P17 stats tab, but scoped to this party.
- Head-to-head only includes sessions in this party.

**Global Stats tab**:
- Same as P17 stats tab (all parties).

**Data source**: `GET /api/parties/:id/users/:userId/stats` (party), `GET /api/users/:userId/stats` (global).

---

## Interaction Patterns

### Toast Notifications

Non-blocking notifications for:
- Successful actions: "Joined party", "Game added to collection", "Session logged", "Invite sent", "Link copied".
- Info messages: "You're already a member".
- Errors: "Something went wrong. Try again."

Toasts appear at the top of the screen, auto-dismiss after 3 seconds, dismissable by tap.

### Confirmation Modals

Required before destructive or significant actions:
- Delete session
- Remove member
- Leave party
- Transfer ownership
- Regenerate invite link

Modal content: title, description of consequence, Cancel + Confirm buttons. Confirm button uses a danger color for destructive actions.

### Loading States

- **Page-level**: centered spinner when fetching initial page data.
- **Button-level**: spinner replaces button text during async action.
- **Skeleton screens**: for dashboard cards and lists while data loads (preferred over blank spinners for data-heavy pages).

### Empty States

Every list/grid has an empty state:
- Parties list: "Create your first party or join one with an invite link."
- Session history: "No sessions yet. Log your first game!" (admin) / "No sessions logged yet." (member).
- Game catalog: "No games in the catalog yet. Add one!"
- Collection: "Your collection is empty. Browse the catalog to add games."
- Leaderboard: "Play some games first to see who's winning!"

### Form Validation

- Client-side validation on blur for individual fields.
- Full validation on submit.
- Field-level error messages appear below the input.
- API error responses mapped to specific fields when possible (via `fields` object).

---

## Responsive Breakpoints

| Breakpoint | Name | Layout changes |
|---|---|---|
| < 640px | Mobile | Single column, bottom nav, FABs, compact cards |
| 640–1024px | Tablet | Two-column grids where appropriate, larger cards |
| > 1024px | Desktop | Sidebar nav replaces bottom nav, three-column grids, wider forms |

---

## Accessibility

- All interactive elements keyboard-navigable.
- ARIA labels on icon-only buttons (gear, remove, close).
- Color is never the sole indicator — icons/text accompany status colors.
- Focus management on modal open/close.
- Form inputs have associated labels (not just placeholders).
- Minimum tap target: 44x44px on mobile.
- Sufficient color contrast (WCAG AA).

---

## Out of Scope (v1)

- Dark mode (follow system preference only if trivial; otherwise light only).
- PWA / offline caching.
- Push notifications.
- Drag-and-drop (except for session ranking, which falls back to number inputs).
- Animations beyond basic transitions.
- i18n / localization (English only).
