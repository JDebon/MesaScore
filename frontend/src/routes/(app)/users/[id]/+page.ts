import { usersApi } from '$api/users';

export const ssr = false;

export async function load({ params }: { params: { id: string } }) {
	const [profile, stats, collection] = await Promise.all([
		usersApi.get(params.id),
		usersApi.stats(params.id),
		usersApi.collection(params.id)
	]);
	return { profile, stats, collection };
}
