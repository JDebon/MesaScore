# Feature Specification

## Scope

This document defines all features for MesaScore v1. Features are organized by area.

**Roles referenced:**
- **Unauthenticated**: can only access login, register, and join-via-invite-link pages.
- **User**: any verified, logged-in user. Can create parties, manage their own collection, view data for parties they belong to.
- **Party Admin**: a user who is the admin of a specific party. Can log sessions, manage members, and send invites for that party.

There is no global admin role. Admin status is always relative to a specific party.

---

## F1 — Authentication

### F1.1 Registration
**As a new user, I want to create an account so I can use MesaScore.**

Acceptance criteria:
- Registration form accepts: username, display name, email, password.
- Username must be unique (checked in real time with debounce).
- Password must be at least 8 characters.
- On submit, the account is created with `email_verified = false` and a verification email is sent.
- User sees a confirmation screen instructing them to check their email.
- User cannot log in until email is verified.

### F1.2 Email verification
**As a new user, I want to verify my email so my account is activated.**

Acceptance criteria:
- Verification email contains a link: `<base_url>/verify-email?token=<token>`.
- Clicking the link sets `email_verified = true`, clears the token, and redirects to login with a success message.
- Tokens expire 24 hours after sending.
- If a user tries to log in with an unverified account, they see an error with a "Resend verification email" option.
- Resending generates a new token and invalidates the previous one.

### F1.3 Login
**As a verified user, I want to log in so I can access the app.**

Acceptance criteria:
- Login form accepts email and password.
- On success, a JWT is issued.
- On failure, a generic error is shown (no indication of which field was wrong).
- JWT expires after 7 days and is refreshed on each authenticated request if expiry is within 24 hours.
- Unverified accounts cannot log in (see F1.2).

### F1.4 Logout
**As any logged-in user, I want to log out.**

Acceptance criteria:
- Logout clears the JWT client-side.
- User is redirected to the login screen.
- All protected routes redirect to login with no valid JWT.

### F1.5 Profile management
**As any user, I want to update my own display name and avatar.**

Acceptance criteria:
- User can update display name and avatar URL.
- User cannot change username, email, or role via profile settings.
- Changes take effect immediately across the app.

---

## F2 — Parties

### F2.1 Create a party
**As any user, I want to create a party so I can track sessions with a friend group.**

Acceptance criteria:
- Any verified user can create a party with a name and optional description.
- The creator is automatically the party admin and the first member.
- An invite code is generated automatically on creation.
- After creation, the user lands on the new party's dashboard.

### F2.2 Join via invite link
**As any user, I want to join a party using an invite link.**

Acceptance criteria:
- Visiting `<base_url>/join/<invite_code>` shows a page with the party name and member count.
- If not logged in, the user is prompted to log in or register first; after auth, they are redirected back to the join page.
- Clicking "Join" adds the user to the party and redirects them to the party dashboard.
- If the user is already a member, they are redirected to the party dashboard with an info message.
- If the invite code is invalid or regenerated, a "Link not found" error is shown.

### F2.3 Invite by username (admin only)
**As a party admin, I want to invite a specific user by searching their username.**

Acceptance criteria:
- Admin searches for users by username or display name.
- Search results exclude users already in the party or with a pending invite.
- Selecting a user sends them a direct invite (creates a PartyInvite record).
- The invited user sees the pending invite on their dashboard.
- Admin sees the invite status (pending/accepted/declined) in the members section.

### F2.4 Accept or decline an invite
**As a user, I want to respond to party invites I receive.**

Acceptance criteria:
- User sees all pending invites on their global dashboard with party name and who invited them.
- Accepting adds them to the party and redirects to the party dashboard.
- Declining marks the invite as declined. No further action required.
- Accepted/declined invites are removed from the pending list.

### F2.5 View party members
**As any party member, I want to see who is in the party.**

