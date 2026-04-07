<script lang="ts">
	import Button from '$components/ui/Button.svelte';
	import Input from '$components/ui/Input.svelte';
	import Avatar from '$components/ui/Avatar.svelte';
	import EmptyState from '$components/ui/EmptyState.svelte';
	import { usersApi } from '$api/users';
	import { addToast } from '$stores/toast.svelte';
	import { getUser, updateUser } from '$stores/auth.svelte';
	import { ApiError } from '$api/client';
	import { validateUrl } from '$lib/validate';
	import type { UserProfile, CollectionItem } from '$api/types';

	let { data } = $props();

	let profile = $state<UserProfile>(data.profile);
	let collection = $state<CollectionItem[]>(data.collection);
	const stats = data.stats;

	let activeTab = $state<'stats' | 'collection'>('stats');

	const currentUser = $derived(getUser());
	const isOwnProfile = $derived(currentUser?.id === profile.id);

	// Edit state
	let editing = $state(false);
	let editDisplayName = $state('');
	let editAvatarUrl = $state('');
	let editAvatarUrlError = $state('');
	let saving = $state(false);

	function startEdit() {
		editDisplayName = profile.display_name;
		editAvatarUrl = profile.avatar_url || '';
		editAvatarUrlError = '';
		editing = true;
	}

	async function saveEdit(event: SubmitEvent) {
		event.preventDefault();
		editAvatarUrlError = validateUrl(editAvatarUrl) ?? '';
		if (editAvatarUrlError) return;
		saving = true;
		try {
			const updated = await usersApi.updateMe({
				display_name: editDisplayName.trim(),
				avatar_url: editAvatarUrl.trim() || null
			});
			profile = updated;
			updateUser({ display_name: updated.display_name, avatar_url: updated.avatar_url });
			editing = false;
			addToast('Profile updated', 'success');
		} catch (e) {
			addToast(e instanceof ApiError ? e.message : 'Failed to update', 'error');
		} finally {
			saving = false;
		}
	}

	async function removeFromCollection(gameId: string) {
		try {
			await usersApi.removeFromCollection(gameId);
			collection = collection.filter((c) => c.game_id !== gameId);
			addToast('Removed from collection', 'info');
		} catch (e) {
			console.error('[profile] Failed to remove from collection:', e);
			addToast('Failed to remove', 'error');
		}
	}

	function formatPercent(n: number): string {
		return (n * 100).toFixed(0) + '%';
	}
</script>

<svelte:head>
	<title>{profile.display_name} - MesaScore</title>
</svelte:head>

