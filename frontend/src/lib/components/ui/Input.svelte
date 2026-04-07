<script lang="ts">
	interface Props {
		type?: string;
		name?: string;
		label?: string;
		placeholder?: string;
		value?: string;
		error?: string;
		disabled?: boolean;
		required?: boolean;
		class?: string;
		oninput?: (e: Event & { currentTarget: HTMLInputElement }) => void;
		onblur?: (e: FocusEvent & { currentTarget: HTMLInputElement }) => void;
	}

	let {
		type = 'text',
		name = '',
		label = '',
		placeholder = '',
		value = $bindable(''),
		error = '',
		disabled = false,
		required = false,
		class: className = '',
		oninput,
		onblur
	}: Props = $props();
</script>

<div class="space-y-1 {className}">
	{#if label}
		<label for={name} class="block text-sm font-medium text-text-primary">
			{label}
			{#if required}<span class="text-danger-500">*</span>{/if}
		</label>
	{/if}
	<input
		{type}
		id={name}
		{name}
		{placeholder}
		bind:value
		{disabled}
		{required}
		{oninput}
		{onblur}
		class="block w-full rounded-lg border bg-surface px-3 py-2 text-sm text-text-primary shadow-sm
			placeholder:text-text-secondary transition-colors
			focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500
			disabled:bg-surface-raised disabled:text-text-secondary
			{error ? 'border-danger-500 focus:ring-danger-500' : 'border-border'}"
	/>
	{#if error}
		<p class="text-sm text-danger-500">{error}</p>
	{/if}
</div>
