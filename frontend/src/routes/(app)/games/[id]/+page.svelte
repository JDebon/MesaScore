<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import Button from '$components/ui/Button.svelte';
	import Avatar from '$components/ui/Avatar.svelte';
	import Spinner from '$components/ui/Spinner.svelte';
	import { gamesApi } from '$api/games';
	import { usersApi } from '$api/users';
	import { statsApi } from '$api/stats';
	import { addToast } from '$stores/toast.svelte';
	import { getUser } from '$stores/auth.svelte';
	import { ApiError } from '$api/client';
	import type { GameDetail, GamePartyStats } from '$api/types';

	let game = $state<GameDetail | null>(null);
	let partyStats = $state<GamePartyStats | null>(null);
	let loading = $state(true);
	let toggling = $state(false);
	let refreshing = $state(false);

	const user = $derived(getUser());
	const ownsGame = $derived(game?.owners.some((o) => o.id === user?.id) ?? false);
	const partyId = $derived(page.url.searchParams.get('party_id'));

	onMount(async () => {
		try {
			const promises: Promise<unknown>[] = [gamesApi.get(page.params.id!)];
			if (partyId) promises.push(statsApi.partyGame(partyId, page.params.id!));
			const [g, ps] = await Promise.allSettled(promises);
			if (g.status === 'fulfilled') game = g.value as GameDetail;
			else addToast('Failed to load game', 'error');
			if (ps && ps.status === 'fulfilled') partyStats = ps.value as GamePartyStats;
		} finally {
			loading = false;
		}
	});

	async function toggleCollection() {
		if (!game) return;
		toggling = true;
		try {
			if (game.in_my_collection) {
				await usersApi.removeFromCollection(game.id);
				game = { ...game, in_my_collection: false, owners: game.owners.filter((o) => o.id !== user?.id) };
				addToast('Removed from collection', 'info');
			} else {
				await usersApi.addToCollection(game.id);
				game = {
					...game,
					in_my_collection: true,
					owners: [...game.owners, { id: user!.id, display_name: user!.display_name, avatar_url: user!.avatar_url }]
				};
				addToast('Added to collection', 'success');
			}
		} catch (e) {
			addToast(e instanceof ApiError ? e.message : 'Failed to update collection', 'error');
		} finally {
			toggling = false;
		}
	}

	async function refreshBgg() {
		if (!game) return;
		refreshing = true;
		try {
			game = await gamesApi.bggRefresh(game.id);
			addToast('BGG data refreshed', 'success');
		} catch (e) {
			addToast(e instanceof ApiError ? e.message : 'Failed to refresh', 'error');
		} finally {
			refreshing = false;
		}
	}
</script>

<svelte:head>
	<title>{game?.name ?? 'Game'} - MesaScore</title>
</svelte:head>

{#if loading}
	<div class="flex justify-center py-20"><Spinner size="lg" /></div>
{:else if game}
	<div class="mx-auto max-w-2xl">
		<!-- Header -->
		<div class="mb-6 flex flex-col gap-4 sm:flex-row">
			{#if game.cover_image_url}
				<img src={game.cover_image_url} alt={game.name} class="h-48 w-36 rounded-xl object-cover shadow-md" />
			{:else}
				<div class="flex h-48 w-36 items-center justify-center rounded-xl bg-surface-raised text-text-secondary">
					<svg class="h-16 w-16" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"/></svg>
				</div>
			{/if}
			<div class="flex-1">
				<h1 class="text-2xl font-bold text-text-primary">{game.name}</h1>
				{#if game.min_players && game.max_players}
					<p class="mt-1 text-text-secondary">{game.min_players}-{game.max_players} players</p>
				{/if}
				{#if game.bgg_rating}
					<p class="mt-1 text-text-secondary">&#9733; {game.bgg_rating.toFixed(1)} BGG rating</p>
				{/if}
				<p class="mt-1 text-sm text-text-secondary">{game.session_count} session{game.session_count === 1 ? '' : 's'} played</p>
				{#if game.bgg_id}
					<a href="https://boardgamegeek.com/boardgame/{game.bgg_id}" target="_blank" rel="noopener noreferrer" class="mt-1 inline-block text-sm text-primary-600 hover:text-primary-700">View on BGG &rarr;</a>
				{/if}

				<div class="mt-4 flex flex-wrap gap-2">
					<Button variant={game.in_my_collection ? 'secondary' : 'primary'} loading={toggling} onclick={toggleCollection}>
						{game.in_my_collection ? 'Remove from collection' : 'Add to collection'}
					</Button>
					{#if ownsGame && game.bgg_id}
						<Button variant="ghost" loading={refreshing} onclick={refreshBgg}>Refresh BGG data</Button>
					{/if}
				</div>
			</div>
		</div>

		<!-- Description -->
		{#if game.description}
			<section class="mb-6">
				<h2 class="mb-2 text-lg font-semibold text-text-primary">Description</h2>
				<p class="text-text-secondary whitespace-pre-line">{game.description}</p>
			</section>
		{/if}

		<!-- Per-party leaderboard (only when party context present) -->
		{#if partyStats}
			<section class="mb-6">
				<h2 class="mb-3 text-lg font-semibold text-text-primary">
					Party Stats
					<span class="ml-1 text-sm font-normal text-text-secondary">({partyStats.total_sessions} session{partyStats.total_sessions === 1 ? '' : 's'})</span>
				</h2>
				{#if partyStats.leaderboard.length > 0}
					<div class="overflow-x-auto rounded-lg bg-surface shadow-sm">
						<table class="w-full text-sm">
							<thead>
								<tr class="border-b text-left text-text-secondary">
									<th class="px-4 py-3 font-medium">Player</th>
									<th class="px-4 py-3 font-medium text-right">Sessions</th>
									<th class="px-4 py-3 font-medium text-right">Wins</th>
									<th class="px-4 py-3 font-medium text-right">Win Rate</th>
								</tr>
							</thead>
							<tbody>
								{#each partyStats.leaderboard as entry}
									<tr class="border-b last:border-0">
										<td class="px-4 py-3">
											<a href="/parties/{partyId}/users/{entry.user.id}" class="flex items-center gap-2">
												<Avatar url={entry.user.avatar_url} name={entry.user.display_name} size="sm" />
												<span class="font-medium text-text-primary">{entry.user.display_name}</span>
											</a>
										</td>
										<td class="px-4 py-3 text-right text-text-secondary">{entry.sessions}</td>
										<td class="px-4 py-3 text-right text-text-secondary">{entry.wins}</td>
										<td class="px-4 py-3 text-right text-text-secondary">{(entry.win_rate * 100).toFixed(0)}%</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</section>
		{/if}
	</div>
{/if}
