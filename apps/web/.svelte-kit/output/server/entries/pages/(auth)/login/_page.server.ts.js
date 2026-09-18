import { t as PUBLIC_API_BASE_URL } from "../../../../chunks/public.js";
import { t as getPostHogClient } from "../../../../chunks/posthog.js";
import { fail } from "@sveltejs/kit";
//#region src/routes/(auth)/login/+page.server.ts
var OAUTH_ERROR_MESSAGES = {
	invalid_state: "Your sign-in attempt expired. Please try again.",
	missing_code: "Sign-in was cancelled or interrupted. Please try again.",
	google_login_failed: "Could not sign in with Google. Please try again.",
	github_login_failed: "Could not sign in with GitHub. Please try again.",
	link_expired: "That sign-in link is invalid or has expired. Request a new one below.",
	verification_failed: "Could not verify your sign-in link. Please request a new one.",
	missing_token: "That sign-in link is missing information. Request a new one below.",
	session_missing: "Something went wrong finishing sign-in. Please try again."
};
var load = async ({ url }) => {
	const errorCode = url.searchParams.get("error");
	return {
		redirectTo: url.searchParams.get("redirectTo") ?? "/",
		errorMessage: errorCode ? OAUTH_ERROR_MESSAGES[errorCode] ?? "Something went wrong. Please try again." : null
	};
};
var actions = { requestMagicLink: async ({ request, fetch }) => {
	const email = ((await request.formData()).get("email") ?? "").toString().trim().toLowerCase();
	if (!email || !email.includes("@")) return fail(400, {
		email,
		error: "Enter a valid email address."
	});
	try {
		const res = await fetch(`${PUBLIC_API_BASE_URL}/api/v1/auth/magic-link`, {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ email })
		});
		if (res.status === 429) return fail(429, {
			email,
			error: "Too many requests. Please wait a few minutes and try again."
		});
		if (!res.ok) return fail(res.status, {
			email,
			error: "Could not send sign-in link. Please try again."
		});
	} catch {
		return fail(502, {
			email,
			error: "Could not reach the server. Please try again in a moment."
		});
	}
	const posthog = getPostHogClient();
	if (posthog) {
		posthog.capture({
			event: "magic_link_requested",
			properties: { source: "email" }
		});
		await posthog.flush();
	}
	return {
		email,
		sent: true
	};
} };
//#endregion
export { actions, load };
