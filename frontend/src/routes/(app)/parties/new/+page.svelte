<script lang="ts">
	import { goto } from '$app/navigation';
	import Button from '$components/ui/Button.svelte';
	import Input from '$components/ui/Input.svelte';
	import { partiesApi } from '$api/parties';
	import { ApiError } from '$api/client';

	let name = $state('');
	let description = $state('');
	let loading = $state(false);
	let error = $state('');

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		if (!name.trim()) return;

		loading = true;
		error = '';
		try {
			const res = await partiesApi.create({ name: name.trim(), description: description.trim() || null });
			goto(`/parties/${res.id}`);
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Failed to create party';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Create Party - MesaScore</title>
</svelte:head>

<div class="mx-auto max-w-lg">
	<h1 class="mb-6 text-2xl font-bold text-text-primary">Create a Party</h1>

	<form onsubmit={handleSubmit} class="rounded-xl bg-surface p-6 shadow-sm space-y-4">
		{#if error}
			<p class="rounded-lg bg-danger-50 p-3 text-sm text-danger-600 dark:bg-danger-600/20 dark:text-danger-400">{error}</p>
		{/if}

		<Input name="name" label="Party name" bind:value={name} required />

		<div class="space-y-1">
			<label for="description" class="block text-sm font-medium text-text-primary">Description</label>
			<textarea
				id="description"
				bind:value={description}
				rows={3}
				class="block w-full rounded-lg border border-border px-3 py-2 text-sm shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500"
				placeholder="What is this party about?"
			></textarea>
		</div>

		<Button type="submit" {loading} class="w-full">Create Party</Button>
	</form>
</div>
