<script lang="ts">
	import { goto } from '$app/navigation';
	import Avatar from '$components/ui/Avatar.svelte';
	import ThemeToggle from '$components/ui/ThemeToggle.svelte';
	import { getUser, clearAuth } from '$stores/auth.svelte';

	let menuOpen = $state(false);
	const user = $derived(getUser());

	function logout() {
		clearAuth();
		goto('/login');
	}
</script>

<header class="sticky top-0 z-40 border-b border-border bg-surface">
	<div class="mx-auto flex h-14 max-w-7xl items-center justify-between px-4">
		<a href="/" class="flex items-center gap-2 font-bold text-primary-600">
			<svg class="h-7 w-7" viewBox="0 0 24 24" fill="currentColor">
				<path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
			</svg>
			MesaScore
		</a>

		<div class="flex items-center gap-3">
			<a href="/games" class="text-sm font-medium text-text-secondary hover:text-text-primary">My Collection</a>

			<ThemeToggle />

			<div class="relative">
				<button
					onclick={() => (menuOpen = !menuOpen)}
					class="flex items-center gap-2 rounded-lg p-1 hover:bg-surface-raised"
					aria-label="User menu"
				>
					<Avatar url={user?.avatar_url} name={user?.display_name ?? '?'} size="sm" />
				</button>

				{#if menuOpen}
					<!-- svelte-ignore a11y_no_static_element_interactions -->
					<div class="fixed inset-0 z-10" onclick={() => (menuOpen = false)} onkeydown={() => {}}></div>
					<div class="absolute right-0 z-20 mt-2 w-56 rounded-lg border border-border bg-surface py-1 shadow-lg">
						<a
							href="/users/{user?.id}"
							class="block px-4 py-2 text-sm text-text-primary hover:bg-surface-raised"
							onclick={() => (menuOpen = false)}
						>
							Profile
						</a>
						<button
							onclick={logout}
							class="block w-full px-4 py-2 text-left text-sm text-text-primary hover:bg-surface-raised"
						>
							Log out
						</button>
					</div>
				{/if}
			</div>
		</div>
	</div>
</header>
