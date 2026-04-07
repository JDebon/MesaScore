<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		variant?: 'default' | 'success' | 'warning' | 'danger' | 'info';
		children: Snippet;
	}

	let { variant = 'default', children }: Props = $props();

	const variants = {
		default: 'bg-surface-raised text-text-secondary',
		success:
			'bg-success-50 text-success-600 dark:bg-success-600/20 dark:text-success-400',
		warning:
			'bg-warning-50 text-warning-500 dark:bg-warning-500/20 dark:text-warning-400',
		danger:
			'bg-danger-50 text-danger-600 dark:bg-danger-600/20 dark:text-danger-400',
		info:
			'bg-primary-50 text-primary-600 dark:bg-primary-600/20 dark:text-primary-400'
	};
</script>

<span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {variants[variant]}">
	{@render children()}
</span>
