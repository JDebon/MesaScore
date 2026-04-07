<script lang="ts">
	import TopBar from '$components/layout/TopBar.svelte';

	let { children } = $props();
</script>

<div class="min-h-screen bg-bg">
	<TopBar />
	<main class="mx-auto max-w-7xl px-4 py-6">
		{@render children()}
	</main>
</div>
