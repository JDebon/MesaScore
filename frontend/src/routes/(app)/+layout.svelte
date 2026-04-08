<script lang="ts">
	import AppSidebar from '$components/layout/AppSidebar.svelte';

	let { children } = $props();
</script>

<div class="min-h-screen bg-bg">
	<AppSidebar />
	<main class="pt-14 lg:pl-60 lg:pt-0">
		<div class="mx-auto max-w-7xl px-4 py-6">
			{@render children()}
		</div>
	</main>
</div>
