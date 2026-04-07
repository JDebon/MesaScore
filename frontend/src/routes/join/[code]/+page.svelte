<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import Button from '$components/ui/Button.svelte';
	import Spinner from '$components/ui/Spinner.svelte';
	import { partiesApi } from '$api/parties';
	import { ApiError } from '$api/client';
	import { isAuthenticated } from '$stores/auth.svelte';
	import { addToast } from '$stores/toast.svelte';

	let partyName = $state('');
	let memberCount = $state(0);
	let loading = $state(true);
	let joining = $state(false);
	let errorMsg = $state('');
	let authenticated = $state(false);

	const code = $derived(page.params.code!);

	onMount(async () => {
		authenticated = isAuthenticated();
		try {
			const res = await partiesApi.joinPreview(code);
			partyName = res.party.name;
			memberCount = res.party.member_count;
		} catch (e) {
			if (e instanceof ApiError && e.status === 404) {
				errorMsg = 'This invite link is no longer valid.';
			} else {
				errorMsg = 'Something went wrong.';
			}
		} finally {
			loading = false;
		}
	});

	async function handleJoin() {
		joining = true;
		try {
			const res = await partiesApi.join(code);
			addToast('Joined party!', 'success');
			goto(`/parties/${res.party_id}`);
		} catch (e) {
			if (e instanceof ApiError) {
				if (e.status === 409) {
					addToast("You're already in this party", 'info');
					goto('/');
				} else {
					errorMsg = e.message;
				}
			}
		} finally {
			joining = false;
		}
	}

	function saveAndRedirect(path: string) {
		if (browser) localStorage.setItem('mesascore_pending_join', code);
		goto(path);
	}
</script>

<div class="flex min-h-screen items-center justify-center bg-bg px-4 py-12">
	<div class="w-full max-w-md">
		<div class="mb-8 text-center">
			<svg class="mx-auto h-12 w-12 text-primary-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
				<path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
			</svg>
			<h1 class="mt-2 text-2xl font-bold text-text-primary">MesaScore</h1>
		</div>

		<div class="rounded-xl bg-surface p-8 shadow-sm text-center">
			{#if loading}
				<Spinner size="lg" class="mx-auto" />
			{:else if errorMsg}
				<p class="text-danger-600">{errorMsg}</p>
				<a href="/login" class="mt-4 inline-block text-sm font-medium text-primary-600">Go to login</a>
			{:else if !authenticated}
				<h2 class="text-xl font-semibold text-text-primary">Join {partyName}</h2>
				<p class="mt-2 text-text-secondary">{memberCount} member{memberCount === 1 ? '' : 's'}</p>
				<p class="mt-4 text-sm text-text-secondary">Log in or register to join this party.</p>
				<div class="mt-6 flex gap-3 justify-center">
					<Button onclick={() => saveAndRedirect('/login')}>Log in</Button>
					<Button variant="secondary" onclick={() => saveAndRedirect('/register')}>Register</Button>
				</div>
			{:else}
				<h2 class="text-xl font-semibold text-text-primary">Join {partyName}?</h2>
				<p class="mt-2 text-text-secondary">{memberCount} member{memberCount === 1 ? '' : 's'}</p>
				<div class="mt-6">
					<Button onclick={handleJoin} loading={joining}>Join Party</Button>
				</div>
			{/if}
		</div>
	</div>
</div>
