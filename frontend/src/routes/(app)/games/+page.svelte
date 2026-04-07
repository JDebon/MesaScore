<script lang="ts">
	import { onMount } from 'svelte';
	import Input from '$components/ui/Input.svelte';
	import Badge from '$components/ui/Badge.svelte';
	import Spinner from '$components/ui/Spinner.svelte';
	import EmptyState from '$components/ui/EmptyState.svelte';
	import { gamesApi } from '$api/games';
	import { addToast } from '$stores/toast.svelte';
	import type { Game } from '$api/types';

	let games = $state<Game[]>([]);
	let loading = $state(true);
	let query = $state('');
	let sort = $state('name');
	let searchTimer: ReturnType<typeof setTimeout>;

	onMount(() => loadGames());

	async function loadGames() {
		loading = true;
		try {
			games = await gamesApi.list({ q: query || undefined, sort });
		} catch (e) {
			console.error('[games] Failed to load:', e);
			addToast('Failed to load games', 'error');
		} finally {
			loading = false;
		}
	}

	function handleSearch() {
		clearTimeout(searchTimer);
		searchTimer = setTimeout(loadGames, 300);
	}

	function handleSort(e: Event) {
		sort = (e.target as HTMLSelectElement).value;
		loadGames();
	}
</script>

<svelte:head>
	<title>Game Catalog - MesaScore</title>
</svelte:head>

<div class="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
	<h1 class="text-2xl font-bold text-text-primary">Game Catalog</h1>
	<a
		href="/games/new"
		class="inline-flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white hover:bg-primary-700"
	>
		<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"/></svg>
		Add game
	</a>
</div>

<div class="mb-4 flex flex-col gap-3 sm:flex-row">
	<div class="flex-1">
		<Input name="search" placeholder="Search games..." bind:value={query} oninput={handleSearch} />
	</div>
	<select
		value={sort}
		onchange={handleSort}
		class="rounded-lg border border-border px-3 py-2 text-sm"
	>
		<option value="name">Name</option>
		<option value="rating">Rating</option>
		<option value="sessions">Sessions</option>
	</select>
</div>

{#if loading}
	<div class="flex justify-center py-20"><Spinner size="lg" /></div>
{:else if games.length === 0}
	<EmptyState message="No games in the catalog yet. Add one!" />
{:else}
	<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
		{#each games as game}
			<a
				href="/games/{game.id}"
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
					{#if game.min_players && game.max_players}
						<p class="text-sm text-text-secondary">{game.min_players}-{game.max_players} players</p>
					{/if}
					<div class="mt-1 flex items-center gap-2">
						{#if game.bgg_rating}
							<span class="text-sm text-text-secondary">&#9733; {game.bgg_rating.toFixed(1)}</span>
						{/if}
						{#if game.in_my_collection}
							<Badge variant="success">In collection</Badge>
						{/if}
					</div>
				</div>
			</a>
		{/each}
	</div>
{/if}
