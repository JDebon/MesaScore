<script lang="ts">
	interface Props {
		url?: string | null;
		name?: string;
		size?: 'sm' | 'md' | 'lg' | 'xl';
		class?: string;
	}

	let { url = null, name = '?', size = 'md', class: className = '' }: Props = $props();

	const sizes = {
		sm: 'h-8 w-8 text-xs',
		md: 'h-10 w-10 text-sm',
		lg: 'h-14 w-14 text-lg',
		xl: 'h-20 w-20 text-2xl'
	};

	const initials = $derived(
		name
			.split(' ')
			.map((n) => n[0])
			.join('')
			.toUpperCase()
			.slice(0, 2)
	);
</script>

{#if url}
	<img src={url} alt={name} class="rounded-full object-cover {sizes[size]} {className}" />
{:else}
	<div
		class="flex items-center justify-center rounded-full bg-primary-100 text-primary-700
			dark:bg-primary-900/40 dark:text-primary-300 font-medium {sizes[size]} {className}"
	>
		{initials}
	</div>
{/if}
