import { browser } from '$app/environment';

export type ThemePreference = 'light' | 'dark' | 'system';

const STORAGE_KEY = 'mesascore_theme';

let preference = $state<ThemePreference>('system');
let resolvedDark = $state(false);

function getSystemDark(): boolean {
	return browser && window.matchMedia('(prefers-color-scheme: dark)').matches;
}

function applyTheme(pref: ThemePreference) {
	const isDark = pref === 'dark' || (pref === 'system' && getSystemDark());
	resolvedDark = isDark;
	if (browser) {
		document.documentElement.classList.toggle('dark', isDark);
	}
}

if (browser) {
	const stored = localStorage.getItem(STORAGE_KEY) as ThemePreference | null;
	preference = stored ?? 'system';
	applyTheme(preference);

	window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
		if (preference === 'system') applyTheme('system');
	});
}

export function getThemePreference(): ThemePreference {
	return preference;
}

export function isResolvedDark(): boolean {
	return resolvedDark;
}

export function setThemePreference(pref: ThemePreference) {
	preference = pref;
	if (browser) localStorage.setItem(STORAGE_KEY, pref);
	applyTheme(pref);
}
