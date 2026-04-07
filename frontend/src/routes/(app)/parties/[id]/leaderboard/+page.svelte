<script lang="ts">
	import { onMount } from 'svelte';
	import Avatar from '$components/ui/Avatar.svelte';
	import Spinner from '$components/ui/Spinner.svelte';
	import EmptyState from '$components/ui/EmptyState.svelte';
	import { statsApi } from '$api/stats';
	import { addToast } from '$stores/toast.svelte';
	import type { LeaderboardEntry } from '$api/types';

	interface Props {
		data: { party: import('$api/types').Party };
	}

	let { data: layoutData }: Props = $props();
	let entries = $state<LeaderboardEntry[]>([]);
	let loading = $state(true);
	let sort = $state('wins');

	onMount(() => loadLeaderboard());

	async function loadLeaderboard() {
		loading = true;
		try {
			entries = await statsApi.partyLeaderboard(layoutData.party.id, sort);
		} catch (e) {
			console.error('[leaderboard] Failed to load:', e);
			addToast('Failed to load leaderboard', 'error');
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

{#if loading}
	<div class="flex justify-center py-20"><Spinner size="lg" /></div>
{:else if entries.length === 0}
	<EmptyState message="Play some games first to see who's winning!" />
{:else}
	<div class="overflow-x-auto rounded-lg bg-surface shadow-sm">
		<table class="w-full text-sm">
			<thead>
				<tr class="border-b text-left text-text-secondary">
					<th class="px-4 py-3 w-12 font-medium">#</th>
					<th class="px-4 py-3 font-medium">Player</th>
					<th class="px-4 py-3 font-medium text-right">
						<button class="hover:text-text-primary {sort === 'wins' ? 'text-primary-600 font-bold' : ''}" onclick={() => changeSort('wins')}>Wins</button>
					</th>
					<th class="px-4 py-3 font-medium text-right">
						<button class="hover:text-text-primary {sort === 'sessions' ? 'text-primary-600 font-bold' : ''}" onclick={() => changeSort('sessions')}>Sessions</button>
					</th>
					<th class="px-4 py-3 font-medium text-right">
						<button class="hover:text-text-primary {sort === 'win_rate' ? 'text-primary-600 font-bold' : ''}" onclick={() => changeSort('win_rate')}>Win Rate</button>
					</th>
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
