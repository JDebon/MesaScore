# Frontend — Technical Tasks

## Stack

- **Framework**: SvelteKit 2 (Svelte 5 with runes)
- **Adapter**: `@sveltejs/adapter-node` (Docker deployment)
- **Styling**: Tailwind CSS 4
- **Charts**: `layerchart` (Svelte-native, lightweight) or `chart.js` with `svelte-chartjs`
- **HTTP**: native `fetch` with typed wrappers
- **State**: Svelte 5 runes (`$state`, `$derived`) + context for shared state
- **Package manager**: pnpm

---

## T10.1 — Project Scaffolding

```bash
pnpm create svelte@latest frontend
# Select: Skeleton project, TypeScript, ESLint, Prettier
cd frontend
pnpm add -D tailwindcss @tailwindcss/vite
pnpm add @sveltejs/adapter-node
```

Configure `svelte.config.js`:
```js
import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

export default {
  preprocess: vitePreprocess(),
  kit: {
    adapter: adapter({ out: 'build' }),
    alias: {
      '$api': 'src/lib/api',
      '$components': 'src/lib/components',
      '$stores': 'src/lib/stores'
    }
  }
};
```

Configure `vite.config.ts`:
```ts
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [tailwindcss(), sveltekit()]
});
```

Add to `src/app.css`:
```css
@import 'tailwindcss';
```

Environment variable in `.env`:
```
PUBLIC_API_URL=http://localhost:8080
```

---

## T10.2 — Directory Structure

```
frontend/
  src/
    app.css                    # Tailwind import
    app.html                   # HTML shell
    lib/
      api/
        client.ts              # Base fetch wrapper with auth
        auth.ts                # Auth endpoints
        users.ts               # User endpoints
        parties.ts             # Party endpoints
        games.ts               # Game + BGG endpoints
        sessions.ts            # Session endpoints
        stats.ts               # Stats endpoints
        types.ts               # Shared API response types
      components/
        ui/                    # Generic UI primitives
          Button.svelte
          Input.svelte
          Modal.svelte
          Toast.svelte
          Spinner.svelte
          Badge.svelte
          Avatar.svelte
          EmptyState.svelte
          Skeleton.svelte
        layout/
          TopBar.svelte
          BottomNav.svelte
          PartyHeader.svelte
        game/
          GameCard.svelte
          GameGrid.svelte
        session/
          SessionCard.svelte
          SessionResultView.svelte
          SessionForm/
            Step1Game.svelte
            Step2Participants.svelte
            Step3Results.svelte
            Step4Review.svelte
        party/
          PartyCard.svelte
          InviteCard.svelte
          MemberRow.svelte
          InviteModal.svelte
        stats/
          StatsStrip.svelte
          LeaderboardTable.svelte
          HeadToHeadTable.svelte
          PerGameTable.svelte
          ActivityChart.svelte
          PlayerHighlight.svelte  # Nemesis, punching bag, etc.
      stores/
        auth.svelte.ts         # Auth state (user, token, login/logout)
        toast.svelte.ts        # Toast notification queue
    routes/
      +layout.svelte           # Root layout (loads auth state)
      +layout.ts               # Root layout load (check auth)
      (auth)/                   # Unauthenticated group layout
        +layout.svelte          # Centered, no nav
        login/
          +page.svelte
        register/
          +page.svelte
        verify-email/
          +page.svelte
      (app)/                    # Authenticated group layout
        +layout.svelte          # Global shell (top bar)
        +layout.ts              # Auth guard (redirect to /login)
        +page.svelte            # Global dashboard (/)
        games/
          +page.svelte          # Game catalog
          [id]/
            +page.svelte        # Game detail
          new/
            +page.svelte        # Add game
        users/
          [id]/
            +page.svelte        # User profile + stats
        parties/
          new/
            +page.svelte        # Create party
          [id]/
            +layout.svelte      # Party shell (bottom nav)
            +layout.ts          # Load party data, check membership
            +page.svelte        # Party dashboard
            sessions/
              +page.svelte      # Session history
              new/
                +page.svelte    # Log session (admin)
              [sessionId]/
                +page.svelte    # Session detail
                edit/
                  +page.svelte  # Edit session (admin)
            leaderboard/
              +page.svelte
            members/
              +page.svelte
            settings/
              +page.svelte      # Party settings (admin)
            users/
              [userId]/
                +page.svelte    # Player stats in party
      join/
        [code]/
          +page.svelte          # Join preview (outside auth layout)
  static/
    favicon.png
  Dockerfile
```

