<script lang="ts">
	import { onMount } from 'svelte';
	import Avatar from '$components/ui/Avatar.svelte';
	import EmptyState from '$components/ui/EmptyState.svelte';
	import LoadState from '$components/ui/LoadState.svelte';
	import Skeleton from '$components/ui/Skeleton.svelte';
	import { statsApi } from '$api/stats';
	import type { LeaderboardEntry } from '$api/types';

	interface Props {
		data: { party: import('$api/types').Party };
	}

	let { data: layoutData }: Props = $props();
	let entries = $state<LeaderboardEntry[]>([]);
	let loading = $state(true);
	let loadError = $state(false);
	let sort = $state('wins');

	onMount(() => loadLeaderboard());

	async function loadLeaderboard() {
		loading = true;
		loadError = false;
		try {
			entries = await statsApi.partyLeaderboard(layoutData.party.id, sort);
		} catch (e) {
			console.error('[leaderboard] Failed to load:', e);
			loadError = true;
		} finally {
			loading = false;
		}
	}

	function changeSort(newSort: string) {
		sort = newSort;
		loadLeaderboard();
	}

	function formatPercent(n: number): string {
		return (n * 100).toFixed(0) + '%';
	}

	// Light / dark medal row styles (T11.8)
	const medalRow = [
		'bg-amber-100 dark:bg-amber-900/40',
		'bg-slate-100 dark:bg-slate-700/40',
		'bg-orange-100 dark:bg-orange-900/40'
	];
	const medalText = [
		'text-amber-800 dark:text-amber-300',
		'text-slate-600 dark:text-slate-300',
		'text-orange-700 dark:text-orange-300'
	];
</script>

<svelte:head>
	<title>Leaderboard - {layoutData.party.name} - MesaScore</title>
</svelte:head>

<!-- Sort controls always visible -->
<div class="mb-4 flex gap-2">
	{#each [['wins', 'Wins'], ['sessions', 'Sessions'], ['win_rate', 'Win Rate']] as [key, label]}
		<button
			onclick={() => changeSort(key)}
			class="rounded-lg px-3 py-1.5 text-sm font-medium transition-colors
				{sort === key ? 'bg-primary-600 text-white' : 'bg-surface text-text-secondary hover:text-text-primary shadow-sm'}"
		>
			{label}
		</button>
	{/each}
</div>

<LoadState {loading} error={loadError} onretry={loadLeaderboard}>
	{#snippet skeleton()}
		<div class="overflow-x-auto rounded-lg bg-surface shadow-sm">
			<div class="divide-y">
				{#each [1, 2, 3, 4, 5, 6, 7, 8] as _}
					<div class="flex items-center gap-4 px-4 py-3">
						<Skeleton class="h-5 w-6 shrink-0" />
						<Skeleton class="h-8 w-8 rounded-full shrink-0" />
						<Skeleton class="h-4 w-32 flex-1" />
						<Skeleton class="h-4 w-8 shrink-0" />
						<Skeleton class="h-4 w-10 shrink-0" />
						<Skeleton class="h-4 w-12 shrink-0" />
					</div>
				{/each}
			</div>
		</div>
	{/snippet}

	{#if entries.length === 0}
		<EmptyState message="Play some games first to see who's winning!" />
	{:else}
		<div class="overflow-x-auto rounded-lg bg-surface shadow-sm">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b text-left text-text-secondary">
						<th class="px-4 py-3 w-12 font-medium">#</th>
						<th class="px-4 py-3 font-medium">Player</th>
						<th class="px-4 py-3 font-medium text-right">Wins</th>
						<th class="px-4 py-3 font-medium text-right">Sessions</th>
						<th class="px-4 py-3 font-medium text-right">Win Rate</th>
					</tr>
				</thead>
				<tbody>
					{#each entries as entry, i}
						<tr class="border-b last:border-0 {i < 3 ? medalRow[i] : ''}">
							<td class="px-4 py-3 font-bold {i < 3 ? medalText[i] : 'text-text-secondary'}">
								{i + 1}
							</td>
							<td class="px-4 py-3">
								<a href="/parties/{layoutData.party.id}/users/{entry.user.id}" class="flex items-center gap-2">
									<Avatar url={entry.user.avatar_url} name={entry.user.display_name} size="sm" />
									<span class="font-medium text-text-primary">{entry.user.display_name}</span>
								</a>
							</td>
							<td class="px-4 py-3 text-right font-medium text-text-primary">{entry.wins}</td>
							<td class="px-4 py-3 text-right text-text-secondary">{entry.sessions}</td>
							<td class="px-4 py-3 text-right text-text-secondary">{formatPercent(entry.win_rate)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</LoadState>
