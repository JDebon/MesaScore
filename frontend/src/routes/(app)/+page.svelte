<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import Button from '$components/ui/Button.svelte';
	import Spinner from '$components/ui/Spinner.svelte';
	import Avatar from '$components/ui/Avatar.svelte';
	import EmptyState from '$components/ui/EmptyState.svelte';
	import { usersApi } from '$api/users';
	import { partiesApi } from '$api/parties';
	import { addToast } from '$stores/toast.svelte';
	import { ApiError } from '$api/client';
	import type { DashboardResponse } from '$api/types';

	let data = $state<DashboardResponse | null>(null);
	let loading = $state(true);

	onMount(async () => {
		try {
			data = await usersApi.dashboard();
		} catch (e) {
			console.error('[dashboard] Failed to load:', e);
			addToast('Failed to load dashboard', 'error');
		} finally {
			loading = false;
		}
	});

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
			}
			addToast('Invite declined', 'info');
		} catch (e) {
			console.error('[dashboard] Failed to decline invite:', e);
			addToast('Failed to decline invite', 'error');
		}
	}

	function formatDate(dateStr: string | null): string {
		if (!dateStr) return 'No sessions yet';
		return new Date(dateStr).toLocaleDateString();
	}
</script>

<svelte:head>
	<title>Dashboard - MesaScore</title>
</svelte:head>

{#if loading}
	<div class="flex justify-center py-20">
		<Spinner size="lg" />
	</div>
{:else if data}
	<!-- Stats strip -->
	<div class="mb-6 grid grid-cols-3 gap-3">
		<div class="rounded-lg bg-surface p-4 text-center shadow-sm">
			<p class="text-2xl font-bold text-text-primary">{data.global_stats.total_sessions}</p>
			<p class="text-xs text-text-secondary">Sessions</p>
		</div>
		<div class="rounded-lg bg-surface p-4 text-center shadow-sm">
			<p class="text-2xl font-bold text-text-primary">{data.global_stats.total_wins}</p>
			<p class="text-xs text-text-secondary">Wins</p>
		</div>
		<div class="rounded-lg bg-surface p-4 text-center shadow-sm">
			<p class="text-2xl font-bold text-text-primary">{data.global_stats.current_streak}</p>
			<p class="text-xs text-text-secondary">Streak</p>
		</div>
	</div>

	<!-- Pending invites -->
	{#if data.pending_invites.length > 0}
		<section class="mb-6">
			<h2 class="mb-3 text-lg font-semibold text-text-primary">Pending Invites</h2>
			<div class="space-y-3">
				{#each data.pending_invites as invite}
					<div class="flex items-center justify-between rounded-lg bg-primary-50 p-4 border border-primary-200">
						<div>
							<p class="font-medium text-text-primary">{invite.party.name}</p>
							<p class="text-sm text-text-secondary">Invited by {invite.invited_by.display_name}</p>
						</div>
						<div class="flex gap-2">
							<Button size="sm" onclick={() => acceptInvite(invite.party.id, invite.id)}>Accept</Button>
							<Button size="sm" variant="secondary" onclick={() => declineInvite(invite.party.id, invite.id)}>Decline</Button>
						</div>
					</div>
				{/each}
			</div>
		</section>
	{/if}

	<!-- Parties -->
	<section>
		<h2 class="mb-3 text-lg font-semibold text-text-primary">My Parties</h2>
		{#if data.parties.length === 0}
			<EmptyState message="Create your first party or ask a friend for an invite link.">
				<Button onclick={() => goto('/parties/new')}>Create a party</Button>
			</EmptyState>
		{:else}
			<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
				{#each data.parties as party}
					<a
						href="/parties/{party.id}"
						class="block rounded-lg bg-surface p-4 shadow-sm transition-shadow hover:shadow-md"
					>
						<h3 class="font-semibold text-text-primary">{party.name}</h3>
						<div class="mt-2 flex items-center gap-4 text-sm text-text-secondary">
							<span>{party.member_count} member{party.member_count === 1 ? '' : 's'}</span>
							<span>{formatDate(party.last_session_at)}</span>
						</div>
					</a>
				{/each}
			</div>
		{/if}
	</section>

	<!-- FAB -->
	<a
		href="/parties/new"
		class="fixed bottom-6 right-6 flex h-14 w-14 items-center justify-center rounded-full bg-primary-600 text-white shadow-lg transition-transform hover:scale-105 hover:bg-primary-700 lg:bottom-8 lg:right-8"
		aria-label="Create a party"
	>
		<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
			<path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"/>
		</svg>
	</a>
{/if}
