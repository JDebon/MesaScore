<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import Button from '$components/ui/Button.svelte';
	import Badge from '$components/ui/Badge.svelte';
	import Avatar from '$components/ui/Avatar.svelte';
	import Modal from '$components/ui/Modal.svelte';
	import Spinner from '$components/ui/Spinner.svelte';
	import { sessionsApi } from '$api/sessions';
	import { addToast } from '$stores/toast.svelte';
	import { getUser } from '$stores/auth.svelte';
	import { ApiError } from '$api/client';
	import type { SessionDetail } from '$api/types';

	interface Props {
		data: { party: import('$api/types').Party };
	}

	let { data: layoutData }: Props = $props();
	let session = $state<SessionDetail | null>(null);
	let loading = $state(true);
	let showDeleteModal = $state(false);
	let deleting = $state(false);

	const user = $derived(getUser());
	const isAdmin = $derived(user?.id === layoutData.party.admin.id);
	const partyId = $derived(layoutData.party.id);

	onMount(async () => {
		try {
			session = await sessionsApi.get(partyId, page.params.sessionId!);
		} catch (e) {
			console.error('[session detail] Failed to load:', e);
			addToast('Failed to load session', 'error');
		} finally {
			loading = false;
		}
	});

	async function deleteSession() {
		if (!session) return;
		deleting = true;
		try {
			await sessionsApi.delete(partyId, session.id);
			addToast('Session deleted', 'success');
			goto(`/parties/${partyId}/sessions`);
		} catch (e) {
			addToast(e instanceof ApiError ? e.message : 'Failed to delete', 'error');
		} finally {
			deleting = false;
		}
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString(undefined, {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	const typeBadge: Record<string, 'default' | 'info' | 'success' | 'warning'> = {
		competitive: 'info',
		team: 'warning',
		cooperative: 'success',
		score: 'default'
	};
</script>

<svelte:head>
	<title>Session Detail - MesaScore</title>
</svelte:head>

{#if loading}
	<div class="flex justify-center py-20"><Spinner size="lg" /></div>
{:else if session}
	<div class="mx-auto max-w-2xl">
		<!-- Header -->
		<div class="mb-6 flex items-start gap-4">
			{#if session.game.cover_image_url}
				<img src={session.game.cover_image_url} alt="" class="h-24 w-18 rounded-lg object-cover shadow" />
			{/if}
			<div class="flex-1">
				<a href="/games/{session.game.id}?party_id={partyId}" class="text-xl font-bold text-text-primary hover:text-primary-600">{session.game.name}</a>
				<div class="mt-1 flex flex-wrap items-center gap-2 text-sm text-text-secondary">
					<Badge variant={typeBadge[session.session_type] ?? 'default'}>{session.session_type}</Badge>
					<span>{formatDate(session.played_at)}</span>
					{#if session.duration_minutes}
						<span>&middot; {session.duration_minutes} min</span>
					{/if}
				</div>
			</div>
		</div>

		<!-- Meta -->
		<div class="mb-6 space-y-2 rounded-lg bg-surface p-4 shadow-sm text-sm text-text-secondary">
			{#if session.notes}
				<p><strong>Notes:</strong> {session.notes}</p>
			{/if}
			{#if session.brought_by}
				<p><strong>Brought by:</strong> {session.brought_by.display_name}</p>
			{/if}
			<p><strong>Logged by:</strong> {session.created_by.display_name}</p>
		</div>

		<!-- Results -->
		<section class="mb-6">
			<h2 class="mb-3 text-lg font-semibold text-text-primary">Results</h2>

			{#if session.session_type === 'cooperative'}
				{@const won = session.participants.some((p) => p.result === 'win')}
				<div class="mb-4 rounded-lg p-4 text-center font-bold text-lg {won ? 'bg-success-50 text-success-600 dark:bg-success-600/20 dark:text-success-400' : 'bg-danger-50 text-danger-600 dark:bg-danger-600/20 dark:text-danger-400'}">
					{won ? 'Victory!' : 'Defeat'}
				</div>
				<div class="space-y-2">
					{#each session.participants as p}
						<div class="flex items-center gap-3 rounded-lg bg-surface p-3 shadow-sm">
							<Avatar url={p.user.avatar_url} name={p.user.display_name} size="sm" />
							<span class="font-medium text-text-primary">{p.user.display_name}</span>
						</div>
					{/each}
				</div>
			{:else}
				{@const sorted = [...session.participants].sort((a, b) => (a.rank ?? 999) - (b.rank ?? 999))}
				<div class="space-y-2">
					{#each sorted as p, i}
						{@const isWinner = p.rank === 1 || p.result === 'win'}
						<div class="flex items-center gap-3 rounded-lg p-3 shadow-sm {isWinner ? 'bg-gold/10 ring-1 ring-gold/30' : 'bg-surface'}">
							<span class="w-8 text-center text-lg font-bold {isWinner ? 'text-gold' : 'text-text-secondary'}">
								{#if p.rank === 1}&#9733;{:else}{p.rank ?? '-'}{/if}
							</span>
							<Avatar url={p.user.avatar_url} name={p.user.display_name} size="sm" />
							<div class="min-w-0 flex-1">
								<span class="font-medium text-text-primary">{p.user.display_name}</span>
								{#if p.team_name}
									<span class="ml-2 text-sm text-text-secondary">({p.team_name})</span>
								{/if}
							</div>
							{#if p.score != null}
								<span class="text-sm font-medium text-text-secondary">{p.score} pts</span>
							{/if}
						</div>
					{/each}
				</div>
			{/if}
		</section>

		<!-- Admin actions -->
		{#if isAdmin}
			<div class="flex gap-3">
				<Button variant="secondary" onclick={() => goto(`/parties/${partyId}/sessions/${session!.id}/edit`)}>Edit</Button>
				<Button variant="danger" onclick={() => (showDeleteModal = true)}>Delete</Button>
			</div>
		{/if}
	</div>
{/if}

<!-- Delete confirm -->
<Modal bind:open={showDeleteModal} title="Delete session" onclose={() => (showDeleteModal = false)}>
	<p class="text-text-secondary">Are you sure? This will permanently delete this session and all participant records.</p>
	{#snippet actions()}
		<Button variant="secondary" onclick={() => (showDeleteModal = false)}>Cancel</Button>
		<Button variant="danger" loading={deleting} onclick={deleteSession}>Delete</Button>
	{/snippet}
</Modal>
