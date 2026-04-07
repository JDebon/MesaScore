<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import Avatar from '$components/ui/Avatar.svelte';
	import Spinner from '$components/ui/Spinner.svelte';
	import { statsApi } from '$api/stats';
	import { usersApi } from '$api/users';
	import { addToast } from '$stores/toast.svelte';
	import type { UserProfile, UserStats } from '$api/types';

	interface Props {
		data: { party: import('$api/types').Party };
	}

	let { data: layoutData }: Props = $props();
	let profile = $state<UserProfile | null>(null);
	let partyStats = $state<UserStats | null>(null);
	let globalStats = $state<UserStats | null>(null);
	let loading = $state(true);
	let activeTab = $state<'party' | 'global'>('party');

	const userId = $derived(page.params.userId!);

	onMount(async () => {
		try {
			const [p, ps, gs] = await Promise.all([
				usersApi.get(userId),
				statsApi.userInParty(layoutData.party.id, userId),
				statsApi.userGlobal(userId)
			]);
			profile = p;
			partyStats = ps;
			globalStats = gs;
		} catch (e) {
			console.error('[player stats] Failed to load:', e);
			addToast('Failed to load stats', 'error');
		} finally {
			loading = false;
		}
	});

	function formatPercent(n: number): string {
		return (n * 100).toFixed(0) + '%';
	}

	const stats = $derived(activeTab === 'party' ? partyStats : globalStats);
</script>

<svelte:head>
	<title>{profile?.display_name ?? 'Player'} - {layoutData.party.name} - MesaScore</title>
</svelte:head>

