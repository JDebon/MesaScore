import { api } from './client';
import type { UserProfile, UserStats, CollectionItem, DashboardResponse, PendingInvite, User } from './types';

export const usersApi = {
	getMe: () => api.get<UserProfile>('/api/users/me'),

	updateMe: (data: { display_name?: string; avatar_url?: string | null }) =>
		api.patch<UserProfile>('/api/users/me', data),

	get: (id: string) => api.get<UserProfile>(`/api/users/${id}`),

	search: (q: string) => api.get<User[]>(`/api/users/search?q=${encodeURIComponent(q)}`),

	dashboard: () => api.get<DashboardResponse>('/api/users/me/dashboard'),

	invites: () => api.get<PendingInvite[]>('/api/users/me/invites'),

	stats: (id: string) => api.get<UserStats>(`/api/users/${id}/stats`),

	collection: (id: string) => api.get<CollectionItem[]>(`/api/users/${id}/collection`),

	addToCollection: (gameId: string) => api.post<void>('/api/users/me/collection', { game_id: gameId }),

	removeFromCollection: (gameId: string) => api.delete<void>(`/api/users/me/collection/${gameId}`)
};