---

## T10.3 — API Client

`src/lib/api/client.ts`:

```ts
import { getToken, clearAuth } from '$stores/auth.svelte';
import { PUBLIC_API_URL } from '$env/static/public';

class ApiError extends Error {
  status: number;
  fields?: Record<string, string>;

  constructor(status: number, message: string, fields?: Record<string, string>) {
    super(message);
    this.status = status;
    this.fields = fields;
  }
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json'
  };

  const token = getToken();
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(`${PUBLIC_API_URL}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined
  });

  // Handle token refresh
  const newToken = res.headers.get('X-New-Token');
  if (newToken) {
    setToken(newToken);
  }

  if (res.status === 401) {
    clearAuth();
    window.location.href = '/login';
    throw new ApiError(401, 'Unauthorized');
  }

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Unknown error' }));
    throw new ApiError(res.status, err.error, err.fields);
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}

export const api = {
  get: <T>(path: string) => request<T>('GET', path),
  post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
  patch: <T>(path: string, body?: unknown) => request<T>('PATCH', path, body),
  delete: <T>(path: string) => request<T>('DELETE', path)
};

export { ApiError };
```

Each resource module wraps the base client with typed functions. Example for `src/lib/api/auth.ts`:

```ts
import { api } from './client';
import type { LoginResponse, User } from './types';

export const authApi = {
  register: (data: { username: string; display_name: string; email: string; password: string }) =>
    api.post<{ message: string }>('/api/auth/register', data),

  login: (email: string, password: string) =>
    api.post<LoginResponse>('/api/auth/login', { email, password }),

  verifyEmail: (token: string) =>
    api.get<{ message: string }>(`/api/auth/verify-email?token=${token}`),

  resendVerification: (email: string) =>
    api.post<void>('/api/auth/resend-verification', { email }),

  checkUsername: (username: string) =>
    api.get<{ available: boolean }>(`/api/auth/check-username?username=${encodeURIComponent(username)}`)
};
```

---

## T10.4 — Auth Store

`src/lib/stores/auth.svelte.ts`:

```ts
import { browser } from '$app/environment';

interface AuthUser {
  id: string;
  username: string;
  display_name: string;
  avatar_url: string | null;
}

let token = $state<string | null>(null);
let user = $state<AuthUser | null>(null);

// Initialize from localStorage on first load
if (browser) {
  token = localStorage.getItem('mesascore_token');
  const savedUser = localStorage.getItem('mesascore_user');
  if (savedUser) user = JSON.parse(savedUser);
}

export function getToken(): string | null { return token; }
export function getUser(): AuthUser | null { return user; }
export function isAuthenticated(): boolean { return !!token; }

export function setAuth(newToken: string, newUser: AuthUser) {
  token = newToken;
  user = newUser;
  if (browser) {
    localStorage.setItem('mesascore_token', newToken);
    localStorage.setItem('mesascore_user', JSON.stringify(newUser));
  }
}

export function setToken(newToken: string) {
  token = newToken;
  if (browser) localStorage.setItem('mesascore_token', newToken);
}

export function clearAuth() {
  token = null;
  user = null;
  if (browser) {
    localStorage.removeItem('mesascore_token');
    localStorage.removeItem('mesascore_user');
  }
}

export function updateUser(updates: Partial<AuthUser>) {
  if (user) {
    user = { ...user, ...updates };
    if (browser) localStorage.setItem('mesascore_user', JSON.stringify(user));
  }
}
```

---

## T10.5 — Toast Store

`src/lib/stores/toast.svelte.ts`:

```ts
interface Toast {
  id: number;
  message: string;
  type: 'success' | 'error' | 'info';
}

let toasts = $state<Toast[]>([]);
let nextId = 0;

export function getToasts(): Toast[] { return toasts; }

export function addToast(message: string, type: Toast['type'] = 'success') {
  const id = nextId++;
  toasts = [...toasts, { id, message, type }];
  setTimeout(() => dismissToast(id), 3000);
}