{#if loading}
	<div class="flex justify-center py-20"><Spinner size="lg" /></div>
{:else if profile && stats}
	<div class="mx-auto max-w-2xl">
		<!-- Header -->
		<div class="mb-6 flex items-center gap-4">
			<Avatar url={profile.avatar_url} name={profile.display_name} size="lg" />
			<div>
				<h1 class="text-xl font-bold text-text-primary">{profile.display_name}</h1>
				<p class="text-text-secondary">@{profile.username}</p>
			</div>
		</div>

		<!-- Tabs -->
		<div class="mb-6 flex border-b">
			<button
				class="px-4 py-2 text-sm font-medium transition-colors border-b-2
					{activeTab === 'party' ? 'border-primary-600 text-primary-600' : 'border-transparent text-text-secondary hover:text-text-primary'}"
				onclick={() => (activeTab = 'party')}
			>Party Stats</button>
			<button
				class="px-4 py-2 text-sm font-medium transition-colors border-b-2
					{activeTab === 'global' ? 'border-primary-600 text-primary-600' : 'border-transparent text-text-secondary hover:text-text-primary'}"
				onclick={() => (activeTab = 'global')}
			>Global Stats</button>
		</div>

		<!-- Stats overview -->
		<div class="mb-6 grid grid-cols-3 gap-3 sm:grid-cols-5">
			<div class="rounded-lg bg-surface p-3 text-center shadow-sm">
				<p class="text-xl font-bold text-text-primary">{stats.total_sessions}</p>
				<p class="text-xs text-text-secondary">Sessions</p>
			</div>
			<div class="rounded-lg bg-surface p-3 text-center shadow-sm">
				<p class="text-xl font-bold text-text-primary">{stats.total_wins}</p>
				<p class="text-xs text-text-secondary">Wins</p>
			</div>
			<div class="rounded-lg bg-surface p-3 text-center shadow-sm">
				<p class="text-xl font-bold text-text-primary">{formatPercent(stats.win_rate)}</p>
				<p class="text-xs text-text-secondary">Win Rate</p>
			</div>
			<div class="rounded-lg bg-surface p-3 text-center shadow-sm">
				<p class="text-xl font-bold text-text-primary">{stats.current_streak}</p>
				<p class="text-xs text-text-secondary">Streak</p>
			</div>
			<div class="rounded-lg bg-surface p-3 text-center shadow-sm">
				<p class="text-xl font-bold text-text-primary">{stats.best_streak}</p>
				<p class="text-xs text-text-secondary">Best</p>
			</div>
		</div>

		<!-- Highlights -->
		<div class="mb-6 grid gap-3 sm:grid-cols-2">
			{#if stats.most_played_game}
				<div class="rounded-lg bg-surface p-4 shadow-sm">
					<p class="text-xs font-medium text-text-secondary uppercase tracking-wide">Most Played</p>
					<p class="mt-1 font-medium text-text-primary">{stats.most_played_game.name}</p>
					<p class="text-sm text-text-secondary">{stats.most_played_game.session_count} sessions</p>
				</div>
			{/if}
			{#if stats.best_win_rate_game}
				<div class="rounded-lg bg-surface p-4 shadow-sm">
					<p class="text-xs font-medium text-text-secondary uppercase tracking-wide">Best Win Rate</p>
					<p class="mt-1 font-medium text-text-primary">{stats.best_win_rate_game.name}</p>
					<p class="text-sm text-text-secondary">{formatPercent(stats.best_win_rate_game.win_rate)} win rate</p>
				</div>
			{/if}
			{#if stats.nemesis}
				<div class="rounded-lg bg-surface p-4 shadow-sm">
					<p class="text-xs font-medium text-text-secondary uppercase tracking-wide">Nemesis</p>
					<p class="mt-1 font-medium text-text-primary">{stats.nemesis.display_name}</p>
					<p class="text-sm text-text-secondary">Lost {stats.nemesis.losses_against} times</p>
				</div>
			{/if}
			{#if stats.punching_bag}
				<div class="rounded-lg bg-surface p-4 shadow-sm">
					<p class="text-xs font-medium text-text-secondary uppercase tracking-wide">Punching Bag</p>
					<p class="mt-1 font-medium text-text-primary">{stats.punching_bag.display_name}</p>
					<p class="text-sm text-text-secondary">Beat them {stats.punching_bag.wins_against} times</p>
				</div>
			{/if}
		</div>

		<!-- Per-game -->
		{#if stats.per_game.length > 0}
			<section class="mb-6">
				<h2 class="mb-3 text-lg font-semibold text-text-primary">Per-Game Breakdown</h2>
				<div class="overflow-x-auto rounded-lg bg-surface shadow-sm">
					<table class="w-full text-sm">
						<thead><tr class="border-b text-left text-text-secondary">
							<th class="px-4 py-3 font-medium">Game</th>
							<th class="px-4 py-3 font-medium text-right">Sessions</th>
							<th class="px-4 py-3 font-medium text-right">Wins</th>
							<th class="px-4 py-3 font-medium text-right">Win Rate</th>
						</tr></thead>
						<tbody>
							{#each stats.per_game as pg}
								<tr class="border-b last:border-0">
									<td class="px-4 py-3 font-medium text-text-primary">{pg.game.name}</td>
									<td class="px-4 py-3 text-right text-text-secondary">{pg.sessions}</td>
									<td class="px-4 py-3 text-right text-text-secondary">{pg.wins}</td>
									<td class="px-4 py-3 text-right text-text-secondary">{formatPercent(pg.win_rate)}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</section>
		{/if}

		<!-- Head-to-head -->
		{#if stats.head_to_head.length > 0}
			<section>
				<h2 class="mb-3 text-lg font-semibold text-text-primary">Head-to-Head</h2>
				<div class="overflow-x-auto rounded-lg bg-surface shadow-sm">
					<table class="w-full text-sm">
						<thead><tr class="border-b text-left text-text-secondary">
							<th class="px-4 py-3 font-medium">Opponent</th>
							<th class="px-4 py-3 font-medium text-right">Together</th>
							<th class="px-4 py-3 font-medium text-right">Wins</th>
							<th class="px-4 py-3 font-medium text-right">Losses</th>
						</tr></thead>
						<tbody>
							{#each stats.head_to_head as h2h}
								<tr class="border-b last:border-0">
									<td class="px-4 py-3">
										<div class="flex items-center gap-2">
											<Avatar url={h2h.opponent.avatar_url} name={h2h.opponent.display_name} size="sm" />
											<span class="font-medium text-text-primary">{h2h.opponent.display_name}</span>
										</div>
									</td>
									<td class="px-4 py-3 text-right text-text-secondary">{h2h.sessions_together}</td>
									<td class="px-4 py-3 text-right text-text-secondary">{h2h.this_user_wins}</td>
									<td class="px-4 py-3 text-right text-text-secondary">{h2h.opponent_wins}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</section>
		{/if}
	</div>
{/if}
