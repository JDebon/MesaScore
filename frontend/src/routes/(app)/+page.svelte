<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import Avatar from '$components/ui/Avatar.svelte';
	import LoadState from '$components/ui/LoadState.svelte';
	import Skeleton from '$components/ui/Skeleton.svelte';
	import { usersApi } from '$api/users';
	import { partiesApi } from '$api/parties';
	import { addToast } from '$stores/toast.svelte';
	import { ApiError } from '$api/client';
	import { getUser } from '$stores/auth.svelte';
	import { setPendingInviteCount } from '$stores/notifications.svelte';
	import type { DashboardResponse } from '$api/types';

	let data = $state<DashboardResponse | null>(null);
	let loading = $state(true);
	let loadError = $state(false);
	let partyPickerOpen = $state(false);

	const user = $derived(getUser());

	const greeting = $derived.by(() => {
		const h = new Date().getHours();
		if (h < 5) return 'Good night';
		if (h < 12) return 'Good morning';
		if (h < 17) return 'Good afternoon';
		return 'Good evening';
	});

	const winRate = $derived(
		data && data.global_stats.total_sessions > 0
			? Math.round((data.global_stats.total_wins / data.global_stats.total_sessions) * 100)
			: 0
	);

	const recentParties = $derived(
		data
			? [...data.parties]
					.filter((p) => p.last_session_at !== null)
					.sort(
						(a, b) =>
							new Date(b.last_session_at!).getTime() - new Date(a.last_session_at!).getTime()
					)
					.slice(0, 4)
			: []
	);

	async function loadData() {
		loading = true;
		loadError = false;
		try {
			data = await usersApi.dashboard();
			setPendingInviteCount(data.pending_invites.length);
		} catch (e) {
			console.error('[dashboard] Failed to load:', e);
			loadError = true;
		} finally {
			loading = false;
		}
	}

	onMount(loadData);

	function handleLogSession() {
		if (!data || data.parties.length === 0) {
			goto('/parties/new');
			return;
		}
		if (data.parties.length === 1) {
			goto(`/parties/${data.parties[0].id}/sessions/new`);
			return;
		}
		partyPickerOpen = true;
	}

	async function acceptInvite(partyId: string, inviteId: string) {
		try {
			const res = await partiesApi.acceptInvite(partyId, inviteId);
			addToast('Joined party!', 'success');
			goto(`/parties/${res.party_id}`);
		} catch (e) {
			addToast(e instanceof ApiError ? e.message : 'Failed to accept invite', 'error');
		}
	}

	async function declineInvite(partyId: string, inviteId: string) {
		try {
			await partiesApi.declineInvite(partyId, inviteId);
			if (data) {
				data = {
					...data,
					pending_invites: data.pending_invites.filter((i) => i.id !== inviteId)
				};
				setPendingInviteCount(data.pending_invites.length);
			}
			addToast('Invite declined', 'info');
		} catch (e) {
			addToast('Failed to decline invite', 'error');
		}
	}

	function timeAgo(dateStr: string | null): string {
		if (!dateStr) return 'Never played';
		const diff = Date.now() - new Date(dateStr).getTime();
		const d = Math.floor(diff / 86400000);
		if (d === 0) return 'Today';
		if (d === 1) return 'Yesterday';
		if (d < 7) return `${d} days ago`;
		if (d < 30) return `${Math.floor(d / 7)}w ago`;
		return `${Math.floor(d / 30)}mo ago`;
	}
</script>

<svelte:head>
	<title>Dashboard — MesaScore</title>
</svelte:head>