export function dismissToast(id: number) {
  toasts = toasts.filter(t => t.id !== id);
}
```

---

## T10.6 — Auth Guard

`src/routes/(app)/+layout.ts`:

```ts
import { redirect } from '@sveltejs/kit';
import { isAuthenticated } from '$stores/auth.svelte';
import { browser } from '$app/environment';

export function load({ url }) {
  if (browser && !isAuthenticated()) {
    const redirectTo = url.pathname + url.search;
    throw redirect(302, `/login?redirect=${encodeURIComponent(redirectTo)}`);
  }
}
```

Post-login redirect: after successful login, check for `redirect` query param and navigate there.

---

## T10.7 — Party Layout and Context

`src/routes/(app)/parties/[id]/+layout.ts`:

```ts
import { partiesApi } from '$api/parties';
import { error } from '@sveltejs/kit';

export async function load({ params }) {
  try {
    const party = await partiesApi.get(params.id);
    return { party };
  } catch (e) {
    if (e.status === 403) throw error(403, 'You are not a member of this party');
    if (e.status === 404) throw error(404, 'Party not found');
    throw e;
  }
}
```

The party layout component reads `data.party` and renders the party shell with bottom nav. It also exposes the party context for child pages.

---

## T10.8 — Invite Link Flow

The `/join/[code]` route lives outside the `(app)` layout group so it works for unauthenticated users.

Flow:
1. Page loads, calls `GET /api/parties/join/:code` for preview.
2. If user is not authenticated:
   - Save `code` to `localStorage` key `mesascore_pending_join`.
   - Show login/register links.
3. After login, root layout checks for `mesascore_pending_join`:
   - If present, redirect to `/join/:code` and clear it.
4. Authenticated user on join page: show "Join" button, POST on click.

---

## T10.9 — Session Form Implementation

The session form (`/parties/:id/sessions/new`) is a multi-step form using local component state.

```ts
// State management within the form component
let step = $state(1);
let formData = $state({
  game_id: '',
  session_type: 'competitive' as SessionType,
  played_at: new Date().toISOString().split('T')[0],
  duration_minutes: null as number | null,
  brought_by_user_id: null as string | null,
  notes: '',
  participants: [] as ParticipantInput[]
});
```

Step transitions validate the current step before advancing. The review step (4) shows a read-only summary before POST.

For **competitive** result entry: render each selected participant with a rank number input. Default ranks based on selection order (1, 2, 3, ...). Ties via same rank number.

For **score** result entry: render each participant with a score input. Ranks auto-compute (sorted by score descending). Override via editable rank field.

For **team** result entry: render a team assignment UI. Each participant gets a team name (text input or select from existing teams). Teams are ranked.

For **cooperative**: single Win/Loss toggle.

---

## T10.10 — Chart Component

`src/lib/components/stats/ActivityChart.svelte`:

Use `layerchart` (Svelte-native):

```bash
pnpm add layerchart
```

The component receives `data: { month: string; count: number }[]` as a prop and renders a vertical bar chart. Months on x-axis, count on y-axis. Bars colored with the app's primary color. Responsive width.

---

## T10.11 — Responsive Navigation

**Bottom nav** (`src/lib/components/layout/BottomNav.svelte`):
- Visible only on screens < 1024px.
- 4 tabs: Dashboard, Sessions, Leaderboard, Members.
- Active tab determined by current route.
- Fixed to bottom of viewport.

**Sidebar nav** (desktop):
- On screens >= 1024px, bottom nav hides and a left sidebar appears.
- Same links, vertical layout.
- Implemented via CSS media queries within the party layout, not separate components.

---

## T10.12 — API Types

`src/lib/api/types.ts` — TypeScript interfaces matching all API response shapes from the API spec. Key types:

```ts
export interface User {
  id: string;
  username: string;
  display_name: string;
  avatar_url: string | null;
}

