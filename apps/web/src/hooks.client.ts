import type { HandleClientError } from '@sveltejs/kit';
import { PUBLIC_POSTHOG_PROJECT_TOKEN } from '$env/static/public';
import posthog from 'posthog-js';

export async function init() {
	if (!PUBLIC_POSTHOG_PROJECT_TOKEN) {
		if (import.meta.env.DEV) {
			throw new Error(
				'PUBLIC_POSTHOG_PROJECT_TOKEN variable required by PostHog is missing or un-configured, this causes events to be silently missed. This error stops appearing once PUBLIC_POSTHOG_PROJECT_TOKEN is configured'
			);
		}

		return;
	}

	posthog.init(PUBLIC_POSTHOG_PROJECT_TOKEN, {
		api_host: '/ingest',
		defaults: '2026-01-30',
		capture_exceptions: true
	});
}

export const handleError: HandleClientError = async ({ error, status, message }) => {
	if (PUBLIC_POSTHOG_PROJECT_TOKEN) {
		posthog.captureException(error);
	}

	return { message, status };
};
