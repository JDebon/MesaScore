<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import Button from '$components/ui/Button.svelte';
	import Input from '$components/ui/Input.svelte';
	import Modal from '$components/ui/Modal.svelte';
	import { partiesApi } from '$api/parties';
	import { addToast } from '$stores/toast.svelte';
	import { getUser } from '$stores/auth.svelte';
	import { ApiError } from '$api/client';
	import type { PartyMember } from '$api/types';

	interface Props {
		data: { party: import('$api/types').Party };
	}

	let { data: layoutData }: Props = $props();
	const user = $derived(getUser());
	const isAdmin = $derived(user?.id === layoutData.party.admin.id);

	let name = $state(layoutData.party.name);
	let description = $state(layoutData.party.description || '');
	let saving = $state(false);
	let inviteCode = $state(layoutData.party.invite_code);

	// Transfer ownership
	let members = $state<PartyMember[]>([]);
	let showTransferModal = $state(false);
	let transferTarget = $state('');
	let transferring = $state(false);

	// Leave
	let showLeaveModal = $state(false);
	let leaving = $state(false);

	// Regenerate
	let showRegenModal = $state(false);
	let regenerating = $state(false);

	onMount(async () => {
		try {
			const res = await partiesApi.members(layoutData.party.id);
			members = res.members;
		} catch (e) {
			console.error('[settings] Failed to load members:', e);
		}
	});

	async function saveParty(event: SubmitEvent) {
		event.preventDefault();
		saving = true;
		try {
			await partiesApi.update(layoutData.party.id, { name: name.trim(), description: description.trim() || null });
			addToast('Party updated', 'success');
		} catch (e) {
			addToast(e instanceof ApiError ? e.message : 'Failed to update', 'error');
		} finally {
			saving = false;
		}
	}

	function copyInviteLink() {
		navigator.clipboard.writeText(`${window.location.origin}/join/${inviteCode}`);
		addToast('Invite link copied!', 'success');
	}

	async function regenerateInvite() {
		regenerating = true;
		try {
			const res = await partiesApi.regenerateInvite(layoutData.party.id);
			inviteCode = res.invite_code;
			addToast('Invite link regenerated', 'success');
			showRegenModal = false;
		} catch (e) {
			addToast(e instanceof ApiError ? e.message : 'Failed to regenerate', 'error');
		} finally {
			regenerating = false;
		}
	}

	async function transferAdmin() {
		if (!transferTarget) return;
		transferring = true;
		try {
			await partiesApi.transferAdmin(layoutData.party.id, transferTarget);
			addToast('Ownership transferred', 'success');
			showTransferModal = false;
			goto(`/parties/${layoutData.party.id}`);
		} catch (e) {
			addToast(e instanceof ApiError ? e.message : 'Failed to transfer', 'error');
		} finally {
			transferring = false;
		}
	}

	async function leaveParty() {
		leaving = true;
		try {
			await partiesApi.leave(layoutData.party.id);
			addToast('You left the party', 'info');
			goto('/');
		} catch (e) {
			addToast(e instanceof ApiError ? e.message : 'Failed to leave', 'error');
		} finally {
			leaving = false;
		}
	}
</script>

<svelte:head>
	<title>Settings - {layoutData.party.name} - MesaScore</title>
</svelte:head>

<div class="mx-auto max-w-lg space-y-8">
	{#if isAdmin}
		<!-- Edit party -->
		<section>
			<h2 class="mb-4 text-lg font-semibold text-text-primary">Edit Party</h2>
			<form onsubmit={saveParty} class="rounded-xl bg-surface p-6 shadow-sm space-y-4">
				<Input name="name" label="Party name" bind:value={name} required />
				<div class="space-y-1">
					<label for="desc" class="block text-sm font-medium text-text-primary">Description</label>
					<textarea
						id="desc"
						bind:value={description}
						rows={3}
						class="block w-full rounded-lg border border-border px-3 py-2 text-sm shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500"
					></textarea>
				</div>
				<Button type="submit" loading={saving}>Save</Button>
			</form>
		</section>

		<!-- Invite link -->
		<section>
			<h2 class="mb-4 text-lg font-semibold text-text-primary">Invite Link</h2>
			<div class="rounded-xl bg-surface p-6 shadow-sm space-y-4">
				<div class="flex items-center gap-2">
					<code class="flex-1 truncate rounded bg-surface-raised px-3 py-2 text-sm text-text-secondary font-mono">{`${typeof window !== 'undefined' ? window.location.origin : ''}/join/${inviteCode}`}</code>
					<Button size="sm" variant="secondary" onclick={copyInviteLink}>Copy</Button>
				</div>
				<Button variant="danger" size="sm" onclick={() => (showRegenModal = true)}>Regenerate</Button>
			</div>
		</section>

		<!-- Transfer ownership -->
		<section>
			<h2 class="mb-4 text-lg font-semibold text-text-primary">Transfer Ownership</h2>
			<div class="rounded-xl bg-surface p-6 shadow-sm">
				<p class="mb-4 text-sm text-text-secondary">Transfer admin rights to another member. You will become a regular member.</p>
				<Button variant="danger" onclick={() => (showTransferModal = true)}>Transfer Ownership</Button>
			</div>
		</section>
	{:else}
		<!-- Non-admin: leave -->
		<section>
			<h2 class="mb-4 text-lg font-semibold text-text-primary">Leave Party</h2>
			<div class="rounded-xl bg-surface p-6 shadow-sm">
				<p class="mb-4 text-sm text-text-secondary">Your historical session records will be preserved.</p>
				<Button variant="danger" onclick={() => (showLeaveModal = true)}>Leave Party</Button>
			</div>
		</section>
	{/if}
</div>

<!-- Regenerate confirm -->
<Modal bind:open={showRegenModal} title="Regenerate invite link" onclose={() => (showRegenModal = false)}>
	<p class="text-text-secondary">This will invalidate the current invite link. Anyone with the old link won't be able to join.</p>
	{#snippet actions()}
		<Button variant="secondary" onclick={() => (showRegenModal = false)}>Cancel</Button>
		<Button variant="danger" loading={regenerating} onclick={regenerateInvite}>Regenerate</Button>
	{/snippet}
</Modal>

<!-- Transfer confirm -->
<Modal bind:open={showTransferModal} title="Transfer ownership" onclose={() => (showTransferModal = false)}>
	<p class="mb-4 text-text-secondary">Are you sure? You will lose admin privileges.</p>
	<select
		bind:value={transferTarget}
		class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text-primary"
	>
		<option value="">Select a member</option>
		{#each members.filter((m) => !m.is_admin) as member}
			<option value={member.id}>{member.display_name} (@{member.username})</option>
		{/each}
	</select>
	{#snippet actions()}
		<Button variant="secondary" onclick={() => (showTransferModal = false)}>Cancel</Button>
		<Button variant="danger" loading={transferring} disabled={!transferTarget} onclick={transferAdmin}>Transfer</Button>
	{/snippet}
</Modal>

<!-- Leave confirm -->
<Modal bind:open={showLeaveModal} title="Leave party" onclose={() => (showLeaveModal = false)}>
	<p class="text-text-secondary">Are you sure you want to leave <strong>{layoutData.party.name}</strong>? Your historical session records will be preserved.</p>
	{#snippet actions()}
		<Button variant="secondary" onclick={() => (showLeaveModal = false)}>Cancel</Button>
		<Button variant="danger" loading={leaving} onclick={leaveParty}>Leave</Button>
	{/snippet}
</Modal>
