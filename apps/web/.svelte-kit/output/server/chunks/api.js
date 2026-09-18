import { t as PUBLIC_API_BASE_URL } from "./public.js";
//#region src/lib/api.ts
var ApiRequestError = class extends Error {
	code;
	status;
	constructor(status, code, message) {
		super(message);
		this.name = "ApiRequestError";
		this.status = status;
		this.code = code;
	}
};
var CSRF_COOKIE = "flocal_csrf";
var CSRF_HEADER = "X-CSRF-Token";
var MUTATING_METHODS = /* @__PURE__ */ new Set([
	"POST",
	"PUT",
	"PATCH",
	"DELETE"
]);
function readCsrfCookie() {
	if (typeof document === "undefined") return null;
	const match = document.cookie.match(new RegExp(`(?:^|;\\s*)${CSRF_COOKIE}=([^;]+)`));
	return match ? decodeURIComponent(match[1]) : null;
}
async function apiFetch(path, init = {}, fetchFn = fetch) {
	const method = (init.method ?? "GET").toUpperCase();
	const headers = new Headers(init.headers);
	if (!headers.has("Content-Type") && !(init.body instanceof FormData)) headers.set("Content-Type", "application/json");
	if (MUTATING_METHODS.has(method)) {
		const csrfToken = readCsrfCookie();
		if (csrfToken) headers.set(CSRF_HEADER, csrfToken);
	}
	const res = await fetchFn(`${PUBLIC_API_BASE_URL}${path}`, {
		...init,
		method,
		credentials: "include",
		headers
	});
	let body;
	try {
		body = await res.json();
	} catch {
		throw new ApiRequestError(res.status, "invalid_response", `Request failed with status ${res.status}`);
	}
	if (!res.ok || !body.success || body.data === void 0) throw new ApiRequestError(res.status, body.error?.code ?? "unknown_error", body.error?.message ?? `Request failed with status ${res.status}`);
	return body.data;
}
function linkAccountUrl(provider) {
	return `${PUBLIC_API_BASE_URL}/api/v1/auth/${provider}/link`;
}
function listLinkedAccounts(fetchFn = fetch) {
	return apiFetch("/api/v1/auth/accounts", {}, fetchFn);
}
function consumeEmailChange(newEmail, token, fetchFn = fetch) {
	return apiFetch("/api/v1/auth/email/verify/confirm", {
		method: "POST",
		body: JSON.stringify({
			newEmail,
			token
		})
	}, fetchFn);
}
function listMySessions(fetchFn = fetch) {
	return apiFetch("/api/v1/auth/sessions", {}, fetchFn);
}
function listCategories(fetchFn = fetch) {
	return apiFetch("/api/v1/topics/categories", {}, fetchFn);
}
function getSession(id, fetchFn = fetch) {
	return apiFetch(`/api/v1/sessions/${id}`, {}, fetchFn);
}
function listSessionHistory(limit = 20, offset = 0, filters = {}, fetchFn = fetch) {
	const params = new URLSearchParams();
	params.set("limit", String(limit));
	params.set("offset", String(offset));
	if (filters.status?.length) params.set("status", filters.status.join(","));
	if (filters.format?.length) params.set("format", filters.format.join(","));
	if (filters.from) params.set("from", filters.from);
	if (filters.to) params.set("to", filters.to);
	if (filters.q) params.set("q", filters.q);
	if (filters.sort) params.set("sort", filters.sort);
	return apiFetch(`/api/v1/sessions/history?${params.toString()}`, {}, fetchFn);
}
function getSessionReport(id, fetchFn = fetch) {
	return apiFetch(`/api/v1/sessions/${id}/report`, {}, fetchFn);
}
function sessionStreamUrl(id) {
	return `${PUBLIC_API_BASE_URL.replace(/^http/, "ws")}/api/v1/sessions/${id}/stream`;
}
function getCurrentPlan(fetchFn = fetch) {
	return apiFetch("/api/v1/plan", {}, fetchFn);
}
function listPlans(fetchFn = fetch) {
	return apiFetch("/api/v1/billing/plans", {}, fetchFn);
}
function getMySubscription(fetchFn = fetch) {
	return apiFetch("/api/v1/billing/subscription", {}, fetchFn);
}
function listMyPayments(limit = 20, offset = 0, fetchFn = fetch) {
	return apiFetch(`/api/v1/billing/payments?limit=${limit}&offset=${offset}`, {}, fetchFn);
}
function getMyStats(fetchFn = fetch) {
	return apiFetch("/api/v1/gamification/stats", {}, fetchFn);
}
function listMyDailyActivity(limit = 30, fetchFn = fetch) {
	return apiFetch(`/api/v1/gamification/activity?limit=${limit}`, {}, fetchFn);
}
function getLeaderboard(limit = 20, offset = 0, fetchFn = fetch) {
	return apiFetch(`/api/v1/gamification/leaderboard?limit=${limit}&offset=${offset}`, {}, fetchFn);
}
function getMyLeaderboardRank(fetchFn = fetch) {
	return apiFetch("/api/v1/gamification/leaderboard/me", {}, fetchFn);
}
//#endregion
export { listSessionHistory as _, getMyLeaderboardRank as a, getSession as c, listCategories as d, listLinkedAccounts as f, listPlans as g, listMySessions as h, getLeaderboard as i, getSessionReport as l, listMyPayments as m, consumeEmailChange as n, getMyStats as o, listMyDailyActivity as p, getCurrentPlan as r, getMySubscription as s, ApiRequestError as t, linkAccountUrl as u, sessionStreamUrl as v };
