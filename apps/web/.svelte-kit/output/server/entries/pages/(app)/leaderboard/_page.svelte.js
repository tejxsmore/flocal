import "../../../../chunks/index-server.js";
import { a as head, b as attr, r as derived, t as attr_class, x as escape_html } from "../../../../chunks/server.js";
import { t as page } from "../../../../chunks/state.js";
import "../../../../chunks/api.js";
//#region src/routes/(app)/leaderboard/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let user = derived(() => page.data.user);
		let entries = [];
		let visibleEntries = derived(() => entries);
		derived(() => visibleEntries().slice(0, 3));
		derived(() => visibleEntries().slice(3));
		head("1226ydt", $$renderer, ($$renderer) => {
			$$renderer.title(($$renderer) => {
				$$renderer.push(`<title>Flocal — Leaderboard</title>`);
			});
		});
		$$renderer.push(`<div class="min-h-screen bg-[#BADF96] px-5 py-8 text-[#1C1124] sm:px-8 sm:py-12 lg:px-12"><div class="mx-auto flex w-full max-w-4xl flex-col gap-8"><header><a href="/profile" class="mb-4 inline-flex items-center gap-1.5 text-sm font-bold text-[#4d2a3a] transition-transform duration-200 hover:-translate-x-1"><span class="text-lg">←</span> Profile</a> <div class="inline-flex rounded-full border-2 border-[#1C1124] bg-[#F7FFCD] px-3.5 py-1.5 text-xs font-black uppercase tracking-[0.12em]">Leaderboard</div> <h1 class="mt-4 text-4xl font-black tracking-[-0.045em] sm:text-5xl md:text-6xl">Rankings</h1> <p class="mt-3 max-w-xl text-sm leading-relaxed text-[#4d2a3a] sm:text-base">See how your XP stacks up against everyone else practising on Flocal.</p></header> `);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--> <div class="flex gap-2"><button type="button"${attr_class(`cursor-pointer rounded-xl border-2 border-[#1C1124] px-4 py-2.5 text-xs font-black transition-transform duration-200 hover:-translate-y-0.5 bg-[#9FA1FF]`)}>🌍 Global</button> <button type="button"${attr("disabled", !user()?.countryCode, true)}${attr_class(`cursor-pointer rounded-xl border-2 border-[#1C1124] px-4 py-2.5 text-xs font-black transition-transform duration-200 hover:-translate-y-0.5 disabled:cursor-not-allowed disabled:opacity-40 bg-[#F7FFCD]`)}>${escape_html(user()?.countryCode ? `🏳️ ${user().countryCode}` : "🏳️ Set country")}</button></div> `);
		$$renderer.push("<!--[0-->");
		$$renderer.push(`<div class="rounded-3xl border-2 border-[#1C1124] bg-[#F7FFCD] px-6 py-12 text-center shadow-[6px_6px_0px_#1C1124]"><p class="text-sm font-bold">Loading rankings…</p></div>`);
		$$renderer.push(`<!--]--></div></div>`);
	});
}
//#endregion
export { _page as default };