Acceptance criteria:
- Members list shows: avatar, display name, username, join date.
- Admin is visually distinguished (e.g., crown icon).
- Party admin sees a "Remove member" button next to each non-admin member.
- Party admin sees pending and declined invites in a separate section.

### F2.6 Remove a member (admin only)
**As a party admin, I want to remove a member from the party.**

Acceptance criteria:
- Admin can remove any non-admin member.
- Confirmation prompt shown before removal.
- Removed member loses access to the party immediately.
- Historical session records are preserved (removed member still appears in past sessions).

### F2.7 Leave a party
**As a non-admin member, I want to leave a party.**

Acceptance criteria:
- Member can leave from the party settings page.
- Confirmation prompt shown.
- After leaving, the party is no longer accessible to them.
- Historical session records are preserved.

### F2.8 Transfer ownership (admin only)
**As a party admin, I want to transfer ownership to another member so I can leave or step down.**

Acceptance criteria:
- Admin selects any current member as the new admin.
- Confirmation prompt shown.
- On confirm, the selected member becomes the new admin.
- The previous admin becomes a regular member and can then leave if they wish.

### F2.9 Regenerate invite link (admin only)
**As a party admin, I want to regenerate the invite link to revoke old links.**

Acceptance criteria:
- Admin can regenerate the invite code from party settings.
- Old invite link immediately stops working.
- New link is shown and can be copied.

---

## F3 — Game Catalog & Collections

### F3.1 Browse my collection
**As any user, I want to browse my own game collection.**

Acceptance criteria:
- Shows only games in the user's own collection, with cover image, name, and date added.
- Filterable by name (client-side text search).
- Links to each game's detail page.

### F3.2 Game detail page
**As any user, I want to see a game's details.**

Acceptance criteria:
- Shows: cover image, name, description, player count range, BGG rating, BGG link (if applicable).
- Shows: all users who own this game (globally).
- Shows: total sessions played with this game across all parties (global count).
- Shows: "Add to my collection" / "Remove from my collection" button.

### F3.3 Add a game to the catalog
**As any verified user, I want to add a game so it's available for everyone.**

Acceptance criteria:
- User searches by name against the BGG API.
- Search results show name, year, and thumbnail.
- Selecting a result pre-fills all game fields from BGG data.
- If the game already exists in the catalog (by BGG ID), the user is prompted to add it to their collection instead.
- User can also add manually (just a name) if not on BGG.
- Adding a game also adds it to the user's collection automatically.

### F3.4 Manage own collection
**As any user, I want to add and remove games from my personal collection.**

Acceptance criteria:
- User can add any catalog game to their collection from the catalog or game detail page.
- User can remove a game from their collection from their profile or game detail page.
- Removing from collection does not affect session history.
- Collection is visible on the user's profile page.

### F3.5 Refresh BGG data
**As any user who owns the game, I want to refresh its BGG data.**

Acceptance criteria:
- Any user who has the game in their collection can trigger a BGG refresh.
- Refresh updates: description, cover image, rating, player counts.
- Name is NOT overwritten on refresh.
- `bgg_fetched_at` is updated.

---

## F4 — Session Logging

### F4.1 Log a session (party admin only)
**As a party admin, I want to log a game session for my party.**

