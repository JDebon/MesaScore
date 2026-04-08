import { env } from '$env/dynamic/private';

export const handleFetch = ({ request, fetch }: { request: Request; fetch: typeof globalThis.fetch }) => {
	const backendUrl = env.BACKEND_URL;
	if (!backendUrl) return fetch(request);

	const url = new URL(request.url);
	if (url.pathname.startsWith('/api/')) {
		return fetch(new Request(`${backendUrl}${url.pathname}${url.search}`, request));
	}
	return fetch(request);
};
