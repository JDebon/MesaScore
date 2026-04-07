<script lang="ts">
	import { getThemePreference, setThemePreference } from '$stores/theme.svelte';
	import type { ThemePreference } from '$stores/theme.svelte';

	const options: { value: ThemePreference; label: string }[] = [
		{ value: 'light', label: 'Light' },
		{ value: 'system', label: 'System' },
		{ value: 'dark', label: 'Dark' }
	];

	const current = $derived(getThemePreference());
</script>

<div
	role="group"
	aria-label="Theme preference"
	class="flex overflow-hidden rounded-lg border border-border"
>
	{#each options as opt}
		<button
			type="button"
			aria-label="{opt.label} theme"
			aria-pressed={current === opt.value}
			class="px-2.5 py-1.5 text-xs font-medium transition-colors
				{current === opt.value
				? 'bg-primary-600 text-white'
				: 'bg-surface text-text-secondary hover:bg-surface-raised'}"
			onclick={() => setThemePreference(opt.value)}
		>
			{#if opt.value === 'light'}
				<!-- Sun icon -->
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
					<circle cx="12" cy="12" r="5"/>
					<path stroke-linecap="round" d="M12 2v2M12 20v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M2 12h2M20 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>
				</svg>
			{:else if opt.value === 'system'}
				<!-- Monitor icon -->
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
					<rect x="2" y="3" width="20" height="14" rx="2"/>
					<path stroke-linecap="round" d="M8 21h8M12 17v4"/>
				</svg>
			{:else}
				<!-- Moon icon -->
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
					<path stroke-linecap="round" d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z"/>
				</svg>
			{/if}
			<span class="sr-only">{opt.label}</span>
		</button>
	{/each}
</div>
