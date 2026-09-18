import { PUBLIC_API_BASE_URL } from '$env/static/public';
import type { ApiEnvelope, SessionResponse, User } from '$lib/types';

export const SESSION_COOKIE = 'flocal_session';

export class ServerApiError extends Error {
	code: string;
	status: number;

	constructor(status: number, code: string, message: string) {
		super(message);
		this.status = status;
		this.code = code;
	}
}

export async function fetchMe(token: string, fetchFn: typeof fetch): Promise<User> {
	const res = await fetchFn(`${PUBLIC_API_BASE_URL}/api/v1/auth/me`, {
		headers: {
			Cookie: `${SESSION_COOKIE}=${token}`
		}
	});

	const body: ApiEnvelope<SessionResponse> = await res.json();

	if (!res.ok || !body.success || body.data === undefined) {
		throw new ServerApiError(
			res.status,
			body.error?.code ?? 'unknown_error',
			body.error?.message ?? `fetchMe failed with status ${res.status}`
		);
	}

	const { user } = body.data;

	if (user.deletedAt) {
		throw new ServerApiError(404, 'user_deleted', 'This account no longer exists.');
	}

	return user;
}
