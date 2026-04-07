/**
 * Validates an email address format.
 * Returns an error message, or null if valid.
 */
export function validateEmail(email: string): string | null {
	if (!email.trim()) return 'Required';
	const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
	if (!re.test(email)) return 'Invalid email address';
	return null;
}

/**
 * Validates a URL is well-formed and uses http or https only.
 * Rejects javascript:, data:, and other dangerous schemes.
 * Returns an error message, or null if valid (empty string is allowed — field is optional).
 */
export function validateUrl(url: string): string | null {
	if (!url.trim()) return null;
	try {
		const parsed = new URL(url);
		if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
			return 'URL must start with http:// or https://';
		}
	} catch {
		return 'Invalid URL';
	}
	return null;
}

/**
 * Validates a username: lowercase letters, numbers, underscores, hyphens; 2–30 chars.
 * Returns an error message, or null if valid.
 */
export function validateUsername(username: string): string | null {
	if (!username.trim()) return 'Required';
	if (username.length < 2) return 'Must be at least 2 characters';
	if (username.length > 30) return 'Must be 30 characters or fewer';
	if (!/^[a-z0-9_-]+$/.test(username)) {
		return 'Only lowercase letters, numbers, underscores, and hyphens allowed';
	}
	return null;
}

/**
 * Validates a password meets minimum requirements.
 * Returns an error message, or null if valid.
 */
export function validatePassword(password: string): string | null {
	if (!password) return 'Required';
	if (password.length < 8) return 'Must be at least 8 characters';
	return null;
}
