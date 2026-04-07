<script lang="ts">
	import Button from '$components/ui/Button.svelte';
	import Input from '$components/ui/Input.svelte';
	import { authApi } from '$api/auth';
	import { ApiError } from '$api/client';
	import { validateEmail, validateUsername, validatePassword } from '$lib/validate';

	let username = $state('');
	let displayName = $state('');
	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let errors = $state<Record<string, string>>({});
	let loading = $state(false);
	let success = $state(false);

	let usernameAvailable = $state<boolean | null>(null);
	let checkingUsername = $state(false);
	let usernameTimer: ReturnType<typeof setTimeout>;

	function checkUsername() {
		usernameAvailable = null;
		clearTimeout(usernameTimer);
		if (username.length < 2) return;
		checkingUsername = true;
		usernameTimer = setTimeout(async () => {
			try {
				const res = await authApi.checkUsername(username);
				usernameAvailable = res.available;
			} catch (e) {
				console.error('[register] Username check failed:', e);
				usernameAvailable = null;
			} finally {
				checkingUsername = false;
			}
		}, 300);
	}

	function validate(): boolean {
		const e: Record<string, string> = {};
		const usernameErr = validateUsername(username);
		if (usernameErr) e.username = usernameErr;
		if (usernameAvailable === false) e.username = 'Username is taken';
		if (!displayName.trim()) e.display_name = 'Required';
		const emailErr = validateEmail(email);
		if (emailErr) e.email = emailErr;
		const passwordErr = validatePassword(password);
		if (passwordErr) e.password = passwordErr;
		if (password !== confirmPassword) e.confirm_password = 'Passwords do not match';
		errors = e;
		return Object.keys(e).length === 0;
	}

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		if (!validate()) return;

		loading = true;
		errors = {};
		try {
			await authApi.register({ username, display_name: displayName, email, password });
			success = true;
		} catch (e) {
			if (e instanceof ApiError) {
				if (e.fields) {
					errors = e.fields;
				} else if (e.status === 409) {
					errors = { email: e.message };
				} else {
					errors = { form: e.message };
				}
			}
		} finally {
			loading = false;
		}
	}
</script>

{#if success}
	<div class="rounded-xl bg-surface p-8 shadow-sm text-center">
		<svg class="mx-auto h-16 w-16 text-success-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
			<path stroke-linecap="round" stroke-linejoin="round" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
		</svg>
		<h2 class="mt-4 text-xl font-semibold text-text-primary">Check your email</h2>
		<p class="mt-2 text-text-secondary">We sent a verification link to <strong>{email}</strong>. Click it to activate your account.</p>
		<a href="/login" class="mt-6 inline-block text-sm font-medium text-primary-600 hover:text-primary-700">Go to login</a>
	</div>
{:else}
	<form onsubmit={handleSubmit} class="rounded-xl bg-surface p-8 shadow-sm space-y-4">
		<h2 class="text-xl font-semibold text-text-primary">Create your account</h2>

		{#if errors.form}
			<p class="rounded-lg bg-danger-50 p-3 text-sm text-danger-600 dark:bg-danger-600/20 dark:text-danger-400">{errors.form}</p>
		{/if}

		<div>
			<Input
				name="username"
				label="Username"
				bind:value={username}
				error={errors.username}
				required
				oninput={checkUsername}
			/>
			{#if checkingUsername}
				<p class="mt-1 text-xs text-text-secondary">Checking...</p>
			{:else if usernameAvailable === true}
				<p class="mt-1 text-xs text-success-600">Available</p>
			{:else if usernameAvailable === false}
				<p class="mt-1 text-xs text-danger-500">Taken</p>
			{/if}
		</div>

		<Input name="display_name" label="Display name" bind:value={displayName} error={errors.display_name} required />
		<Input name="email" label="Email" type="email" bind:value={email} error={errors.email} required />
		<Input name="password" label="Password" type="password" bind:value={password} error={errors.password} required />
		<Input
			name="confirm_password"
			label="Confirm password"
			type="password"
			bind:value={confirmPassword}
			error={errors.confirm_password}
			required
		/>

		<Button type="submit" {loading} class="w-full">Register</Button>

		<p class="text-center text-sm text-text-secondary">
			Already have an account? <a href="/login" class="font-medium text-primary-600 hover:text-primary-700">Log in</a>
		</p>
	</form>
{/if}
