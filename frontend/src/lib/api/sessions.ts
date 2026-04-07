import { api } from './client';
import type { SessionDetail, SessionSummary, PaginatedResponse, ParticipantInput } from './types';

export const sessionsApi = {
	list: (
		partyId: string,
		params?: {
			game_id?: string;
			user_id?: string;
			type?: string;
			from?: string;
			to?: string;
			page?: number;
			per_page?: number;
		}
	) => {
		const search = new URLSearchParams();
		if (params) {
			for (const [k, v] of Object.entries(params)) {
				if (v !== undefined && v !== '') search.set(k, String(v));
			}
		}
		const qs = search.toString();
		return api.get<PaginatedResponse<SessionSummary>>(
			`/api/parties/${partyId}/sessions${qs ? '?' + qs : ''}`
		);
	},

	get: (partyId: string, sessionId: string) =>
		api.get<SessionDetail>(`/api/parties/${partyId}/sessions/${sessionId}`),

	create: (
		partyId: string,
		data: {
			game_id: string;
			session_type: string;
			played_at: string;
			duration_minutes?: number | null;
			notes?: string | null;
			brought_by_user_id?: string | null;
			participants: ParticipantInput[];
		}
	) => api.post<{ id: string }>(`/api/parties/${partyId}/sessions`, data),

	update: (
		partyId: string,
		sessionId: string,
		data: {
			game_id: string;
			session_type: string;
			played_at: string;
			duration_minutes?: number | null;
			notes?: string | null;
			brought_by_user_id?: string | null;
			participants: ParticipantInput[];
		}
	) => api.patch<SessionDetail>(`/api/parties/${partyId}/sessions/${sessionId}`, data),

	delete: (partyId: string, sessionId: string) =>
		api.delete<void>(`/api/parties/${partyId}/sessions/${sessionId}`)
};
