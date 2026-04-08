<script lang="ts">
	import { onMount } from 'svelte';
	import Avatar from '$components/ui/Avatar.svelte';
	import EmptyState from '$components/ui/EmptyState.svelte';
	import LoadState from '$components/ui/LoadState.svelte';
	import Skeleton from '$components/ui/Skeleton.svelte';
	import { partiesApi } from '$api/parties';
	import { addToast } from '$stores/toast.svelte';
	import { getUser } from '$stores/auth.svelte';
	import type { PartyDashboard } from '$api/types';

	interface Props {
		data: { party: import('$api/types').Party };
	}

	let { data: layoutData }: Props = $props();
	let dashboard = $state<PartyDashboard | null>(null);
	let loading = $state(true);
	let loadError = $state(false);
	const user = $derived(getUser());
	const isAdmin = $derived(user?.id === layoutData.party.admin.id);

	async function loadData() {
		loading = true;
		loadError = false;
		try {
			dashboard = await partiesApi.dashboard(layoutData.party.id);
		} catch (e) {
			console.error('[party dashboard] Failed to load:', e);
			loadError = true;
		} finally {
			loading = false;
		}
	}

	onMount(loadData);

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString();
	}

	// Build a 12-month window with counts, oldest to newest
	const chartData = $derived.by(() => {
		if (!dashboard) return [];
		const map = new Map(dashboard.sessions_per_month.map((m) => [m.month, m.count]));
		const now = new Date();
		return Array.from({ length: 12 }, (_, i) => {
			const d = new Date(now.getFullYear(), now.getMonth() - (11 - i), 1);
			const key = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
			const label = d.toLocaleDateString('en', { month: 'short' });
			return { key, label, count: map.get(key) ?? 0 };
		});
	});

	const chartMax = $derived(Math.max(1, ...chartData.map((d) => d.count)));
</script>

<svelte:head>
	<title>{layoutData.party.name} - MesaScore</title>
</svelte:head>

