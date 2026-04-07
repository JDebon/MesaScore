import { api } from './client';
import type { Game, GameDetail, BggSearchResult, BGGGameDetail, AvailableGame } from './types';

export const gamesApi = {
	list: (params?: { q?: string; sort?: string }) => {
		const search = new URLSearchParams();
		if (params?.q) search.set('q', params.q);
		if (params?.sort) search.set('sort', params.sort);
		const qs = search.toString();
		return api.get<Game[]>(`/api/games${qs ? '?' + qs : ''}`);
	},

	get: (id: string) => api.get<GameDetail>(`/api/games/${id}`),

	create: (data: { bgg_id?: number | null; name?: string | null }) =>
		api.post<{ id: string }>('/api/games', data),

	update: (
		id: string,
		data: {
			name?: string;
			description?: string | null;
			cover_image_url?: string | null;
			min_players?: number | null;
			max_players?: number | null;
		}
	) => api.patch<GameDetail>(`/api/games/${id}`, data),

	bggRefresh: (id: string) => api.post<GameDetail>(`/api/games/${id}/bgg-refresh`),

	bggSearch: (q: string) =>
		api.get<BggSearchResult[]>(`/api/bgg/search?q=${encodeURIComponent(q)}`),

	bggThing: (id: number) =>
		api.get<BGGGameDetail>(`/api/bgg/thing?id=${id}`),

	availableForParty: (partyId: string) =>
		api.get<AvailableGame[]>(`/api/parties/${partyId}/available-games`)
};
