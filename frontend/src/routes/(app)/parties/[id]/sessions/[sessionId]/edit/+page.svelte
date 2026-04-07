<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import Button from '$components/ui/Button.svelte';
	import Input from '$components/ui/Input.svelte';
	import Avatar from '$components/ui/Avatar.svelte';
	import Spinner from '$components/ui/Spinner.svelte';
	import { gamesApi } from '$api/games';
	import { partiesApi } from '$api/parties';
	import { sessionsApi } from '$api/sessions';
	import { addToast } from '$stores/toast.svelte';
	import { getUser } from '$stores/auth.svelte';
	import { ApiError } from '$api/client';
	import type { AvailableGame, PartyMember, SessionDetail, ParticipantInput } from '$api/types';

	interface Props {
		data: { party: import('$api/types').Party };
	}

	let { data: layoutData }: Props = $props();
	const user = $derived(getUser());
	const isAdmin = $derived(user?.id === layoutData.party.admin.id);

	let loadingData = $state(true);
	let saving = $state(false);
	let availableGames = $state<AvailableGame[]>([]);
	let members = $state<PartyMember[]>([]);

	let gameId = $state('');
	let sessionType = $state<'competitive' | 'team' | 'cooperative' | 'score'>('competitive');
	let playedAt = $state('');
	let durationMinutes = $state('');
	let broughtByUserId = $state('');
	let notes = $state('');
	let participants = $state<(ParticipantInput & { display_name: string })[]>([]);

	const selectedGame = $derived(availableGames.find((g) => g.id === gameId));
	const partyId = $derived(layoutData.party.id);
	const sessionId = $derived(page.params.sessionId!);

	onMount(async () => {
		if (!isAdmin) {
			goto(`/parties/${partyId}`);
			return;
		}
		try {
			const [g, m, s] = await Promise.all([
				gamesApi.availableForParty(partyId),
				partiesApi.members(partyId),
				sessionsApi.get(partyId, sessionId)
			]);
			availableGames = g;
			members = m.members;
			populateForm(s);
		} catch (e) {
			console.error('[sessions/edit] Failed to load:', e);
			addToast('Failed to load session', 'error');
		} finally {
			loadingData = false;
		}
	});

	function populateForm(s: SessionDetail) {
		gameId = s.game.id;
		sessionType = s.session_type;
		playedAt = s.played_at.split('T')[0];
		durationMinutes = s.duration_minutes?.toString() ?? '';
		broughtByUserId = s.brought_by?.id ?? '';
		notes = s.notes ?? '';
		participants = s.participants.map((p) => ({
			user_id: p.user.id,
			display_name: p.user.display_name,
			team_name: p.team_name,
			rank: p.rank,
			score: p.score,
			result: p.result
		}));
	}

	function updateRank(idx: number, val: string) {
		participants[idx].rank = val ? parseInt(val) : null;
	}
	function updateScore(idx: number, val: string) {
		participants[idx].score = val ? parseFloat(val) : null;
	}
	function updateTeam(idx: number, val: string) {
		participants[idx].team_name = val || null;
	}
	function setCoopResult(result: 'win' | 'loss') {
		participants = participants.map((p) => ({ ...p, result }));
	}

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		saving = true;
		try {
			await sessionsApi.update(partyId, sessionId, {
				game_id: gameId,
				session_type: sessionType,
				played_at: playedAt + 'T00:00:00Z',
				duration_minutes: durationMinutes ? parseInt(durationMinutes) : null,
				notes: notes.trim() || null,
				brought_by_user_id: broughtByUserId || null,
				participants: participants.map(({ user_id, team_name, rank, score, result }) => ({
					user_id, team_name, rank, score, result
				}))
			});
			addToast('Session updated!', 'success');
			goto(`/parties/${partyId}/sessions/${sessionId}`);
		} catch (e) {
			if (e instanceof ApiError && e.fields) {
				addToast(Object.values(e.fields).join(', '), 'error');
			} else {
				addToast(e instanceof ApiError ? e.message : 'Failed to update', 'error');
			}
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>Edit Session - MesaScore</title>
</svelte:head>

{#if loadingData}
	<div class="flex justify-center py-20"><Spinner size="lg" /></div>
{:else}
	<div class="mx-auto max-w-lg">
		<h1 class="mb-6 text-2xl font-bold text-text-primary">Edit Session</h1>

		<form onsubmit={submit} class="space-y-6">
			<!-- Game & basics -->
			<div class="rounded-xl bg-surface p-6 shadow-sm space-y-4">
				<div class="space-y-1">
					<label for="game" class="block text-sm font-medium text-text-primary">Game</label>
					<select id="game" bind:value={gameId} class="w-full rounded-lg border border-border px-3 py-2 text-sm">
						{#each availableGames as game}
							<option value={game.id}>{game.name}</option>
						{/each}
					</select>
				</div>

				<div class="space-y-1">
					<label class="block text-sm font-medium text-text-primary">Session type</label>
					<div class="grid grid-cols-2 gap-2">
						{#each ['competitive', 'team', 'cooperative', 'score'] as t}
							<button
								type="button"
								class="rounded-lg border px-3 py-2 text-sm font-medium capitalize transition-colors
									{sessionType === t ? 'border-primary-500 bg-primary-50 text-primary-700' : 'border-border text-text-secondary hover:bg-bg'}"
								onclick={() => (sessionType = t as typeof sessionType)}
							>{t}</button>
						{/each}
					</div>
				</div>

				<Input name="played_at" label="Date played" type="date" bind:value={playedAt} required />
				<Input name="duration" label="Duration (minutes)" type="number" bind:value={durationMinutes} />
				<div class="space-y-1">
					<label for="notes" class="block text-sm font-medium text-text-primary">Notes</label>
					<textarea id="notes" bind:value={notes} rows={2} class="block w-full rounded-lg border border-border px-3 py-2 text-sm shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500"></textarea>
				</div>
			</div>

			<!-- Results -->
			<div class="rounded-xl bg-surface p-6 shadow-sm space-y-4">
				<h3 class="font-semibold text-text-primary">Results</h3>
				{#if sessionType === 'cooperative'}
					<div class="flex gap-3">
						<Button
							type="button"
							variant={participants[0]?.result === 'win' ? 'primary' : 'secondary'}
							class="flex-1"
							onclick={() => setCoopResult('win')}
						>Win</Button>
						<Button
							type="button"
							variant={participants[0]?.result === 'loss' ? 'danger' : 'secondary'}
							class="flex-1"
							onclick={() => setCoopResult('loss')}
						>Loss</Button>
					</div>
				{:else}
					{#each participants as p, i}
						<div class="flex items-center gap-3 rounded-lg border border-border p-3">
							<Avatar name={p.display_name} size="sm" />
							<span class="min-w-0 flex-1 truncate font-medium text-text-primary">{p.display_name}</span>
							{#if sessionType === 'team'}
								<input type="text" placeholder="Team" value={p.team_name ?? ''} oninput={(e) => updateTeam(i, (e.target as HTMLInputElement).value)} class="w-20 rounded border border-border px-2 py-1 text-sm" />
							{/if}
							<input type="number" placeholder="Rank" value={p.rank ?? ''} oninput={(e) => updateRank(i, (e.target as HTMLInputElement).value)} class="w-16 rounded border border-border px-2 py-1 text-sm text-center" min="1" />
							{#if sessionType === 'score' || sessionType === 'competitive'}
								<input type="number" placeholder="Score" value={p.score ?? ''} oninput={(e) => updateScore(i, (e.target as HTMLInputElement).value)} class="w-20 rounded border border-border px-2 py-1 text-sm text-center" />
							{/if}
						</div>
					{/each}
				{/if}
			</div>

			<div class="flex gap-3">
				<Button variant="secondary" onclick={() => goto(`/parties/${partyId}/sessions/${sessionId}`)}>Cancel</Button>
				<Button type="submit" class="flex-1" loading={saving}>Save Changes</Button>
			</div>
		</form>
	</div>
{/if}
