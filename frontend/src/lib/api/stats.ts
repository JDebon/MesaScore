import { api } from './client';
import type { LeaderboardEntry, UserStats, GamePartyStats } from './types';

export const statsApi = {
	partyLeaderboard: (partyId: string, sort?: string) => {
		const qs = sort ? `?sort=${sort}` : '';
		return api.get<LeaderboardEntry[]>(`/api/parties/${partyId}/stats/leaderboard${qs}`);
	},

	partyGame: (partyId: string, gameId: string) =>
		api.get<GamePartyStats>(`/api/parties/${partyId}/stats/games/${gameId}`),

	partyActivity: (partyId: string) =>
		api.get<{ month: string; count: number }[]>(`/api/parties/${partyId}/stats/activity`),

	userInParty: (partyId: string, userId: string) =>
		api.get<UserStats>(`/api/parties/${partyId}/users/${userId}/stats`),

	userGlobal: (userId: string) => api.get<UserStats>(`/api/users/${userId}/stats`)
};