Acceptance criteria:
- Admin selects: game (from any participating member's collection), session type, date played, party members who participated.
- Admin optionally sets: duration, notes, which member brought the game (must own it).
- Admin fills in results according to session type:
  - **Competitive**: assign a rank to each player. Ties allowed. Score optional.
  - **Team**: group players into named teams, rank each team. Score optional per team.
  - **Cooperative**: mark group win or group loss. Applied to all players.
  - **Score**: enter score per player, ranks auto-calculated but editable.
- Only current party members can be added as participants.
- Session cannot be saved with fewer than 2 participants.
- Session date cannot be in the future.

### F4.2 Edit a session (party admin only)
**As a party admin, I want to edit a logged session to correct mistakes.**

Acceptance criteria:
- Admin can edit all session fields.
- Changing session type resets all participant result fields.

### F4.3 Delete a session (party admin only)
**As a party admin, I want to delete a session logged by mistake.**

Acceptance criteria:
- Confirmation prompt shown.
- Deletion is permanent and cascades to participant records.

### F4.4 View session history (party)
**As any party member, I want to browse all sessions logged in the party.**

Acceptance criteria:
- List sorted by date descending.
- Each entry shows: game, date, session type, participants, winner(s).
- Filterable by game, player, session type, and date range.
- Each entry links to session detail.

### F4.5 Session detail page
**As any party member, I want to see the full details of a session.**

Acceptance criteria:
- Shows all session fields and full participant results.
- Party admin sees Edit and Delete buttons.

---

## F5 — Stats & Leaderboards

All stats exist in two scopes:
- **Party scope**: filtered to one party's sessions. Visible to party members.
- **Global scope**: across all sessions a user participated in, regardless of party. Visible on user profile.

### F5.1 Party leaderboard
**As any party member, I want to see who wins the most in our group.**

Acceptance criteria:
- Table of all party members sorted by wins within that party.
- Shows per player: wins, sessions played in this party, win rate.
- Co-op sessions count as a win for all players when the group wins.

### F5.2 Per-game party stats
**As any party member, I want to see who performs best at a specific game within our party.**

Acceptance criteria:
- Shown on the game detail page when viewed in a party context.
- Table of players who played this game in this party, sorted by win rate.

### F5.3 Player global stats
**As any user, I want to see a player's global stats across all parties.**

Acceptance criteria:
- Total sessions, total wins, overall win rate (all parties combined).
- Current win streak and best win streak (global, across all parties).
- Most played game (global).
- Best win rate game (global, min 3 sessions).
- Nemesis (global).
- Punching bag (global).
- Per-game breakdown (global).
- Head-to-head record against each other player (global).

### F5.4 Player party stats
**As any party member, I want to see how a player performs specifically within our party.**

Acceptance criteria:
- Shown as a tab or section on the player profile when viewed from within a party.
- Same metrics as F5.3 but filtered to sessions of this party only.
- Head-to-head only includes sessions in this party.

### F5.5 Activity over time
**As any party member, I want to see how often the party plays.**

Acceptance criteria:
- Bar chart of sessions per month for the last 12 months.
- Scoped to the current party.
- Shown on the party dashboard.

### F5.6 Win streak tracking
**As any user, I want to see my current and best win streaks.**

Acceptance criteria:
- A win streak is consecutive sessions (ordered by played_at, across all parties globally) where the user won.
- Co-op wins count.
- Sessions the user did not participate in do not break or extend the streak.
- Streak resets on any participated session where the user did not win.

---

## F6 — Dashboard

### F6.1 Global user dashboard
**As any user, I want a home screen showing my parties and overall activity.**

Acceptance criteria:
- Shows: list of parties the user belongs to (with name, member count, last session date).
- Shows: pending party invites with Accept/Decline actions.
- Shows: user's global stats at a glance (total sessions across all parties, total wins, current streak).
- Shows: "Create a party" button.
- Mobile-first layout.

### F6.2 Party dashboard
**As any party member, I want to see a summary of the party's activity.**

Acceptance criteria:
- Shows: last 5 sessions in this party (game, date, winner(s)).
- Shows: party stats at a glance (total sessions, unique games played, total members).
- Shows: current party leader (member with most wins in this party).
- Shows: sessions per month chart (last 12 months, party-scoped).
- Shows: most played game in this party.
- Mobile-first layout.

---

## Out of Scope for v1

- Achievements and badge system.
- Push/email notifications for new sessions.
- Password reset via email.
- Session media uploads.
- Public/discoverable party directories.
- Multiple party admins.
- Import from BoardGameGeek user collections.
- Real-time updates (page refresh is sufficient).
- Party deletion.
