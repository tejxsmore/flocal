import "../../../../chunks/index-server.js";
import { a as head, r as derived } from "../../../../chunks/server.js";
import { t as page } from "../../../../chunks/state.js";
import "../../../../chunks/api.js";
import "../../../../chunks/types.js";
//#region src/routes/(app)/profile/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		derived(() => page.data.user);
		let activity = [];
		function buildHeatmapWeeks(days) {
			const byDate = new Map(days.map((d) => [d.activityDate, d.sessionsCount]));
			const cells = [];
			const today = /* @__PURE__ */ new Date();
			for (let i = 90; i >= 0; i--) {
				const d = new Date(today);
				d.setDate(d.getDate() - i);
				const key = d.toISOString().slice(0, 10);
				cells.push({
					date: key,
					count: byDate.get(key) ?? 0
				});
			}
			const weeks = [];
			for (let i = 0; i < cells.length; i += 7) weeks.push(cells.slice(i, i + 7));
			return weeks;
		}
		derived(() => buildHeatmapWeeks(activity));
		head("b6jup3", $$renderer, ($$renderer) => {
			$$renderer.title(($$renderer) => {
				$$renderer.push(`<title>Flocal — Profile</title>`);
			});
		});
		$$renderer.push(`<div class="min-h-screen bg-[#BADF96] px-5 py-7 text-[#1C1124] sm:px-8 sm:py-10"><div class="mx-auto w-full max-w-5xl"><header class="flex items-center justify-between"><a href="/" class="text-2xl font-black tracking-tight transition-transform duration-200 hover:-translate-y-0.5">Flocal</a> <div class="flex items-center gap-2"><a href="/session" class="rounded-xl border-2 border-[#1C1124] bg-[#F7FFCD] px-4 py-2 text-xs font-black transition-transform duration-200 hover:-translate-y-0.5 sm:text-sm">Sessions</a> <a href="/profile/settings" class="hidden rounded-xl border-2 border-[#1C1124] bg-[#F7FFCD] px-4 py-2 text-xs font-black transition-transform duration-200 hover:-translate-y-0.5 sm:block sm:text-sm">Settings</a></div></header> `);
		$$renderer.push("<!--[0-->");
		$$renderer.push(`<div class="flex min-h-[70vh] items-center justify-center"><div class="rounded-3xl border-2 border-[#1C1124] bg-[#F7FFCD] px-8 py-6 text-center shadow-[7px_7px_0px_#1C1124]"><p class="text-sm font-black">Loading your profile…</p></div></div>`);
		$$renderer.push(`<!--]--></div></div>`);
	});
}
//#endregion
export { _page as default };
