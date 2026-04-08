<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		loading: boolean;
		error: boolean;
		onretry?: () => void;
		/** Optional skeleton snippet shown while loading. Falls back to a centered spinner. */
		skeleton?: Snippet;
		children: Snippet;
	}

	let { loading, error, onretry, skeleton, children }: Props = $props();
</script>

{#if loading}
	{#if skeleton}
		{@render skeleton()}
	{:else}
		<div class="flex justify-center py-20">
			<svg
				class="h-12 w-12 animate-spin text-primary-600"
				xmlns="http://www.w3.org/2000/svg"
				fill="none"
				viewBox="0 0 24 24"
			>
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
				<path
					class="opacity-75"
					fill="currentColor"
					d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
				></path>
			</svg>
		</div>
	{/if}
{:else if error}
	<div class="flex flex-col items-center gap-3 py-16 text-center">
		<svg
			class="h-10 w-10 text-text-disabled"
			fill="none"
			viewBox="0 0 24 24"
			stroke="currentColor"
			stroke-width="1.5"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z"
			/>
		</svg>
		<p class="text-sm text-text-secondary">Something went wrong.</p>
		{#if onretry}
			<button
				onclick={onretry}
				class="rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white hover:bg-primary-700 active:bg-primary-800"
			>
				Try again
			</button>
		{/if}
	</div>
{:else}
	{@render children()}
{/if}
