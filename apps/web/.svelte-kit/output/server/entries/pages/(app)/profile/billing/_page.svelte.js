import "../../../../../chunks/index-server.js";
import { a as head } from "../../../../../chunks/server.js";
import "../../../../../chunks/api.js";
import "posthog-js";
//#region src/routes/(app)/profile/billing/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		head("eyj1ab", $$renderer, ($$renderer) => {
			$$renderer.title(($$renderer) => {
				$$renderer.push(`<title>Plan &amp; Billing | Flocal</title>`);
			});
			$$renderer.push(`<meta name="description" content="Manage your Flocal subscription and billing."/>`);
		});
		$$renderer.push(`<div class="min-h-screen bg-[#BADF96] px-5 py-10 text-[#1C1124] sm:px-8 lg:px-12"><div class="mx-auto max-w-6xl"><header class="mb-10"><a href="/profile" class="mb-7 inline-flex items-center gap-2 text-sm font-bold text-[#4d2a3a] transition-opacity hover:opacity-60"><span class="text-lg">←</span> Back to profile</a> <div class="max-w-2xl"><div class="mb-4 inline-flex rounded-full border-2 border-[#1C1124] bg-[#F7FFCD] px-3 py-1.5 text-[11px] font-black uppercase tracking-[0.14em]">Account billing</div> <h1 class="text-4xl font-black tracking-tighter sm:text-5xl lg:text-6xl">Your plan. <br/> Your progress.</h1> <p class="mt-4 max-w-xl text-base leading-relaxed text-[#4d2a3a]">Manage your Flocal plan, payment method and subscription.</p></div></header> `);
		$$renderer.push("<!--[0-->");
		$$renderer.push(`<div class="rounded-4xl border-2 border-[#1C1124] bg-white p-12 text-center shadow-[7px_7px_0_#1C1124]"><div class="mx-auto h-9 w-9 animate-spin rounded-full border-4 border-[#1C1124]/15 border-t-[#1C1124]"></div> <p class="mt-4 font-bold">Loading your billing...</p></div>`);
		$$renderer.push(`<!--]--></div></div>`);
	});
}
//#endregion
export { _page as default };
