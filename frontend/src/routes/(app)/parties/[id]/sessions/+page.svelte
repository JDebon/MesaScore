<script lang="ts">
	import { onMount } from 'svelte';
	import Button from '$components/ui/Button.svelte';
	import Badge from '$components/ui/Badge.svelte';
	import EmptyState from '$components/ui/EmptyState.svelte';
	import LoadState from '$components/ui/LoadState.svelte';
	import Skeleton from '$components/ui/Skeleton.svelte';
	import { sessionsApi } from '$api/sessions';
	import { addToast } from '$stores/toast.svelte';
	import { getUser } from '$stores/auth.svelte';
	import type { SessionSummary } from '$api/types';

	interface Props {
		data: { party: import('$api/types').Party };
	}

	let { data: layoutData }: Props = $props();
	let sessions = $state<SessionSummary[]>([]);
	let loading = $state(true);
	let loadError = $state(false);
	let loadingMore = $state(false);
	let total = $state(0);
	let currentPage = $state(1);
	const perPage = 20;

	// Filters
	let filterType = $state('');
	let filterFrom = $state('');
	let filterTo = $state('');

	const user = $derived(getUser());
	const isAdmin = $derived(user?.id === layoutData.party.admin.id);
	const hasMore = $derived(sessions.length < total);

	onMount(() => loadSessions());

	async function loadSessions(append = false) {
		if (append) loadingMore = true;
		else {
			loading = true;
			loadError = false;
		}

		try {
			const res = await sessionsApi.list(layoutData.party.id, {
				type: filterType || undefined,
				from: filterFrom || undefined,
				to: filterTo || undefined,
				page: currentPage,
				per_page: perPage
			});
			if (append) {
				sessions = [...sessions, ...res.data];
			} else {
				sessions = res.data;
			}
			total = res.total;
		} catch (e) {
			console.error('[sessions] Failed to load:', e);
			if (!append) loadError = true;
			else addToast('Failed to load more sessions', 'error');
		} finally {
			loading = false;
			loadingMore = false;
		}
	}

	function applyFilters() {
		currentPage = 1;
		loadSessions();
	}

	function loadMore() {
		currentPage++;
		loadSessions(true);
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString();
	}

	const typeBadge: Record<string, 'default' | 'info' | 'success' | 'warning'> = {
		competitive: 'info',
		team: 'warning',
		cooperative: 'success',
		score: 'default'
	};
</script>

<svelte:head>
	<title>Sessions - {layoutData.party.name} - MesaScore</title>
</svelte:head>

<!-- Filters always visible -->
<div class="mb-4 flex flex-wrap gap-3">
	<select
		bind:value={filterType}
		onchange={applyFilters}
		class="rounded-lg border border-border px-3 py-2 text-sm"
	>
		<option value="">All types</option>
		<option value="competitive">Competitive</option>
		<option value="team">Team</option>
		<option value="cooperative">Cooperative</option>
		<option value="score">Score</option>
	</select>
	<input type="date" bind:value={filterFrom} onchange={applyFilters} class="rounded-lg border border-border px-3 py-2 text-sm" />
	<input type="date" bind:value={filterTo} onchange={applyFilters} class="rounded-lg border border-border px-3 py-2 text-sm" />
</div>

<LoadState {loading} error={loadError} onretry={() => loadSessions()}>
	{#snippet skeleton()}
		<div class="space-y-2">
			{#each [1, 2, 3, 4, 5] as _}
				<div class="flex items-center gap-3 rounded-lg bg-surface p-3 shadow-sm">
					<Skeleton class="h-12 w-10 rounded shrink-0" />
					<div class="flex-1 space-y-2">
						<Skeleton class="h-4 w-40" />
						<Skeleton class="h-3 w-32" />
					</div>
					<Skeleton class="h-5 w-20 rounded-full shrink-0" />
				</div>
			{/each}
		</div>
	{/snippet}

	{#if sessions.length === 0}
		<EmptyState message={isAdmin ? 'No sessions yet. Log your first game!' : 'No sessions logged yet.'} />
	{:else}
		<div class="space-y-2">
			{#each sessions as session}
				<a
					href="/parties/{layoutData.party.id}/sessions/{session.id}"
					class="flex items-center gap-3 rounded-lg bg-surface p-3 shadow-sm transition-shadow hover:shadow-md"
				>
					{#if session.game.cover_image_url}
						<img src={session.game.cover_image_url} alt="" class="h-12 w-10 rounded object-cover" />
					{:else}
						<div class="flex h-12 w-10 items-center justify-center rounded bg-surface-raised text-text-secondary">
							<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"/></svg>
						</div>
					{/if}
					<div class="min-w-0 flex-1">
						<p class="truncate font-medium text-text-primary">{session.game.name}</p>
						<p class="text-sm text-text-secondary">
							{formatDate(session.played_at)} &mdash;
							{#if session.winners.length > 0}
								{session.winners.map((w) => w.display_name).join(', ')} won
							{:else}
								{session.participant_count} players
							{/if}
						</p>
					</div>
					<Badge variant={typeBadge[session.session_type] ?? 'default'}>{session.session_type}</Badge>
				</a>
			{/each}
		</div>

		{#if hasMore}
			<div class="mt-4 text-center">
				<Button variant="secondary" loading={loadingMore} onclick={loadMore}>
					Load more
				</Button>
			</div>
		{/if}
	{/if}
</LoadState>
