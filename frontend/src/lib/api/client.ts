import { getToken, setToken, clearAuth } from '$stores/auth.svelte';

export class ApiError extends Error {
	status: number;
	fields?: Record<string, string>;

	constructor(status: number, message: string, fields?: Record<string, string>) {
		super(message);
		this.status = status;
		this.fields = fields;
	}
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
	const headers: Record<string, string> = {
		'Content-Type': 'application/json'
	};

	const token = getToken();
	if (token) {
		headers['Authorization'] = `Bearer ${token}`;
	}

	const controller = new AbortController();
	const timeoutId = setTimeout(() => controller.abort(), 15_000);

	let res: Response;
	try {
		res = await fetch(path, {
			method,
			headers,
			body: body ? JSON.stringify(body) : undefined,
			signal: controller.signal
		});
	} catch (fetchErr) {
		clearTimeout(timeoutId);
		if ((fetchErr as DOMException)?.name === 'AbortError') {
			throw new ApiError(0, 'Request timed out. Please try again.');
		}
		throw fetchErr;
	}
	clearTimeout(timeoutId);

	const newToken = res.headers.get('X-New-Token');
	if (newToken) {
		setToken(newToken);
	}

	if (res.status === 401) {
		clearAuth();
		if (typeof window !== 'undefined') {
			window.location.href = '/login';
		}
		throw new ApiError(401, 'Unauthorized');
	}

	if (!res.ok) {
		const err = await res.json().catch((parseErr) => {
			console.error('[API] Failed to parse error response:', parseErr, 'for', method, path);
			return { error: 'Unknown error' };
		});
		const apiError = new ApiError(res.status, err.error, err.fields);
		console.error('[API] Error:', res.status, method, path, apiError.message);
		throw apiError;
	}

	if (res.status === 204 || !res.headers.get('content-type')?.includes('application/json')) return undefined as T;
	return res.json();
}

export const api = {
	get: <T>(path: string) => request<T>('GET', path),
	post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
	patch: <T>(path: string, body?: unknown) => request<T>('PATCH', path, body),
	delete: <T>(path: string) => request<T>('DELETE', path)
};
