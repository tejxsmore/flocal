import { t as PUBLIC_API_BASE_URL } from "../chunks/public.js";
import { redirect } from "@sveltejs/kit";
//#region src/lib/server/api.ts
var SESSION_COOKIE = "flocal_session";
var ServerApiError = class extends Error {
	code;
	status;
	constructor(status, code, message) {
		super(message);
		this.status = status;
		this.code = code;
	}
};
async function fetchMe(token, fetchFn) {
	const res = await fetchFn(`${PUBLIC_API_BASE_URL}/api/v1/auth/me`, { headers: { Cookie: `${SESSION_COOKIE}=${token}` } });
	const body = await res.json();
	if (!res.ok || !body.success || body.data === void 0) throw new ServerApiError(res.status, body.error?.code ?? "unknown_error", body.error?.message ?? `fetchMe failed with status ${res.status}`);
	const { user } = body.data;
	if (user.deletedAt) throw new ServerApiError(404, "user_deleted", "This account no longer exists.");
	return user;
}
//#endregion
//#region src/hooks.server.ts
var PUBLIC_ROUTES = ["/login", "/verify"];
var ONBOARDING_ROUTE = "/onboarding";
function isPublicRoute(pathname) {
	return PUBLIC_ROUTES.some((route) => pathname === route || pathname.startsWith(route + "/"));
}
var handle = async ({ event, resolve }) => {
	const { pathname } = event.url;
	if (pathname.startsWith("/ingest")) {
		const hostname = pathname.startsWith("/ingest/static/") || pathname.startsWith("/ingest/array/") ? "us-assets.i.posthog.com" : "us.i.posthog.com";
		const url = new URL(event.request.url);
		url.protocol = "https:";
		url.hostname = hostname;
		url.port = "443";
		url.pathname = pathname.replace(/^\/ingest/, "");
		const headers = new Headers(event.request.headers);
		headers.set("host", hostname);
		headers.set("accept-encoding", "");
		const clientIp = event.request.headers.get("x-forwarded-for") || event.getClientAddress();
		if (clientIp) headers.set("x-forwarded-for", clientIp);
		return await fetch(url.toString(), {
			method: event.request.method,
			headers,
			body: event.request.body,
			duplex: "half"
		});
	}
	const token = event.cookies.get(SESSION_COOKIE);
	event.locals.user = null;
	if (token) try {
		event.locals.user = await fetchMe(token, event.fetch);
	} catch {
		event.locals.user = null;
		event.cookies.delete(SESSION_COOKIE, { path: "/" });
	}
	if (!event.locals.user && !isPublicRoute(event.url.pathname)) redirect(303, `/login?redirectTo=${encodeURIComponent(event.url.pathname)}`);
	if (event.locals.user && event.url.pathname === "/login") redirect(303, "/");
	if (event.locals.user && !event.locals.user.onboardingCompleted && !isPublicRoute(event.url.pathname) && event.url.pathname !== ONBOARDING_ROUTE) redirect(303, ONBOARDING_ROUTE);
	if (event.locals.user && event.locals.user.onboardingCompleted && event.url.pathname === ONBOARDING_ROUTE) redirect(303, "/");
	return resolve(event);
};
//#endregion
export { handle };