<LoadState {loading} error={loadError} onretry={loadData}>
	{#snippet skeleton()}
		<!-- Stats strip -->
		<div class="mb-6 grid grid-cols-3 gap-3">
			{#each [1, 2, 3] as _}
				<div class="rounded-lg bg-surface p-4 shadow-sm space-y-2">
					<Skeleton class="h-7 w-12 mx-auto" />
					<Skeleton class="h-3 w-16 mx-auto" />
				</div>
			{/each}
		</div>
		<!-- Leader card -->
		<div class="mb-6 rounded-lg bg-surface p-4 shadow-sm flex items-center gap-3">
			<Skeleton class="h-10 w-10 rounded-full" />
			<div class="space-y-2 flex-1">
				<Skeleton class="h-4 w-32" />
				<Skeleton class="h-3 w-20" />
			</div>
		</div>
		<!-- Activity chart -->
		<div class="mb-6 rounded-lg bg-surface p-4 shadow-sm">
			<Skeleton class="h-4 w-20 mb-3" />
			<Skeleton class="h-20 w-full" />
		</div>
		<!-- Recent sessions -->
		<Skeleton class="h-5 w-36 mb-3" />
		<div class="space-y-2">
			{#each [1, 2, 3] as _}
				<div class="flex items-center gap-3 rounded-lg bg-surface p-3 shadow-sm">
					<Skeleton class="h-10 w-10 rounded" />
					<div class="flex-1 space-y-2">
						<Skeleton class="h-4 w-40" />
						<Skeleton class="h-3 w-28" />
					</div>
				</div>
			{/each}
		</div>
	{/snippet}

	<!-- Stats strip -->
	<div class="mb-6 grid grid-cols-3 gap-3">
		<div class="rounded-lg bg-surface p-4 text-center shadow-sm">
			<p class="text-2xl font-bold text-text-primary">{dashboard!.total_sessions}</p>
			<p class="text-xs text-text-secondary">Sessions</p>
		</div>
		<div class="rounded-lg bg-surface p-4 text-center shadow-sm">
			<p class="text-2xl font-bold text-text-primary">{dashboard!.total_unique_games}</p>
			<p class="text-xs text-text-secondary">Games</p>
		</div>
		<div class="rounded-lg bg-surface p-4 text-center shadow-sm">
			<p class="text-2xl font-bold text-text-primary">{dashboard!.total_members}</p>
			<p class="text-xs text-text-secondary">Members</p>
		</div>
	</div>

	<!-- Current leader -->
	{#if dashboard!.current_leader}
		<div class="mb-6 rounded-lg bg-gradient-to-r from-primary-50 to-primary-100 dark:from-primary-900/20 dark:to-primary-800/20 p-4">
			<p class="mb-2 text-xs font-medium text-primary-600 uppercase tracking-wide">Current Leader</p>
			<a href="/parties/{layoutData.party.id}/users/{dashboard!.current_leader.user.id}" class="flex items-center gap-3">
				<Avatar url={dashboard!.current_leader.user.avatar_url} name={dashboard!.current_leader.user.display_name} size="md" />
				<div>
					<p class="font-semibold text-text-primary">{dashboard!.current_leader.user.display_name}</p>
					<p class="text-sm text-text-secondary">{dashboard!.current_leader.wins} win{dashboard!.current_leader.wins === 1 ? '' : 's'}</p>
				</div>
			</a>
		</div>
	{/if}

	<!-- Most played game -->
	{#if dashboard!.most_played_game}
		<div class="mb-6 rounded-lg bg-surface p-4 shadow-sm">
			<p class="mb-2 text-xs font-medium text-text-secondary uppercase tracking-wide">Most Played</p>
			<a href="/games/{dashboard!.most_played_game.id}?party_id={layoutData.party.id}" class="flex items-center gap-3">
				{#if dashboard!.most_played_game.cover_image_url}
					<img src={dashboard!.most_played_game.cover_image_url} alt="" class="h-12 w-12 rounded-lg object-cover" />
				{:else}
					<div class="flex h-12 w-12 items-center justify-center rounded-lg bg-surface-raised text-text-secondary">
						<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"/></svg>
					</div>
				{/if}
				<div>
					<p class="font-medium text-text-primary">{dashboard!.most_played_game.name}</p>
					<p class="text-sm text-text-secondary">{dashboard!.most_played_game.session_count} session{dashboard!.most_played_game.session_count === 1 ? '' : 's'}</p>
				</div>
			</a>
		</div>
	{/if}

	<!-- Activity chart -->
	<section class="mb-6">
		<h2 class="mb-3 text-lg font-semibold text-text-primary">Activity</h2>
		<div class="rounded-lg bg-surface p-4 shadow-sm">
			<div class="flex h-20 items-end gap-0.5">
				{#each chartData as bar}
					<div
						class="flex-1 rounded-t bg-primary-400 dark:bg-primary-500 transition-all min-h-[2px]"
						style="height: {Math.max(2, (bar.count / chartMax) * 80)}px"
						title="{bar.label}: {bar.count} session{bar.count === 1 ? '' : 's'}"
					></div>
				{/each}
			</div>
			<div class="mt-1 flex gap-0.5">
				{#each chartData as bar}
					<div class="flex-1 truncate text-center text-[9px] text-text-secondary">{bar.label}</div>
				{/each}
			</div>
		</div>
	</section>

	<!-- Recent sessions -->
	<section class="mb-6">
		<div class="mb-3 flex items-center justify-between">
			<h2 class="text-lg font-semibold text-text-primary">Recent Sessions</h2>
			<a href="/parties/{layoutData.party.id}/sessions" class="text-sm font-medium text-primary-600 hover:text-primary-700">View all</a>
		</div>

		{#if dashboard!.recent_sessions.length === 0}
			<EmptyState message={isAdmin ? 'No sessions yet. Log your first game!' : 'No sessions logged yet.'} />
		{:else}
			<div class="space-y-2">
				{#each dashboard!.recent_sessions as session}
					<a
						href="/parties/{layoutData.party.id}/sessions/{session.id}"
						class="flex items-center gap-3 rounded-lg bg-surface p-3 shadow-sm transition-shadow hover:shadow-md"
					>
						{#if session.game.cover_image_url}
							<img src={session.game.cover_image_url} alt="" class="h-10 w-10 rounded object-cover" />
						{:else}
							<div class="flex h-10 w-10 items-center justify-center rounded bg-surface-raised text-text-secondary">
								<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"/></svg>
							</div>
						{/if}
						<div class="min-w-0 flex-1">
							<p class="truncate font-medium text-text-primary">{session.game.name}</p>
							<p class="text-sm text-text-secondary">
								{formatDate(session.played_at)}
								{#if session.winners.length > 0}
									&mdash; {session.winners.map((w) => w.display_name).join(', ')}
								{/if}
							</p>
						</div>
					</a>
				{/each}
			</div>
		{/if}
	</section>

	<!-- Admin FAB for logging sessions -->
	{#if isAdmin}
		<a
			href="/parties/{layoutData.party.id}/sessions/new"
			class="fixed bottom-20 right-6 flex h-14 w-14 items-center justify-center rounded-full bg-primary-600 text-white shadow-lg transition-transform hover:scale-105 hover:bg-primary-700 lg:bottom-8 lg:right-8"
			aria-label="Log session"
		>
			<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"/>
			</svg>
		</a>
	{/if}
</LoadState>
