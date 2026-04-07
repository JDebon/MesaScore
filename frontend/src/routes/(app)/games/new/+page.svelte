<script lang="ts">
	import { goto } from '$app/navigation';
	import Button from '$components/ui/Button.svelte';
	import Input from '$components/ui/Input.svelte';
	import Spinner from '$components/ui/Spinner.svelte';
	import { gamesApi } from '$api/games';
	import { addToast } from '$stores/toast.svelte';
	import { ApiError } from '$api/client';
	import type { BggSearchResult, BGGGameDetail } from '$api/types';

	type View = 'search' | 'preview' | 'manual';

	let view = $state<View>('search');
	let searchQuery = $state('');
	let searchResults = $state<BggSearchResult[]>([]);
	let searching = $state(false);
	let searchTimer: ReturnType<typeof setTimeout>;

	let selectedBggId = $state<number | null>(null);
	let preview = $state<BGGGameDetail | null>(null);
	let loadingPreview = $state(false);

	let manualName = $state('');
	let saving = $state(false);

	// For 409 conflict case
	let conflictGameId = $state<string | null>(null);

	function searchBgg() {
		clearTimeout(searchTimer);
		conflictGameId = null;
		if (searchQuery.length < 3) {
			searchResults = [];
			return;
		}
		searching = true;
		searchTimer = setTimeout(async () => {
			try {
				searchResults = await gamesApi.bggSearch(searchQuery);
			} catch (e) {
				console.error('[games/new] BGG search failed:', e);
				searchResults = [];
			} finally {
				searching = false;
			}
		}, 300);
	}

	async function selectResult(result: BggSearchResult) {
		selectedBggId = result.bgg_id;
		view = 'preview';
		loadingPreview = true;
		try {
			preview = await gamesApi.bggThing(result.bgg_id);
		} catch (e) {
			console.error('[games/new] BGG detail fetch failed:', e);
			addToast('Failed to fetch game details from BGG', 'error');
			view = 'search';
		} finally {
			loadingPreview = false;
		}
	}

	function backToSearch() {
		view = 'search';
		preview = null;
		selectedBggId = null;
		conflictGameId = null;
	}

	async function addToCatalog() {
		if (!selectedBggId) return;
		saving = true;
		conflictGameId = null;
		try {
			const res = await gamesApi.create({ bgg_id: selectedBggId });
			addToast('Game added to catalog!', 'success');
			goto(`/games/${res.id}`);
		} catch (e) {
			if (e instanceof ApiError && e.status === 409) {
				// Fetch the existing game id to link to collection
				const games = await gamesApi.list({ q: preview?.name ?? '' }).catch(() => []);
				const match = games.find((g) => g.bgg_id === selectedBggId);
				conflictGameId = match?.id ?? null;
			} else {
				addToast(e instanceof ApiError ? e.message : 'Failed to add game', 'error');
			}
		} finally {
			saving = false;
		}
	}

	async function addManually() {
		if (!manualName.trim()) return;
		saving = true;
		try {
			const res = await gamesApi.create({ name: manualName.trim() });
			addToast('Game added!', 'success');
			goto(`/games/${res.id}`);
		} catch (e) {
			addToast(e instanceof ApiError ? e.message : 'Failed to add game', 'error');
		} finally {
			saving = false;
		}
	}

	function formatPlayers(min: number | null, max: number | null): string {
		if (min && max) return min === max ? `${min} players` : `${min}–${max} players`;
		if (min) return `${min}+ players`;
		if (max) return `Up to ${max} players`;
		return '';
	}
</script>

<svelte:head>
	<title>Add Game - MesaScore</title>
</svelte:head>

<div class="mx-auto max-w-lg">
	<h1 class="mb-6 text-2xl font-bold text-text-primary">Add a Game</h1>

	{#if view === 'search'}
		<!-- State 1: BGG search -->
		<Input
			name="bgg_search"
			placeholder="Search BoardGameGeek..."
			bind:value={searchQuery}
			oninput={searchBgg}
		/>

		{#if searching}
			<div class="mt-6 flex justify-center"><Spinner size="sm" /></div>
		{:else if searchResults.length > 0}
			<div class="mt-4 max-h-80 space-y-2 overflow-y-auto">
				{#each searchResults as result}
					<button
						class="flex w-full items-center gap-3 rounded-lg bg-surface p-3 text-left shadow-sm hover:shadow-md"
						onclick={() => selectResult(result)}
					>
						{#if result.thumbnail_url}
							<img src={result.thumbnail_url} alt="" class="h-12 w-10 rounded object-cover shrink-0" />
						{:else}
							<div class="flex h-12 w-10 shrink-0 items-center justify-center rounded bg-surface-raised text-text-secondary text-lg">?</div>
						{/if}
						<div class="min-w-0">
							<p class="font-medium text-text-primary truncate">{result.name}</p>
							{#if result.year}
								<p class="text-sm text-text-secondary">({result.year})</p>
							{/if}
						</div>
					</button>
				{/each}
			</div>
		{:else if searchQuery.length >= 3}
			<p class="mt-6 text-center text-sm text-text-secondary">No results found</p>
		{/if}

		<p class="mt-6 text-center text-sm text-text-secondary">
			Can't find it?
			<button class="text-primary-600 hover:underline" onclick={() => (view = 'manual')}>
				Add manually
			</button>
		</p>

	{:else if view === 'preview'}
		<!-- State 2: BGG preview -->
		<button class="mb-4 text-sm text-primary-600 hover:text-primary-700" onclick={backToSearch}>
			&larr; Back
		</button>

		{#if loadingPreview}
			<div class="flex justify-center py-20"><Spinner size="lg" /></div>
		{:else if preview}
			<div class="rounded-xl bg-surface p-5 shadow-sm space-y-4">
				{#if preview.cover_image_url}
					<img
						src={preview.cover_image_url}
						alt={preview.name}
						class="mx-auto h-48 w-36 rounded-lg object-cover shadow"
					/>
				{/if}

				<div>
					<h2 class="text-xl font-bold text-text-primary">{preview.name}</h2>
					{#if preview.year}
						<p class="text-sm text-text-secondary">{preview.year}</p>
					{/if}
				</div>

				<div class="flex flex-wrap gap-4 text-sm text-text-secondary">
					{#if preview.min_players || preview.max_players}
						<span>{formatPlayers(preview.min_players, preview.max_players)}</span>
					{/if}
					{#if preview.bgg_rating}
						<span>BGG rating: {preview.bgg_rating.toFixed(1)}</span>
					{/if}
				</div>

				{#if conflictGameId}
					<div class="rounded-lg bg-warning-50 border border-warning-200 p-3 text-sm">
						<p class="font-medium text-warning-800">This game is already in the catalog.</p>
						<a href="/games/{conflictGameId}" class="mt-1 block text-primary-600 hover:underline">
							View game &rarr;
						</a>
					</div>
				{:else}
					<Button onclick={addToCatalog} loading={saving} class="w-full">Add to catalog</Button>
				{/if}
			</div>
		{/if}

	{:else}
		<!-- State 3: Manual add -->
		<button class="mb-4 text-sm text-primary-600 hover:text-primary-700" onclick={() => (view = 'search')}>
			&larr; Back to search
		</button>

		<div class="rounded-xl bg-surface p-5 shadow-sm space-y-4">
			<Input name="name" label="Game name" bind:value={manualName} required />
			<Button onclick={addManually} loading={saving} class="w-full" disabled={!manualName.trim()}>
				Add
			</Button>
		</div>
	{/if}
</div>
