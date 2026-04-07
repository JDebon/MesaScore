import { partiesApi } from '$api/parties';
import { error } from '@sveltejs/kit';
import { ApiError } from '$api/client';

export async function load({ params }) {
	try {
		const party = await partiesApi.get(params.id);
		return { party };
	} catch (e) {
		if (e instanceof ApiError) {
			if (e.status === 403) throw error(403, 'You are not a member of this party');
			if (e.status === 404) throw error(404, 'Party not found');
		}
		throw e;
	}
}
