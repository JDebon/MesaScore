<script lang="ts">
	import { onMount } from 'svelte';
	import Input from '$components/ui/Input.svelte';
	import EmptyState from '$components/ui/EmptyState.svelte';
	import LoadState from '$components/ui/LoadState.svelte';
	import Skeleton from '$components/ui/Skeleton.svelte';
	import { usersApi } from '$api/users';
	import { getUser } from '$stores/auth.svelte';
	import { addToast } from '$stores/toast.svelte';
	import type { CollectionItem } from '$api/types';

	let allGames = $state<CollectionItem[]>([]);
	let loading = $state(true);
	let loadError = $state(false);
	let query = $state('');

	const filteredGames = $derived(
		query
			? allGames.filter((g) => g.name.toLowerCase().includes(query.toLowerCase()))
			: allGames
	);

	onMount(() => loadCollection());

	async function loadCollection() {
		loading = true;
		loadError = false;
		const user = getUser();
		if (!user) return;
		try {
			allGames = await usersApi.collection(user.id);
		} catch (e) {
			console.error('[games] Failed to load collection:', e);
			loadError = true;
			addToast('Failed to load your collection', 'error');
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>My Games - MesaScore</title>
</svelte:head>

<div class="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
	<h1 class="text-2xl font-bold text-text-primary">My Games</h1>
	<a
		href="/games/new"
		class="inline-flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white hover:bg-primary-700"
	>
		<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"/></svg>
		Add game
	</a>
</div>

<div class="mb-4">
	<Input name="search" placeholder="Search your games..." bind:value={query} />
</div>

<LoadState {loading} error={loadError} onretry={loadCollection}>
	{#snippet skeleton()}
		<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
			{#each [1, 2, 3, 4, 5, 6] as _}
				<div class="flex gap-3 rounded-lg bg-surface p-3 shadow-sm">
					<Skeleton class="h-20 w-16 rounded shrink-0" />
					<div class="flex-1 space-y-2 pt-1">
						<Skeleton class="h-4 w-4/5" />
						<Skeleton class="h-3 w-24" />
					</div>
				</div>
			{/each}
		</div>
	{/snippet}

	{#if filteredGames.length === 0}
		<EmptyState message={query ? 'No games match your search.' : 'Your collection is empty. Add a game!'} />
	{:else}
		<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
			{#each filteredGames as game}
				<a
					href="/games/{game.game_id}"
					class="flex gap-3 rounded-lg bg-surface p-3 shadow-sm transition-shadow hover:shadow-md"
				>
					{#if game.cover_image_url}
						<img src={game.cover_image_url} alt="" class="h-20 w-16 rounded object-cover" />
					{:else}
						<div class="flex h-20 w-16 items-center justify-center rounded bg-surface-raised text-text-secondary">
							<svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"/></svg>
						</div>
					{/if}
					<div class="min-w-0 flex-1">
						<p class="truncate font-medium text-text-primary">{game.name}</p>
						<p class="mt-1 text-xs text-text-secondary">Added {new Date(game.added_at).toLocaleDateString('en-GB')}</p>
					</div>
				</a>
			{/each}
		</div>
	{/if}
</LoadState>
