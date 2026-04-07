import { api } from './client';
import type { Party, MembersResponse, PartyDashboard } from './types';

export const partiesApi = {
	create: (data: { name: string; description?: string | null }) =>
		api.post<{ id: string; invite_code: string }>('/api/parties', data),

	get: (id: string) => api.get<Party>(`/api/parties/${id}`),

	update: (id: string, data: { name?: string; description?: string | null }) =>
		api.patch<Party>(`/api/parties/${id}`, data),

	regenerateInvite: (id: string) =>
		api.post<{ invite_code: string }>(`/api/parties/${id}/regenerate-invite`),

	joinPreview: (inviteCode: string) =>
		api.get<{ party: { id: string; name: string; member_count: number } }>(
			`/api/parties/join/${inviteCode}`
		),

	join: (inviteCode: string) => api.post<{ party_id: string }>(`/api/parties/join/${inviteCode}`),

	members: (id: string) => api.get<MembersResponse>(`/api/parties/${id}/members`),

	removeMember: (partyId: string, userId: string) =>
		api.delete<void>(`/api/parties/${partyId}/members/${userId}`),

	leave: (id: string) => api.post<void>(`/api/parties/${id}/leave`),

	transferAdmin: (id: string, newAdminUserId: string) =>
		api.post<{ message: string }>(`/api/parties/${id}/transfer-admin`, {
			new_admin_user_id: newAdminUserId
		}),

	sendInvite: (partyId: string, userId: string) =>
		api.post<void>(`/api/parties/${partyId}/invites`, { user_id: userId }),

	acceptInvite: (partyId: string, inviteId: string) =>
		api.post<{ party_id: string }>(`/api/parties/${partyId}/invites/${inviteId}/accept`),

	declineInvite: (partyId: string, inviteId: string) =>
		api.post<void>(`/api/parties/${partyId}/invites/${inviteId}/decline`),

	dashboard: (id: string) => api.get<PartyDashboard>(`/api/parties/${id}/dashboard`)
};
