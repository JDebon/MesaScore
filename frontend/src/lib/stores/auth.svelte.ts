import { browser } from '$app/environment';

interface AuthUser {
	id: string;
	username: string;
	display_name: string;
	avatar_url: string | null;
}

let token = $state<string | null>(null);
let user = $state<AuthUser | null>(null);

if (browser) {
	token = localStorage.getItem('mesascore_token');
	const savedUser = localStorage.getItem('mesascore_user');
	if (savedUser) {
		try {
			user = JSON.parse(savedUser);
		} catch (e) {
			console.error('[auth] Failed to parse stored user, clearing:', e);
			localStorage.removeItem('mesascore_user');
		}
	}
}

export function getToken(): string | null {
	return token;
}
export function getUser(): AuthUser | null {
	return user;
}
export function isAuthenticated(): boolean {
	return !!token;
}

export function setAuth(newToken: string, newUser: AuthUser) {
	token = newToken;
	user = newUser;
	if (browser) {
		localStorage.setItem('mesascore_token', newToken);
		localStorage.setItem('mesascore_user', JSON.stringify(newUser));
	}
}

export function setToken(newToken: string) {
	token = newToken;
	if (browser) localStorage.setItem('mesascore_token', newToken);
}

export function clearAuth() {
	token = null;
	user = null;
	if (browser) {
		localStorage.removeItem('mesascore_token');
		localStorage.removeItem('mesascore_user');
	}
}

export function updateUser(updates: Partial<AuthUser>) {
	if (user) {
		user = { ...user, ...updates };
		if (browser) localStorage.setItem('mesascore_user', JSON.stringify(user));
	}
}
