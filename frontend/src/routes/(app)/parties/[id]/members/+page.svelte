<script lang="ts">
	import { onMount } from 'svelte';
	import Button from '$components/ui/Button.svelte';
	import Avatar from '$components/ui/Avatar.svelte';
	import Badge from '$components/ui/Badge.svelte';
	import Modal from '$components/ui/Modal.svelte';
	import Input from '$components/ui/Input.svelte';
	import Spinner from '$components/ui/Spinner.svelte';
	import { partiesApi } from '$api/parties';
	import { usersApi } from '$api/users';
	import { addToast } from '$stores/toast.svelte';
	import { getUser } from '$stores/auth.svelte';
	import type { PartyMember, PartyInvite, User } from '$api/types';
	import { ApiError } from '$api/client';

	interface Props {
		data: { party: import('$api/types').Party };
	}

	let { data: layoutData }: Props = $props();
	let members = $state<PartyMember[]>([]);
	let invites = $state<PartyInvite[]>([]);
	let loading = $state(true);
	const user = $derived(getUser());
	const isAdmin = $derived(user?.id === layoutData.party.admin.id);

	// Invite modal
	let showInviteModal = $state(false);
	let searchQuery = $state('');
	let searchResults = $state<User[]>([]);
	let searching = $state(false);
	let searchTimer: ReturnType<typeof setTimeout>;

	// Confirm remove modal
	let showRemoveModal = $state(false);
	let memberToRemove = $state<PartyMember | null>(null);
	let removing = $state(false);

	onMount(loadMembers);

	async function loadMembers() {
		try {
			const res = await partiesApi.members(layoutData.party.id);
			members = res.members;
			invites = res.invites;
		} catch (e) {
			console.error('[members] Failed to load:', e);
			addToast('Failed to load members', 'error');
		} finally {
			loading = false;
		}
	}

	function searchUsers() {
		clearTimeout(searchTimer);
		if (searchQuery.length < 2) {
			searchResults = [];
			return;
		}
		searching = true;
		searchTimer = setTimeout(async () => {
			try {
				const results = await usersApi.search(searchQuery);
				const memberIds = new Set(members.map((m) => m.id));
				const invitedIds = new Set(invites.filter((i) => i.status === 'pending').map((i) => i.invited_user.id));
				searchResults = results.filter((u) => !memberIds.has(u.id) && !invitedIds.has(u.id));
			} catch (e) {
				console.error('[members] User search failed:', e);
				searchResults = [];
			} finally {
				searching = false;
			}
		}, 300);
	}

	async function sendInvite(userId: string) {
		try {
			await partiesApi.sendInvite(layoutData.party.id, userId);
			addToast('Invite sent', 'success');
			showInviteModal = false;
			searchQuery = '';
			searchResults = [];
			await loadMembers();
		} catch (e) {
			addToast(e instanceof ApiError ? e.message : 'Failed to send invite', 'error');
		}
	}

	function confirmRemove(member: PartyMember) {
		memberToRemove = member;
		showRemoveModal = true;
	}

	async function removeMember() {
		if (!memberToRemove) return;
		removing = true;
		try {
			await partiesApi.removeMember(layoutData.party.id, memberToRemove.id);
			addToast('Member removed', 'success');
			showRemoveModal = false;
			memberToRemove = null;
			await loadMembers();
		} catch (e) {
			addToast(e instanceof ApiError ? e.message : 'Failed to remove member', 'error');
		} finally {
			removing = false;
		}
	}

	function copyInviteLink() {
		const url = `${window.location.origin}/join/${layoutData.party.invite_code}`;
		navigator.clipboard.writeText(url);
		addToast('Invite link copied!', 'success');
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString();
	}
</script>

<svelte:head>
	<title>Members - {layoutData.party.name} - MesaScore</title>
</svelte:head>

