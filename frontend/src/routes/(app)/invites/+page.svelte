<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import LoadState from '$components/ui/LoadState.svelte';
	import Skeleton from '$components/ui/Skeleton.svelte';
	import { usersApi } from '$api/users';
	import { partiesApi } from '$api/parties';
	import { addToast } from '$stores/toast.svelte';
	import { ApiError } from '$api/client';
	import { setPendingInviteCount } from '$stores/notifications.svelte';
	import type { PendingInvite } from '$api/types';

	let invites = $state<PendingInvite[]>([]);
	let loading = $state(true);
	let loadError = $state(false);

	async function loadData() {
		loading = true;
		loadError = false;
		try {
			const data = await usersApi.dashboard();
			invites = data.pending_invites;
			setPendingInviteCount(data.pending_invites.length);
		} catch (e) {
			console.error('[invites] Failed to load:', e);
			loadError = true;
		} finally {
			loading = false;
		}
	}

	onMount(loadData);

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
			invites = invites.filter((i) => i.id !== inviteId);
			setPendingInviteCount(invites.length);
			addToast('Invite declined', 'info');
		} catch (e) {
			addToast('Failed to decline invite', 'error');
		}
	}

	function timeAgo(dateStr: string): string {
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
	<title>Invites — MesaScore</title>
</svelte:head>

<LoadState {loading} error={loadError} onretry={loadData}>
	{#snippet skeleton()}
		<div class="space-y-4">
			<Skeleton class="h-7 w-32" />
			{#each [1, 2] as _}
				<div class="rounded-2xl bg-surface border border-border p-4 space-y-2">
					<Skeleton class="h-5 w-40" />
					<Skeleton class="h-3 w-28" />
				</div>
			{/each}
		</div>
	{/snippet}

	<div class="space-y-5">
		<h1 class="text-xl font-bold text-text-primary" style="font-family: var(--font-display)">
			Invites
		</h1>

		{#if invites.length === 0}
			<section class="rounded-2xl border border-dashed border-border px-6 py-12 text-center">
				<div class="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-surface-raised">
					<svg class="h-7 w-7 text-text-secondary" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
					</svg>
				</div>
				<p class="font-semibold text-text-primary" style="font-family: var(--font-display)">No pending invites</p>
				<p class="mt-1 text-sm text-text-secondary">Ask a party admin to invite you or use a join link.</p>
			</section>
		{:else}
			<div class="space-y-3">
				{#each invites as invite}
					<div class="flex items-center justify-between rounded-2xl border border-primary-500/20 bg-primary-500/5 px-4 py-4">
						<div class="min-w-0">
							<p class="font-semibold text-text-primary">{invite.party.name}</p>
							<p class="mt-0.5 text-xs text-text-secondary">
								from {invite.invited_by.display_name}
								<span class="mx-1.5 opacity-40">·</span>
								{timeAgo(invite.created_at)}
							</p>
						</div>
						<div class="ml-3 flex shrink-0 gap-2">
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
		{/if}
	</div>
</LoadState>
