import { PostHog } from 'posthog-node';
import { PUBLIC_POSTHOG_PROJECT_TOKEN, PUBLIC_POSTHOG_HOST } from '$env/static/public';

let posthogClient: PostHog | null = null;

export function getPostHogClient(): PostHog | null {
	if (!PUBLIC_POSTHOG_PROJECT_TOKEN || !PUBLIC_POSTHOG_HOST) {
		if (import.meta.env.DEV) {
			const variable = !PUBLIC_POSTHOG_PROJECT_TOKEN
				? 'PUBLIC_POSTHOG_PROJECT_TOKEN'
				: 'PUBLIC_POSTHOG_HOST';
			console.error(
				`${variable} variable required by PostHog is missing or un-configured, this causes events to be silently missed. This error stops appearing once ${variable} is configured`
			);
		}

		return null;
	}

	if (!posthogClient) {
		posthogClient = new PostHog(PUBLIC_POSTHOG_PROJECT_TOKEN, {
			host: PUBLIC_POSTHOG_HOST,
			flushAt: 1,
			flushInterval: 0
		});
	}

	return posthogClient;
}
