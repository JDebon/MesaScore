<script lang="ts">
	import { page } from '$app/state';

	interface Props {
		partyId: string;
	}

	let { partyId }: Props = $props();

	const tabs = $derived([
		{ label: 'Dashboard', href: `/parties/${partyId}`, icon: 'home' },
		{ label: 'Sessions', href: `/parties/${partyId}/sessions`, icon: 'play' },
		{ label: 'Leaderboard', href: `/parties/${partyId}/leaderboard`, icon: 'trophy' },
		{ label: 'Members', href: `/parties/${partyId}/members`, icon: 'users' }
	]);

	function isActive(href: string): boolean {
		const path = page.url.pathname;
		if (href === `/parties/${partyId}`) return path === href;
		return path.startsWith(href);
	}
</script>

<!-- Mobile bottom nav only — desktop nav is handled by AppSidebar -->
<nav class="fixed bottom-0 left-0 right-0 z-30 border-t border-border bg-surface lg:hidden">
	<div class="flex">
		{#each tabs as tab}
			<a
				href={tab.href}
				class="flex flex-1 flex-col items-center gap-1 py-2 text-xs transition-colors
					{isActive(tab.href) ? 'text-primary-500' : 'text-text-secondary hover:text-text-primary'}"
			>
				{#if tab.icon === 'home'}
					<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/></svg>
				{:else if tab.icon === 'play'}
					<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/><path stroke-linecap="round" stroke-linejoin="round" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
				{:else if tab.icon === 'trophy'}
					<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M5 3h14a1 1 0 011 1v2a7 7 0 01-7 7 7 7 0 01-7-7V4a1 1 0 011-1zM8 21h8m-4-4v4m-5-8a2 2 0 01-2-2V8m14 3a2 2 0 002-2V8"/></svg>
				{:else}
					<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"/></svg>
				{/if}
				{tab.label}
			</a>
		{/each}
	</div>
</nav>
