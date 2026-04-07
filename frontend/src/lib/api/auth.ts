import { api } from './client';
import type { LoginResponse } from './types';

export const authApi = {
	register: (data: { username: string; display_name: string; email: string; password: string }) =>
		api.post<{ message: string }>('/api/auth/register', data),

	login: (email: string, password: string) =>
		api.post<LoginResponse>('/api/auth/login', { email, password }),

	verifyEmail: (token: string) =>
		api.get<{ message: string }>(`/api/auth/verify-email?token=${token}`),

	resendVerification: (email: string) =>
		api.post<void>('/api/auth/resend-verification', { email }),

	checkUsername: (username: string) =>
		api.get<{ available: boolean }>(
			`/api/auth/check-username?username=${encodeURIComponent(username)}`
		)
};
