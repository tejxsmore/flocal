// @ts-nocheck
import type { PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { PUBLIC_API_BASE_URL } from '$env/static/public';
import type { ApiEnvelope, SessionResponse } from '$lib/types';
import { getPostHogClient } from '$lib/server/posthog';

export const load = async ({ url, cookies, fetch }: Parameters<PageServerLoad>[0]) => {
	const errorCode = url.searchParams.get('error');

	if (errorCode) {
		redirect(303, `/login?error=${encodeURIComponent(errorCode)}`);
	}

	const email = url.searchParams.get('email');
	const token = url.searchParams.get('token');
	const redirectTo = url.searchParams.get('redirectTo') ?? '/';

	if (!email || !token) {
		redirect(303, '/login?error=missing_token');
	}

	const res = await fetch(`${PUBLIC_API_BASE_URL}/api/v1/auth/magic-link/consume`, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify({
			email,
			token
		})
	});

	const body: ApiEnvelope<SessionResponse> = await res.json().catch(() => ({ success: false }));

	if (!res.ok || !body.success || !body.data) {
		redirect(303, '/login?error=verification_failed');
	}

	const { token: sessionToken, csrfToken, expiresAt } = body.data;

	if (!sessionToken || !csrfToken) {
		redirect(303, '/login?error=session_missing');
	}

	const maxAge = Math.max(1, Math.floor((new Date(expiresAt).getTime() - Date.now()) / 1000));

	cookies.set('flocal_session', sessionToken, {
		path: '/',
		httpOnly: true,
		secure: false,
		sameSite: 'lax',
		maxAge
	});

	cookies.set('flocal_csrf', csrfToken, {
		path: '/',
		httpOnly: false,
		secure: false,
		sameSite: 'lax',
		maxAge
	});

	const { user } = body.data;
	if (user) {
		const posthog = getPostHogClient();
		if (posthog) {
			posthog.identify({
				distinctId: user.id,
				properties: { name: user.name, username: user.username }
			});
			posthog.capture({
				distinctId: user.id,
				event: 'user_signed_in',
				properties: { method: 'magic_link' }
			});
			await posthog.flush();
		}
	}

	redirect(303, redirectTo);
};