<div class="mx-auto max-w-2xl">
	<!-- Profile header -->
	<div class="mb-6 flex items-start gap-4">
		<Avatar url={profile.avatar_url} name={profile.display_name} size="xl" />
		<div class="flex-1">
			{#if editing}
				<form onsubmit={saveEdit} class="space-y-3">
					<Input name="display_name" label="Display name" bind:value={editDisplayName} required />
					<Input name="avatar_url" label="Avatar URL" bind:value={editAvatarUrl} error={editAvatarUrlError} />
					<div class="flex gap-2">
						<Button type="submit" size="sm" loading={saving}>Save</Button>
						<Button size="sm" variant="secondary" onclick={() => (editing = false)}>Cancel</Button>
					</div>
				</form>
			{:else}
				<h1 class="text-2xl font-bold text-text-primary">{profile.display_name}</h1>
				<p class="text-text-secondary">@{profile.username}</p>
				<p class="text-sm text-text-secondary">Member since {new Date(profile.created_at).toLocaleDateString()}</p>
				{#if isOwnProfile}
					<Button size="sm" variant="ghost" class="mt-2" onclick={startEdit}>Edit profile</Button>
				{/if}
			{/if}
		</div>
	</div>

	<!-- Tabs -->
	<div class="mb-6 flex border-b">
		<button
			class="px-4 py-2 text-sm font-medium transition-colors border-b-2
				{activeTab === 'stats' ? 'border-primary-600 text-primary-600' : 'border-transparent text-text-secondary hover:text-text-primary'}"
			onclick={() => (activeTab = 'stats')}
		>Stats</button>
		<button
			class="px-4 py-2 text-sm font-medium transition-colors border-b-2
				{activeTab === 'collection' ? 'border-primary-600 text-primary-600' : 'border-transparent text-text-secondary hover:text-text-primary'}"
			onclick={() => (activeTab = 'collection')}
		>Collection ({collection.length})</button>
	</div>

	{#if activeTab === 'stats'}
		<!-- Overview -->
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
					<a href="/games/{stats.most_played_game.id}" class="mt-1 block font-medium text-text-primary hover:text-primary-600">{stats.most_played_game.name}</a>
					<p class="text-sm text-text-secondary">{stats.most_played_game.session_count} sessions</p>
				</div>
			{/if}
			{#if stats.best_win_rate_game}
				<div class="rounded-lg bg-surface p-4 shadow-sm">
					<p class="text-xs font-medium text-text-secondary uppercase tracking-wide">Best Win Rate</p>
					<a href="/games/{stats.best_win_rate_game.id}" class="mt-1 block font-medium text-text-primary hover:text-primary-600">{stats.best_win_rate_game.name}</a>
					<p class="text-sm text-text-secondary">{formatPercent(stats.best_win_rate_game.win_rate)} win rate</p>
				</div>
			{/if}
			{#if stats.nemesis}
				<div class="rounded-lg bg-surface p-4 shadow-sm">
					<p class="text-xs font-medium text-text-secondary uppercase tracking-wide">Nemesis</p>
					<a href="/users/{stats.nemesis.id}" class="mt-1 block font-medium text-text-primary hover:text-primary-600">{stats.nemesis.display_name}</a>
					<p class="text-sm text-text-secondary">Lost {stats.nemesis.losses_against} times against them</p>
				</div>
			{/if}
			{#if stats.punching_bag}
				<div class="rounded-lg bg-surface p-4 shadow-sm">
					<p class="text-xs font-medium text-text-secondary uppercase tracking-wide">Punching Bag</p>
					<a href="/users/{stats.punching_bag.id}" class="mt-1 block font-medium text-text-primary hover:text-primary-600">{stats.punching_bag.display_name}</a>
					<p class="text-sm text-text-secondary">Beat them {stats.punching_bag.wins_against} times</p>
				</div>
			{/if}
		</div>

		<!-- Per-game breakdown -->
		{#if stats.per_game.length > 0}
			<section class="mb-6">
				<h2 class="mb-3 text-lg font-semibold text-text-primary">Per-Game Breakdown</h2>
				<div class="overflow-x-auto rounded-lg bg-surface shadow-sm">
					<table class="w-full text-sm">
						<thead>
							<tr class="border-b text-left text-text-secondary">
								<th class="px-4 py-3 font-medium">Game</th>
								<th class="px-4 py-3 font-medium text-right">Sessions</th>
								<th class="px-4 py-3 font-medium text-right">Wins</th>
								<th class="px-4 py-3 font-medium text-right">Win Rate</th>
							</tr>
						</thead>
						<tbody>
							{#each stats.per_game as pg}
								<tr class="border-b last:border-0">
									<td class="px-4 py-3">
										<a href="/games/{pg.game.id}" class="font-medium text-text-primary hover:text-primary-600">{pg.game.name}</a>
									</td>
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
						<thead>
							<tr class="border-b text-left text-text-secondary">
								<th class="px-4 py-3 font-medium">Opponent</th>
								<th class="px-4 py-3 font-medium text-right">Together</th>
								<th class="px-4 py-3 font-medium text-right">My Wins</th>
								<th class="px-4 py-3 font-medium text-right">Their Wins</th>
							</tr>
						</thead>
						<tbody>
							{#each stats.head_to_head as h2h}
								<tr class="border-b last:border-0">
									<td class="px-4 py-3">
										<a href="/users/{h2h.opponent.id}" class="flex items-center gap-2">
											<Avatar url={h2h.opponent.avatar_url} name={h2h.opponent.display_name} size="sm" />
											<span class="font-medium text-text-primary">{h2h.opponent.display_name}</span>
										</a>
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
	{:else}
		<!-- Collection tab -->
		{#if collection.length === 0}
			<EmptyState message={isOwnProfile ? 'Your collection is empty. Browse the catalog to add games.' : 'No games in collection.'} />
		{:else}
			<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
				{#each collection as item}
					<div class="flex items-center gap-3 rounded-lg bg-surface p-3 shadow-sm">
						<a href="/games/{item.game_id}" class="flex items-center gap-3 min-w-0 flex-1">
							{#if item.cover_image_url}
								<img src={item.cover_image_url} alt="" class="h-14 w-10 rounded object-cover" />
							{:else}
								<div class="flex h-14 w-10 items-center justify-center rounded bg-surface-raised text-text-secondary">
									<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"/></svg>
								</div>
							{/if}
							<span class="truncate font-medium text-text-primary">{item.name}</span>
						</a>
						{#if isOwnProfile}
							<button
								class="text-text-secondary hover:text-danger-500"
								onclick={() => removeFromCollection(item.game_id)}
								aria-label="Remove from collection"
							>
								<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/></svg>
							</button>
						{/if}
					</div>
				{/each}
			</div>
		{/if}
	{/if}
</div>
