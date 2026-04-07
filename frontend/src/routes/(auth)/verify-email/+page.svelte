<script lang="ts">
	import { page } from '$app/state';
	import { authApi } from '$api/auth';
	import { ApiError } from '$api/client';
	import Spinner from '$components/ui/Spinner.svelte';
	import { onMount } from 'svelte';

	let status = $state<'loading' | 'success' | 'error'>('loading');
	let errorMessage = $state('');

	onMount(async () => {
		const token = page.url.searchParams.get('token');
		if (!token) {
			status = 'error';
			errorMessage = 'No verification token provided.';
			return;
		}
		try {
			await authApi.verifyEmail(token);
			status = 'success';
		} catch (e) {
			status = 'error';
			errorMessage =
				e instanceof ApiError ? e.message : 'This link is invalid or has expired.';
		}
	});
</script>

<div class="rounded-xl bg-surface p-8 shadow-sm text-center">
	{#if status === 'loading'}
		<Spinner size="lg" class="mx-auto" />
		<p class="mt-4 text-text-secondary">Verifying your email...</p>
	{:else if status === 'success'}
		<svg class="mx-auto h-16 w-16 text-success-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
			<path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
		</svg>
		<h2 class="mt-4 text-xl font-semibold text-text-primary">Email verified!</h2>
		<p class="mt-2 text-text-secondary">You can now log in to your account.</p>
		<a href="/login" class="mt-4 inline-block text-sm font-medium text-primary-600 hover:text-primary-700">Go to login</a>
	{:else}
		<svg class="mx-auto h-16 w-16 text-danger-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
			<path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
		</svg>
		<h2 class="mt-4 text-xl font-semibold text-text-primary">Verification failed</h2>
		<p class="mt-2 text-text-secondary">{errorMessage}</p>
		<a href="/login" class="mt-4 inline-block text-sm font-medium text-primary-600 hover:text-primary-700">Go to login</a>
	{/if}
</div>
