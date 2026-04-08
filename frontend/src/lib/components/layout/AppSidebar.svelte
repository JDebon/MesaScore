<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import Avatar from '$components/ui/Avatar.svelte';
	import ThemeToggle from '$components/ui/ThemeToggle.svelte';
	import { getUser, clearAuth } from '$stores/auth.svelte';
	import { getPendingInviteCount } from '$stores/notifications.svelte';

	let drawerOpen = $state(false);
	const user = $derived(getUser());
	const inviteCount = $derived(getPendingInviteCount());

	function navClass(href: string, exact = false): string {
		const path = page.url.pathname;
		const active = exact ? path === href : path === href || path.startsWith(href + '/');
		return [
			'flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-all duration-150',
			active
				? 'bg-primary-500/10 text-primary-500 dark:bg-primary-400/10 dark:text-primary-400'
				: 'text-text-secondary hover:bg-surface-raised hover:text-text-primary'
		].join(' ');
	}

	function closeDrawer() {
		drawerOpen = false;
	}

	function logout() {
		clearAuth();
		goto('/login');
	}
</script>

<!-- ── Mobile top bar ─────────────────────────────────────────────── -->
<header
	class="fixed left-0 right-0 top-0 z-40 flex h-14 items-center justify-between border-b border-border bg-surface px-4 lg:hidden"
>
	<button
		onclick={() => (drawerOpen = !drawerOpen)}
		class="rounded-lg p-2 text-text-secondary transition-colors hover:bg-surface-raised hover:text-text-primary"
		aria-label="Open navigation"
	>
		{#if drawerOpen}
			<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
			</svg>
		{:else}
			<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
			</svg>
		{/if}
	</button>

	<a href="/" class="flex items-center gap-2 text-text-primary" onclick={closeDrawer}>
		<div class="flex h-7 w-7 items-center justify-center rounded-lg bg-primary-600">
			<svg class="h-4 w-4 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
				<path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" />
			</svg>
		</div>
		<span class="text-base font-bold" style="font-family: var(--font-display)">MesaScore</span>
	</a>

	<div class="flex items-center gap-2">
		{#if inviteCount > 0}
			<a
				href="/invites"
				class="relative rounded-lg p-2 text-text-secondary hover:text-text-primary"
				aria-label="{inviteCount} pending invites"
			>
				<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
				</svg>
				<span class="absolute right-1 top-1 flex h-4 w-4 items-center justify-center rounded-full bg-primary-600 text-[10px] font-bold text-white">
					{inviteCount > 9 ? '9+' : inviteCount}
				</span>
			</a>
		{/if}
		<ThemeToggle />
	</div>
</header>

<!-- ── Backdrop ───────────────────────────────────────────────────── -->
{#if drawerOpen}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-40 bg-black/50 backdrop-blur-sm lg:hidden"
		onclick={closeDrawer}
		onkeydown={() => {}}
	></div>
{/if}

<!-- ── Sidebar ─────────────────────────────────────────────────────── -->
<aside
	class="fixed inset-y-0 left-0 z-50 flex w-60 flex-col border-r border-border bg-surface transition-transform duration-300 ease-in-out lg:translate-x-0
		{drawerOpen ? 'translate-x-0' : '-translate-x-full'}"
>
	<!-- Logo -->
	<div class="flex h-16 shrink-0 items-center border-b border-border px-5">
		<a href="/" class="flex items-center gap-3" onclick={closeDrawer}>
			<div class="flex h-8 w-8 items-center justify-center rounded-xl bg-primary-600 shadow-sm shadow-primary-900/50">
				<svg class="h-4.5 w-4.5 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
					<path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" />
				</svg>
			</div>
			<span class="text-lg font-bold text-text-primary" style="font-family: var(--font-display)">
				MesaScore
			</span>
		</a>
	</div>

	<!-- Navigation -->
	<nav class="flex flex-1 flex-col gap-1 overflow-y-auto p-3">
		<a href="/" class={navClass('/', true)} onclick={closeDrawer}>
			<svg class="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
			</svg>
			<span>Home</span>
		</a>

		<a href="/parties" class={navClass('/parties')} onclick={closeDrawer}>
			<svg class="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
			</svg>
			<span>Parties</span>
		</a>

		<a href="/invites" class={navClass('/invites')} onclick={closeDrawer}>
			<svg class="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
			</svg>
			<span class="flex-1">Invites</span>
			{#if inviteCount > 0}
				<span class="flex h-5 min-w-5 items-center justify-center rounded-full bg-primary-600 px-1.5 text-[11px] font-bold text-white">
					{inviteCount > 9 ? '9+' : inviteCount}
				</span>
			{/if}
		</a>

		<a href="/games" class={navClass('/games')} onclick={closeDrawer}>
			<svg class="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M9 3H5a2 2 0 00-2 2v4m6-6h10a2 2 0 012 2v4M9 3v18m0 0h10a2 2 0 002-2V9M9 21H5a2 2 0 01-2-2V9m0 0h18" />
			</svg>
			<span>My Games</span>
		</a>
	</nav>

	<!-- Divider -->
	<div class="mx-3 shrink-0 border-t border-border"></div>

	<!-- Bottom: theme + user -->
	<div class="shrink-0 space-y-1 p-3">
		<div class="flex items-center justify-between rounded-xl px-3 py-2">
			<span class="text-xs font-medium text-text-secondary">Theme</span>
			<ThemeToggle />
		</div>

		<a
			href="/users/{user?.id}"
			class="flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm transition-colors hover:bg-surface-raised"
			onclick={closeDrawer}
		>
			<Avatar url={user?.avatar_url} name={user?.display_name ?? '?'} size="sm" />
			<div class="min-w-0 flex-1">
				<p class="truncate font-medium text-text-primary">{user?.display_name}</p>
				<p class="truncate text-xs text-text-secondary">@{user?.username}</p>
			</div>
		</a>

		<button
			onclick={logout}
			class="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium text-text-secondary transition-colors hover:bg-surface-raised hover:text-text-primary"
		>
			<svg class="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
			</svg>
			Log out
		</button>
	</div>
</aside>
