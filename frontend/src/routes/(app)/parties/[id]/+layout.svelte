<script lang="ts">
	import BottomNav from '$components/layout/BottomNav.svelte';
	import { getUser } from '$stores/auth.svelte';
	import type { Snippet } from 'svelte';

	interface Props {
		data: { party: import('$api/types').Party };
		children: Snippet;
	}

	let { data, children }: Props = $props();
	const user = $derived(getUser());
	const isAdmin = $derived(user?.id === data.party.admin.id);
</script>

<div class="lg:pl-56">
	<div class="mb-4 flex items-center justify-between">
		<div class="flex items-center gap-3">
			<a href="/" class="text-text-secondary hover:text-text-primary" aria-label="Back to dashboard">
				<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7"/>
				</svg>
			</a>
			<h1 class="text-xl font-bold text-text-primary">{data.party.name}</h1>
		</div>
		{#if isAdmin}
			<a
				href="/parties/{data.party.id}/settings"
				class="rounded-lg p-2 text-text-secondary hover:bg-surface-raised hover:text-text-primary"
				aria-label="Party settings"
			>
				<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
					<path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
				</svg>
			</a>
		{/if}
	</div>
	<div class="pb-20 lg:pb-0">
		{@render children()}
	</div>
</div>
<BottomNav partyId={data.party.id} />
