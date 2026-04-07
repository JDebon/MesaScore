<script lang="ts">
	import { getToasts, dismissToast } from '$stores/toast.svelte';

	const typeClasses = {
		success:
			'bg-success-50 text-success-600 border-success-500 dark:bg-success-600/20 dark:text-success-400 dark:border-success-500/50',
		error:
			'bg-danger-50 text-danger-600 border-danger-500 dark:bg-danger-600/20 dark:text-danger-400 dark:border-danger-500/50',
		info:
			'bg-primary-50 text-primary-600 border-primary-500 dark:bg-primary-600/20 dark:text-primary-400 dark:border-primary-500/50'
	};
</script>

<div class="fixed top-4 left-1/2 z-[100] flex -translate-x-1/2 flex-col gap-2">
	{#each getToasts() as toast (toast.id)}
		<button
			class="rounded-lg border-l-4 px-4 py-3 shadow-lg {typeClasses[toast.type]} min-w-[250px] text-left text-sm font-medium"
			onclick={() => dismissToast(toast.id)}
		>
			{toast.message}
		</button>
	{/each}
</div>
