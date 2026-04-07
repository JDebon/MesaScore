# Game Search UX — Feature Specification

## Scope

This spec improves the game search and catalog addition experience. It supersedes the relevant parts of F3.3 in `docs/02-features/spec.md`.

---

## F6.1 — BGG search result ranking

**As a user searching for a game, I want base games to appear before expansions and editions.**

Acceptance criteria:
- Search results are ranked server-side before being returned to the client.
- Ranking tiers (applied in order):
  1. Exact name match (case-insensitive, trimmed).
  2. Name starts with the query and contains no colon (`:`).
  3. Name contains the query and contains no colon.
  4. All other matches (names with `:` typically indicate expansions or editions).
- Within each tier, results are sorted by year ascending (oldest first; nulls last).
- The client does not perform any additional sorting.

---

## F6.2 — Game addition flow without redundant form fields

**As a user adding a game from BGG, I should not have to enter data that BGG already provides.**

Acceptance criteria:
- After selecting a BGG result, the frontend fetches full game details from BGG and shows a read-only preview card containing: cover image, name, year, player count range, and BGG rating.
- There are no editable form fields for cover image URL, player counts, description, or rating when adding a BGG-backed game.
- The user confirms by clicking "Add to catalog" — no further input required.
- If the game already exists in the catalog (matched by BGG ID), the user is prompted to add it to their collection instead.
- The manual add path (no BGG) only asks for a name.
- Adding a game also adds it to the user's collection automatically.