export interface Party {
  id: string;
  name: string;
  description: string | null;
  admin: User;
  invite_code: string;
  member_count: number;
  created_at: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface DashboardResponse {
  parties: { id: string; name: string; member_count: number; last_session_at: string | null }[];
  pending_invites: PendingInvite[];
  global_stats: { total_sessions: number; total_wins: number; current_streak: number };
}

export interface PendingInvite {
  id: string;
  party: { id: string; name: string };
  invited_by: { id: string; display_name: string };
  created_at: string;
}

export interface SessionDetail {
  id: string;
  game: { id: string; name: string; cover_image_url: string | null };
  session_type: 'competitive' | 'team' | 'cooperative' | 'score';
  played_at: string;
  duration_minutes: number | null;
  notes: string | null;
  brought_by: { id: string; display_name: string } | null;
  created_by: { id: string; display_name: string };
  created_at: string;
  participants: SessionParticipant[];
}

export interface SessionParticipant {
  user: User;
  team_name: string | null;
  rank: number | null;
  score: number | null;
  result: 'win' | 'loss' | 'draw' | null;
}

export interface Game {
  id: string;
  bgg_id: number | null;
  name: string;
  description: string | null;
  cover_image_url: string | null;
  min_players: number | null;
  max_players: number | null;
  bgg_rating: number | null;
  session_count: number;
  in_my_collection: boolean;
}

export interface LeaderboardEntry {
  user: User;
  wins: number;
  sessions: number;
  win_rate: number;
}

export interface UserStats {
  user: User;
  total_sessions: number;
  total_wins: number;
  win_rate: number;
  current_streak: number;
  best_streak: number;
  most_played_game: { id: string; name: string; session_count: number } | null;
  best_win_rate_game: { id: string; name: string; win_rate: number } | null;
  nemesis: { id: string; display_name: string; losses_against: number } | null;
  punching_bag: { id: string; display_name: string; wins_against: number } | null;
  per_game: PerGameStat[];
  head_to_head: HeadToHeadEntry[];
}

export interface PerGameStat {
  game: { id: string; name: string; cover_image_url: string | null };
  sessions: number;
  wins: number;
  win_rate: number;
}

export interface HeadToHeadEntry {
  opponent: User;
  sessions_together: number;
  this_user_wins: number;
  opponent_wins: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
}

export interface PartyDashboard {
  total_sessions: number;
  total_unique_games: number;
  total_members: number;
  current_leader: { user: User; wins: number } | null;
  most_played_game: { id: string; name: string; cover_image_url: string | null; session_count: number } | null;
  sessions_per_month: { month: string; count: number }[];
  recent_sessions: {
    id: string;
    game: { id: string; name: string; cover_image_url: string | null };
    played_at: string;
    session_type: string;
    winners: { id: string; display_name: string }[];
  }[];
}
```

---

## T10.13 — Dockerfile

```dockerfile
FROM node:22-alpine AS build
WORKDIR /app
RUN corepack enable
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY . .
RUN pnpm build

FROM node:22-alpine
WORKDIR /app
RUN corepack enable
COPY --from=build /app/build ./build
COPY --from=build /app/package.json ./
COPY --from=build /app/pnpm-lock.yaml ./
RUN pnpm install --prod --frozen-lockfile
ENV PORT=3000
EXPOSE 3000
CMD ["node", "build"]
```

---

## T10.14 — Environment & Proxy

The frontend needs a single public env var:

```
PUBLIC_API_URL=http://localhost:8080
```

In production (behind Caddy), this becomes the same origin (`/api` proxied by Caddy). The `PUBLIC_API_URL` should be set to the base URL without trailing slash.

Caddy routes:
- `/api/*` → backend:8080
- `/*` → frontend:3000

---

## T10.15 — Implementation Order

Follow the same feature order as the backend. Each phase includes its API module, types, components, and route pages.

1. **Project setup**: scaffolding, Tailwind, API client, auth store, toast store, layouts (T10.1–T10.6)
2. **Auth pages**: register, verify-email, login, auth guard, invite link redirect (P1–P3, T10.8)
3. **Global dashboard**: dashboard page, party cards, invite cards, stats strip (P4)
4. **Party basics**: create party, party layout with bottom nav, party dashboard (P5, P7, T10.7, T10.11)
5. **Members & invites**: members page, invite modal, user search, accept/decline (P12)
6. **Party settings**: settings page, transfer ownership, regenerate link, leave (P13)
7. **Game catalog**: catalog page, game detail, add game flow, collection toggle, BGG search (P14–P16)
8. **User profile**: profile page, stats tab, collection tab, profile edit (P17)
9. **Session logging**: session form (multi-step), session list, session detail, edit/delete (P8–P10)
10. **Stats & leaderboard**: leaderboard page, player-in-party stats, activity chart (P11, P18, T10.10)
11. **Polish**: empty states, loading skeletons, error handling review, responsive breakpoints, accessibility audit
