import { redirect } from '@sveltejs/kit';
import { isAuthenticated } from '$stores/auth.svelte';
import { browser } from '$app/environment';

export function load({ url }) {
	if (browser && !isAuthenticated()) {
		// Store the intended destination in sessionStorage instead of the URL
		// to avoid exposing sensitive paths (party IDs, invite codes, etc.) in
		// browser history, server logs, and referrer headers.
		const redirectTo = url.pathname + url.search;
		if (redirectTo !== '/') {
			sessionStorage.setItem('mesascore_redirect', redirectTo);
		}
		throw redirect(302, '/login');
	}
}
