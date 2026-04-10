<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
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
	import type { AvailableGame, PartyMember, ParticipantInput } from '$api/types';

	interface Props {
		data: { party: import('$api/types').Party };
	}

	let { data: layoutData }: Props = $props();
	const user = $derived(getUser());
	const isAdmin = $derived(user?.id === layoutData.party.admin.id);

	// ── SessionStorage persistence ──────────────────────────────────────
	const storageKey = `mesascore_new_session_${layoutData.party.id}`;

	const savedState = browser
		? (() => {
				try {
					const s = sessionStorage.getItem(storageKey);
					return s ? JSON.parse(s) : null;
				} catch {
					return null;
				}
			})()
		: null;

	// ── Form state (restored from sessionStorage if available) ───────────
	let step = $state<number>(savedState?.step ?? 1);
	let loadingData = $state(true);
	let saving = $state(false);

	// Data loaded from API
	let availableGames = $state<AvailableGame[]>([]);
	let members = $state<PartyMember[]>([]);

	// Form fields
	let gameId = $state<string>(savedState?.gameId ?? '');
	let sessionType = $state<'competitive' | 'team' | 'cooperative' | 'score'>(
		savedState?.sessionType ?? 'competitive'
	);
	let playedAt = $state<string>(savedState?.playedAt ?? new Date().toISOString().split('T')[0]);
	let durationMinutes = $state<string>(savedState?.durationMinutes ?? '');
	let broughtByUserId = $state<string>(savedState?.broughtByUserId ?? '');
	let notes = $state<string>(savedState?.notes ?? '');
	let selectedMembers = $state<Set<string>>(new Set(savedState?.selectedMembers ?? []));
	let participants = $state<(ParticipantInput & { display_name: string })[]>(
		savedState?.participants ?? []
	);

	// Per-step errors shown inline
	let step2Error = $state('');
	let submitError = $state('');

	const selectedGame = $derived(availableGames.find((g) => g.id === gameId));

	// ── Persist state to sessionStorage on every meaningful change ───────
	$effect(() => {
		if (!browser || loadingData) return;
		try {
			sessionStorage.setItem(
				storageKey,
				JSON.stringify({
					step,
					gameId,
					sessionType,
					playedAt,
					durationMinutes,
					broughtByUserId,
					notes,
					selectedMembers: [...selectedMembers],
					participants
				})
			);
		} catch {
			// Ignore storage errors (private mode, quota exceeded)
		}
	});

	onMount(async () => {
		try {
			const [g, m] = await Promise.all([
				gamesApi.availableForParty(layoutData.party.id),
				partiesApi.members(layoutData.party.id)
			]);
			availableGames = g;
			members = m.members;
		} catch (e) {
			console.error('[sessions/new] Failed to load data:', e);
			addToast('Failed to load data', 'error');
		} finally {
			loadingData = false;
		}
	});

	function toggleMember(id: string) {
		const next = new Set(selectedMembers);
		if (next.has(id)) next.delete(id);
		else next.add(id);
		selectedMembers = next;
		step2Error = '';
	}

	function goToStep3() {
		if (selectedMembers.size < 2) {
			step2Error = 'Select at least 2 participants.';
			return;
		}
		step2Error = '';
		participants = [...selectedMembers].map((id, i) => {
			const m = members.find((m) => m.id === id)!;
			return {
				user_id: id,
				display_name: m.display_name,
				team_name: null,
				rank: sessionType === 'cooperative' ? null : i + 1,
				score: null,
				result: null
			};
		});
		step = 3;
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

	async function submit() {
		saving = true;
		submitError = '';
		try {
			const res = await sessionsApi.create(layoutData.party.id, {
				game_id: gameId,
				session_type: sessionType,
				played_at: playedAt + 'T00:00:00Z',
				duration_minutes: durationMinutes ? parseInt(durationMinutes) : null,
				notes: notes.trim() || null,
				brought_by_user_id: broughtByUserId || null,
				participants: participants.map(({ user_id, team_name, rank, score, result }) => ({
					user_id,
					team_name,
					rank,
					score,
					result
				}))
			});
			// Clear persisted draft on success
			try { sessionStorage.removeItem(storageKey); } catch { /* ignore */ }
			addToast('Session logged!', 'success');
			goto(`/parties/${layoutData.party.id}/sessions/${res.id}`);
		} catch (e) {
			if (e instanceof ApiError && e.fields) {
				submitError = Object.values(e.fields).join(' ');
			} else {
				submitError = e instanceof ApiError ? e.message : 'Failed to log session. Please try again.';
			}
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>Log Session - {layoutData.party.name} - MesaScore</title>
</svelte:head>

{#if loadingData}
	<div class="flex justify-center py-20"><Spinner size="lg" /></div>
{:else}
	<div class="mx-auto max-w-lg">
		<!-- Step indicator -->
		<div class="mb-6 flex items-center gap-2">
			{#each [1, 2, 3, 4] as s}
				<div class="flex-1 h-1 rounded-full {s <= step ? 'bg-primary-600' : 'bg-gray-200'}"></div>
			{/each}
		</div>

		{#if step === 1}
			<!-- Step 1: Game & basics -->
			<h2 class="mb-4 text-xl font-bold text-text-primary">Game & Details</h2>
			<div class="space-y-4 rounded-xl bg-surface p-6 shadow-sm">
				<div class="space-y-1">
					<label for="game" class="block text-sm font-medium text-text-primary">Game <span class="text-danger-500">*</span></label>
					<select id="game" bind:value={gameId} class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text-primary" required>
						<option value="">Select a game</option>
						{#each availableGames as game}
							<option value={game.id}>{game.name}</option>
						{/each}
					</select>
				</div>

				<div class="space-y-1">
					<label class="block text-sm font-medium text-text-primary">Session type <span class="text-danger-500">*</span></label>
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

				{#if selectedGame && selectedGame.owners.length > 0}
					<div class="space-y-1">
						<label for="brought_by" class="block text-sm font-medium text-text-primary">Brought by</label>
						<select id="brought_by" bind:value={broughtByUserId} class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text-primary">
							<option value="">None</option>
							{#each selectedGame.owners as owner}
								<option value={owner.id}>{owner.display_name}</option>
							{/each}
						</select>
					</div>
				{/if}

				<div class="space-y-1">
					<label for="notes" class="block text-sm font-medium text-text-primary">Notes</label>
					<textarea id="notes" bind:value={notes} rows={2} class="block w-full rounded-lg border border-border px-3 py-2 text-sm shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500"></textarea>
				</div>

				<Button class="w-full" disabled={!gameId} onclick={() => (step = 2)}>Next</Button>
			</div>

		{:else if step === 2}
			<!-- Step 2: Participants -->
			<h2 class="mb-4 text-xl font-bold text-text-primary">Select Participants</h2>
			<div class="rounded-xl bg-surface p-6 shadow-sm">
				<p class="mb-4 text-sm text-text-secondary">Select at least 2 players.</p>
				<div class="space-y-2">
					{#each members as member}
						<button
							class="flex w-full items-center gap-3 rounded-lg border p-3 text-left transition-colors
								{selectedMembers.has(member.id) ? 'border-primary-500 bg-primary-50' : 'border-border hover:bg-bg'}"
							onclick={() => toggleMember(member.id)}
						>
							<Avatar url={member.avatar_url} name={member.display_name} size="sm" />
							<span class="font-medium text-text-primary">{member.display_name}</span>
							{#if selectedMembers.has(member.id)}
								<svg class="ml-auto h-5 w-5 text-primary-600" fill="currentColor" viewBox="0 0 20 20">
									<path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"/>
								</svg>
							{/if}
						</button>
					{/each}
				</div>

				{#if step2Error}
					<p class="mt-3 text-sm text-danger-600">{step2Error}</p>
				{/if}

				<div class="mt-4 flex gap-3">
					<Button variant="secondary" onclick={() => (step = 1)}>Back</Button>
					<Button class="flex-1" onclick={goToStep3}>Next ({selectedMembers.size} selected)</Button>
				</div>
			</div>

		{:else if step === 3}
			<!-- Step 3: Results -->
			<h2 class="mb-4 text-xl font-bold text-text-primary">Results</h2>
			<div class="rounded-xl bg-surface p-6 shadow-sm space-y-4">
				{#if sessionType === 'cooperative'}
					<p class="text-sm text-text-secondary">Did the group win or lose?</p>
					<div class="flex gap-3">
						<Button
							variant={participants[0]?.result === 'win' ? 'primary' : 'secondary'}
							class="flex-1"
							onclick={() => setCoopResult('win')}
						>Win</Button>
						<Button
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
								<input
									type="text"
									placeholder="Team"
									value={p.team_name ?? ''}
									oninput={(e) => updateTeam(i, (e.target as HTMLInputElement).value)}
									class="w-20 rounded border border-border px-2 py-1 text-sm"
								/>
							{/if}
							<input
								type="number"
								placeholder="Rank"
								value={p.rank ?? ''}
								oninput={(e) => updateRank(i, (e.target as HTMLInputElement).value)}
								class="w-16 rounded border border-border px-2 py-1 text-sm text-center"
								min="1"
							/>
							{#if sessionType === 'score' || sessionType === 'competitive'}
								<input
									type="number"
									placeholder="Score"
									value={p.score ?? ''}
									oninput={(e) => updateScore(i, (e.target as HTMLInputElement).value)}
									class="w-20 rounded border border-border px-2 py-1 text-sm text-center"
								/>
							{/if}
						</div>
					{/each}
				{/if}

				<div class="flex gap-3">
					<Button variant="secondary" onclick={() => (step = 2)}>Back</Button>
					<Button class="flex-1" onclick={() => (step = 4)}>Review</Button>
				</div>
			</div>

		{:else}
			<!-- Step 4: Review & submit -->
			<h2 class="mb-4 text-xl font-bold text-text-primary">Review</h2>
			<div class="rounded-xl bg-surface p-6 shadow-sm space-y-4">
				<div class="text-sm text-text-secondary space-y-1">
					<p><strong>Game:</strong> {selectedGame?.name}</p>
					<p><strong>Type:</strong> {sessionType}</p>
					<p><strong>Date:</strong> {playedAt}</p>
					{#if durationMinutes}<p><strong>Duration:</strong> {durationMinutes} min</p>{/if}
					{#if notes}<p><strong>Notes:</strong> {notes}</p>{/if}
				</div>

				<div>
					<p class="mb-2 font-medium text-text-primary">Participants ({participants.length})</p>
					{#each participants as p}
						<div class="flex items-center gap-2 py-1 text-sm">
							<span class="w-8 text-center font-medium text-text-secondary">{p.rank ?? '-'}</span>
							<span class="text-text-primary">{p.display_name}</span>
							{#if p.team_name}<span class="text-text-secondary">({p.team_name})</span>{/if}
							{#if p.score != null}<span class="text-text-secondary">{p.score} pts</span>{/if}
							{#if p.result}<span class="text-text-secondary">{p.result}</span>{/if}
						</div>
					{/each}
				</div>

				{#if submitError}
					<div class="rounded-lg bg-danger-50 px-3 py-2 text-sm text-danger-600 dark:bg-danger-600/20 dark:text-danger-400">
						{submitError}
					</div>
				{/if}

				<div class="flex gap-3">
					<Button variant="secondary" onclick={() => (step = 3)}>Back</Button>
					<Button class="flex-1" loading={saving} onclick={submit}>Save Session</Button>
				</div>
			</div>
		{/if}
	</div>
{/if}
