// src/hooks.server.ts
import type { Handle, HandleServerError } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';
import { fetchMe, SESSION_COOKIE } from '$lib/server/api';

const PUBLIC_ROUTES = ['/login', '/verify'];
const ONBOARDING_ROUTE = '/onboarding';

function isPublicRoute(pathname: string) {
	return PUBLIC_ROUTES.some((route) => pathname === route || pathname.startsWith(route + '/'));
}

export const handle: Handle = async ({ event, resolve }) => {
	const { pathname } = event.url;

	// Reverse proxy for PostHog — route /ingest requests to PostHog servers
	if (pathname.startsWith('/ingest')) {
		const useAssetHost =
			pathname.startsWith('/ingest/static/') || pathname.startsWith('/ingest/array/');
		const hostname = useAssetHost ? 'us-assets.i.posthog.com' : 'us.i.posthog.com';

		const url = new URL(event.request.url);
		url.protocol = 'https:';
		url.hostname = hostname;
		url.port = '443';
		url.pathname = pathname.replace(/^\/ingest/, '');

		const headers = new Headers(event.request.headers);
		headers.set('host', hostname);
		headers.set('accept-encoding', '');

		const clientIp =
			event.request.headers.get('x-forwarded-for') || event.getClientAddress();
		if (clientIp) {
			headers.set('x-forwarded-for', clientIp);
		}

		const response = await fetch(url.toString(), {
			method: event.request.method,
			headers,
			body: event.request.body,
			// @ts-expect-error - duplex is required for streaming request bodies
			duplex: 'half'
		});

		return response;
	}

	const token = event.cookies.get(SESSION_COOKIE);

	event.locals.user = null;

	if (token) {
		try {
			event.locals.user = await fetchMe(token, event.fetch);
		} catch {
			event.locals.user = null;
			// stale/invalid cookie (or soft-deleted account) — clear it so we don't keep retrying every request
			event.cookies.delete(SESSION_COOKIE, { path: '/' });
		}
	}

	if (!event.locals.user && !isPublicRoute(event.url.pathname)) {
		redirect(303, `/login?redirectTo=${encodeURIComponent(event.url.pathname)}`);
	}

	if (event.locals.user && event.url.pathname === '/login') {
		redirect(303, '/');
	}

	if (
		event.locals.user &&
		!event.locals.user.onboardingCompleted &&
		!isPublicRoute(event.url.pathname) &&
		event.url.pathname !== ONBOARDING_ROUTE
	) {
		redirect(303, ONBOARDING_ROUTE);
	}

	if (
		event.locals.user &&
		event.locals.user.onboardingCompleted &&
		event.url.pathname === ONBOARDING_ROUTE
	) {
		redirect(303, '/');
	}

	return resolve(event);
};