<!-- Party picker modal -->
{#if partyPickerOpen}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-end justify-center bg-black/60 backdrop-blur-sm sm:items-center"
		onclick={() => (partyPickerOpen = false)}
		onkeydown={(e) => e.key === 'Escape' && (partyPickerOpen = false)}
	>
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="w-full max-w-sm rounded-t-2xl border border-border bg-surface p-5 sm:rounded-2xl"
			onclick={(e) => e.stopPropagation()}
			onkeydown={() => {}}
		>
			<div class="mb-4 flex items-center justify-between">
				<h2 class="text-base font-semibold text-text-primary" style="font-family: var(--font-display)">
					Which party?
				</h2>
				<button
					onclick={() => (partyPickerOpen = false)}
					class="rounded-lg p-1.5 text-text-secondary hover:bg-surface-raised"
					aria-label="Close"
				>
					<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>
			<div class="space-y-2">
				{#each data!.parties as party}
					<button
						onclick={() => {
							partyPickerOpen = false;
							goto(`/parties/${party.id}/sessions/new`);
						}}
						class="flex w-full items-center justify-between rounded-xl border border-border px-4 py-3 text-left transition-colors hover:border-primary-500/50 hover:bg-surface-raised"
					>
						<div>
							<p class="font-medium text-text-primary">{party.name}</p>
							<p class="text-xs text-text-secondary">{party.member_count} member{party.member_count === 1 ? '' : 's'}</p>
						</div>
						<svg class="h-4 w-4 text-text-secondary" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
						</svg>
					</button>
				{/each}
			</div>
		</div>
	</div>
{/if}

<LoadState {loading} error={loadError} onretry={loadData}>
	{#snippet skeleton()}
		<div class="space-y-6">
			<!-- hero skeleton -->
			<div class="rounded-2xl bg-surface p-6 space-y-4">
				<Skeleton class="h-7 w-48" />
				<Skeleton class="h-4 w-32" />
				<Skeleton class="h-12 w-full rounded-xl" />
			</div>
			<!-- stats skeleton -->
			<div class="grid grid-cols-3 gap-3">
				{#each [1, 2, 3] as _}
					<div class="rounded-2xl bg-surface p-4 space-y-2">
						<Skeleton class="h-7 w-10 mx-auto" />
						<Skeleton class="h-3 w-14 mx-auto" />
					</div>
				{/each}
			</div>
			<!-- activity skeleton -->
			<div class="space-y-3">
				<Skeleton class="h-4 w-32" />
				{#each [1, 2, 3] as _}
					<div class="rounded-2xl bg-surface p-4 space-y-2">
						<Skeleton class="h-4 w-40" />
						<Skeleton class="h-3 w-24" />
					</div>
				{/each}
			</div>
		</div>
	{/snippet}

	<div class="space-y-6">
		<!-- ── Hero banner ─────────────────────────────────────────────── -->
		<section
			class="relative overflow-hidden rounded-2xl border border-border bg-surface p-6"
			style="background-image: radial-gradient(ellipse 80% 60% at 100% 0%, color-mix(in srgb, var(--color-primary-500) 8%, transparent), transparent);"
		>
			<!-- Decorative dots grid -->
			<div
				class="pointer-events-none absolute inset-0 opacity-[0.03] dark:opacity-[0.06]"
				style="background-image: radial-gradient(circle, var(--color-text-primary) 1px, transparent 1px); background-size: 24px 24px;"
			></div>

			<div class="relative flex flex-col gap-5 sm:flex-row sm:items-center sm:justify-between">
				<div>
					<p class="text-xs font-medium uppercase tracking-widest text-text-secondary">
						{greeting}
					</p>
					<h1
						class="mt-1 text-2xl font-bold text-text-primary sm:text-3xl"
						style="font-family: var(--font-display)"
					>
						{user?.display_name?.split(' ')[0] ?? user?.username}
					</h1>
					<p class="mt-1 text-sm text-text-secondary">
						{#if data!.parties.length === 0}
							Create a party to start tracking your games.
						{:else if data!.global_stats.total_sessions === 0}
							Log your first session to get started.
						{:else}
							Ready for another game night?
						{/if}
					</p>
				</div>

				<button
					onclick={handleLogSession}
					class="group flex items-center gap-3 self-start rounded-xl bg-primary-500 px-5 py-3.5 text-sm font-semibold text-white shadow-md shadow-primary-900/20 transition-all duration-150 hover:bg-primary-600 hover:shadow-primary-900/30 active:scale-95 sm:self-center dark:shadow-primary-900/50"
				>
					<svg class="h-5 w-5 transition-transform duration-150 group-hover:scale-110" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
					</svg>
					Log a Session
				</button>
			</div>
		</section>

		<!-- ── Stats strip ─────────────────────────────────────────────── -->
		{#if data!.global_stats.total_sessions > 0}
			<div class="grid grid-cols-3 gap-3">
				<div class="rounded-2xl border border-border bg-surface px-4 py-4 text-center">
					<p
						class="text-2xl font-bold text-text-primary"
						style="font-family: var(--font-display)"
					>
						{data!.global_stats.total_sessions}
					</p>
					<p class="mt-0.5 text-[11px] font-medium uppercase tracking-wider text-text-secondary">
						Sessions
					</p>
				</div>

				<div class="rounded-2xl border border-border bg-surface px-4 py-4 text-center">
					<p
						class="text-2xl font-bold text-text-primary"
						style="font-family: var(--font-display)"
					>
						{winRate}<span class="text-base font-normal text-text-secondary">%</span>
					</p>
					<p class="mt-0.5 text-[11px] font-medium uppercase tracking-wider text-text-secondary">
						Win rate
					</p>
				</div>

				<div class="rounded-2xl border border-border bg-surface px-4 py-4 text-center">
					<p
						class="text-2xl font-bold text-text-primary"
						style="font-family: var(--font-display)"
					>
						{#if data!.global_stats.current_streak > 0}
							<span style="color: var(--color-gold)">{data!.global_stats.current_streak}</span>
						{:else}
							{data!.global_stats.current_streak}
						{/if}
					</p>
					<p class="mt-0.5 text-[11px] font-medium uppercase tracking-wider text-text-secondary">
						{data!.global_stats.current_streak > 1 ? '🔥 Streak' : 'Streak'}
					</p>
				</div>
			</div>
		{/if}

		<!-- ── Pending invites ─────────────────────────────────────────── -->
		{#if data!.pending_invites.length > 0}
			<section>
				<h2 class="mb-3 text-[11px] font-semibold uppercase tracking-widest text-text-secondary">
					Pending Invites
				</h2>
				<div class="space-y-2">
					{#each data!.pending_invites as invite}
						<div
							class="flex items-center justify-between rounded-2xl border border-primary-500/20 bg-primary-500/5 px-4 py-3.5"
						>
							<div>
								<p class="font-semibold text-text-primary">{invite.party.name}</p>
								<p class="text-xs text-text-secondary">
									from {invite.invited_by.display_name}
								</p>
							</div>
							<div class="flex gap-2">
								<button
									onclick={() => acceptInvite(invite.party.id, invite.id)}
									class="rounded-lg bg-primary-500 px-3 py-1.5 text-xs font-semibold text-white transition-colors hover:bg-primary-600"
								>
									Accept
								</button>
								<button
									onclick={() => declineInvite(invite.party.id, invite.id)}
									class="rounded-lg border border-border px-3 py-1.5 text-xs font-medium text-text-secondary transition-colors hover:bg-surface-raised hover:text-text-primary"
								>
									Decline
								</button>
							</div>
						</div>
					{/each}
				</div>
			</section>
		{/if}

		<!-- ── Recent activity ─────────────────────────────────────────── -->
		{#if recentParties.length > 0}
			<section>
				<div class="mb-3 flex items-center justify-between">
					<h2 class="text-[11px] font-semibold uppercase tracking-widest text-text-secondary">
						Recent Activity
					</h2>
					<a
						href="/parties"
						class="text-xs font-medium text-primary-500 hover:text-primary-600"
					>
						All parties →
					</a>
				</div>

				<div class="grid gap-3 sm:grid-cols-2">
					{#each recentParties as party}
						<a
							href="/parties/{party.id}"
							class="group flex items-center justify-between rounded-2xl border border-border bg-surface px-4 py-4 transition-all duration-150 hover:border-primary-500/30 hover:shadow-sm"
						>
							<div class="min-w-0">
								<p class="truncate font-semibold text-text-primary">{party.name}</p>
								<p class="mt-0.5 text-xs text-text-secondary">
									{party.member_count} member{party.member_count === 1 ? '' : 's'}
									<span class="mx-1.5 opacity-40">·</span>
									{timeAgo(party.last_session_at)}
								</p>
							</div>
							<svg
								class="ml-3 h-4 w-4 shrink-0 text-text-secondary transition-transform duration-150 group-hover:translate-x-0.5 group-hover:text-primary-500"
								fill="none"
								viewBox="0 0 24 24"
								stroke="currentColor"
								stroke-width="2"
							>
								<path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
							</svg>
						</a>
					{/each}
				</div>
			</section>
		{/if}

		<!-- ── Empty state ─────────────────────────────────────────────── -->
		{#if data!.parties.length === 0}
			<section class="rounded-2xl border border-dashed border-border px-6 py-10 text-center">
				<div class="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-surface-raised">
					<svg class="h-7 w-7 text-text-secondary" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
					</svg>
				</div>
				<p class="font-semibold text-text-primary" style="font-family: var(--font-display)">
					No parties yet
				</p>
				<p class="mt-1 text-sm text-text-secondary">
					Create a party or ask a friend for an invite link.
				</p>
				<a
					href="/parties/new"
					class="mt-4 inline-flex items-center gap-2 rounded-xl bg-primary-500 px-5 py-2.5 text-sm font-semibold text-white hover:bg-primary-600"
				>
					Create a party
				</a>
			</section>
		{/if}
	</div>
</LoadState>
