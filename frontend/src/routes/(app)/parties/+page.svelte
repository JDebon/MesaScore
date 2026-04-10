<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import LoadState from '$components/ui/LoadState.svelte';
	import Skeleton from '$components/ui/Skeleton.svelte';
	import { usersApi } from '$api/users';
	import type { DashboardResponse } from '$api/types';

	let data = $state<DashboardResponse | null>(null);
	let loading = $state(true);
	let loadError = $state(false);

	async function loadData() {
		loading = true;
		loadError = false;
		try {
			data = await usersApi.dashboard();
		} catch (e) {
			console.error('[parties] Failed to load:', e);
			loadError = true;
		} finally {
			loading = false;
		}
	}

	onMount(loadData);

	function timeAgo(dateStr: string | null): string {
		if (!dateStr) return 'Never played';
		const diff = Date.now() - new Date(dateStr).getTime();
		const d = Math.floor(diff / 86400000);
		if (d === 0) return 'Today';
		if (d === 1) return 'Yesterday';
		if (d < 7) return `${d} days ago`;
		if (d < 30) return `${Math.floor(d / 7)}w ago`;
		return `${Math.floor(d / 30)}mo ago`;
	}
</script>

<svelte:head>
	<title>Parties — MesaScore</title>
</svelte:head>

<LoadState {loading} error={loadError} onretry={loadData}>
	{#snippet skeleton()}
		<div class="space-y-4">
			<div class="flex items-center justify-between">
				<Skeleton class="h-7 w-32" />
				<Skeleton class="h-9 w-32 rounded-xl" />
			</div>
			{#each [1, 2, 3] as _}
				<div class="rounded-2xl bg-surface border border-border p-4 space-y-2">
					<Skeleton class="h-5 w-40" />
					<Skeleton class="h-3 w-24" />
				</div>
			{/each}
		</div>
	{/snippet}

	<div class="space-y-5">
		<div class="flex items-center justify-between">
			<h1 class="text-xl font-bold text-text-primary" style="font-family: var(--font-display)">
				Parties
			</h1>
			<a
				href="/parties/new"
				class="flex items-center gap-2 rounded-xl bg-primary-500 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-600 transition-colors"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
				</svg>
				New Party
			</a>
		</div>

		{#if data!.parties.length === 0}
			<section class="rounded-2xl border border-dashed border-border px-6 py-12 text-center">
				<div class="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-surface-raised">
					<svg class="h-7 w-7 text-text-secondary" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
					</svg>
				</div>
				<p class="font-semibold text-text-primary" style="font-family: var(--font-display)">No parties yet</p>
				<p class="mt-1 text-sm text-text-secondary">Create a party or ask a friend for an invite link.</p>
				<a
					href="/parties/new"
					class="mt-4 inline-flex items-center gap-2 rounded-xl bg-primary-500 px-5 py-2.5 text-sm font-semibold text-white hover:bg-primary-600"
				>
					Create a party
				</a>
			</section>
		{:else}
			<div class="grid gap-3 sm:grid-cols-2">
				{#each data!.parties as party}
					<a
						href="/parties/{party.id}"
						class="group flex items-center justify-between rounded-2xl border border-border bg-surface px-4 py-4 transition-all duration-150 hover:border-primary-500/30 hover:shadow-sm"
					>
						<div class="min-w-0">
							<p class="truncate font-semibold text-text-primary">{party.name}</p>
							<p class="mt-0.5 text-xs text-text-secondary">
								{party.member_count} member{party.member_count === 1 ? '' : 's'}
								<span class="mx-1.5 opacity-40">·</span>
								{timeAgo(party.last_session_at)}
							</p>
						</div>
						<svg
							class="ml-3 h-4 w-4 shrink-0 text-text-secondary transition-transform duration-150 group-hover:translate-x-0.5 group-hover:text-primary-500"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							stroke-width="2"
						>
							<path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
						</svg>
					</a>
				{/each}
			</div>
		{/if}
	</div>
</LoadState>
