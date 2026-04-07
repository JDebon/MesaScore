<script lang="ts">
	import ThemeToggle from '$components/ui/ThemeToggle.svelte';

	let { children } = $props();
</script>

<div class="relative flex min-h-screen items-center justify-center bg-bg px-4 py-12">
	<div class="absolute right-4 top-4">
		<ThemeToggle />
	</div>
	<div class="w-full max-w-md">
		<div class="mb-8 text-center">
			<svg class="mx-auto h-12 w-12 text-primary-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
				<path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
			</svg>
			<h1 class="mt-2 text-2xl font-bold text-text-primary">MesaScore</h1>
		</div>
		{@render children()}
	</div>
</div>