{#if loading}
	<div class="flex justify-center py-20"><Spinner size="lg" /></div>
{:else}
	<!-- Admin invite controls -->
	{#if isAdmin}
		<div class="mb-6 flex flex-wrap gap-3">
			<Button onclick={() => (showInviteModal = true)}>Invite by username</Button>
			<Button variant="secondary" onclick={copyInviteLink}>Copy invite link</Button>
		</div>
	{/if}

	<!-- Members list -->
	<section class="mb-8">
		<h2 class="mb-3 text-lg font-semibold text-text-primary">Members ({members.length})</h2>
		<div class="space-y-2">
			{#each members as member}
				<div class="flex items-center justify-between rounded-lg bg-surface p-3 shadow-sm">
					<a href="/parties/{layoutData.party.id}/users/{member.id}" class="flex items-center gap-3 min-w-0">
						<Avatar url={member.avatar_url} name={member.display_name} size="sm" />
						<div class="min-w-0">
							<div class="flex items-center gap-2">
								<span class="truncate font-medium text-text-primary">{member.display_name}</span>
								{#if member.is_admin}
									<span class="text-gold" title="Admin">&#9812;</span>
								{/if}
							</div>
							<p class="text-sm text-text-secondary">@{member.username} &middot; joined {formatDate(member.joined_at)}</p>
						</div>
					</a>
					{#if isAdmin && !member.is_admin}
						<Button size="sm" variant="danger" onclick={() => confirmRemove(member)}>Remove</Button>
					{/if}
				</div>
			{/each}
		</div>
	</section>

	<!-- Pending invites (admin only) -->
	{#if isAdmin && invites.length > 0}
		<section>
			<h2 class="mb-3 text-lg font-semibold text-text-primary">Invites</h2>
			<div class="space-y-2">
				{#each invites as invite}
					<div class="flex items-center justify-between rounded-lg bg-surface p-3 shadow-sm">
						<div class="flex items-center gap-3">
							<Avatar name={invite.invited_user.display_name} size="sm" />
							<div>
								<p class="font-medium text-text-primary">{invite.invited_user.display_name}</p>
								<p class="text-sm text-text-secondary">@{invite.invited_user.username}</p>
							</div>
						</div>
						<Badge variant={invite.status === 'pending' ? 'info' : invite.status === 'declined' ? 'danger' : 'success'}>
							{invite.status}
						</Badge>
					</div>
				{/each}
			</div>
		</section>
	{/if}
{/if}

<!-- Invite modal -->
<Modal bind:open={showInviteModal} title="Invite by username" onclose={() => { showInviteModal = false; searchQuery = ''; searchResults = []; }}>
	<Input
		name="search"
		placeholder="Search by username or name..."
		bind:value={searchQuery}
		oninput={searchUsers}
	/>
	{#if searching}
		<div class="mt-4 flex justify-center"><Spinner size="sm" /></div>
	{:else if searchResults.length > 0}
		<div class="mt-3 max-h-60 space-y-2 overflow-y-auto">
			{#each searchResults as result}
				<button
					class="flex w-full items-center gap-3 rounded-lg p-2 text-left hover:bg-bg"
					onclick={() => sendInvite(result.id)}
				>
					<Avatar url={result.avatar_url} name={result.display_name} size="sm" />
					<div>
						<p class="font-medium text-text-primary">{result.display_name}</p>
						<p class="text-sm text-text-secondary">@{result.username}</p>
					</div>
				</button>
			{/each}
		</div>
	{:else if searchQuery.length >= 2}
		<p class="mt-4 text-center text-sm text-text-secondary">No users found</p>
	{/if}
</Modal>

<!-- Remove confirm modal -->
<Modal bind:open={showRemoveModal} title="Remove member" onclose={() => { showRemoveModal = false; memberToRemove = null; }}>
	<p class="text-text-secondary">Are you sure you want to remove <strong>{memberToRemove?.display_name}</strong> from this party?</p>
	<p class="mt-2 text-sm text-text-secondary">Their historical session records will be preserved.</p>
	{#snippet actions()}
		<Button variant="secondary" onclick={() => (showRemoveModal = false)}>Cancel</Button>
		<Button variant="danger" loading={removing} onclick={removeMember}>Remove</Button>
	{/snippet}
</Modal>
