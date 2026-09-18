import { n as PUBLIC_POSTHOG_HOST, r as PUBLIC_POSTHOG_PROJECT_TOKEN } from "./public.js";
import { PostHog } from "posthog-node";
//#region src/lib/server/posthog.ts
var posthogClient = null;
function getPostHogClient() {
	if (!posthogClient) posthogClient = new PostHog(PUBLIC_POSTHOG_PROJECT_TOKEN, {
		host: PUBLIC_POSTHOG_HOST,
		flushAt: 1,
		flushInterval: 0
	});
	return posthogClient;
}
//#endregion
export { getPostHogClient as t };
