<script lang="ts">
	import { goto } from '$app/navigation';
	import Button from '$components/ui/Button.svelte';
	import Input from '$components/ui/Input.svelte';
	import { authApi } from '$api/auth';
	import { ApiError } from '$api/client';
	import { setAuth } from '$stores/auth.svelte';
	import { addToast } from '$stores/toast.svelte';
	import { browser } from '$app/environment';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);
	let showResend = $state(false);
	let resendEmail = $state('');
	let resending = $state(false);

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		loading = true;
		error = '';
		showResend = false;

		try {
			const res = await authApi.login(email, password);
			setAuth(res.token, res.user);

			// Check for pending join
			if (browser) {
				const pendingJoin = localStorage.getItem('mesascore_pending_join');
				if (pendingJoin) {
					localStorage.removeItem('mesascore_pending_join');
					goto(`/join/${pendingJoin}`);
					return;
				}
			}

			// Read redirect target from sessionStorage (set by the auth guard in +layout.ts).
			// This avoids exposing the destination path in the URL.
			const redirectTo = browser
				? (sessionStorage.getItem('mesascore_redirect') ?? '/')
				: '/';
			if (redirectTo !== '/') sessionStorage.removeItem('mesascore_redirect');
			goto(redirectTo);
		} catch (e) {
			if (e instanceof ApiError) {
				if (e.status === 403 && e.message === 'email_not_verified') {
					error = 'Your email is not verified.';
					showResend = true;
					resendEmail = email;
				} else {
					error = 'Invalid email or password';
				}
			}
		} finally {
			loading = false;
		}
	}

	async function handleResend() {
		resending = true;
		try {
			await authApi.resendVerification(resendEmail);
			addToast('If an account exists, we sent a new verification email.', 'info');
			showResend = false;
		} catch {
			// Always show success message (no info leak)
			addToast('If an account exists, we sent a new verification email.', 'info');
		} finally {
			resending = false;
		}
	}
</script>

<form onsubmit={handleSubmit} class="rounded-xl bg-surface p-8 shadow-sm space-y-4">
	<h2 class="text-xl font-semibold text-text-primary">Log in</h2>

	{#if error}
		<div class="rounded-lg bg-danger-50 p-3 text-sm text-danger-600 dark:bg-danger-600/20 dark:text-danger-400">
			{error}
			{#if showResend}
				<button
					type="button"
					onclick={handleResend}
					disabled={resending}
					class="ml-1 font-medium underline hover:no-underline"
				>
					{resending ? 'Sending...' : 'Resend verification email'}
				</button>
			{/if}
		</div>
	{/if}

	<Input name="email" label="Email" type="email" bind:value={email} required />
	<Input name="password" label="Password" type="password" bind:value={password} required />

	<Button type="submit" {loading} class="w-full">Log in</Button>

	<p class="text-center text-sm text-text-secondary">
		Don't have an account? <a href="/register" class="font-medium text-primary-600 hover:text-primary-700">Register</a>
	</p>
</form>
